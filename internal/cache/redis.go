package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedis(addr, password string, db int) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return rdb.Ping(context.Background()).Err()
}

func GetClient() *redis.Client {
	if rdb == nil {
		panic("Redis client not initialized. Call InitRedis first.")
	}
	return rdb
}
