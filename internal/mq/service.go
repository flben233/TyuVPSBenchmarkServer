package mq

import (
	"VPSBenchmarkBackend/internal/cache"
	"VPSBenchmarkBackend/internal/common"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

const (
	TaskPending = "pending"
	TaskRunning = "running"
	TaskDone    = "done"
)

type Task[T any] struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Progress float32 `json:"progress"` // 0.0 to 1.0
	Result   T       `json:"result"`   // Additional data for the task
}

func HandleQuery(taskID string) (*Task[any], error) {
	if taskID == "" {
		return nil, &common.InvalidParamError{Message: "Task ID is required"}
	}
	var task Task[any]
	err := cache.GetJSON(context.Background(), taskID, &task)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, &common.InvalidParamError{Message: "Invalid task ID or task has expired"}
		}
		return nil, fmt.Errorf("failed to get task status from Redis: %w", err)
	}
	return &task, nil
}
