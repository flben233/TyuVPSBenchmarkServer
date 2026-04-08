package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/service"
	"VPSBenchmarkBackend/internal/webssh/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: websocketOriginAllowed,
}

func websocketOriginAllowed(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" {
		return false
	}

	frontendURL := strings.TrimSpace(config.Get().FrontendURL)
	if frontendURL != "" {
		allowURL, err := url.Parse(frontendURL)
		if err != nil || allowURL.Host == "" {
			return false
		}
		return strings.EqualFold(originURL.Host, allowURL.Host)
	}

	return strings.EqualFold(originURL.Host, r.Host)
}

func HandleWebSocket(ctx *gin.Context) {
	tokenString := ctx.Query("token")
	if tokenString == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token is required"})
		return
	}

	cfg := config.Get()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.JwtSecret), nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	var userID int64 = -1
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if id, ok := claims["github_id"].(float64); ok {
			userID = int64(id)
		}
	}
	if userID == -1 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		return
	}

	sessionID := ctx.Query("session_id")
	if sessionID == "" {
		sessionID = strconv.FormatInt(userID, 10)
	}

	if err := service.AcquireSession(userID); err != nil {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}
	defer service.ReleaseSession(userID)

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("WebSocket accept error: %v", err)
		return
	}
	defer conn.Close()

	wsHandler(ctx, conn, userID, sessionID)
}

func wsHandler(ctx *gin.Context, conn *websocket.Conn, userID int64, sessionID string) {
	var writeMu sync.Mutex

	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocket handler panic (user %d): %v", userID, r)
			writeTimeout(conn, &writeMu, model.ServerMessage{
				Type:    model.TypeError,
				Message: "internal error",
			})
		}
	}()

	var sshSession *service.SSHSession
	requestContext := context.Background()
	if ctx != nil && ctx.Request != nil {
		requestContext = ctx.Request.Context()
	}

	defer func() {
		if sshSession != nil {
			service.UnregisterSession(sessionID)
			sshSession.Close()
		}
	}()

	sendMsg := func(msg *model.ServerMessage) {
		writeTimeout(conn, &writeMu, *msg)
	}

	sendOutput := func(data []byte) {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.WriteMessage(websocket.BinaryMessage, data)
	}

	for {
		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocket read error:", err)
			break
		}

		if msgType == websocket.BinaryMessage {
			if sshSession != nil {
				if err := sshSession.WriteInput(payload); err != nil {
					sendMsg(&model.ServerMessage{
						Type:    model.TypeError,
						Message: err.Error(),
					})
				}
			}
			continue
		}

		var clientMsg model.ClientMessage
		if err := json.Unmarshal(payload, &clientMsg); err != nil {
			break
		}

		switch clientMsg.Type {
		case model.TypeConnect:
			if sshSession != nil {
				service.UnregisterSession(sessionID)
				sshSession.Close()
				sshSession = nil
			}

			sshSession = service.NewSSHSession()
			if err := sshSession.Connect(&clientMsg); err != nil {
				sendMsg(&model.ServerMessage{
					Type:    model.TypeError,
					Message: err.Error(),
				})
				sshSession = nil
				continue
			}
			service.RegisterSession(sessionID, sshSession)
			sendMsg(&model.ServerMessage{
				Type: model.TypeConnected,
			})
			go sshSession.ReadOutput(sendOutput, sendMsg)

		case model.TypeResize:
			if sshSession != nil {
				if err := sshSession.Resize(clientMsg.Rows, clientMsg.Cols); err != nil {
					log.Printf("Resize error: %v", err)
				}
			}

		case model.TypePing:
			// keepalive, no-op

		case model.TypeAgentTask:
			prompt := strings.TrimSpace(clientMsg.Message)
			agentClient := newAgentClient()
			createResult, err := agentClient.CreateTask(requestContext, prompt, map[string]any{
				"session_id": sessionID,
			})
			if err != nil {
				sendAgentErrorMessage(sendMsg, "", err, "failed to create task")
				continue
			}
			taskID := strings.TrimSpace(createResult.TaskID)
			if taskID == "" {
				sendAgentErrorMessage(sendMsg, "", service.ErrAgentBadResponse, "agent service returned empty task id")
				continue
			}

			status := strings.TrimSpace(createResult.Status)
			if status == "" {
				status = model.AgentStatusRunning
			}
			binding := store.TaskBinding{
				UserID:       userID,
				SessionID:    sessionID,
				Status:       status,
				CreatedAt:    time.Now().UTC(),
				CommandCount: 0,
			}
			if err := saveTaskBinding(requestContext, taskID, binding); err != nil {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "failed to create task"})
				continue
			}

			sendMsg(&model.ServerMessage{Type: model.TypeAgentUpdate, TaskID: taskID, Status: status, Message: "task created"})

		case model.TypeAgentMsg:
			taskID := strings.TrimSpace(clientMsg.TaskID)
			if taskID == "" {
				sendMsg(&model.ServerMessage{
					Type:    model.TypeAgentError,
					TaskID:  unknownTaskID,
					Status:  model.AgentStatusFailed,
					Message: "task_id is required",
				})
				continue
			}
			binding, err := getTaskBinding(requestContext, taskID)
			if err != nil {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "failed to query task"})
				continue
			}
			if binding == nil || binding.UserID != userID {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "task not found"})
				continue
			}

			message := strings.TrimSpace(clientMsg.Message)
			if message == "" {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "message is required"})
				continue
			}

			agentClient := newAgentClient()
			reply, err := agentClient.SendTaskMessage(requestContext, taskID, message)
			if err != nil {
				sendAgentErrorMessage(sendMsg, taskID, err, "failed to send task message")
				continue
			}
			if !reply.OK {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "agent service rejected task message"})
				continue
			}

			nextStatus := deriveTaskStatus(binding.Status, reply)
			if nextStatus != binding.Status {
				binding.Status = nextStatus
				if err := saveTaskBinding(requestContext, taskID, *binding); err != nil {
					sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "failed to update task status"})
					continue
				}
			}

			data := reply.Data
			if data == nil {
				data = map[string]any{}
			}
			sendMsg(&model.ServerMessage{Type: model.TypeAgentUpdate, TaskID: taskID, Status: binding.Status, Message: coalesceAgentMessage(reply.Message, "message accepted")})
			if awaiting, _ := data["awaiting_approval"].(bool); awaiting {
				question := extractAgentText(data["final_response"])
				if question == "" {
					question = coalesceAgentMessage(reply.Message, "approval required")
				}
				sendMsg(&model.ServerMessage{Type: model.TypeAgentApproval, TaskID: taskID, Status: model.AgentStatusAwaitingApproval, Message: "approval required", Question: question})
			} else if done, _ := data["task_complete"].(bool); done {
				summary := extractAgentText(data["final_response"])
				if summary == "" {
					summary = coalesceAgentMessage(reply.Message, "task completed")
				}
				sendMsg(&model.ServerMessage{Type: model.TypeAgentDone, TaskID: taskID, Status: model.AgentStatusCompleted, Message: summary, Summary: summary})
			}

		case model.TypeAgentAck:
			taskID := strings.TrimSpace(clientMsg.TaskID)
			if taskID == "" {
				sendMsg(&model.ServerMessage{
					Type:    model.TypeAgentError,
					TaskID:  unknownTaskID,
					Status:  model.AgentStatusFailed,
					Message: "task_id is required",
				})
				continue
			}
			binding, err := getTaskBinding(requestContext, taskID)
			if err != nil {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "failed to query task"})
				continue
			}
			if binding == nil || binding.UserID != userID {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "task not found"})
				continue
			}

			approved := false
			if clientMsg.Approved != nil {
				approved = *clientMsg.Approved
			}
			agentClient := newAgentClient()
			reply, err := agentClient.ApproveTask(requestContext, taskID, approved)
			if err != nil {
				sendAgentErrorMessage(sendMsg, taskID, err, "failed to approve task")
				continue
			}
			if !reply.OK {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "agent service rejected task approval"})
				continue
			}

			if approved {
				binding.Status = deriveTaskStatus(binding.Status, reply)
				if binding.Status == "" {
					binding.Status = model.AgentStatusRunning
				}
			} else {
				binding.Status = model.AgentStatusFailed
			}
			if err := saveTaskBinding(requestContext, taskID, *binding); err != nil {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: "failed to update task approval"})
				continue
			}

			if !approved {
				sendMsg(&model.ServerMessage{Type: model.TypeAgentError, TaskID: taskID, Status: model.AgentStatusFailed, Message: coalesceAgentMessage(reply.Message, "approval denied")})
				continue
			}

			data := reply.Data
			if data == nil {
				data = map[string]any{}
			}
			sendMsg(&model.ServerMessage{Type: model.TypeAgentUpdate, TaskID: taskID, Status: binding.Status, Message: coalesceAgentMessage(reply.Message, "approval accepted")})
			if awaiting, _ := data["awaiting_approval"].(bool); awaiting {
				question := extractAgentText(data["final_response"])
				if question == "" {
					question = coalesceAgentMessage(reply.Message, "approval required")
				}
				sendMsg(&model.ServerMessage{Type: model.TypeAgentApproval, TaskID: taskID, Status: model.AgentStatusAwaitingApproval, Message: "approval required", Question: question})
			} else if done, _ := data["task_complete"].(bool); done {
				summary := extractAgentText(data["final_response"])
				if summary == "" {
					summary = coalesceAgentMessage(reply.Message, "task completed")
				}
				sendMsg(&model.ServerMessage{Type: model.TypeAgentDone, TaskID: taskID, Status: model.AgentStatusCompleted, Message: summary, Summary: summary})
			}
		}
	}
}

