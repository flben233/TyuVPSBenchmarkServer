package report

import (
	"VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/report/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(base string, r *gin.Engine) {
	base += "/report"
	{
		reportAPI := r.Group(base + "/data")
		reportAPI.GET("/list", handler.ListReports)
		reportAPI.GET("/details", handler.GetReportDetails)
		reportAPI.GET("/search", handler.SearchReports)
		reportAPI.POST("/search", handler.SearchReports)
		reportAPI.GET("/media-names", handler.GetAllMediaNames)
		reportAPI.GET("/virtualizations", handler.GetAllVirtualizations)
		reportAPI.GET("/backroute-types", handler.GetAllBackRouteTypes)
	}
	{
		adminAPI := r.Group(base + "/admin")
		adminAPI.Use(auth.GetJWTMiddleware()) // Protect admin routes with JWT authentication
		adminAPI.POST("/add", handler.AddReport)
		adminAPI.POST("/delete", handler.DeleteReport)
	}
}
