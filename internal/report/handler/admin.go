package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/report/request"
	"VPSBenchmarkBackend/internal/report/response"
	"VPSBenchmarkBackend/internal/report/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// QueryReportTaskStatus handles GET /report/admin/status
// @Summary Query Report Task Status (Admin)
// @Description Query the status of an asynchronous report task by ID. Requires admin authentication.
// @Tags report
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query string true "Report Task ID"
// @Success 200 {object} common.APIResponse[string]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/admin/status [get]
func QueryReportTaskStatus(ctx *gin.Context) {
	id := ctx.Query("id")
	status, err := service.QueryReportTaskStatus(id)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success(status))
}

// AddReport handles POST /report/admin/add
// @Summary Add Report (Admin)
// @Description Async submit reports. Request should be JSON with `html` field. Requires admin authentication.
// @Tags report
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.AddReportRequest false "Report HTML or raw body"
// @Success 201 {object} common.APIResponse[[]response.ReportIDResponse]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/admin/add [post]
func AddReport(ctx *gin.Context) {
	var req []request.AddReportRequest

	// Try to bind JSON first
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "invalid request format"))
		return
	}

	reportID, err := service.AddReportsAsync(req)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, common.SuccessWithMessage("Report added successfully", response.ReportIDResponse{Id: reportID}))
}

// DeleteReport handles POST /report/admin/delete
// @Summary Delete Report (Admin)
// @Description Delete an existing report by ID. Requires admin authentication.
// @Tags report
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.DeleteReportRequest false "Report ID"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/admin/delete [post]
func DeleteReport(ctx *gin.Context) {
	var req request.DeleteReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Try query parameter
		req.ID = ctx.Query("id")
		if req.ID == "" {
			ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "report ID is required"))
			return
		}
	}

	if err := service.DeleteReport(req.ID); err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessWithMessage[any]("Report deleted successfully", nil))
}

// UpdateReport handles POST /report/admin/update
// @Summary Update the Monitor ID of a Report (Admin)
// @Description Update the monitor ID of a report by ID. Requires admin authentication.
// @Tags report
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UpdateReportRequest false "Report ID"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/admin/update [post]
func UpdateReport(ctx *gin.Context) {
	var req request.UpdateReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "invalid request format"))
		return
	}

	if req.ID == "" {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "report ID and monitor ID are required"))
		return
	}

	if err := service.UpdateReport(req.ID, req.MonitorID, req.OtherInfo); err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessWithMessage[any]("Report monitor ID updated successfully", nil))
}
