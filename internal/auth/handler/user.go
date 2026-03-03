package handler

import (
	"VPSBenchmarkBackend/internal/auth/model"
	"VPSBenchmarkBackend/internal/auth/service"
	"VPSBenchmarkBackend/internal/common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUser handles GET /auth/admin/user/:id
// @Summary Get User
// @Description Get a user by ID
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} common.APIResponse[model.User]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/admin/user/{id} [get]
func GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "invalid user id"))
		return
	}
	user, err := service.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Success(user))
}

// ListUsers handles GET /auth/admin/users
// @Summary List Users
// @Description List all users
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[[]model.User]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/admin/users [get]
func ListUsers(c *gin.Context) {
	users, err := service.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Success(users))
}

// UpdateUser handles POST /auth/admin/user/update
// @Summary Update User
// @Description Update an existing user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.User true "User"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/admin/user [post]
func UpdateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "invalid request body"))
		return
	}
	if err := service.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.SuccessWithMessage[any]("user updated", nil))
}

// DeleteUser handles POST /auth/admin/user/delete
// @Summary Delete User
// @Description Delete a user by ID
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "User ID" example({"id": 1})
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/admin/user/delete [post]
func DeleteUser(c *gin.Context) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		c.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "invalid user id"))
		return
	}
	if err := service.DeleteUser(req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.SuccessWithMessage[any]("user deleted", nil))
}

// CreateUserGroup handles POST /auth/admin/group
// @Summary Create User Group
// @Description Create a new user group
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.UserGroup true "UserGroup"
// @Success 201 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/admin/group/create [post]
func CreateUserGroup(c *gin.Context) {
	var group model.UserGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "invalid request body"))
		return
	}
	if err := service.CreateUserGroup(group); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}
	c.JSON(http.StatusCreated, common.SuccessWithMessage[any]("group created", nil))
}

// ListUserGroups handles GET /auth/admin/groups
// @Summary List User Groups
// @Description List all user groups
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[[]model.UserGroup]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/admin/groups [get]
func ListUserGroups(c *gin.Context) {
	groups, err := service.ListUserGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Success(groups))
}

// UpdateUserGroup handles POST /auth/admin/group/update
// @Summary Update User Group
// @Description Update an existing user group
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.UserGroup true "UserGroup"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/admin/group [post]
func UpdateUserGroup(c *gin.Context) {
	var group model.UserGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "invalid request body"))
		return
	}
	if err := service.UpdateUserGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.SuccessWithMessage[any]("group updated", nil))
}

// DeleteUserGroup handles POST /auth/admin/group/delete
// @Summary Delete User Group
// @Description Delete a user group by ID
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Group ID" example({"id": 1})
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /auth/admin/group/delete [post]
func DeleteUserGroup(c *gin.Context) {
	var req struct {
		ID uint32 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		c.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "invalid group id"))
		return
	}
	if err := service.DeleteUserGroup(req.ID); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.SuccessWithMessage[any]("group deleted", nil))
}
