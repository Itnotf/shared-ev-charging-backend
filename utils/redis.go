package utils

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis(addr, password string, db int) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func GetRedis() *redis.Client {
	return RedisClient
}

func RedisCtx() context.Context {
	return context.Background()
}
