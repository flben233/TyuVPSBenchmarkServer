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
	}

	c.JSON(http.StatusOK, common.Success(gin.H{"tools": tools}))
}
