package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/inspector/request"
	"VPSBenchmarkBackend/internal/inspector/response"
	"VPSBenchmarkBackend/internal/inspector/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserSettings
//
// @Summary Get Inspector User Settings
// @Tags inspector
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[response.SettingData]
// @Router /inspector/settings [get]
func GetUserSettings(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	data := service.GetUserSettings(userID.(int64))
	ctx.JSON(http.StatusOK, common.Success(data))
}

// UpdateUserSettings
//
// @Summary Update Inspector User Settings
// @Tags inspector
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UpdateInspectorSettingRequest true "User setting information"
// @Success 200 {object} common.APIResponse[response.SettingData]
// @Router /inspector/settings/update [post]
func UpdateUserSettings(ctx *gin.Context) {
	var req request.UpdateInspectorSettingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	if err := service.UpdateUserSettings(userID.(int64), req.NotifyURL, req.BgURL); err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	data := service.GetUserSettings(userID.(int64))
	ctx.JSON(http.StatusOK, common.Success[response.SettingData](*data))
}
