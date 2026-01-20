package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
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
	userID, exists := ctx.Get("user_login")
	userName, exists := ctx.Get("user_name")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}
	isAdmin := config.Get().AdminID == userID.(string)
	id, err := service.AddHost(userID.(string), userName.(string), req.Target, req.Name, isAdmin)
	if err != nil {
		if _, ok := err.(*service.HostLimitError); ok {
			ctx.JSON(http.StatusForbidden, common.Error(common.ForbiddenCode, "Host limit reached"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
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
	userID, exists := ctx.Get("user_login")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	// Check if user is admin
	isAdmin := false
	cfg := config.Get()
	if userID.(string) == cfg.AdminID {
		isAdmin = true
	}

	err = service.RemoveHost(userID.(string), id, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
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
	userID, exists := ctx.Get("user_login")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	// Check if user is admin
	isAdmin := false
	cfg := config.Get()
	if userID.(string) == cfg.AdminID {
		isAdmin = true
	}

	hosts, err := service.ListHosts(userID.(string), isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
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
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
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
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(status))
}
