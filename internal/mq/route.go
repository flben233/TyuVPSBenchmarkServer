package mq

import (
	"VPSBenchmarkBackend/internal/common"
	"github.com/gin-gonic/gin"
)

func init() {
	// Register the routes
	common.RegisterRoutes(RegisterRoute)
}

func RegisterRoute(base string, r *gin.Engine) {
	base += "/task"
	reportGroup := r.Group(base)
	{
		reportGroup.GET("/status/:id", QueryTaskStatus)
	}
}
