package mq

import (
	"VPSBenchmarkBackend/internal/common"
	"github.com/gin-gonic/gin"
	"net/http"
)

// QueryTaskStatus QueryReportTaskStatus handles GET /task/status/{id}
// @Summary Query Report Task Status
// @Description Query the status of an asynchronous report task by ID.
// @Tags report
// @Accept json
// @Produce json
// @Param id path string true "Report Task ID"
// @Success 200 {object} common.APIResponse[Task[any]]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /task/status/{id} [get]
func QueryTaskStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	status, err := HandleQuery(id)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success(status))
}
