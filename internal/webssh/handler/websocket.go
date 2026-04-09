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
	"sync/atomic"
	"time"

	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/service"
	"VPSBenchmarkBackend/internal/webssh/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var wsStreamOwnerSeq uint64

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
	streamOwnerID := fmt.Sprintf("ws-%d", atomic.AddUint64(&wsStreamOwnerSeq, 1))
	registeredTaskIDs := make(map[string]struct{})

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
		for taskID := range registeredTaskIDs {
			agentStreamBridge.Unregister(taskID, streamOwnerID)
		}
		if sshSession != nil {
			service.UnregisterSession(sessionID)
			sshSession.Close()
		}
	}()

	writeJSONNoLock := func(msg model.ServerMessage) {
		_ = conn.WriteJSON(msg)
	}

	sendMsg := func(msg *model.ServerMessage) {
		writeServerMessageLocked(&writeMu, writeJSONNoLock, *msg)
	}

	sendOutput := func(data []byte) {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.WriteMessage(websocket.BinaryMessage, data)
	}

	registerTaskStream := func(taskID string) {
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			return
		}
		if _, ok := registeredTaskIDs[taskID]; ok {
			return
		}
		agentStreamBridge.Register(taskID, streamOwnerID, func(msg model.ServerMessage) { sendMsg(&msg) })
		registeredTaskIDs[taskID] = struct{}{}
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
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "failed to create task")
				continue
			}
			registerTaskStream(taskID)
			sendAgentState(sendMsg, taskID, mapStatusToAgentState(status), "task created")

		case model.TypeAgentMsg:
			taskID := strings.TrimSpace(clientMsg.TaskID)
			if taskID == "" {
				sendAgentState(sendMsg, unknownTaskID, mapStatusToAgentState(model.AgentStatusFailed), "task_id is required")
				continue
			}
			binding, err := getTaskBinding(requestContext, taskID)
			if err != nil {
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "failed to query task")
				continue
			}
			if binding == nil || binding.UserID != userID {
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "task not found")
				continue
			}
			registerTaskStream(taskID)

			message := strings.TrimSpace(clientMsg.Message)
			if message == "" {
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "message is required")
				continue
			}

			agentClient := newAgentClient()
			reply, err := agentClient.SendTaskMessage(requestContext, taskID, message)
			if err != nil {
				sendAgentErrorMessage(sendMsg, taskID, err, "failed to send task message")
				continue
			}
			if !reply.OK {
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "agent service rejected task message")
				continue
			}

			nextStatus := deriveTaskStatus(binding.Status, reply)
			if nextStatus != binding.Status {
				binding.Status = nextStatus
				if err := saveTaskBinding(requestContext, taskID, *binding); err != nil {
					sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "failed to update task status")
					continue
				}
			}

			handleAgentReplyResult(sendMsg, taskID, binding.Status, reply, "message accepted")

		case model.TypeAgentAck:
			taskID := strings.TrimSpace(clientMsg.TaskID)
			if taskID == "" {
				sendAgentState(sendMsg, unknownTaskID, mapStatusToAgentState(model.AgentStatusFailed), "task_id is required")
				continue
			}
			binding, err := getTaskBinding(requestContext, taskID)
			if err != nil {
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "failed to query task")
				continue
			}
			if binding == nil || binding.UserID != userID {
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "task not found")
				continue
			}
			registerTaskStream(taskID)

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
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "agent service rejected task approval")
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
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), "failed to update task approval")
				continue
			}

			if !approved {
				sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusFailed), coalesceAgentMessage(reply.Message, "approval denied"))
				continue
			}

			handleAgentReplyResult(sendMsg, taskID, binding.Status, reply, "approval accepted")
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

	sendAgentState(sendMsg, normalizeTaskID(taskID), mapStatusToAgentState(model.AgentStatusFailed), message)
}

type agentStreamEmitter interface {
	Register(taskID string, ownerID string, sender service.AgentStreamSender)
	Unregister(taskID string, ownerID string)
	EmitMessageStart(taskID string, messageID string)
	EmitToken(taskID string, messageID string, delta string)
	EmitMessageEnd(taskID string, messageID string, reason string)
	EmitState(taskID string, state string, message string)
}

var agentStreamBridge agentStreamEmitter = service.NewAgentStreamBridge()

func writeServerMessageLocked(writeMu *sync.Mutex, writeNoLock func(model.ServerMessage), msg model.ServerMessage) {
	writeMu.Lock()
	defer writeMu.Unlock()
	writeNoLock(msg)
}

func writeAgentMessageTripletLocked(writeMu *sync.Mutex, writeNoLock func(model.ServerMessage), taskID string, messageID string, delta string) {
	writeMu.Lock()
	defer writeMu.Unlock()
	writeNoLock(model.ServerMessage{Type: model.TypeAgentMessageStart, TaskID: taskID, MessageID: messageID})
	writeNoLock(model.ServerMessage{Type: model.TypeAgentToken, TaskID: taskID, MessageID: messageID, Delta: delta})
	writeNoLock(model.ServerMessage{Type: model.TypeAgentMessageEnd, TaskID: taskID, MessageID: messageID, FinishReason: "stop"})
}

func sendAgentState(sendMsg func(*model.ServerMessage), taskID string, state string, message string) {
	sendMsg(&model.ServerMessage{
		Type:    model.TypeAgentState,
		TaskID:  normalizeTaskID(taskID),
		State:   state,
		Message: message,
	})
}

func handleAgentReplyResult(sendMsg func(*model.ServerMessage), taskID string, status string, reply service.AgentReply, acceptFallback string) {
	data := reply.Data
	if data == nil {
		data = map[string]any{}
	}
	if awaiting, _ := data["awaiting_approval"].(bool); awaiting {
		question := extractAgentText(data["final_response"])
		if question == "" {
			question = coalesceAgentMessage(reply.Message, "approval required")
		}
		sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusAwaitingApproval), question)
		return
	}
	if done, _ := data["task_complete"].(bool); done {
		summary := extractAgentText(data["final_response"])
		if summary == "" {
			summary = coalesceAgentMessage(reply.Message, "task completed")
		}
		sendAgentState(sendMsg, taskID, mapStatusToAgentState(model.AgentStatusCompleted), summary)
		return
	}
	sendAgentState(sendMsg, taskID, mapStatusToAgentState(status), coalesceAgentMessage(reply.Message, acceptFallback))
}

func mapStatusToAgentState(status string) string {
	switch strings.TrimSpace(status) {
	case model.AgentStatusAwaitingApproval:
		return "awaiting_approval"
	case model.AgentStatusCompleted:
		return "done"
	case model.AgentStatusFailed:
		return "error"
	case model.AgentStatusRunning:
		return "running_command"
	default:
		return "thinking"
	}
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
	writeServerMessageLocked(writeMu, func(m model.ServerMessage) { _ = conn.WriteJSON(m) }, msg)
}
