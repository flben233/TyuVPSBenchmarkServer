package lookingglass

import (
	"VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/lookingglass/handler"

	"github.com/gin-gonic/gin"
)

func init() {
	// Register the routes
	common.RegisterRoutes(RegisterRoute)
}

func RegisterRoute(base string, r *gin.Engine) {
	base += "/lookingglass"
	{
		// Public routes - list all records
		publicAPI := r.Group(base)
		publicAPI.GET("/list", handler.ListAllRecords)
	}
	{
		// Protected routes - CRUD operations
		protectedAPI := r.Group(base)
		protectedAPI.Use(auth.GetJWTMiddleware())
		{
			protectedAPI.GET("/records", handler.ListRecords)
			protectedAPI.POST("/records", handler.AddRecord)
			protectedAPI.PUT("/records/:id", handler.UpdateRecord)
			protectedAPI.DELETE("/records/:id", handler.RemoveRecord)
		}
	}
	{
		// Admin routes - review records
		adminAPI := r.Group(base + "/admin")
		adminAPI.Use(auth.GetJWTMiddleware(), auth.GetAdminMiddleware())
		{
			adminAPI.GET("/pending", handler.ListPendingRecords)
			adminAPI.POST("/approve/:id", handler.ApproveRecord)
			adminAPI.POST("/reject/:id", handler.RejectRecord)
		}
	}
}
