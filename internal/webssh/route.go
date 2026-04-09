package webssh

import (
	"VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/handler"
	"VPSBenchmarkBackend/internal/webssh/middleware"

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

	agentGroup := r.Group(base + "/agent")
	agentGroup.Use(middleware.InternalToken())
	{
		agentGroup.POST("/safety-check", handler.HandleSafetyCheck)
		agentGroup.POST("/execute", handler.HandleExecute)
		agentGroup.POST("/stream-event", handler.HandleStreamEvent)
		agentGroup.GET("/tools", handler.HandleTools)
	}

	agentJWTGroup := r.Group(base + "/agent")
	agentJWTGroup.Use(auth.GetJWTMiddleware())
	{
		agentJWTGroup.POST("/tasks", handler.HandleCreateTask)
		agentJWTGroup.POST("/tasks/:task_id/message", handler.HandleTaskMessage)
		agentJWTGroup.POST("/tasks/:task_id/approve", handler.HandleTaskApprove)
		agentJWTGroup.GET("/tasks/:task_id", handler.HandleGetTask)
		agentJWTGroup.GET("/tasks", handler.HandleListTasks)
	}
}
