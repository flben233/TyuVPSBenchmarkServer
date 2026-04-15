package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/webssh/model"
	"VPSBenchmarkBackend/internal/webssh/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

	wsHandler(ctx, conn, userID)
}

func wsHandler(ctx *gin.Context, conn *websocket.Conn, userID int64) {
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

	defer func() {
		if sshSession != nil {
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
			sshSession = service.NewSSHSession()
			if err := sshSession.Connect(&clientMsg); err != nil {
				sendMsg(&model.ServerMessage{
					Type:    model.TypeError,
					Message: err.Error(),
				})
				sshSession = nil
				continue
			}
			sendMsg(&model.ServerMessage{
				Type:    model.TypeConnected,
				Message: sshSession.ID,
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
		}
	}
}

func writeTimeout(conn *websocket.Conn, writeMu *sync.Mutex, msg model.ServerMessage) {
	writeMu.Lock()
	defer writeMu.Unlock()
	_ = conn.WriteJSON(msg)
}
