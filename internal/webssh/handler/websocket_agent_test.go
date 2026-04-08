package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/service"
	"VPSBenchmarkBackend/internal/webssh/store"

	"github.com/gorilla/websocket"
)

func TestWebSocketAgentApprovalSequence(t *testing.T) {
	originalNewAgentClient := newAgentClient
	originalGetTaskBinding := getTaskBinding
	originalSaveTaskBinding := saveTaskBinding
	defer func() {
		newAgentClient = originalNewAgentClient
		getTaskBinding = originalGetTaskBinding
		saveTaskBinding = originalSaveTaskBinding
	}()

	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{
			createTask: func(_ context.Context, prompt string, metadata map[string]any) (service.CreateTaskResult, error) {
				if prompt != "" {
					t.Fatalf("expected empty prompt from websocket test task, got %q", prompt)
				}
				if metadata["session_id"] != "session-test" {
					t.Fatalf("expected session metadata session-test, got %v", metadata["session_id"])
				}
				return service.CreateTaskResult{TaskID: "py-task-1", Status: model.AgentStatusRunning}, nil
			},
		}
	}
	getTaskBinding = func(_ context.Context, taskID string) (*store.TaskBinding, error) {
		if taskID != "py-task-1" {
			t.Fatalf("unexpected task id lookup: %s", taskID)
		}
		return &store.TaskBinding{UserID: 1, SessionID: "session-test", Status: model.AgentStatusRunning}, nil
	}
	saveTaskBinding = func(_ context.Context, taskID string, binding store.TaskBinding) error {
		if taskID != "py-task-1" {
			t.Fatalf("unexpected task id save: %s", taskID)
		}
		if binding.UserID != 1 || binding.SessionID != "session-test" {
			t.Fatalf("unexpected binding saved: %+v", binding)
		}
		return nil
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }}
		conn, err := u.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		wsHandler(nil, conn, 1, "session-test")
	}))
	t.Cleanup(server.Close)

	wsURL := "ws" + server.URL[len("http"):]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if err := conn.WriteJSON(model.ClientMessage{Type: model.TypeAgentTask}); err != nil {
		t.Fatalf("write agent_task failed: %v", err)
	}

	createdMsg := readAgentServerMessage(t, conn)
	if createdMsg.Type != model.TypeAgentUpdate {
		t.Fatalf("expected first message type %q, got %q", model.TypeAgentUpdate, createdMsg.Type)
	}
	if createdMsg.TaskID != "py-task-1" {
		t.Fatalf("expected task_id py-task-1, got %q", createdMsg.TaskID)
	}
	if createdMsg.Message == "" {
		t.Fatal("expected task creation update message")
	}
	if createdMsg.Type == model.TypeAgentDone {
		t.Fatal("expected create flow to emit task update before any completion")
	}

	if err := conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond)); err != nil {
		t.Fatalf("set read deadline failed: %v", err)
	}
	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Fatal("expected no synthetic follow-up approval/done message on task creation")
	}

}

func TestWebSocketAgentOriginPolicy(t *testing.T) {
	loadWebSocketTestConfig(t, "https://frontend.example.com")

	req := httptest.NewRequest(http.MethodGet, "http://backend.local/webssh/ws", nil)
	req.Host = "backend.local"
	req.Header.Set("Origin", "https://frontend.example.com")
	if !websocketOriginAllowed(req) {
		t.Fatal("expected configured frontend origin to be allowed")
	}

	req.Header.Set("Origin", "https://evil.example.com")
	if websocketOriginAllowed(req) {
		t.Fatal("expected non-allowlisted origin to be denied")
	}

	loadWebSocketTestConfig(t, "")
	req = httptest.NewRequest(http.MethodGet, "http://backend.local/webssh/ws", nil)
	req.Host = "backend.local"
	req.Header.Set("Origin", "http://backend.local")
	if !websocketOriginAllowed(req) {
		t.Fatal("expected same-host origin allowed when frontendUrl is empty")
	}

	req.Header.Set("Origin", "http://other.local")
	if websocketOriginAllowed(req) {
		t.Fatal("expected different host origin denied when frontendUrl is empty")
	}
}

func readAgentServerMessage(t *testing.T, conn *websocket.Conn) model.ServerMessage {
	t.Helper()
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline failed: %v", err)
	}
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read server message failed: %v", err)
	}
	var msg model.ServerMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		t.Fatalf("unmarshal server message failed: %v", err)
	}
	return msg
}

func loadWebSocketTestConfig(t *testing.T, frontendURL string) {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := []byte(`{"jwtSecret":"test-jwt-secret","frontendUrl":"` + frontendURL + `"}`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	if err := config.Load(configPath); err != nil {
		t.Fatalf("load config failed: %v", err)
	}
}
