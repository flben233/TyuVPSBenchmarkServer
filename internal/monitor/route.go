package monitor

import (
	"VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/monitor/handler"

	"github.com/gin-gonic/gin"
)

func init() {
	// Register the routes
	common.RegisterRoutes(RegisterRoute)
}

func RegisterRoute(base string, r *gin.Engine) {
	base += "/monitor"
	{
		// Public routes - get statistics
		publicAPI := r.Group(base)
		publicAPI.GET("/statistics", handler.GetStatistics)
		publicAPI.GET("/status", handler.GetServerStatus)
	}
	{
		// Protected routes - manage hosts
		protectedAPI := r.Group(base)
		protectedAPI.Use(auth.GetJWTMiddleware())
		{
			protectedAPI.GET("/hosts", handler.ListHosts)
			protectedAPI.POST("/hosts", handler.AddHost)
			protectedAPI.POST("/hosts/:id", handler.RemoveHost)
		}
	}
	{
		// Admin routes - review hosts
		adminAPI := r.Group(base + "/admin")
		adminAPI.Use(auth.GetJWTMiddleware(), auth.GetAdminMiddleware())
		{
			adminAPI.GET("/pending", handler.ListPendingHosts)
			adminAPI.POST("/approve/:id", handler.ApproveHost)
			adminAPI.POST("/reject/:id", handler.RejectHost)
		}
	}
}
