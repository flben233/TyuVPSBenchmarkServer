package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/service"

	"github.com/gin-gonic/gin"
)

const taskIDHeader = "X-Task-ID"

var executeTaskCommand = service.ExecuteTaskCommand

var allowedAgentStates = map[string]struct{}{
	"thinking":          {},
	"running_command":   {},
	"awaiting_approval": {},
	"done":              {},
	"error":             {},
}

func ExecuteTaskCommandForTest(fn func(context.Context, string, string, bool) (service.ExecuteCommandResult, error)) {
	executeTaskCommand = fn
}

func ResetExecuteTaskCommandForTest() {
	executeTaskCommand = service.ExecuteTaskCommand
}

type SafetyCheckRequest struct {
	Command string `json:"command" binding:"required"`
}

type ExecuteRequest struct {
	Command  string `json:"command" binding:"required"`
	Approved bool   `json:"approved"`
}

type ToolDefinition struct {
	Name        string `json:"name"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type streamEventPayload struct {
	Type         string `json:"type" binding:"required"`
	TaskID       string `json:"task_id" binding:"required"`
	MessageID    string `json:"message_id"`
	Delta        string `json:"delta"`
	FinishReason string `json:"finish_reason"`
	State        string `json:"state"`
	Message      string `json:"message"`
}

func HandleSafetyCheck(c *gin.Context) {
	var req SafetyCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "invalid request: "+err.Error()))
		return
	}

	result := service.ClassifyCommand(req.Command)
	c.JSON(http.StatusOK, common.Success(result))
}

func HandleExecute(c *gin.Context) {
	taskID := strings.TrimSpace(c.GetHeader(taskIDHeader))
	if taskID == "" {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "X-Task-ID header is required"))
		return
	}

	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "invalid request: "+err.Error()))
		return
	}

	result, err := executeTaskCommand(c.Request.Context(), taskID, req.Command, req.Approved)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			c.JSON(http.StatusNotFound, common.Error(common.BadRequestCode, "task binding not found"))
			return
		case errors.Is(err, service.ErrSessionNotFound):
			c.JSON(http.StatusNotFound, common.Error(common.BadRequestCode, "session not found"))
			return
		case errors.Is(err, service.ErrCommandLimitExceeded):
			c.JSON(http.StatusTooManyRequests, common.Error(common.LimitExceededCode, "command limit exceeded"))
			return
		default:
			c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "execute command failed"))
			return
		}
	}

	c.JSON(http.StatusOK, common.Success(result))
}

func HandleStreamEvent(c *gin.Context) {
	var req streamEventPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "invalid request: "+err.Error()))
		return
	}

	taskID := strings.TrimSpace(req.TaskID)
	if taskID == "" {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "task_id is required"))
		return
	}
	// Stream events accept task_id from body for compatibility with async emitters.
	// If X-Task-ID is present, it must match body task_id to prevent cross-task routing mistakes.
	headerTaskID := strings.TrimSpace(c.GetHeader(taskIDHeader))
	if headerTaskID != "" && headerTaskID != taskID {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "task_id mismatch between header and body"))
		return
	}

	eventType := strings.TrimSpace(req.Type)
	switch eventType {
	case "agent_message_start":
		messageID := strings.TrimSpace(req.MessageID)
		if messageID == "" {
			c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "message_id is required for agent_message_start"))
			return
		}
		agentStreamBridge.EmitMessageStart(taskID, messageID)
	case "agent_token":
		messageID := strings.TrimSpace(req.MessageID)
		if messageID == "" {
			c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "message_id is required for agent_token"))
			return
		}
		agentStreamBridge.EmitToken(taskID, messageID, req.Delta)
	case "agent_message_end":
		messageID := strings.TrimSpace(req.MessageID)
		if messageID == "" {
			c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "message_id is required for agent_message_end"))
			return
		}
		reason := strings.TrimSpace(req.FinishReason)
		if reason == "" {
			reason = "stop"
		}
		agentStreamBridge.EmitMessageEnd(taskID, messageID, reason)
	case "agent_state":
		state := strings.TrimSpace(req.State)
		if state == "" {
			c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "state is required for agent_state"))
			return
		}
		if _, ok := allowedAgentStates[state]; !ok {
			c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "unsupported state for agent_state"))
			return
		}
		agentStreamBridge.EmitState(taskID, state, req.Message)
	default:
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "unsupported stream event type"))
		return
	}

	c.JSON(http.StatusOK, common.Success(gin.H{"accepted": true}))
}

func HandleTools(c *gin.Context) {
	tools := []ToolDefinition{
		{
			Name:        "safety-check",
			Method:      http.MethodPost,
			Path:        "/agent/safety-check",
			Description: "classify command safety without execution",
		},
		{
			Name:        "execute",
			Method:      http.MethodPost,
			Path:        "/agent/execute",
			Description: "execute command by task-bound webssh session",
		},
		{
			Name:        "tools",
			Method:      http.MethodGet,
			Path:        "/agent/tools",
			Description: "list available internal agent tools",
		},
		{
			Name:        "stream-event",
			Method:      http.MethodPost,
			Path:        "/agent/stream-event",
			Description: "ingest async agent stream events for websocket dispatch",
		},
	}

	c.JSON(http.StatusOK, common.Success(gin.H{"tools": tools}))
}
