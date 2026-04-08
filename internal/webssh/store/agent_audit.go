package store

import (
	"VPSBenchmarkBackend/internal/cache"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const auditTTL = 24 * time.Hour

type AuditEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Command     string    `json:"command"`
	RiskLevel   string    `json:"risk_level"`
	Approved    bool      `json:"approved"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
	CommandNo   int       `json:"command_no,omitempty"`
	ExecutedBy  string    `json:"executed_by,omitempty"`
	SessionID   string    `json:"session_id,omitempty"`
	TaskStatus  string    `json:"task_status,omitempty"`
	MessageType string    `json:"message_type,omitempty"`
}

func BuildAuditKey(taskID string) string {
	return "audit:" + taskID
}

func AppendAuditEntry(ctx context.Context, taskID string, entry AuditEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	key := BuildAuditKey(taskID)
	client := cache.GetClient()
	pipe := client.TxPipeline()
	pipe.RPush(ctx, key, string(data))
	pipe.Expire(ctx, key, auditTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func ListAuditEntries(ctx context.Context, taskID string) ([]AuditEntry, error) {
	key := BuildAuditKey(taskID)
	values, err := cache.GetClient().LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	return decodeAuditEntries(values)
}

func decodeAuditEntries(values []string) ([]AuditEntry, error) {
	entries := make([]AuditEntry, 0, len(values))
	malformed := 0
	for _, value := range values {
		var entry AuditEntry
		if err := json.Unmarshal([]byte(value), &entry); err != nil {
			malformed++
			continue
		}
		entries = append(entries, entry)
	}

	if malformed > 0 {
		return entries, fmt.Errorf("skipped %d malformed audit entr%s", malformed, pluralSuffix(malformed))
	}

	return entries, nil
}

func pluralSuffix(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
