package auth

import (
	"VPSBenchmarkBackend/internal/auth/handler"
	"VPSBenchmarkBackend/internal/auth/middleware"
	"VPSBenchmarkBackend/internal/common"

	"github.com/gin-gonic/gin"
)

func init() {
	// Register the routes
	common.RegisterRoutes(RegisterRoute)
}

// RegisterRoute registers auth routes
func RegisterRoute(base string, r *gin.Engine) {
	authGroup := r.Group(base + "/auth")
	{
		// Public routes (no authentication required)
		authGroup.GET("/github/login", handler.GithubLogin)

		// Protected routes (authentication required)
		protected := authGroup.Group("")
		protected.Use(GetJWTMiddleware())
		{
			protected.GET("/user", handler.GetUserInfo)
		}

		adminRoute := authGroup.Group("")
		adminRoute.Use(GetJWTMiddleware(), GetAdminMiddleware())
		{
			adminRoute.GET("/admin", handler.CheckAdminUser)
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

func GetAllowCORSMiddleware() gin.HandlerFunc {
	return middleware.AllowCORS()
}
