package webssh

import (
	"VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/handler"

	"github.com/gin-gonic/gin"
)

func init() {
	common.RegisterRoutes(RegisterRoute)
}

func RegisterRoute(base string, r *gin.Engine) {
	group1 := r.Group(base + "/webssh")
	group1.GET("/ws", handler.HandleWebSocket)
	group2 := r.Group(base + "/webssh/sync")
	group2.Use(auth.GetJWTMiddleware())
	{
		group2.POST("/upload", handler.HandleUpload)
		group2.GET("/download", handler.HandleDownload)
		group2.POST("/reset", handler.HandleReset)
	}
}
