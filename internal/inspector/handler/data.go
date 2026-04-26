package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/inspector/request"
	"VPSBenchmarkBackend/internal/inspector/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateHost
//
// @Summary Create Inspect Host
// @Tags inspector
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateHostRequest true "Host information, including target and optional monitor_type (ping|tcp|http)"
// @Success 201 {object} common.APIResponse[any]
// @Router /inspector/hosts/create [post]
func CreateHost(ctx *gin.Context) {
	var req request.CreateHostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	id, err := service.CreateHost(userID.(int64), req.Target, req.MonitorType, req.Name, req.Tags, req.Notify, req.NotifyTolerance, req.TrafficSettlementDay, req.MonthlyTrafficLimit)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, common.Success(gin.H{"id": id}))
}

// UpdateHost
//
// @Summary Update Inspect Host
// @Tags inspector
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Host ID"
// @Param request body request.UpdateHostRequest true "Host update information, including optional monitor_type (ping|tcp|http)"
// @Success 200 {object} common.APIResponse[any]
// @Router /inspector/hosts/update/{id} [post]
func UpdateHost(ctx *gin.Context) {
	hostID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid host ID"))
		return
	}

	var req request.UpdateHostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	err = service.UpdateHost(userID.(int64), hostID, req.Name, req.Tags, req.Target, req.MonitorType, req.Notify, req.NotifyTolerance, req.TrafficSettlementDay, req.MonthlyTrafficLimit)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}

// DeleteHost
//
// @Summary Delete Inspect Host
// @Tags inspector
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Host ID"
// @Success 200 {object} common.APIResponse[any]
// @Router /inspector/hosts/delete/{id} [post]
func DeleteHost(ctx *gin.Context) {
	hostID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid host ID"))
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	err = service.DeleteHost(userID.(int64), hostID)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}

// ListHosts
//
// @Summary List Inspect Hosts
// @Tags inspector
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[[]response.HostListResponse]
// @Router /inspector/hosts [get]
func ListHosts(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	hosts, err := service.ListHosts(userID.(int64))
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success(hosts))
}

// PutData
//
// @Summary Put Inspect Data
// @Tags inspector
// @Accept json
// @Produce json
// @Security None
// @Param request body request.PutDataRequest true "Traffic and ping data"
// @Success 200 {object} common.APIResponse[any]
// @Router /inspector/data/put [post]
func PutData(ctx *gin.Context) {
	var req request.PutDataRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	err := service.PutData(req.Traffic, req.HostInfo, req.HostID)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}

// QueryData
//
// @Summary Query Inspect Data
// @Description Query monitor data for the current user. Supports ping, TCPing, and HTTPing based hosts.
// @Tags inspector
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start query int true "Start timestamp (nanoseconds)"
// @Param end query int true "End timestamp (nanoseconds)"
// @Param interval query string true "Aggregation interval (e.g. 1h, 30m)"
// @Success 200 {object} common.APIResponse[[]response.HostData]
// @Router /inspector/data [get]
func QueryData(ctx *gin.Context) {
	var req request.QueryDataRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	data, err := service.QueryData(userID.(int64), req.Start, req.End, req.Interval)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success(data))
}

// GetVisitorPage
//
// @Summary Get Inspector Visitor Page
// @Description Query the public visitor page data for a user's selected inspector hosts.
// @Tags inspector
// @Accept json
// @Produce json
// @Param start query int true "Start timestamp (nanoseconds)"
// @Param end query int true "End timestamp (nanoseconds)"
// @Param interval query string true "Aggregation interval (e.g. 1h, 30m)"
// @Success 200 {object} common.APIResponse[response.VisitorPageData]
// @Router /inspector/visitor/{id} [get]
func GetVisitorPage(ctx *gin.Context) {
	var req request.VisitorPageRequest
	ownerIDStr := ctx.Param("id")
	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid owner ID"))
		return
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	data, err := service.GetVisitorPage(ownerID, req.Start, req.End, req.Interval)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success(data))
}
