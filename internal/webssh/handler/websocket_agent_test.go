package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/service"
	"VPSBenchmarkBackend/internal/webssh/store"

	"github.com/gorilla/websocket"
)

var legacyAgentEventTypes = map[model.MessageType]struct{}{
	model.MessageType("agent_update"):   {},
	model.MessageType("agent_approval"): {},
	model.MessageType("agent_error"):    {},
	model.MessageType("agent_done"):     {},
}

type spyAgentStreamBridge struct {
	mu              sync.Mutex
	owners          map[string]string
	senders         map[string]func(model.ServerMessage)
	unregisterCalls map[string]int
	unregisterCh    chan string
	deliveries      int32
}

func newSpyAgentStreamBridge() *spyAgentStreamBridge {
	return &spyAgentStreamBridge{
		owners:          make(map[string]string),
		senders:         make(map[string]func(model.ServerMessage)),
		unregisterCalls: make(map[string]int),
		unregisterCh:    make(chan string, 8),
	}
}

func (b *spyAgentStreamBridge) Register(taskID string, ownerID string, sender service.AgentStreamSender) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.owners[taskID] = ownerID
	b.senders[taskID] = sender
}

func (b *spyAgentStreamBridge) Unregister(taskID string, ownerID string) {
	b.mu.Lock()
	if b.owners[taskID] != ownerID {
		b.mu.Unlock()
		return
	}
	delete(b.owners, taskID)
	delete(b.senders, taskID)
	b.unregisterCalls[taskID]++
	b.mu.Unlock()
	b.unregisterCh <- taskID
}

func (b *spyAgentStreamBridge) EmitMessageStart(taskID string, messageID string) {
	b.emit(taskID, model.ServerMessage{Type: model.TypeAgentMessageStart, TaskID: taskID, MessageID: messageID})
}

func (b *spyAgentStreamBridge) EmitToken(taskID string, messageID string, delta string) {
	b.emit(taskID, model.ServerMessage{Type: model.TypeAgentToken, TaskID: taskID, MessageID: messageID, Delta: delta})
}

func (b *spyAgentStreamBridge) EmitMessageEnd(taskID string, messageID string, reason string) {
	b.emit(taskID, model.ServerMessage{Type: model.TypeAgentMessageEnd, TaskID: taskID, MessageID: messageID, FinishReason: reason})
}

func (b *spyAgentStreamBridge) EmitState(taskID string, state string, message string) {
	b.emit(taskID, model.ServerMessage{Type: model.TypeAgentState, TaskID: taskID, State: state, Message: message})
}

func (b *spyAgentStreamBridge) emit(taskID string, msg model.ServerMessage) {
	b.mu.Lock()
	sender := b.senders[taskID]
	b.mu.Unlock()
	if sender == nil {
		return
	}
	atomic.AddInt32(&b.deliveries, 1)
	sender(msg)
}

func (b *spyAgentStreamBridge) unregisterCount(taskID string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.unregisterCalls[taskID]
}

func (b *spyAgentStreamBridge) deliveriesCount() int32 {
	return atomic.LoadInt32(&b.deliveries)
}

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

	stateMsg := readAgentServerMessage(t, conn)
	assertNotLegacyAgentEventType(t, stateMsg.Type)
	if stateMsg.Type != model.TypeAgentState {
		t.Fatalf("expected message type %q, got %q", model.TypeAgentState, stateMsg.Type)
	}
	if stateMsg.TaskID != "py-task-1" {
		t.Fatalf("expected task_id py-task-1, got %q", stateMsg.TaskID)
	}
	if stateMsg.State == "" {
		t.Fatal("expected non-empty state")
	}

	if err := conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond)); err != nil {
		t.Fatalf("set read deadline failed: %v", err)
	}
	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Fatal("expected no extra websocket agent event after state ack")
	}

}

