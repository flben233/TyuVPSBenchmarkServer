package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/service"
	"VPSBenchmarkBackend/internal/webssh/store"

	"github.com/gin-gonic/gin"
)

type fakeAgentClient struct {
	createTask func(ctx context.Context, prompt string, metadata map[string]any) (service.CreateTaskResult, error)
	sendTask   func(ctx context.Context, taskID string, message string) (service.AgentReply, error)
	approve    func(ctx context.Context, taskID string, approved bool) (service.AgentReply, error)
}

func (f *fakeAgentClient) CreateTask(ctx context.Context, prompt string, metadata map[string]any) (service.CreateTaskResult, error) {
	if f.createTask != nil {
		return f.createTask(ctx, prompt, metadata)
	}
	return service.CreateTaskResult{TaskID: "py-task-1", Status: model.AgentStatusRunning}, nil
}

func (f *fakeAgentClient) SendTaskMessage(ctx context.Context, taskID string, message string) (service.AgentReply, error) {
	if f.sendTask != nil {
		return f.sendTask(ctx, taskID, message)
	}
	return service.AgentReply{OK: true, Message: "ok", Data: map[string]any{}}, nil
}

func (f *fakeAgentClient) ApproveTask(ctx context.Context, taskID string, approved bool) (service.AgentReply, error) {
	if f.approve != nil {
		return f.approve(ctx, taskID, approved)
	}
	return service.AgentReply{OK: true, Message: "ok", Data: map[string]any{}}, nil
}

func TestHandleTaskMessageForwardsToPythonAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalGet := getTaskBinding
	originalSave := saveTaskBinding
	originalNewAgentClient := newAgentClient
	defer func() {
		getTaskBinding = originalGet
		saveTaskBinding = originalSave
		newAgentClient = originalNewAgentClient
	}()

	getTaskBinding = func(_ context.Context, taskID string) (*store.TaskBinding, error) {
		if taskID != "task-1" {
			t.Fatalf("unexpected task id %q", taskID)
		}
		return &store.TaskBinding{UserID: 42, SessionID: "session-a", Status: model.AgentStatusRunning}, nil
	}
	saveTaskBinding = func(_ context.Context, _ string, _ store.TaskBinding) error {
		return nil
	}

	var capturedTaskID string
	var capturedMessage string
	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{
			sendTask: func(_ context.Context, taskID string, message string) (service.AgentReply, error) {
				capturedTaskID = taskID
				capturedMessage = message
				return service.AgentReply{OK: true, Message: "accepted", Data: map[string]any{"awaiting_approval": true}}, nil
			},
		}
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	r.POST("/api/agent/tasks/:task_id/message", HandleTaskMessage)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/task-1/message", bytes.NewReader([]byte(`{"message":"run uptime"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if capturedTaskID != "task-1" {
		t.Fatalf("expected task id task-1, got %q", capturedTaskID)
	}
	if capturedMessage != "run uptime" {
		t.Fatalf("expected message run uptime, got %q", capturedMessage)
	}
}

func TestHandleTaskMessageOwnershipGate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalGet := getTaskBinding
	originalNewAgentClient := newAgentClient
	defer func() {
		getTaskBinding = originalGet
		newAgentClient = originalNewAgentClient
	}()

	getTaskBinding = func(_ context.Context, _ string) (*store.TaskBinding, error) {
		return &store.TaskBinding{UserID: 7, SessionID: "session-a", Status: model.AgentStatusRunning}, nil
	}
	newAgentClient = func() service.AgentClient {
		t.Fatal("agent client should not be called when ownership check fails")
		return &fakeAgentClient{}
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	r.POST("/api/agent/tasks/:task_id/message", HandleTaskMessage)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/task-1/message", bytes.NewReader([]byte(`{"message":"run uptime"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestHandleTaskMessageMappedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalGet := getTaskBinding
	originalNewAgentClient := newAgentClient
	defer func() {
		getTaskBinding = originalGet
		newAgentClient = originalNewAgentClient
	}()

	getTaskBinding = func(_ context.Context, _ string) (*store.TaskBinding, error) {
		return &store.TaskBinding{UserID: 42, SessionID: "session-a", Status: model.AgentStatusRunning}, nil
	}
	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{
			sendTask: func(_ context.Context, _ string, _ string) (service.AgentReply, error) {
				return service.AgentReply{}, service.ErrAgentTaskNotFound
			},
		}
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	r.POST("/api/agent/tasks/:task_id/message", HandleTaskMessage)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/task-1/message", bytes.NewReader([]byte(`{"message":"run uptime"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, rec.Code, rec.Body.String())
	}

	var resp common.APIResponse[any]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != common.BadRequestCode {
		t.Fatalf("expected response code %d, got %d", common.BadRequestCode, resp.Code)
	}
}

func TestHandleTaskMessageReplyNotOKReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalGet := getTaskBinding
	originalSave := saveTaskBinding
	originalNewAgentClient := newAgentClient
	defer func() {
		getTaskBinding = originalGet
		saveTaskBinding = originalSave
		newAgentClient = originalNewAgentClient
	}()

	getTaskBinding = func(_ context.Context, _ string) (*store.TaskBinding, error) {
		return &store.TaskBinding{UserID: 42, SessionID: "session-a", Status: model.AgentStatusRunning}, nil
	}
	saveTaskBinding = func(_ context.Context, _ string, _ store.TaskBinding) error {
		t.Fatal("saveTaskBinding should not be called when agent reply is not ok")
		return nil
	}
	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{
			sendTask: func(_ context.Context, _ string, _ string) (service.AgentReply, error) {
				return service.AgentReply{OK: false, Message: "agent rejected", Data: map[string]any{}}, nil
			},
		}
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	r.POST("/api/agent/tasks/:task_id/message", HandleTaskMessage)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/task-1/message", bytes.NewReader([]byte(`{"message":"run uptime"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadGateway, rec.Code, rec.Body.String())
	}

	var resp common.APIResponse[any]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != common.InternalErrorCode {
		t.Fatalf("expected response code %d, got %d", common.InternalErrorCode, resp.Code)
	}
}

func TestHandleTaskApproveForwardsToPythonAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalGet := getTaskBinding
	originalSave := saveTaskBinding
	originalNewAgentClient := newAgentClient
	defer func() {
		getTaskBinding = originalGet
		saveTaskBinding = originalSave
		newAgentClient = originalNewAgentClient
	}()

	getTaskBinding = func(_ context.Context, taskID string) (*store.TaskBinding, error) {
		if taskID != "task-1" {
			t.Fatalf("unexpected task id %q", taskID)
		}
		return &store.TaskBinding{UserID: 42, SessionID: "session-a", Status: model.AgentStatusAwaitingApproval}, nil
	}
	saveTaskBinding = func(_ context.Context, taskID string, binding store.TaskBinding) error {
		if taskID != "task-1" {
			t.Fatalf("unexpected task id %q", taskID)
		}
		if binding.Status != model.AgentStatusRunning {
			t.Fatalf("expected status running after approval, got %q", binding.Status)
		}
		return nil
	}

	var capturedTaskID string
	var capturedApproved bool
	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{
			approve: func(_ context.Context, taskID string, approved bool) (service.AgentReply, error) {
				capturedTaskID = taskID
				capturedApproved = approved
				return service.AgentReply{OK: true, Message: "approval accepted", Data: map[string]any{}}, nil
			},
		}
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	r.POST("/api/agent/tasks/:task_id/approve", HandleTaskApprove)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/task-1/approve", bytes.NewReader([]byte(`{"approved":true}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if capturedTaskID != "task-1" {
		t.Fatalf("expected task id task-1, got %q", capturedTaskID)
	}
	if !capturedApproved {
		t.Fatal("expected approved=true forwarded")
	}
}

