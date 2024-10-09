package cache

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string, passwd string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	return &RedisClient{client: rdb}
}

func (r *RedisClient) SetCache(key string, value string) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

func (r *RedisClient) GetCache(key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // 这个数据库中没有这个key
	} else if err != nil {
		return "", err
	}
	return value, nil

}
