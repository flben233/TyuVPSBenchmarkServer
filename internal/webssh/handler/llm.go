package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

type NewConversationRequest struct {
	SSHSessionID string           `json:"sshSessionId" binding:"required"`
	LLMAPI       *NewLLMAPIConfig `json:"llmApi,omitempty"`
}

type NewLLMAPIConfig struct {
	APIBase string `json:"apiBase,omitempty"`
	APIKey  string `json:"apiKey,omitempty"`
	Model   string `json:"model,omitempty"`
}

type NewConversationResponse struct {
	ConversationID string `json:"conversationId"`
}

type ChatRequest struct {
	ConversationID         string        `json:"conversationId,omitempty"`
	Message                string        `json:"message,omitempty"`
	Messages               []ChatMessage `json:"messages,omitempty"`
	ApprovalGranted        *bool         `json:"approval_granted,omitempty"`
	AllowedCommands        []string      `json:"allowed_commands,omitempty"`
	SessionAllowedCommands []string      `json:"session_allowed_commands,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StopRequest struct {
	ConversationID string `json:"conversationId" binding:"required"`
}

type StopResponse struct {
	Stopped bool `json:"stopped"`
}

func proxyRequest(c *gin.Context, path string) {
	parsedURL, err := url.Parse(config.Get().LLMBackendURL + path)
	if err != nil {
		common.DefaultErrorHandler(c, err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.Director = func(req *http.Request) {
		req.URL = parsedURL
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func proxyRequestWithBody(c *gin.Context, path string, body []byte) {
	parsedURL, err := url.Parse(config.Get().LLMBackendURL + path)
	if err != nil {
		common.DefaultErrorHandler(c, err)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.Director = func(req *http.Request) {
		req.URL = parsedURL
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.ContentLength = int64(len(body))
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

// NewConversation creates a new LLM conversation bound to an SSH session.
// @Summary Create WebSSH LLM conversation
// @Description Proxy request to Python backend to create a conversation bound to an SSH session.
// Injects the authenticated user ID for rate limiting.
// @Tags webssh
// @Accept json
// @Produce json
// @Param request body NewConversationRequest true "Create conversation request"
// @Success 200 {object} NewConversationResponse
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /webssh/llm/new [post]
func NewConversation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "Failed to read request body"))
		return
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "Invalid JSON"))
		return
	}
	bodyMap["userId"] = fmt.Sprintf("%d", userID.(int64))

	modifiedBody, err := json.Marshal(bodyMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "Failed to marshal request"))
		return
	}

	proxyRequestWithBody(c, "/new", modifiedBody)
}

// Chat sends one user turn and proxies SSE stream from Python backend.
// @Summary Stream WebSSH LLM chat
// @Description Proxy chat request to Python backend and stream SSE events back to client.
// @Tags webssh
// @Accept json
// @Produce text/event-stream
// @Param request body ChatRequest true "Chat request"
// @Success 200 {string} string "SSE stream"
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /webssh/llm/chat [post]
func Chat(c *gin.Context) {
	proxyRequest(c, "/chat")
}

// Stop signals the LLM agent to stop generating the current response.
// @Summary Stop WebSSH LLM response
// @Description Proxy stop request to Python backend to abort an in-progress LLM response.
// @Tags webssh
// @Accept json
// @Produce json
// @Param request body StopRequest true "Stop request"
// @Success 200 {object} StopResponse
// @Failure 400 {object} common.APIResponse[any]
// @Failure 404 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /webssh/llm/stop [post]
func Stop(c *gin.Context) {
	proxyRequest(c, "/stop")
}
