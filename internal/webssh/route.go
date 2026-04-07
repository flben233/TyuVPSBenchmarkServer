package webssh

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/handler"

	"github.com/gin-gonic/gin"
)

func init() {
	common.RegisterRoutes(RegisterRoute)
}

func RegisterRoute(base string, r *gin.Engine) {
	group := r.Group(base + "/webssh")
	group.GET("/ws", handler.HandleWebSocket)
}
