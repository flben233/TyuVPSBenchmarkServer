package tool

import (
	"VPSBenchmarkBackend/internal/tool/handler"

	"github.com/gin-gonic/gin"
)

// RegisterRouter wires tool endpoints.
func RegisterRouter(base string, r *gin.Engine) {
	group := r.Group(base + "/tool")
	group.GET("/ip", handler.IPLookup)
	group.POST("/ip", handler.IPLookup)

	group.GET("/traceroute", handler.Traceroute)
	group.POST("/traceroute", handler.Traceroute)

	group.GET("/whois", handler.Whois)
	group.POST("/whois", handler.Whois)
}
