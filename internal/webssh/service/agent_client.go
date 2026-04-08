package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"VPSBenchmarkBackend/internal/config"
)

const defaultAgentHTTPTimeout = 10 * time.Second

var (
	ErrAgentURLNotConfigured = errors.New("agent url is not configured")
	ErrAgentTaskNotFound     = errors.New("agent task not found")
	ErrAgentRequestInvalid   = errors.New("agent request invalid")
	ErrAgentUnavailable      = errors.New("agent service unavailable")
	ErrAgentBadResponse      = errors.New("agent service bad response")
)

type AgentClient interface {
	CreateTask(ctx context.Context, prompt string, metadata map[string]any) (CreateTaskResult, error)
	SendTaskMessage(ctx context.Context, taskID string, message string) (AgentReply, error)
	ApproveTask(ctx context.Context, taskID string, approved bool) (AgentReply, error)
}

type HTTPAgentClient struct {
	httpClient *http.Client
	baseURL    string
}

type createTaskPayload struct {
	Prompt   string         `json:"prompt"`
	Metadata map[string]any `json:"metadata"`
}

type CreateTaskResult struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type taskMessagePayload struct {
	Message string `json:"message"`
}

type approveTaskPayload struct {
	Approved bool `json:"approved"`
}

type AgentReply struct {
	OK      bool           `json:"ok"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

func NewAgentClient() AgentClient {
	return &HTTPAgentClient{
		httpClient: &http.Client{Timeout: defaultAgentHTTPTimeout},
		baseURL:    strings.TrimRight(strings.TrimSpace(config.Get().AgentURL), "/"),
	}
}

func (c *HTTPAgentClient) CreateTask(ctx context.Context, prompt string, metadata map[string]any) (CreateTaskResult, error) {
	payload := createTaskPayload{Prompt: prompt, Metadata: metadata}

	body, err := c.doJSON(ctx, http.MethodPost, "/api/agent/tasks", payload)
	if err != nil {
		return CreateTaskResult{}, err
	}

	var result CreateTaskResult
	if err := json.Unmarshal(body, &result); err != nil {
		return CreateTaskResult{}, fmt.Errorf("%w: decode create task response: %v", ErrAgentBadResponse, err)
	}

	return result, nil
}

func (c *HTTPAgentClient) SendTaskMessage(ctx context.Context, taskID string, message string) (AgentReply, error) {
	payload := taskMessagePayload{Message: message}
	escapedTaskID := url.PathEscape(strings.TrimSpace(taskID))
	body, err := c.doJSON(ctx, http.MethodPost, "/api/agent/tasks/"+escapedTaskID+"/message", payload)
	if err != nil {
		return AgentReply{}, err
	}

	var result AgentReply
	if err := json.Unmarshal(body, &result); err != nil {
		return AgentReply{}, fmt.Errorf("%w: decode task message response: %v", ErrAgentBadResponse, err)
	}

	return result, nil
}

func (c *HTTPAgentClient) ApproveTask(ctx context.Context, taskID string, approved bool) (AgentReply, error) {
	payload := approveTaskPayload{Approved: approved}
	escapedTaskID := url.PathEscape(strings.TrimSpace(taskID))
	body, err := c.doJSON(ctx, http.MethodPost, "/api/agent/tasks/"+escapedTaskID+"/approve", payload)
	if err != nil {
		return AgentReply{}, err
	}

	var result AgentReply
	if err := json.Unmarshal(body, &result); err != nil {
		return AgentReply{}, fmt.Errorf("%w: decode approve task response: %v", ErrAgentBadResponse, err)
	}

	return result, nil
}

func (c *HTTPAgentClient) doJSON(ctx context.Context, method string, path string, payload any) ([]byte, error) {
	if c.baseURL == "" {
		return nil, ErrAgentURLNotConfigured
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(rawPayload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAgentUnavailable, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}

	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		snippet = http.StatusText(resp.StatusCode)
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s", ErrAgentTaskNotFound, snippet)
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return nil, fmt.Errorf("%w: %s", ErrAgentRequestInvalid, snippet)
	default:
		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("%w: %s", ErrAgentUnavailable, snippet)
		}
		return nil, fmt.Errorf("%w: status=%d body=%s", ErrAgentBadResponse, resp.StatusCode, snippet)
	}
}
