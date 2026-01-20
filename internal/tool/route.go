package tool

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/tool/handler"

	"github.com/gin-gonic/gin"
)

func init() {
	// Register the routes
	common.RegisterRoutes(RegisterRoute)
}

// RegisterRouter wires tool endpoints.
func RegisterRoute(base string, r *gin.Engine) {
	group := r.Group(base + "/tool")
	group.GET("/traceroute", handler.Traceroute)
	group.GET("/whois", handler.Whois)
}
