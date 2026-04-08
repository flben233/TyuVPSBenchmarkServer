package middleware

import (
	"net/http"

	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"

	"github.com/gin-gonic/gin"
)

const internalTokenHeader = "X-Internal-Token"

func InternalToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(internalTokenHeader)
		if token == "" || token != config.Get().AgentInternalToken {
			c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "invalid internal token"))
			c.Abort()
			return
		}

		c.Next()
	}
}
