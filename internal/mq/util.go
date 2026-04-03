package mq

import (
	"VPSBenchmarkBackend/internal/cache"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const taskPrefix = "task:"

func SetTask[T any](task Task[T]) error {
	var taskData []byte
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal report task data: %w", err)
	}
	cache.GetClient().Set(context.Background(), taskPrefix+task.ID, string(taskData), 30*time.Minute)
	return nil
}

func GetTask[T any](taskID string) (*Task[T], error) {
	var task Task[T]
	err := cache.GetJSON(context.Background(), taskPrefix+taskID, &task)
	if err != nil {
		return nil, fmt.Errorf("failed to get task from Redis: %w", err)
	}
	return &task, nil
}