func TestHandleTaskApproveOwnershipGate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalGet := getTaskBinding
	originalNewAgentClient := newAgentClient
	defer func() {
		getTaskBinding = originalGet
		newAgentClient = originalNewAgentClient
	}()

	getTaskBinding = func(_ context.Context, _ string) (*store.TaskBinding, error) {
		return &store.TaskBinding{UserID: 7, SessionID: "session-a", Status: model.AgentStatusAwaitingApproval}, nil
	}
	newAgentClient = func() service.AgentClient {
		t.Fatal("agent client should not be called when ownership check fails")
		return &fakeAgentClient{}
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	r.POST("/api/agent/tasks/:task_id/approve", HandleTaskApprove)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/task-1/approve", bytes.NewReader([]byte(`{"approved":true}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestHandleTaskApproveMappedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalGet := getTaskBinding
	originalNewAgentClient := newAgentClient
	defer func() {
		getTaskBinding = originalGet
		newAgentClient = originalNewAgentClient
	}()

	getTaskBinding = func(_ context.Context, _ string) (*store.TaskBinding, error) {
		return &store.TaskBinding{UserID: 42, SessionID: "session-a", Status: model.AgentStatusAwaitingApproval}, nil
	}
	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{
			approve: func(_ context.Context, _ string, _ bool) (service.AgentReply, error) {
				return service.AgentReply{}, service.ErrAgentUnavailable
			},
		}
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	r.POST("/api/agent/tasks/:task_id/approve", HandleTaskApprove)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/task-1/approve", bytes.NewReader([]byte(`{"approved":true}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadGateway, rec.Code, rec.Body.String())
	}

	var resp common.APIResponse[any]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != common.InternalErrorCode {
		t.Fatalf("expected response code %d, got %d", common.InternalErrorCode, resp.Code)
	}
}

func TestHandleTaskApproveRequiresExplicitApproved(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalGet := getTaskBinding
	originalNewAgentClient := newAgentClient
	defer func() {
		getTaskBinding = originalGet
		newAgentClient = originalNewAgentClient
	}()

	getTaskBinding = func(_ context.Context, _ string) (*store.TaskBinding, error) {
		return &store.TaskBinding{UserID: 42, SessionID: "session-a", Status: model.AgentStatusAwaitingApproval}, nil
	}
	newAgentClient = func() service.AgentClient {
		t.Fatal("agent client should not be called when approved is omitted")
		return &fakeAgentClient{}
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	r.POST("/api/agent/tasks/:task_id/approve", HandleTaskApprove)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/task-1/approve", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	var resp common.APIResponse[any]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != common.InvalidParamCode {
		t.Fatalf("expected response code %d, got %d", common.InvalidParamCode, resp.Code)
	}
}

func TestHandleTaskApproveReplyNotOKReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalGet := getTaskBinding
	originalSave := saveTaskBinding
	originalNewAgentClient := newAgentClient
	defer func() {
		getTaskBinding = originalGet
		saveTaskBinding = originalSave
		newAgentClient = originalNewAgentClient
	}()

	getTaskBinding = func(_ context.Context, _ string) (*store.TaskBinding, error) {
		return &store.TaskBinding{UserID: 42, SessionID: "session-a", Status: model.AgentStatusAwaitingApproval}, nil
	}
	saveTaskBinding = func(_ context.Context, _ string, _ store.TaskBinding) error {
		t.Fatal("saveTaskBinding should not be called when agent reply is not ok")
		return nil
	}
	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{
			approve: func(_ context.Context, _ string, _ bool) (service.AgentReply, error) {
				return service.AgentReply{OK: false, Message: "denied", Data: map[string]any{}}, nil
			},
		}
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	r.POST("/api/agent/tasks/:task_id/approve", HandleTaskApprove)

	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/task-1/approve", bytes.NewReader([]byte(`{"approved":true}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadGateway, rec.Code, rec.Body.String())
	}

	var resp common.APIResponse[any]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != common.InternalErrorCode {
		t.Fatalf("expected response code %d, got %d", common.InternalErrorCode, resp.Code)
	}
}

func TestCreateTaskBindsUserAndSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedTaskID string
	var capturedBinding store.TaskBinding

	originalSave := saveTaskBinding
	originalAgentClientFactory := newAgentClient
	saveTaskBinding = func(_ context.Context, taskID string, binding store.TaskBinding) error {
		capturedTaskID = taskID
		capturedBinding = binding
		return nil
	}
	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{}
	}
	t.Cleanup(func() {
		saveTaskBinding = originalSave
		newAgentClient = originalAgentClientFactory
	})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	group := r.Group("/api/agent")
	group.POST("/tasks", HandleCreateTask)

	body := []byte(`{"session_id":"session-a","prompt":"check disk usage"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	if capturedTaskID != "py-task-1" {
		t.Fatalf("expected forwarded task id py-task-1, got %q", capturedTaskID)
	}
	if capturedBinding.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", capturedBinding.UserID)
	}
	if capturedBinding.SessionID != "session-a" {
		t.Fatalf("expected session id session-a, got %q", capturedBinding.SessionID)
	}
	if capturedBinding.Status != model.AgentStatusRunning {
		t.Fatalf("expected status %q, got %q", model.AgentStatusRunning, capturedBinding.Status)
	}
	if capturedBinding.CommandCount != 0 {
		t.Fatalf("expected command count 0, got %d", capturedBinding.CommandCount)
	}
	if capturedBinding.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be populated")
	}
	if time.Since(capturedBinding.CreatedAt) > 2*time.Second {
		t.Fatalf("expected recent created_at, got %s", capturedBinding.CreatedAt)
	}
}

func TestCreateTaskRequiresUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	saveCalled := false
	originalSave := saveTaskBinding
	originalAgentClientFactory := newAgentClient
	saveTaskBinding = func(_ context.Context, _ string, _ store.TaskBinding) error {
		saveCalled = true
		return nil
	}
	newAgentClient = func() service.AgentClient {
		t.Fatal("agent client should not be created without user context")
		return &fakeAgentClient{}
	}
	t.Cleanup(func() {
		saveTaskBinding = originalSave
		newAgentClient = originalAgentClientFactory
	})

	r := gin.New()
	group := r.Group("/api/agent")
	group.POST("/tasks", HandleCreateTask)

	body := []byte(`{"session_id":"session-a","prompt":"check disk usage"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusUnauthorized, rec.Code, rec.Body.String())
	}
	if saveCalled {
		t.Fatal("expected binding store not called when user context is missing")
	}

	var resp common.APIResponse[any]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != common.BadRequestCode {
		t.Fatalf("expected response code %d, got %d", common.BadRequestCode, resp.Code)
	}
}

func TestCreateTaskRejectsWhitespaceSessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	saveCalled := false
	originalSave := saveTaskBinding
	originalAgentClientFactory := newAgentClient
	saveTaskBinding = func(_ context.Context, _ string, _ store.TaskBinding) error {
		saveCalled = true
		return nil
	}
	newAgentClient = func() service.AgentClient {
		t.Fatal("agent client should not be created for invalid session id")
		return &fakeAgentClient{}
	}
	t.Cleanup(func() {
		saveTaskBinding = originalSave
		newAgentClient = originalAgentClientFactory
	})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	group := r.Group("/api/agent")
	group.POST("/tasks", HandleCreateTask)

	body := []byte(`{"session_id":"   \t  ","prompt":"check disk usage"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
	if saveCalled {
		t.Fatal("expected binding store not called for invalid session id")
	}

	var resp common.APIResponse[any]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != common.InvalidParamCode {
		t.Fatalf("expected response code %d, got %d", common.InvalidParamCode, resp.Code)
	}
}
