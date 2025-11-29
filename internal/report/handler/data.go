package handler

import (
	"VPSBenchmarkBackend/internal/report/request"
	"VPSBenchmarkBackend/internal/report/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListReports handles GET /report/data/list
func ListReports(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	reports, total, err := service.ListReports(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetReportDetails handles GET /report/data/details?id=xxx
func GetReportDetails(ctx *gin.Context) {
	reportID := ctx.Query("id")
	if reportID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "report ID is required",
		})
		return
	}

	report, err := service.GetReportDetails(reportID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": report,
	})
}

// TODO: 添加获取流媒体列表、虚拟化类型列表接口
// SearchReports handles GET /report/data/search
func SearchReports(ctx *gin.Context) {
	var searchReq request.SearchRequest
	if err := ctx.ShouldBindJSON(&searchReq); err != nil {
		// If no JSON body, try query parameters
		keyword := ctx.Query("keyword")
		if keyword != "" {
			searchReq.Keyword = &keyword
		}
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	reports, total, err := service.SearchReports(&searchReq, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
