package handler

import (
	"VPSBenchmarkBackend/internal/report/request"
	"VPSBenchmarkBackend/internal/report/response"
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
		ctx.JSON(http.StatusInternalServerError, response.Error(response.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessPaginated(reports, total, page, pageSize))
}

// GetReportDetails handles GET /report/data/details?id=xxx
func GetReportDetails(ctx *gin.Context) {
	reportID := ctx.Query("id")
	if reportID == "" {
		ctx.JSON(http.StatusBadRequest, response.Error(response.BadRequestCode, "report ID is required"))
		return
	}

	report, err := service.GetReportDetails(reportID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error(404, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(report))
}

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
		ctx.JSON(http.StatusInternalServerError, response.Error(response.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessPaginated(reports, total, page, pageSize))
}

// GetAllMediaNames handles GET /report/data/media-names
func GetAllMediaNames(ctx *gin.Context) {
	mediaNames, err := service.GetAllMediaNames()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error(response.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(mediaNames))
}

// GetAllVirtualizations handles GET /report/data/virtualizations
func GetAllVirtualizations(ctx *gin.Context) {
	virtualizations, err := service.GetAllVirtualizations()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error(response.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(virtualizations))
}

// GetAllBackRouteTypes handles GET /report/data/backroute-types
func GetAllBackRouteTypes(ctx *gin.Context) {
	routeTypes, err := service.GetAllBackRouteTypes()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error(response.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(routeTypes))
}
