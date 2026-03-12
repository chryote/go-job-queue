package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

func Push(ctx context.Context, rdb *redis.Client, data []byte) error {
	return rdb.LPush(ctx, "jobs_queue", data).Err()
}

func Pop(ctx context.Context, rdb *redis.Client) ([]string, error) {
	return rdb.BRPop(ctx, 0, "jobs_queue").Result()
}
