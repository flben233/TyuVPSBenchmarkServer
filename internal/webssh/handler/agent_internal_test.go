package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh"
	"VPSBenchmarkBackend/internal/webssh/handler"
	"VPSBenchmarkBackend/internal/webssh/service"

	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
)

func TestExecuteRejectsMissingInternalToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loadTestConfig(t)

	r := gin.New()
	webssh.RegisterRoute("/api", r)

	body := []byte(`{"command":"ls -la"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestHandleToolsReturnsRelativePaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	loadTestConfig(t)

	r := gin.New()
	webssh.RegisterRoute("/api", r)

	req := httptest.NewRequest(http.MethodGet, "/api/agent/tools", nil)
	req.Header.Set("X-Internal-Token", "test-internal-token")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Tools []struct {
				Path string `json:"path"`
			} `json:"tools"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if len(payload.Data.Tools) == 0 {
		t.Fatal("expected tools to be returned")
	}

	for _, tool := range payload.Data.Tools {
		if tool.Path == "" {
			t.Fatal("expected non-empty tool path")
		}
		if len(tool.Path) >= 4 && tool.Path[:4] == "/api" {
			t.Fatalf("expected relative path, got %q", tool.Path)
		}
	}
}

func TestHandleExecuteTaskNotFoundMapping(t *testing.T) {
	r := setupRouterForExecuteTest(t, func(_ context.Context, _ string, _ string, _ bool) (service.ExecuteCommandResult, error) {
		return service.ExecuteCommandResult{}, service.ErrTaskNotFound
	})

	rec := performExecuteRequest(r, "test-internal-token", "task-1", `{"command":"ls -la"}`)
	assertErrorResponse(t, rec, http.StatusNotFound, common.BadRequestCode)
}

func TestHandleExecuteSessionNotFoundMapping(t *testing.T) {
	r := setupRouterForExecuteTest(t, func(_ context.Context, _ string, _ string, _ bool) (service.ExecuteCommandResult, error) {
		return service.ExecuteCommandResult{}, service.ErrSessionNotFound
	})

	rec := performExecuteRequest(r, "test-internal-token", "task-1", `{"command":"ls -la"}`)
	assertErrorResponse(t, rec, http.StatusNotFound, common.BadRequestCode)
}

func TestHandleExecuteCommandLimitExceededMapping(t *testing.T) {
	r := setupRouterForExecuteTest(t, func(_ context.Context, _ string, _ string, _ bool) (service.ExecuteCommandResult, error) {
		return service.ExecuteCommandResult{}, service.ErrCommandLimitExceeded
	})

	rec := performExecuteRequest(r, "test-internal-token", "task-1", `{"command":"ls -la"}`)
	assertErrorResponse(t, rec, http.StatusTooManyRequests, common.LimitExceededCode)
}

func setupRouterForExecuteTest(t *testing.T, fn func(context.Context, string, string, bool) (service.ExecuteCommandResult, error)) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	loadTestConfig(t)

	handler.ExecuteTaskCommandForTest(fn)
	t.Cleanup(func() {
		handler.ResetExecuteTaskCommandForTest()
	})

	r := gin.New()
	webssh.RegisterRoute("/api", r)
	return r
}

func performExecuteRequest(r *gin.Engine, token string, taskID string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/agent/execute", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", token)
	req.Header.Set("X-Task-ID", taskID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, wantHTTP int, wantCode int) {
	t.Helper()
	if rec.Code != wantHTTP {
		t.Fatalf("expected status %d, got %d", wantHTTP, rec.Code)
	}

	var payload struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if payload.Code != wantCode {
		t.Fatalf("expected code %d, got %d", wantCode, payload.Code)
	}
}

func loadTestConfig(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	configContent := []byte(`{"agentInternalToken":"test-internal-token"}`)
	if err := os.WriteFile(configPath, configContent, 0o600); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	if err := config.Load(configPath); err != nil {
		t.Fatalf("load config failed: %v", err)
	}
}
