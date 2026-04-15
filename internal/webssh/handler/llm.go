package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type NewConversationRequest struct {
	SSHSessionID string `json:"sshSessionId" binding:"required"`
}

type NewConversationResponse struct {
	ConversationID string `json:"conversationId"`
}

type ChatRequest struct {
	ConversationID  string `json:"conversationId,omitempty"`
	Message         string `json:"message,omitempty"`
	ApprovalGranted *bool  `json:"approval_granted,omitempty"`
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

// NewConversation creates a new LLM conversation bound to an SSH session.
// @Summary Create WebSSH LLM conversation
// @Description Proxy request to Python backend to create a conversation bound to an SSH session.
// @Tags webssh
// @Accept json
// @Produce json
// @Param request body NewConversationRequest true "Create conversation request"
// @Success 200 {object} NewConversationResponse
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /webssh/llm/new [post]
func NewConversation(c *gin.Context) {
	proxyRequest(c, "/new")
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
