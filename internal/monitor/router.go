package monitor

import (
	"VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/monitor/handler"
	"VPSBenchmarkBackend/internal/monitor/store"

	"github.com/gin-gonic/gin"
)

func InitMonitorStore(dbPath string) error {
	return store.InitMonitorStore(dbPath)
}

func RegisterRouter(base string, r *gin.Engine) {
	base += "/monitor"
	{
		// Public routes - get statistics
		publicAPI := r.Group(base)
		publicAPI.GET("/statistics", handler.GetStatistics)
	}
	{
		// Protected routes - manage hosts
		protectedAPI := r.Group(base)
		protectedAPI.Use(auth.GetJWTMiddleware())
		{
			protectedAPI.GET("/hosts", handler.ListHosts)
			protectedAPI.POST("/hosts", handler.AddHost)
			protectedAPI.DELETE("/hosts/:id", handler.RemoveHost)
		}
	}
}
