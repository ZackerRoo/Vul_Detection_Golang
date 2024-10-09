package server

import (
	"backend/cache"
	"backend/database"
	pb "backend/vulnerability"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

const (
	// rabbitMQURL = "amqp://guest:guest@localhost:5672/" // RabbitMQ 连接地址
	rabbitMQURL = "amqp://guest:guest@rabbitmq-server:5672/"
	queueName   = "vulnerability_queue" // 队列名称
)

// 定义 gRPC 服务
type server struct {
	pb.UnimplementedVulnerabilityServer
}

var db *database.GromDataBase

// 用于存储检测结果的共享 map
var resultMap sync.Map
var redisClient *cache.RedisClient

// 发送消息到 RabbitMQ 队列（生产者部分）
func sendToRabbitMQ(messageID string, contractCode string) error {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // 队列名称
		false,     // 是否持久化
		false,     // 是否自动删除
		false,     // 是否为独占
		false,     // 是否阻塞
		nil,       // 额外属性
	)
	if err != nil {
		return err
	}

	// 将任务ID和合约代码组合成消息
	message := fmt.Sprintf("%s|%s", messageID, contractCode)

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key (队列名称)
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	return err
}

// 从 RabbitMQ 队列中消费消息并执行检测任务（消费者部分）
func consumeFromRabbitMQ() {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // 队列名称
		false,     // 是否持久化
		false,     // 是否自动删除
		false,     // 是否为独占
		false,     // 是否阻塞
		nil,       // 额外属性
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // 队列名称
		"",     // 消费者标签
		true,   // 是否自动应答
		false,  // 是否独占
		false,  // 是否阻塞
		false,
		nil, // 额外属性
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			// 解析消息，提取任务ID和合约代码
			messageParts := strings.SplitN(string(d.Body), "|", 2)
			if len(messageParts) != 2 {
				log.Printf("Invalid message format: %s", d.Body)
				continue
			}

			messageID := messageParts[0]
			contractCode := messageParts[1]

			log.Printf("Received a message for ID: %s", messageID)

			// 调用模型执行漏洞检测
			hasVulnerability, message, lines := detectVulnerability(contractCode)
			log.Printf("检测结果: %v, 信息: %s, 行数: %s", hasVulnerability, message, lines)
			new := fmt.Sprintf("%s|%s", message, lines)

			// 将检测结果存储到共享 map 中
			resultMap.Store(messageID, &pb.DetectResponse{
				HasVulnerability: hasVulnerability,
				Message:          new,
			})
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	select {} // 保持消费者一直运行
}

// 模型检测函数
func detectVulnerability(contractCode string) (bool, string, string) {
	// 将合约代码写入临时文件
	// detectScriptPath := "../demo/detect.py"
	// detectScriptDir := "../demo"
	detectScriptPath := "/app/demo/detect.py"
	detectScriptDir := "/app/demo"

	tmpFile, err := ioutil.TempFile("/tmp", "contract_*.sol")
	if err != nil {
		log.Printf("Error creating temporary file: %v", err)
		return false, "创建临时文件失败", ""
	}
	defer os.Remove(tmpFile.Name()) // 在函数结束时删除临时文件

	if _, err := tmpFile.Write([]byte(contractCode)); err != nil {
		log.Printf("Error writing to temporary file: %v", err)
		return false, "写入临时文件失败", ""
	}
	tmpFile.Close()
	contractFilePath := tmpFile.Name()
	// PythonInterfertorPath := "/root/anaconda3/envs/VD/bin/python"
	PythonInterfertorPath := "/app/demo/venv/bin/python"

	// 执行模型检测的命令，传递临时文件路径
	cmd := exec.Command(PythonInterfertorPath, detectScriptPath, "dev", contractFilePath)
	cmd.Dir = detectScriptDir

	// 执行命令并获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing detection model: %v, output: %s", err, string(output))
		return false, "模型执行出错", ""
	}

	// 处理模型输出
	result := string(output)
	log.Printf("Model output: %s", result)

	if strings.Contains(result, "Detected vulnerabilities at lines") {
		re := regexp.MustCompile(`Detected vulnerabilities at lines: \[(.*?)\]`)
		match := re.FindStringSubmatch(result)
		lineNumber := match[1]
		return true, "检测到漏洞", lineNumber
	}

	// // 根据模型的输出结果判断是否存在漏洞
	// if result == "检测到漏洞" {
	// 	return true, "检测到漏洞"
	// }
	return false, "未检测到漏洞", ""
}