func sendAgentErrorMessage(sendMsg func(*model.ServerMessage), taskID string, err error, defaultMessage string) {
	message := defaultMessage
	if errors.Is(err, service.ErrAgentURLNotConfigured) {
		message = "agent url is not configured"
	} else if errors.Is(err, service.ErrAgentTaskNotFound) {
		message = "task not found in agent service"
	} else if errors.Is(err, service.ErrAgentRequestInvalid) {
		message = err.Error()
	} else if errors.Is(err, service.ErrAgentUnavailable) {
		message = "agent service unavailable"
	} else if errors.Is(err, service.ErrAgentBadResponse) {
		message = "invalid response from agent service"
	}

	sendMsg(&model.ServerMessage{
		Type:    model.TypeAgentError,
		TaskID:  normalizeTaskID(taskID),
		Status:  model.AgentStatusFailed,
		Message: message,
	})
}

const unknownTaskID = "unknown"

func normalizeTaskID(taskID string) string {
	trimmed := strings.TrimSpace(taskID)
	if trimmed == "" {
		return unknownTaskID
	}
	return trimmed
}

func coalesceAgentMessage(msg string, fallback string) string {
	trimmed := strings.TrimSpace(msg)
	if trimmed != "" {
		return trimmed
	}
	return fallback
}

func extractAgentText(value any) string {
	asText, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(asText)
}

func writeTimeout(conn *websocket.Conn, writeMu *sync.Mutex, msg model.ServerMessage) {
	writeMu.Lock()
	defer writeMu.Unlock()
	_ = conn.WriteJSON(msg)
}
