package handler

import (
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

// TODO: 这里直接用请求体提交，不需要包一层
// AddReport handles POST /report/admin/add
func AddReport(ctx *gin.Context) {
	var req AddReportRequest

	// Try to bind JSON first
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// If JSON binding fails, try to read raw HTML from body
		body, readErr := io.ReadAll(ctx.Request.Body)
		if readErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request format",
			})
			return
		}
		req.HTML = string(body)
	}

	if req.HTML == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "HTML content is required",
		})
		return
	}

	reportID, err := service.AddReport(req.HTML)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":   "Report added successfully",
		"report_id": reportID,
	})
}

// DeleteReport handles POST /report/admin/delete
func DeleteReport(ctx *gin.Context) {
	var req DeleteReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Try query parameter
		req.ID = ctx.Query("id")
		if req.ID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "report ID is required",
			})
			return
		}
	}

	if err := service.DeleteReport(req.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Report deleted successfully",
	})
}