func TestWebSocketAgentTaskCreateSendsOnlyStateAck(t *testing.T) {
	originalNewAgentClient := newAgentClient
	originalSaveTaskBinding := saveTaskBinding
	defer func() {
		newAgentClient = originalNewAgentClient
		saveTaskBinding = originalSaveTaskBinding
	}()

	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{
			createTask: func(_ context.Context, prompt string, metadata map[string]any) (service.CreateTaskResult, error) {
				if prompt != "check disk" {
					t.Fatalf("expected prompt check disk, got %q", prompt)
				}
				if metadata["session_id"] != "session-test" {
					t.Fatalf("expected session metadata session-test, got %v", metadata["session_id"])
				}
				return service.CreateTaskResult{TaskID: "py-task-stream", Status: model.AgentStatusRunning}, nil
			},
		}
	}
	saveTaskBinding = func(_ context.Context, taskID string, binding store.TaskBinding) error {
		if taskID != "py-task-stream" {
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

	if err := conn.WriteJSON(model.ClientMessage{Type: model.TypeAgentTask, Message: "check disk"}); err != nil {
		t.Fatalf("write agent_task failed: %v", err)
	}

	stateMsg := readAgentServerMessage(t, conn)
	assertNotLegacyAgentEventType(t, stateMsg.Type)
	if stateMsg.Type != model.TypeAgentState {
		t.Fatalf("expected first message type %q, got %q", model.TypeAgentState, stateMsg.Type)
	}
	if stateMsg.TaskID != "py-task-stream" {
		t.Fatalf("expected task_id py-task-stream, got %q", stateMsg.TaskID)
	}
	if stateMsg.State == "" {
		t.Fatal("expected non-empty state")
	}

	if err := conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond)); err != nil {
		t.Fatalf("set read deadline failed: %v", err)
	}
	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Fatal("expected no synthetic stream triplet while async stream events are enabled")
	}
}

func TestWebSocketAgentDisconnectUnregistersTaskStream(t *testing.T) {
	originalNewAgentClient := newAgentClient
	originalSaveTaskBinding := saveTaskBinding
	originalBridge := agentStreamBridge
	spyBridge := newSpyAgentStreamBridge()
	agentStreamBridge = spyBridge
	defer func() {
		newAgentClient = originalNewAgentClient
		saveTaskBinding = originalSaveTaskBinding
		agentStreamBridge = originalBridge
	}()

	newAgentClient = func() service.AgentClient {
		return &fakeAgentClient{
			createTask: func(_ context.Context, _ string, _ map[string]any) (service.CreateTaskResult, error) {
				return service.CreateTaskResult{TaskID: "py-task-disconnect", Status: model.AgentStatusRunning}, nil
			},
		}
	}
	saveTaskBinding = func(_ context.Context, taskID string, binding store.TaskBinding) error {
		if taskID != "py-task-disconnect" {
			t.Fatalf("unexpected task id save: %s", taskID)
		}
		if binding.UserID != 1 {
			t.Fatalf("unexpected user id in binding: %+v", binding)
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

	if err := conn.WriteJSON(model.ClientMessage{Type: model.TypeAgentTask, Message: "check disk"}); err != nil {
		t.Fatalf("write agent_task failed: %v", err)
	}

	for i := 0; i < 1; i++ {
		_ = readAgentServerMessage(t, conn)
	}

	if err := conn.Close(); err != nil {
		t.Fatalf("close conn failed: %v", err)
	}

	select {
	case taskID := <-spyBridge.unregisterCh:
		if taskID != "py-task-disconnect" {
			t.Fatalf("expected unregister for py-task-disconnect, got %q", taskID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for unregister on disconnect")
	}

	before := spyBridge.deliveriesCount()
	spyBridge.EmitToken("py-task-disconnect", "m1", "should-not-deliver")
	after := spyBridge.deliveriesCount()
	if after != before {
		t.Fatalf("expected no post-disconnect delivery, before=%d after=%d", before, after)
	}
	if spyBridge.unregisterCount("py-task-disconnect") != 1 {
		t.Fatalf("expected exactly one unregister call, got %d", spyBridge.unregisterCount("py-task-disconnect"))
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

func assertNotLegacyAgentEventType(t *testing.T, msgType model.MessageType) {
	t.Helper()
	if _, found := legacyAgentEventTypes[msgType]; found {
		t.Fatalf("legacy agent event type must not be emitted: %q", msgType)
	}
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
