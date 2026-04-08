package store

import (
	"VPSBenchmarkBackend/internal/cache"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const taskTTL = 30 * time.Minute

type TaskBinding struct {
	UserID       int64     `json:"user_id"`
	SessionID    string    `json:"session_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	CommandCount int       `json:"command_count"`
}

func BuildTaskKey(taskID string) string {
	return "task:" + taskID
}

func BuildCheckpointKey(taskID string) string {
	return "checkpoint:" + taskID
}

func SaveTaskBinding(ctx context.Context, taskID string, binding TaskBinding) error {
	return cache.SetJSON(ctx, BuildTaskKey(taskID), binding, taskTTL)
}

func GetTaskBinding(ctx context.Context, taskID string) (*TaskBinding, error) {
	var binding TaskBinding
	found, err := normalizeTaskBindingGetError(cache.GetJSON(ctx, BuildTaskKey(taskID), &binding))
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &binding, nil
}

func normalizeTaskBindingGetError(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	return false, err
}
