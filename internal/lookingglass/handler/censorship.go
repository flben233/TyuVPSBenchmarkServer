package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/lookingglass/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListPendingRecords handles GET /lookingglass/admin/pending - lists records awaiting review
// @Summary List Pending Records
// @Description List all looking glass records awaiting review (admin only)
// @Tags lookingglass
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[[]response.LookingGlassResponse]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /lookingglass/admin/pending [get]
func ListPendingRecords(ctx *gin.Context) {
	records, err := service.ListPendingRecords()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(records))
}

// ApproveRecord handles POST /lookingglass/admin/approve/:id - approves a record
// @Summary Approve Record
// @Description Approve a looking glass record for public display (admin only)
// @Tags lookingglass
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Record ID"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /lookingglass/admin/approve/{id} [post]
func ApproveRecord(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid record ID"))
		return
	}

	err = service.ApproveRecord(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}

// RejectRecord handles POST /lookingglass/admin/reject/:id - rejects a record
// @Summary Reject Record
// @Description Reject a looking glass record (admin only)
// @Tags lookingglass
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Record ID"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /lookingglass/admin/reject/{id} [post]
func RejectRecord(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid record ID"))
		return
	}

	err = service.RejectRecord(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}
