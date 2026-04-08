package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/store"
)

var (
	ErrTaskNotFound         = errors.New("task binding not found")
	ErrSessionNotFound      = errors.New("ssh session not found")
	ErrCommandLimitExceeded = errors.New("max command limit exceeded")

	getTaskBinding = store.GetTaskBinding
	saveTask       = store.SaveTaskBinding
	appendAudit    = store.AppendAuditEntry
	resolveSession = func(sessionID string) (sessionInputWriter, bool) {
		session, ok := GetSession(sessionID)
		if !ok || session == nil {
			return nil, false
		}
		return session, true
	}
)

const (
	defaultAgentCommandTimeoutSeconds = 30
	defaultAgentMaxCommands           = 20
)

type sessionInputWriter interface {
	WriteInput([]byte) error
}

type ExecuteCommandResult struct {
	TaskID            string       `json:"task_id"`
	SessionID         string       `json:"session_id,omitempty"`
	Status            string       `json:"status"`
	Executed          bool         `json:"executed"`
	RequiresApproval  bool         `json:"requires_approval"`
	ApprovalAccepted  bool         `json:"approval_accepted"`
	Safety            SafetyResult `json:"safety"`
	Message           string       `json:"message"`
	ExecutedAtUnixSec int64        `json:"executed_at_unix_sec,omitempty"`
}

func ExecuteTaskCommand(ctx context.Context, taskID string, command string, approved bool) (ExecuteCommandResult, error) {
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(agentCommandTimeoutSeconds())*time.Second)
	defer cancel()

	trimmedTaskID := strings.TrimSpace(taskID)
	trimmedCommand := strings.TrimSpace(command)

	binding, err := getTaskBinding(execCtx, trimmedTaskID)
	if err != nil {
		return ExecuteCommandResult{}, err
	}
	if binding == nil {
		return ExecuteCommandResult{}, ErrTaskNotFound
	}

	if binding.CommandCount >= agentMaxCommands() {
		return ExecuteCommandResult{}, ErrCommandLimitExceeded
	}

	safety := ClassifyCommand(trimmedCommand)
	result := ExecuteCommandResult{
		TaskID:           trimmedTaskID,
		SessionID:        binding.SessionID,
		Safety:           safety,
		RequiresApproval: safety.RequiresApproval,
		ApprovalAccepted: approved,
	}

	if safety.Level == model.SafetyLevelBlocked || safety.Level == model.SafetyLevelDangerous {
		result.Status = model.AgentStatusAwaitingApproval
		result.Message = "command requires manual handling due to safety policy"
		if err := appendAuditRecord(execCtx, trimmedTaskID, binding.SessionID, trimmedCommand, safety.Level, approved, result.Status); err != nil {
			return ExecuteCommandResult{}, fmt.Errorf("append audit entry: %w", err)
		}
		return result, nil
	}

	if safety.RequiresApproval && !approved {
		result.Status = model.AgentStatusAwaitingApproval
		result.Message = "command requires approval"
		if err := appendAuditRecord(execCtx, trimmedTaskID, binding.SessionID, trimmedCommand, safety.Level, approved, result.Status); err != nil {
			return ExecuteCommandResult{}, fmt.Errorf("append audit entry: %w", err)
		}
		return result, nil
	}

	sshSession, ok := resolveSession(binding.SessionID)
	if !ok || sshSession == nil {
		return ExecuteCommandResult{}, ErrSessionNotFound
	}

	if err := sshSession.WriteInput([]byte(trimmedCommand + "\n")); err != nil {
		return ExecuteCommandResult{}, fmt.Errorf("write command to session: %w", err)
	}

	result.Status = model.AgentStatusRunning
	result.Executed = true
	result.Message = "command dispatched"
	result.ExecutedAtUnixSec = time.Now().Unix()

	binding.Status = model.AgentStatusRunning
	binding.CommandCount++
	if err := saveTask(execCtx, trimmedTaskID, *binding); err != nil {
		return ExecuteCommandResult{}, fmt.Errorf("save task binding: %w", err)
	}
	if err := appendAuditRecord(execCtx, trimmedTaskID, binding.SessionID, trimmedCommand, safety.Level, approved, result.Status); err != nil {
		return ExecuteCommandResult{}, fmt.Errorf("append audit entry: %w", err)
	}

	return result, nil
}

func appendAuditRecord(ctx context.Context, taskID, sessionID, command, riskLevel string, approved bool, status string) error {
	return appendAudit(ctx, taskID, store.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Command:    command,
		RiskLevel:  riskLevel,
		Approved:   approved,
		SessionID:  sessionID,
		TaskStatus: status,
	})
}

func IsApprovalStatus(status string) bool {
	return status == model.AgentStatusAwaitingApproval
}

func agentCommandTimeoutSeconds() int {
	v := config.Get().AgentCommandTimeout
	if v <= 0 {
		return defaultAgentCommandTimeoutSeconds
	}
	return v
}

func agentMaxCommands() int {
	v := config.Get().AgentMaxCommands
	if v <= 0 {
		return defaultAgentMaxCommands
	}
	return v
}
