package client

import (
	"context"
	"log"
	"time"
	pb "backend/vulnerability" // 替换为实际的pb文件路径

	"google.golang.org/grpc"
)

func ClientMain() {
	// 连接到 gRPC 服务器
	conn, err := grpc.Dial("10.50.1.207:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	select {}

	// client := pb.NewVulnerabilityClient(conn)
	// reader := bufio.NewReader(os.Stdin) //标准读入

	// for {
	// 	fmt.Println("请输入合约代码（以 'END' 结束输入，输入 'exit' 退出）：")
	// 	var contractLines []string

	// 	for {
	// 		line, err := reader.ReadString('\n')
	// 		if err != nil {
	// 			log.Fatalf("读取输入失败: %v", err)
	// 		}
	// 		line = strings.TrimSpace(line)
	// 		if line == "exit" {
	// 			fmt.Println("退出客户端")
	// 			return
	// 		}
	// 		if line == "END" {
	// 			break
	// 		}
	// 		contractLines = append(contractLines, line)
	// 	}

	// 	contractSourcecode := strings.Join(contractLines, "\n")

	// 	fmt.Printf("contractSourcecode: %v\n", contractSourcecode)

	// 	request := &pb.DetectRequest{
	// 		ContractSourcecode: contractSourcecode,
	// 	}

	// 	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	// 	defer cancel()

	// 	response, err := client.Detect(ctx, request)
	// 	if err != nil {
	// 		log.Fatalf("could not detect vulnerability: %v", err)
	// 	}

	// 	log.Printf("Response: %v", response.Message)

	// }
}

func DetectContract(sourcecode string) (string, error) {
	conn, err := grpc.Dial("10.50.1.207:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	client := pb.NewVulnerabilityClient(conn)

	request := &pb.DetectRequest{
		ContractSourcecode: sourcecode,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	response, err := client.Detect(ctx, request)

	if err != nil {
		return "", err
	}

	return response.Message, nil

}
