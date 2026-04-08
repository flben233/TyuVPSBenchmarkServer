package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/store"
)

type testSessionWriter struct {
	writes [][]byte
	err    error
}

func (w *testSessionWriter) WriteInput(data []byte) error {
	if w.err != nil {
		return w.err
	}
	w.writes = append(w.writes, append([]byte(nil), data...))
	return nil
}

func TestExecuteTaskCommandMissingTaskReturnsDomainError(t *testing.T) {
	loadServiceConfigForTest(30, 20)
	defer resetExecHooks()

	getTaskBinding = func(ctx context.Context, taskID string) (*store.TaskBinding, error) {
		return nil, nil
	}

	_, err := ExecuteTaskCommand(context.Background(), "missing-task", "ls -la", true)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestExecuteTaskCommandMissingSessionReturnsDomainError(t *testing.T) {
	loadServiceConfigForTest(30, 20)
	defer resetExecHooks()

	getTaskBinding = func(ctx context.Context, taskID string) (*store.TaskBinding, error) {
		return &store.TaskBinding{SessionID: "session-missing", Status: model.AgentStatusRunning}, nil
	}
	resolveSession = func(sessionID string) (sessionInputWriter, bool) {
		return nil, false
	}

	_, err := ExecuteTaskCommand(context.Background(), "task-1", "ls -la", true)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestExecuteTaskCommandApprovalGate(t *testing.T) {
	loadServiceConfigForTest(30, 20)
	defer resetExecHooks()

	getTaskBinding = func(ctx context.Context, taskID string) (*store.TaskBinding, error) {
		return &store.TaskBinding{SessionID: "session-1", Status: model.AgentStatusRunning}, nil
	}
	resolveSession = func(sessionID string) (sessionInputWriter, bool) {
		t.Fatalf("session should not be resolved when approval is missing")
		return nil, false
	}
	appendAudit = func(ctx context.Context, taskID string, entry store.AuditEntry) error {
		return nil
	}

	result, err := ExecuteTaskCommand(context.Background(), "task-1", "apt install nginx", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Executed {
		t.Fatal("expected command not executed due to approval gate")
	}
	if !result.RequiresApproval {
		t.Fatal("expected requires approval true")
	}
	if !IsApprovalStatus(result.Status) {
		t.Fatalf("expected approval status, got %q", result.Status)
	}
}

func TestExecuteTaskCommandRespectsMaxCommandLimit(t *testing.T) {
	loadServiceConfigForTest(30, 1)
	defer resetExecHooks()

	getTaskBinding = func(ctx context.Context, taskID string) (*store.TaskBinding, error) {
		return &store.TaskBinding{SessionID: "session-1", Status: model.AgentStatusRunning, CommandCount: 1}, nil
	}

	_, err := ExecuteTaskCommand(context.Background(), "task-1", "ls -la", true)
	if !errors.Is(err, ErrCommandLimitExceeded) {
		t.Fatalf("expected ErrCommandLimitExceeded, got %v", err)
	}
}

func TestExecuteTaskCommandReturnsAuditError(t *testing.T) {
	loadServiceConfigForTest(30, 20)
	defer resetExecHooks()

	getTaskBinding = func(ctx context.Context, taskID string) (*store.TaskBinding, error) {
		return &store.TaskBinding{SessionID: "session-1", Status: model.AgentStatusRunning}, nil
	}
	appendAudit = func(ctx context.Context, taskID string, entry store.AuditEntry) error {
		return errors.New("audit boom")
	}

	_, err := ExecuteTaskCommand(context.Background(), "task-1", "apt install nginx", false)
	if err == nil {
		t.Fatal("expected audit error")
	}
}

func TestExecuteTaskCommandRespectsConfigTimeoutAndPersists(t *testing.T) {
	loadServiceConfigForTest(1, 20)
	defer resetExecHooks()

	writer := &testSessionWriter{}
	getTaskBinding = func(ctx context.Context, taskID string) (*store.TaskBinding, error) {
		return &store.TaskBinding{SessionID: "session-1", Status: model.AgentStatusRunning, CommandCount: 0}, nil
	}
	resolveSession = func(sessionID string) (sessionInputWriter, bool) {
		return writer, true
	}
	ctxDeadlineChecked := false
	saveTask = func(ctx context.Context, taskID string, binding store.TaskBinding) error {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected execution context with deadline")
		}
		if time.Until(deadline) > 1500*time.Millisecond {
			t.Fatalf("expected short timeout context, got %v until deadline", time.Until(deadline))
		}
		ctxDeadlineChecked = true
		return nil
	}
	appendAudit = func(ctx context.Context, taskID string, entry store.AuditEntry) error {
		return nil
	}

	result, err := ExecuteTaskCommand(context.Background(), "task-1", "ls -la", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !ctxDeadlineChecked {
		t.Fatal("expected save task to validate deadline")
	}
	if !result.Executed {
		t.Fatal("expected executed true")
	}
	if len(writer.writes) != 1 {
		t.Fatalf("expected 1 command write, got %d", len(writer.writes))
	}
}

func TestIsApprovalStatus(t *testing.T) {
	if !IsApprovalStatus(model.AgentStatusAwaitingApproval) {
		t.Fatal("expected awaiting_approval to be approval status")
	}
	if IsApprovalStatus(model.AgentStatusRunning) {
		t.Fatal("expected running to not be approval status")
	}
}

func resetExecHooks() {
	getTaskBinding = store.GetTaskBinding
	saveTask = store.SaveTaskBinding
	appendAudit = store.AppendAuditEntry
	resolveSession = func(sessionID string) (sessionInputWriter, bool) {
		session, ok := GetSession(sessionID)
		if !ok || session == nil {
			return nil, false
		}
		return session, true
	}
}

func loadServiceConfigForTest(timeoutSeconds, maxCommands int) {
	dir, err := os.MkdirTemp("", "agent-exec-config")
	if err != nil {
		panic(err)
	}

	path := filepath.Join(dir, "config.json")
	content := []byte(`{"agentCommandTimeout":` + strconv.Itoa(timeoutSeconds) + `,"agentMaxCommands":` + strconv.Itoa(maxCommands) + `}`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		panic(err)
	}

	if err := config.Load(path); err != nil {
		panic(err)
	}
}
