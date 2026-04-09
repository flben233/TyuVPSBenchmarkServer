package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh/middleware"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/service"

	"github.com/gin-gonic/gin"
)

type streamEventBridgeSpy struct {
	messageStarts []struct {
		taskID    string
		messageID string
	}
	tokens []struct {
		taskID    string
		messageID string
		delta     string
	}
	messageEnds []struct {
		taskID    string
		messageID string
		reason    string
	}
	states []struct {
		taskID  string
		state   string
		message string
	}
}

func (s *streamEventBridgeSpy) Register(string, string, service.AgentStreamSender) {}
func (s *streamEventBridgeSpy) Unregister(string, string)                          {}

func (s *streamEventBridgeSpy) EmitMessageStart(taskID string, messageID string) {
	s.messageStarts = append(s.messageStarts, struct {
		taskID    string
		messageID string
	}{taskID: taskID, messageID: messageID})
}

func (s *streamEventBridgeSpy) EmitToken(taskID string, messageID string, delta string) {
	s.tokens = append(s.tokens, struct {
		taskID    string
		messageID string
		delta     string
	}{taskID: taskID, messageID: messageID, delta: delta})
}

func (s *streamEventBridgeSpy) EmitMessageEnd(taskID string, messageID string, reason string) {
	s.messageEnds = append(s.messageEnds, struct {
		taskID    string
		messageID string
		reason    string
	}{taskID: taskID, messageID: messageID, reason: reason})
}

func (s *streamEventBridgeSpy) EmitState(taskID string, state string, message string) {
	s.states = append(s.states, struct {
		taskID  string
		state   string
		message string
	}{taskID: taskID, state: state, message: message})
}

func TestHandleStreamEventDispatchesValidPayloads(t *testing.T) {
	g := setupInternalRouterForStreamEvent(t)

	tests := []struct {
		name    string
		payload map[string]any
		assert  func(t *testing.T, spy *streamEventBridgeSpy)
	}{
		{
			name: "message start",
			payload: map[string]any{
				"type":       string(model.TypeAgentMessageStart),
				"task_id":    "task-1",
				"message_id": "msg-1",
			},
			assert: func(t *testing.T, spy *streamEventBridgeSpy) {
				t.Helper()
				if len(spy.messageStarts) != 1 {
					t.Fatalf("expected 1 message start, got %d", len(spy.messageStarts))
				}
				got := spy.messageStarts[0]
				if got.taskID != "task-1" || got.messageID != "msg-1" {
					t.Fatalf("unexpected message start dispatch: %+v", got)
				}
			},
		},
		{
			name: "token",
			payload: map[string]any{
				"type":       string(model.TypeAgentToken),
				"task_id":    "task-1",
				"message_id": "msg-1",
				"delta":      "hello",
			},
			assert: func(t *testing.T, spy *streamEventBridgeSpy) {
				t.Helper()
				if len(spy.tokens) != 1 {
					t.Fatalf("expected 1 token, got %d", len(spy.tokens))
				}
				got := spy.tokens[0]
				if got.taskID != "task-1" || got.messageID != "msg-1" || got.delta != "hello" {
					t.Fatalf("unexpected token dispatch: %+v", got)
				}
			},
		},
		{
			name: "message end",
			payload: map[string]any{
				"type":          string(model.TypeAgentMessageEnd),
				"task_id":       "task-1",
				"message_id":    "msg-1",
				"finish_reason": "stop",
			},
			assert: func(t *testing.T, spy *streamEventBridgeSpy) {
				t.Helper()
				if len(spy.messageEnds) != 1 {
					t.Fatalf("expected 1 message end, got %d", len(spy.messageEnds))
				}
				got := spy.messageEnds[0]
				if got.taskID != "task-1" || got.messageID != "msg-1" || got.reason != "stop" {
					t.Fatalf("unexpected message end dispatch: %+v", got)
				}
			},
		},
		{
			name: "state",
			payload: map[string]any{
				"type":    string(model.TypeAgentState),
				"task_id": "task-1",
				"state":   "thinking",
				"message": "planning",
			},
			assert: func(t *testing.T, spy *streamEventBridgeSpy) {
				t.Helper()
				if len(spy.states) != 1 {
					t.Fatalf("expected 1 state, got %d", len(spy.states))
				}
				got := spy.states[0]
				if got.taskID != "task-1" || got.state != "thinking" || got.message != "planning" {
					t.Fatalf("unexpected state dispatch: %+v", got)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spy := &streamEventBridgeSpy{}
			originalBridge := agentStreamBridge
			agentStreamBridge = spy
			t.Cleanup(func() {
				agentStreamBridge = originalBridge
			})

			reqBody, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("marshal payload failed: %v", err)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/agent/stream-event", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Internal-Token", "test-internal-token")
			rec := httptest.NewRecorder()

			g.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
			}

			tc.assert(t, spy)
		})
	}
}

