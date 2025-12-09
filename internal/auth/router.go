package auth

import (
	"VPSBenchmarkBackend/internal/auth/handler"
	"VPSBenchmarkBackend/internal/auth/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRouter registers auth routes
func RegisterRouter(base string, r *gin.Engine) {
	authGroup := r.Group(base + "/auth")
	{
		// Public routes (no authentication required)
		authGroup.POST("/github/login", handler.GithubLogin)

		// Protected routes (authentication required)
		protected := authGroup.Group("")
		protected.Use(middleware.JWTAuth())
		{
			protected.GET("/user", handler.GetUserInfo)
		}
	}
}

// GetJWTMiddleware returns the JWT authentication middleware
func GetJWTMiddleware() gin.HandlerFunc {
	return middleware.JWTAuth()
}

// GetOptionalJWTMiddleware returns the optional JWT authentication middleware
func GetOptionalJWTMiddleware() gin.HandlerFunc {
	return middleware.OptionalJWTAuth()
}

// GetAdminMiddleware returns the admin authentication middleware
func GetAdminMiddleware() gin.HandlerFunc {
	return middleware.CheckAdmin()
}
