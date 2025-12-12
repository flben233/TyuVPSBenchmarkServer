package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/monitor/request"
	"VPSBenchmarkBackend/internal/monitor/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddHost handles POST /monitor/hosts - adds a new monitoring host
func AddHost(ctx *gin.Context) {
	var req request.HostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	// Get username from context (set by auth middleware)
	username, exists := ctx.Get("user_login")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	id, err := service.AddHost(username.(string), req.Target, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, common.Success(map[string]int64{"id": id}))
}

// RemoveHost handles DELETE /monitor/hosts/:id - removes a monitoring host
func RemoveHost(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid host ID"))
		return
	}

	// Get username from context
	username, exists := ctx.Get("user_login")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	// Check if user is admin
	isAdmin := false
	cfg := config.Get()
	if username.(string) == cfg.AdminID {
		isAdmin = true
	}

	err = service.RemoveHost(username.(string), id, isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(nil))
}

// ListHosts handles GET /monitor/hosts - lists monitoring hosts
func ListHosts(ctx *gin.Context) {
	// Get username from context
	username, exists := ctx.Get("user_login")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	// Check if user is admin
	isAdmin := false
	cfg := config.Get()
	if username.(string) == cfg.AdminID {
		isAdmin = true
	}

	hosts, err := service.ListHosts(username.(string), isAdmin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(hosts))
}

// GetStatistics handles GET /monitor/statistics - gets monitoring statistics
func GetStatistics(ctx *gin.Context) {
	stats, err := service.GetStatistics()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(stats))
}
