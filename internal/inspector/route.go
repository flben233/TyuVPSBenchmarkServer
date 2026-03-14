package inspector

import (
	"VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/inspector/handler"

	"github.com/gin-gonic/gin"
)

func init() {
	common.RegisterRoutes(RegisterRoute)
}

func RegisterRoute(base string, r *gin.Engine) {
	base += "/inspector"

	protectedAPI := r.Group(base)
	protectedAPI.Use(auth.GetJWTMiddleware())
	{
		protectedAPI.GET("/hosts", handler.ListHosts)
		protectedAPI.POST("/hosts/create", handler.CreateHost)
		protectedAPI.POST("/hosts/update/:id", handler.UpdateHost)
		protectedAPI.POST("/hosts/delete/:id", handler.DeleteHost)

		protectedAPI.GET("/data", handler.QueryData)
		protectedAPI.GET("/settings", handler.GetUserSettings)
		protectedAPI.POST("/settings/update", handler.UpdateUserSettings)
		protectedAPI.POST("/notify/test", handler.TestNotify)
	}

	publicAPI := r.Group(base)
	{
		publicAPI.POST("/data/put", handler.PutData)
	}
}
