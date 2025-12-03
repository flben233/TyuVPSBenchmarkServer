package handler

import (
	"VPSBenchmarkBackend/internal/report/response"
	"VPSBenchmarkBackend/internal/report/service"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddReportRequest represents the request body for adding a report
type AddReportRequest struct {
	HTML string `json:"html" binding:"required"`
}

// DeleteReportRequest represents the request body for deleting a report
type DeleteReportRequest struct {
	ID string `json:"id" binding:"required"`
}

// AddReport handles POST /report/admin/add
func AddReport(ctx *gin.Context) {
	var req AddReportRequest

	// Try to bind JSON first
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// If JSON binding fails, try to read raw HTML from body
		body, readErr := io.ReadAll(ctx.Request.Body)
		if readErr != nil {
			ctx.JSON(http.StatusBadRequest, response.Error(response.BadRequestCode, "invalid request format"))
			return
		}
		req.HTML = string(body)
	}

	if req.HTML == "" {
		ctx.JSON(http.StatusBadRequest, response.Error(response.BadRequestCode, "HTML content is required"))
		return
	}

	reportID, err := service.AddReport(req.HTML)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error(response.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, response.SuccessWithMessage("Report added successfully", gin.H{
		"report_id": reportID,
	}))
}

// DeleteReport handles POST /report/admin/delete
func DeleteReport(ctx *gin.Context) {
	var req DeleteReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Try query parameter
		req.ID = ctx.Query("id")
		if req.ID == "" {
			ctx.JSON(http.StatusBadRequest, response.Error(response.BadRequestCode, "report ID is required"))
			return
		}
	}

	if err := service.DeleteReport(req.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error(response.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.SuccessWithMessage("Report deleted successfully", nil))
}
