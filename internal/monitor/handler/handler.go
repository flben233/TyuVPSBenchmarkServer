package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/monitor/request"
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/internal/monitor/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddHost handles POST /monitor/hosts - adds a new monitoring host
// @Summary Add Monitoring Host
// @Description Add a new host to monitoring list. Requires authentication.
// @Tags monitor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.HostRequest true "Host information"
// @Success 201 {object} common.APIResponse[response.HostIDResponse]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 403 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /monitor/hosts [post]
func AddHost(ctx *gin.Context) {
	var req request.HostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	// Get userID from context (set by auth middleware)
	userID, exists := ctx.Get("user_id")
	userName, exists := ctx.Get("user_name")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}
	id, err := service.AddHost(userID.(int64), userName.(string), req.Target, req.Name)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, common.Success(response.HostIDResponse{Id: id}))
}

// RemoveHost handles POST /monitor/hosts/:id - removes a monitoring host
// @Summary Remove Monitoring Host
// @Description Remove a host from monitoring by ID. Requires authentication.
// @Tags monitor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Host ID"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /monitor/hosts/{id} [post]
func RemoveHost(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid host ID"))
		return
	}

	// Get userID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	err = service.RemoveHost(userID.(int64), id)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}

// ListHosts handles GET /monitor/hosts - lists monitoring hosts
// @Summary List Monitoring Hosts
// @Description List monitoring hosts for current user. Requires authentication.
// @Tags monitor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[[]response.HostResponse]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /monitor/hosts [get]
func ListHosts(ctx *gin.Context) {
	// Get userID from context
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

// GetStatistics handles GET /monitor/statistics - gets monitoring statistics
// @Summary Get Monitoring Statistics
// @Description Retrieve overall monitoring statistics (public).
// @Tags monitor
// @Accept json
// @Produce json
// @Success 200 {object} common.APIResponse[[]response.StatisticsResponse]
// @Failure 500 {object} common.APIResponse[any]
// @Router /monitor/statistics [get]
func GetStatistics(ctx *gin.Context) {
	id := ctx.Query("id")
	stats, err := service.GetStatistics(id)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success(stats))
}

// GetServerStatus handles GET /monitor/status - gets current server status
// @Summary Get Server Status
// @Description Get current server uptime, CPU usage, memory usage, and network throughput.
// @Tags monitor
// @Accept json
// @Produce json
// @Success 200 {object} common.APIResponse[response.ServerStatusResponse]
// @Failure 500 {object} common.APIResponse[any]
// @Router /monitor/status [get]
func GetServerStatus(ctx *gin.Context) {
	status, err := service.GetServerStatus()
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success(status))
}