// 实现 gRPC 服务的方法
func (s *server) Detect(ctx context.Context, in *pb.DetectRequest) (*pb.DetectResponse, error) {
	contractCode := in.ContractSourcecode

	// 在这个代码中结合redis 高速缓冲的方法
	cacheKey := fmt.Sprintf("contract_%x", sha256.Sum256([]byte(contractCode)))
	fmt.Printf("Generated Cache Key: %s\n", cacheKey)
	cachedResult, err := redisClient.GetCache(cacheKey)
	// cachedResult = strings.TrimSpace(cachedResult)
	if err == nil && cachedResult != "" {
		log.Printf("Cache hit for contract code, returning cached result.")
		return &pb.DetectResponse{
			HasVulnerability: strings.Contains(cachedResult, "检测到漏洞"),
			Message:          cachedResult,
		}, nil
	}

	// 如果cache 没有命中的话，生成一个唯一的任务ID
	messageID := fmt.Sprintf("%d", time.Now().UnixNano())

	// 将任务ID和合约代码发送到 RabbitMQ 队列
	err = sendToRabbitMQ(messageID, contractCode)
	if err != nil {
		log.Printf("Error sending message to RabbitMQ: %v", err)
		return &pb.DetectResponse{
			HasVulnerability: false,
			Message:          "无法将消息发送到 RabbitMQ",
		}, err
	}

	// 等待检测结果
	log.Printf("等待检测任务完成，任务ID: %s", messageID)
	timeout := time.After(180 * time.Second) // 设置超时时间，避免无限等待

	// 使用 channel 来等待检测结果
	for {
		select {
		case <-timeout:
			log.Printf("检测任务超时，任务ID: %s", messageID)
			return &pb.DetectResponse{
				HasVulnerability: false,
				Message:          "检测任务超时",
			}, nil
		default:
			if result, ok := resultMap.Load(messageID); ok {
				resultMap.Delete(messageID) // 移除已处理的结果
				detectResponse := result.(*pb.DetectResponse)

				err := redisClient.SetCache(cacheKey, detectResponse.Message)
				if err != nil {
					log.Printf("Error setting cache for contract code: %v", err)
				} else {
					log.Printf("Cache successfully updated with detection result")
				}

				record := database.VulnerabilityRecord{
					ContractSource:   contractCode,
					HasVulnerability: detectResponse.HasVulnerability,
					DetectedLines:    "",
					Message:          detectResponse.Message,
					CreatedAt:        time.Now(),
				}

				if err := db.InsertRecord(&record); err != nil {
					log.Printf("Error inserting record into database: %v", err)
				} else {
					log.Printf("Record successfully inserted into database")
				}

				return result.(*pb.DetectResponse), nil
			}
			time.Sleep(1 * time.Second) // 等待 1 秒钟后再次检查结果
		}
	}
}

func ServerMain() {
	// 启动 RabbitMQ 消费者以执行检测任务
	var err error
	go consumeFromRabbitMQ()

	dsn := "root:123456789@tcp(mysql-new-server:3306)/vul_detect?charset=utf8mb4&parseTime=True&loc=Local"
	redisClient = cache.NewRedisClient("redis-new-server:6379", "", 0)

	db, err = database.NewDatabase(dsn)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	// 创建 gRPC 服务器
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterVulnerabilityServer(s, &server{})

	log.Printf("gRPC server listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
