package handler

import (
	"VPSBenchmarkBackend/internal/auth/response"
	"VPSBenchmarkBackend/internal/auth/service"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// GithubLogin handles GitHub OAuth login
// @Summary GitHub OAuth Login
// @Description Exchange GitHub OAuth code for JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "GitHub OAuth code"
// @Success 200 {object} common.APIResponse[response.LoginResponse]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/github/login [get]
func GithubLogin(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid request: code is required"))
		return
	}

	token, err := service.GithubLogin(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "Failed to login with GitHub: "+err.Error()))
		return
	}
	query := url.Values{}
	query.Add("token", token)
	rawQuery := query.Encode()
	frontendURL := config.Get().FrontendURL + "?" + rawQuery
	c.Redirect(http.StatusPermanentRedirect, frontendURL)
	// c.JSON(http.StatusOK, common.Success(response.LoginResponse{Token: token}))
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

// CheckAdminUser is a dummy handler to check if the user is admin
// @Summary Check Admin User
// @Description Check if the current user is an admin
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[any]
// @Failure 403 {object} common.APIResponse[any]
// @Router /auth/admin [get]
func CheckAdminUser(c *gin.Context) {
	c.JSON(http.StatusOK, common.Success[any](nil))
}
