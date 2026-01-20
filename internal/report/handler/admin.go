package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/report/request"
	"VPSBenchmarkBackend/internal/report/response"
	"VPSBenchmarkBackend/internal/report/service"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddReport handles POST /report/admin/add
// @Summary Add Report (Admin)
// @Description Add a new report. Request can be JSON with `html` field or raw HTML body. Requires admin authentication.
// @Tags report
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.AddReportRequest false "Report HTML or raw body"
// @Success 201 {object} common.APIResponse[response.ReportIDResponse]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /report/admin/add [post]
func AddReport(ctx *gin.Context) {
	var req request.AddReportRequest

	// Try to bind JSON first
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// If JSON binding fails, try to read raw HTML from body
		body, readErr := io.ReadAll(ctx.Request.Body)
		if readErr != nil {
			ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "invalid request format"))
			return
		}
		req.HTML = string(body)
	}

	if req.HTML == "" {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "HTML content is required"))
		return
	}

	reportID, err := service.AddReport(req.HTML, req.MonitorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
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
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.SuccessWithMessage[any]("Report deleted successfully", nil))
}
