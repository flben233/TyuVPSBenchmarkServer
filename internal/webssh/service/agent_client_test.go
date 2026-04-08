package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"VPSBenchmarkBackend/internal/config"
)

func TestCreateTaskForwardsToPythonAgent(t *testing.T) {
	var (
		capturedMethod string
		capturedPath   string
		capturedBody   createTaskPayload
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("decode request body failed: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"task_id":"py-task-1","status":"running"}`))
	}))
	defer server.Close()

	loadAgentClientConfigForTest(t, server.URL)

	client := NewAgentClient()
	result, err := client.CreateTask(context.Background(), "check disk usage", map[string]any{"session_id": "session-a"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedMethod != http.MethodPost {
		t.Fatalf("expected method %s, got %s", http.MethodPost, capturedMethod)
	}
	if capturedPath != "/api/agent/tasks" {
		t.Fatalf("expected path /api/agent/tasks, got %s", capturedPath)
	}
	if capturedBody.Prompt != "check disk usage" {
		t.Fatalf("expected prompt to be forwarded, got %q", capturedBody.Prompt)
	}
	if capturedBody.Metadata["session_id"] != "session-a" {
		t.Fatalf("expected metadata session_id=session-a, got %#v", capturedBody.Metadata["session_id"])
	}
	if result.TaskID != "py-task-1" {
		t.Fatalf("expected task_id py-task-1, got %q", result.TaskID)
	}
	if result.Status != "running" {
		t.Fatalf("expected status running, got %q", result.Status)
	}
}

func TestSendTaskMessageForwardsToPythonAgent(t *testing.T) {
	var (
		capturedMethod string
		capturedPath   string
		capturedBody   taskMessagePayload
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("decode request body failed: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"message":"accepted","data":{"task_id":"py-task-2"}}`))
	}))
	defer server.Close()

	loadAgentClientConfigForTest(t, server.URL)

	client := NewAgentClient()
	result, err := client.SendTaskMessage(context.Background(), "py-task-2", "run uptime")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedMethod != http.MethodPost {
		t.Fatalf("expected method %s, got %s", http.MethodPost, capturedMethod)
	}
	if capturedPath != "/api/agent/tasks/py-task-2/message" {
		t.Fatalf("expected path /api/agent/tasks/py-task-2/message, got %s", capturedPath)
	}
	if capturedBody.Message != "run uptime" {
		t.Fatalf("expected message to be forwarded, got %q", capturedBody.Message)
	}
	if !result.OK {
		t.Fatal("expected ok=true from agent response")
	}
}

func TestSendTaskMessageEscapesTaskIDPathSegment(t *testing.T) {
	var capturedRequestURI string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"message":"accepted","data":{}}`))
	}))
	defer server.Close()

	loadAgentClientConfigForTest(t, server.URL)

	client := NewAgentClient()
	_, err := client.SendTaskMessage(context.Background(), "task with/slash", "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedRequestURI != "/api/agent/tasks/task%20with%2Fslash/message" {
		t.Fatalf("expected escaped task id request uri, got %s", capturedRequestURI)
	}
}

func TestApproveTaskForwardsToPythonAgent(t *testing.T) {
	var (
		capturedMethod string
		capturedPath   string
		capturedBody   approveTaskPayload
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("decode request body failed: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"message":"approval accepted","data":{"task_id":"py-task-3"}}`))
	}))
	defer server.Close()

	loadAgentClientConfigForTest(t, server.URL)

	client := NewAgentClient()
	result, err := client.ApproveTask(context.Background(), "py-task-3", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedMethod != http.MethodPost {
		t.Fatalf("expected method %s, got %s", http.MethodPost, capturedMethod)
	}
	if capturedPath != "/api/agent/tasks/py-task-3/approve" {
		t.Fatalf("expected path /api/agent/tasks/py-task-3/approve, got %s", capturedPath)
	}
	if !capturedBody.Approved {
		t.Fatal("expected approved=true to be forwarded")
	}
	if !result.OK {
		t.Fatal("expected ok=true from agent response")
	}
}

func TestApproveTaskEscapesTaskIDPathSegment(t *testing.T) {
	var capturedRequestURI string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"message":"accepted","data":{}}`))
	}))
	defer server.Close()

	loadAgentClientConfigForTest(t, server.URL)

	client := NewAgentClient()
	_, err := client.ApproveTask(context.Background(), "task with/slash", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedRequestURI != "/api/agent/tasks/task%20with%2Fslash/approve" {
		t.Fatalf("expected escaped task id request uri, got %s", capturedRequestURI)
	}
}

func loadAgentClientConfigForTest(t *testing.T, agentURL string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{"agentUrl":"` + agentURL + `"}`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	if err := config.Load(path); err != nil {
		t.Fatalf("load config failed: %v", err)
	}
}
