package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
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

func SetJSON(ctx context.Context, key string, value any, exp time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, string(data), exp).Err()
}

func GetJSON(ctx context.Context, key string, dest any) error {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}
