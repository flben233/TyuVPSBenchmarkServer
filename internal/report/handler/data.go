package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/report/request"
	"VPSBenchmarkBackend/internal/report/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListReports handles GET /report/data/list
// @Summary List Reports
// @Description Get paginated list of reports. Supports `page` and `page_size` query parameters.
// @Tags report
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} common.PaginatedResponse[[]response.ReportInfoResponse]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/data/list [get]
func ListReports(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	reports, total, err := service.ListReports(page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessPaginated(reports, total, page, pageSize))
}

// GetReportDetails handles GET /report/data/details?id=xxx
// @Summary Get Report Details
// @Description Get full details of a report by ID.
// @Tags report
// @Accept json
// @Produce json
// @Param id query string true "Report ID"
// @Success 200 {object} common.APIResponse[model.BenchmarkResult]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 404 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/data/details [get]
func GetReportDetails(ctx *gin.Context) {
	reportID := ctx.Query("id")
	if reportID == "" {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "report ID is required"))
		return
	}

	report, err := service.GetReportDetails(reportID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, common.Error(404, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(report))
}

// SearchReports handles GET /report/data/search
// @Summary Search Reports
// @Description Search reports by criteria. Accepts JSON body or query parameters. Supports pagination.
// @Tags report
// @Accept json
// @Produce json
// @Param request body request.SearchRequest false "Search filters"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} common.PaginatedResponse[[]response.ReportInfoResponse]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/data/search [get]
// @Router /report/data/search [post]
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
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessPaginated(reports, total, page, pageSize))
}

// GetAllMediaNames handles GET /report/data/media-names
// @Summary Get Media Names
// @Description Return all available media names for reports.
// @Tags report
// @Accept json
// @Produce json
// @Success 200 {object} common.APIResponse[[]string]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/data/media-names [get]
func GetAllMediaNames(ctx *gin.Context) {
	mediaNames, err := service.GetAllMediaNames()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(mediaNames))
}

// GetAllVirtualizations handles GET /report/data/virtualizations
// @Summary Get Virtualization Types
// @Description Return all virtualization types used in reports.
// @Tags report
// @Accept json
// @Produce json
// @Success 200 {object} common.APIResponse[[]string]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/data/virtualizations [get]
func GetAllVirtualizations(ctx *gin.Context) {
	virtualizations, err := service.GetAllVirtualizations()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(virtualizations))
}

// GetAllBackRouteTypes handles GET /report/data/backroute-types
// @Summary Get Backroute Types
// @Description Return all backroute types used in reports.
// @Tags report
// @Accept json
// @Produce json
// @Success 200 {object} common.APIResponse[[]string]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/data/backroute-types [get]
func GetAllBackRouteTypes(ctx *gin.Context) {
	routeTypes, err := service.GetAllBackRouteTypes()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(routeTypes))
}
