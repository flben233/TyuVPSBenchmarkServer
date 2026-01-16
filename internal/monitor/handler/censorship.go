package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/monitor/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListPendingHosts handles GET /monitor/admin/pending - lists hosts awaiting review
// @Summary List Pending Hosts
// @Description List all hosts awaiting review (admin only)
// @Tags monitor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[[]response.HostResponse]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /monitor/admin/pending [get]
func ListPendingHosts(ctx *gin.Context) {
	hosts, err := service.ListPendingHosts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(hosts))
}

// ApproveHost handles POST /monitor/admin/approve/:id - approves a host
// @Summary Approve Host
// @Description Approve a host for public display (admin only)
// @Tags monitor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Host ID"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /monitor/admin/approve/{id} [post]
func ApproveHost(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid host ID"))
		return
	}

	err = service.ApproveHost(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}

// RejectHost handles POST /monitor/admin/reject/:id - rejects a host
// @Summary Reject Host
// @Description Reject a host (admin only)
// @Tags monitor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Host ID"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /monitor/admin/reject/{id} [post]
func RejectHost(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid host ID"))
		return
	}

	err = service.RejectHost(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}
