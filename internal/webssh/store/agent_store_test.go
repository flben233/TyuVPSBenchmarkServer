package store

import (
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestTaskRedisKeyFormat(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "basic", in: "task-123", want: "task:task-123"},
		{name: "empty", in: "", want: "task:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildTaskKey(tt.in)
			if got != tt.want {
				t.Fatalf("unexpected key: %s", got)
			}
		})
	}

}

func TestAgentKeyBuilders(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "audit", got: BuildAuditKey("task-123"), want: "audit:task-123"},
		{name: "checkpoint", got: BuildCheckpointKey("task-123"), want: "checkpoint:task-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("unexpected key: %s", tt.got)
			}
		})
	}
}

func TestNormalizeTaskBindingGetError(t *testing.T) {
	found, err := normalizeTaskBindingGetError(nil)
	if !found || err != nil {
		t.Fatalf("expected found=true and nil error, got found=%v err=%v", found, err)
	}

	found, err = normalizeTaskBindingGetError(redis.Nil)
	if found || err != nil {
		t.Fatalf("expected not found and nil error for redis.Nil, got found=%v err=%v", found, err)
	}

	expectedErr := errors.New("boom")
	found, err = normalizeTaskBindingGetError(expectedErr)
	if found || !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped original error, got found=%v err=%v", found, err)
	}
}

func TestDecodeAuditEntriesSkipsMalformed(t *testing.T) {
	entries, err := decodeAuditEntries([]string{
		`{"timestamp":"2026-01-01T00:00:00Z","command":"ls","risk_level":"safe","approved":true}`,
		`{not-json}`,
		`{"timestamp":"2026-01-01T00:00:01Z","command":"df -h","risk_level":"safe","approved":true}`,
	})

	if len(entries) != 2 {
		t.Fatalf("expected 2 valid entries, got %d", len(entries))
	}
	if err == nil {
		t.Fatal("expected parse warning error for malformed audit entry")
	}
}
