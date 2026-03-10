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
		authGroup.POST("/refresh", handler.RefreshToken)

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

			adminRoute.GET("/admin/user/:id", handler.GetUser)
			adminRoute.GET("/admin/users", handler.ListUsers)
			adminRoute.POST("/admin/user/update", handler.UpdateUser)
			adminRoute.POST("/admin/user/delete", handler.DeleteUser)

			adminRoute.POST("/admin/group/create", handler.CreateUserGroup)
			adminRoute.GET("/admin/groups", handler.ListUserGroups)
			adminRoute.POST("/admin/group/update", handler.UpdateUserGroup)
			adminRoute.POST("/admin/group/delete", handler.DeleteUserGroup)
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