func TestHandleStreamEventRejectsMalformedPayloads(t *testing.T) {
	g := setupInternalRouterForStreamEvent(t)

	tests := []struct {
		name    string
		payload map[string]any
		headers map[string]string
	}{
		{
			name: "unknown event type",
			payload: map[string]any{
				"type":    "agent_update",
				"task_id": "task-1",
			},
		},
		{
			name: "missing task_id",
			payload: map[string]any{
				"type":       string(model.TypeAgentMessageStart),
				"message_id": "msg-1",
			},
		},
		{
			name: "token missing message_id",
			payload: map[string]any{
				"type":    string(model.TypeAgentToken),
				"task_id": "task-1",
				"delta":   "x",
			},
		},
		{
			name: "token delta not string",
			payload: map[string]any{
				"type":       string(model.TypeAgentToken),
				"task_id":    "task-1",
				"message_id": "msg-1",
				"delta":      123,
			},
		},
		{
			name: "state missing state field",
			payload: map[string]any{
				"type":    string(model.TypeAgentState),
				"task_id": "task-1",
			},
		},
		{
			name: "state invalid enum value",
			payload: map[string]any{
				"type":    string(model.TypeAgentState),
				"task_id": "task-1",
				"state":   "idle",
			},
		},
		{
			name: "mismatched task id header",
			payload: map[string]any{
				"type":    string(model.TypeAgentState),
				"task_id": "task-body",
				"state":   "thinking",
			},
			headers: map[string]string{"X-Task-ID": "task-header"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spy := &streamEventBridgeSpy{}
			originalBridge := agentStreamBridge
			agentStreamBridge = spy
			t.Cleanup(func() {
				agentStreamBridge = originalBridge
			})

			reqBody, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("marshal payload failed: %v", err)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/agent/stream-event", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Internal-Token", "test-internal-token")
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()

			g.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
			}

			if len(spy.messageStarts)+len(spy.tokens)+len(spy.messageEnds)+len(spy.states) != 0 {
				t.Fatal("expected no dispatch for malformed payload")
			}
		})
	}
}

func TestHandleStreamEventAcceptsBodyTaskIDWithoutHeader(t *testing.T) {
	g := setupInternalRouterForStreamEvent(t)
	spy := &streamEventBridgeSpy{}
	originalBridge := agentStreamBridge
	agentStreamBridge = spy
	t.Cleanup(func() {
		agentStreamBridge = originalBridge
	})

	body := map[string]any{
		"type":    string(model.TypeAgentState),
		"task_id": "task-body",
		"state":   "thinking",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/agent/stream-event", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", "test-internal-token")
	rec := httptest.NewRecorder()

	g.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if len(spy.states) != 1 {
		t.Fatalf("expected one state dispatch, got %d", len(spy.states))
	}
}

func setupInternalRouterForStreamEvent(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	loadInternalConfigForStreamEventTest(t)

	r := gin.New()
	agentGroup := r.Group("/api/agent")
	agentGroup.Use(middleware.InternalToken())
	agentGroup.POST("/stream-event", HandleStreamEvent)
	return r
}

func loadInternalConfigForStreamEventTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{"agentInternalToken":"test-internal-token"}`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	if err := config.Load(path); err != nil {
		t.Fatalf("load config failed: %v", err)
	}
}
