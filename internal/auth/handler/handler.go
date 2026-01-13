package handler

import (
	"VPSBenchmarkBackend/internal/auth/response"
	"VPSBenchmarkBackend/internal/auth/service"
	"VPSBenchmarkBackend/internal/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GithubLoginRequest is the request body for GitHub login
type GithubLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// GithubLogin handles GitHub OAuth login
// @Summary GitHub OAuth Login
// @Description Exchange GitHub OAuth code for JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GithubLoginRequest true "GitHub OAuth code"
// @Success 200 {object} common.APIResponse[response.LoginResponse]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/github/login [post]
func GithubLogin(c *gin.Context) {
	var req GithubLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid request: code is required"))
		return
	}

	token, err := service.GithubLogin(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "Failed to login with GitHub: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.Success(response.LoginResponse{Token: token}))
}

// GetUserInfo returns the current user's information from JWT
// @Summary Get User Info
// @Description Get current user information from JWT token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[response.UserInfo]
// @Failure 401 {object} common.APIResponse[any]
// @Router /auth/user [get]
func GetUserInfo(c *gin.Context) {
	// Get user info from context (set by middleware)
	name, exists := c.Get("user_name")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	avatarURL, _ := c.Get("user_avatar_url")

	userInfo := response.UserInfo{
		Name:      name.(string),
		AvatarURL: avatarURL.(string),
	}

	c.JSON(http.StatusOK, common.Success(userInfo))
}
