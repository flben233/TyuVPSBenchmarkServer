package handler

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"VPSBenchmarkBackend/internal/cache"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/service"
	"VPSBenchmarkBackend/internal/webssh/store"

	"github.com/gin-gonic/gin"
)

var (
	getTaskBinding      = store.GetTaskBinding
	saveTaskBinding     = store.SaveTaskBinding
	listTaskBindingsFor = listTaskBindingsForUser
	newAgentClient      = service.NewAgentClient
)

type CreateTaskRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Prompt    string `json:"prompt" binding:"required"`
}

type AgentMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

type AgentApprovalRequest struct {
	Approved *bool `json:"approved" binding:"required"`
}

type TaskSummary struct {
	TaskID       string    `json:"task_id"`
	SessionID    string    `json:"session_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	CommandCount int       `json:"command_count"`
}

func HandleCreateTask(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "Invalid request: "+err.Error()))
		return
	}
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "session_id is required"))
		return
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "prompt is required"))
		return
	}

	agentClient := newAgentClient()
	createResult, err := agentClient.CreateTask(c.Request.Context(), prompt, map[string]any{
		"session_id": sessionID,
	})
	if err != nil {
		respondAgentError(c, err, "failed to create task")
		return
	}
	taskID := strings.TrimSpace(createResult.TaskID)
	if taskID == "" {
		c.JSON(http.StatusBadGateway, common.Error(common.InternalErrorCode, "agent service returned empty task id"))
		return
	}

	status := strings.TrimSpace(createResult.Status)
	if status == "" {
		status = model.AgentStatusRunning
	}
	binding := store.TaskBinding{
		UserID:       userID.(int64),
		SessionID:    sessionID,
		Status:       status,
		CreatedAt:    time.Now().UTC(),
		CommandCount: 0,
	}

	if err := saveTaskBinding(c.Request.Context(), taskID, binding); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "failed to create task"))
		return
	}

	c.JSON(http.StatusOK, common.Success(TaskSummary{
		TaskID:       taskID,
		SessionID:    binding.SessionID,
		Status:       binding.Status,
		CreatedAt:    binding.CreatedAt,
		CommandCount: binding.CommandCount,
	}))
}

func HandleTaskMessage(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	binding, ok := loadOwnedTaskBinding(c, taskID)
	if !ok {
		return
	}

	var req AgentMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "Invalid request: "+err.Error()))
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "message is required"))
		return
	}

	agentClient := newAgentClient()
	reply, err := agentClient.SendTaskMessage(c.Request.Context(), taskID, message)
	if err != nil {
		respondAgentError(c, err, "failed to send task message")
		return
	}
	if !reply.OK {
		c.JSON(http.StatusBadGateway, common.Error(common.InternalErrorCode, "agent service rejected task message"))
		return
	}

	nextStatus := deriveTaskStatus(binding.Status, reply)
	if nextStatus != binding.Status {
		binding.Status = nextStatus
		if err := saveTaskBinding(c.Request.Context(), taskID, *binding); err != nil {
			c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "failed to update task status"))
			return
		}
	}

	c.JSON(http.StatusOK, common.Success(gin.H{
		"task_id": taskID,
		"status":  binding.Status,
		"agent":   reply,
	}))
}

func HandleTaskApprove(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	binding, ok := loadOwnedTaskBinding(c, taskID)
	if !ok {
		return
	}

	var req AgentApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "Invalid request: "+err.Error()))
		return
	}
	if req.Approved == nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "approved is required"))
		return
	}

	agentClient := newAgentClient()
	reply, err := agentClient.ApproveTask(c.Request.Context(), taskID, *req.Approved)
	if err != nil {
		respondAgentError(c, err, "failed to approve task")
		return
	}
	if !reply.OK {
		c.JSON(http.StatusBadGateway, common.Error(common.InternalErrorCode, "agent service rejected task approval"))
		return
	}

	if *req.Approved {
		binding.Status = model.AgentStatusRunning
	} else {
		binding.Status = model.AgentStatusFailed
	}

	if err := saveTaskBinding(c.Request.Context(), taskID, *binding); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "failed to update task approval"))
		return
	}

	c.JSON(http.StatusOK, common.Success(gin.H{
		"task_id":  taskID,
		"status":   binding.Status,
		"approved": *req.Approved,
		"agent":    reply,
	}))
}

func HandleGetTask(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	binding, ok := loadOwnedTaskBinding(c, taskID)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, common.Success(TaskSummary{
		TaskID:       taskID,
		SessionID:    binding.SessionID,
		Status:       binding.Status,
		CreatedAt:    binding.CreatedAt,
		CommandCount: binding.CommandCount,
	}))
}

func HandleListTasks(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	tasks, err := listTaskBindingsFor(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "failed to list tasks"))
		return
	}

	c.JSON(http.StatusOK, common.Success(tasks))
}

func loadOwnedTaskBinding(c *gin.Context, taskID string) (*store.TaskBinding, bool) {
	if taskID == "" {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "task_id is required"))
		return nil, false
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return nil, false
	}

	binding, err := getTaskBinding(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "failed to query task"))
		return nil, false
	}
	if binding == nil || binding.UserID != userID.(int64) {
		c.JSON(http.StatusNotFound, common.Error(common.BadRequestCode, "task not found"))
		return nil, false
	}

	return binding, true
}

func listTaskBindingsForUser(ctx context.Context, userID int64) ([]TaskSummary, error) {
	redisClient := cache.GetClient()
	cursor := uint64(0)
	out := make([]TaskSummary, 0)

	for {
		keys, nextCursor, err := redisClient.Scan(ctx, cursor, "task:*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			taskID := strings.TrimPrefix(key, "task:")
			binding, err := getTaskBinding(ctx, taskID)
			if err != nil {
				return nil, err
			}
			if binding == nil || binding.UserID != userID {
				continue
			}
			out = append(out, TaskSummary{
				TaskID:       taskID,
				SessionID:    binding.SessionID,
				Status:       binding.Status,
				CreatedAt:    binding.CreatedAt,
				CommandCount: binding.CommandCount,
			})
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})

	return out, nil
}

func respondAgentError(c *gin.Context, err error, defaultMessage string) {
	switch {
	case errors.Is(err, service.ErrAgentURLNotConfigured):
		c.JSON(http.StatusServiceUnavailable, common.Error(common.InternalErrorCode, "agent url is not configured"))
	case errors.Is(err, service.ErrAgentTaskNotFound):
		c.JSON(http.StatusNotFound, common.Error(common.BadRequestCode, "task not found in agent service"))
	case errors.Is(err, service.ErrAgentRequestInvalid):
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, err.Error()))
	case errors.Is(err, service.ErrAgentUnavailable):
		c.JSON(http.StatusBadGateway, common.Error(common.InternalErrorCode, "agent service unavailable"))
	case errors.Is(err, service.ErrAgentBadResponse):
		c.JSON(http.StatusBadGateway, common.Error(common.InternalErrorCode, "invalid response from agent service"))
	default:
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, defaultMessage))
	}
}

func deriveTaskStatus(current string, reply service.AgentReply) string {
	if reply.Data == nil {
		return current
	}

	if done, ok := reply.Data["task_complete"].(bool); ok && done {
		return model.AgentStatusCompleted
	}
	if waiting, ok := reply.Data["awaiting_approval"].(bool); ok && waiting {
		return model.AgentStatusAwaitingApproval
	}

	if current == "" {
		return model.AgentStatusRunning
	}
	return current
}
