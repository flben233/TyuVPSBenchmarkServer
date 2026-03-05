package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/lookingglass/request"
	"VPSBenchmarkBackend/internal/lookingglass/response"
	"VPSBenchmarkBackend/internal/lookingglass/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddRecord handles POST /lookingglass/records - adds a new looking glass record
// @Summary Add Looking Glass Record
// @Description Add a new looking glass record. Requires authentication.
// @Tags lookingglass
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.LookingGlassRequest true "Looking glass record information"
// @Success 201 {object} common.APIResponse[response.LookingGlassIDResponse]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 403 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /lookingglass/records [post]
func AddRecord(ctx *gin.Context) {
	var req request.LookingGlassRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	userID, exists := ctx.Get("user_id")
	userName, exists := ctx.Get("user_name")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}
	id, err := service.AddRecord(userID.(int64), userName.(string), req.ServerName, req.TestURL)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, common.Success(response.LookingGlassIDResponse{Id: id}))
}

// UpdateRecord handles POST /lookingglass/records/delete/:id - updates a looking glass record
// @Summary Update Looking Glass Record
// @Description Update a looking glass record by ID. Requires authentication.
// @Tags lookingglass
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Record ID"
// @Param request body request.LookingGlassRequest true "Updated record information"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /lookingglass/records/{id} [post]
func UpdateRecord(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid record ID"))
		return
	}

	var req request.LookingGlassRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, err.Error()))
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	err = service.UpdateRecord(userID.(int64), id, req.ServerName, req.TestURL)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}

// RemoveRecord handles POST /lookingglass/records/delete/:id - removes a looking glass record
// @Summary Remove Looking Glass Record
// @Description Remove a looking glass record by ID. Requires authentication.
// @Tags lookingglass
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Record ID"
// @Success 200 {object} common.APIResponse[any]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /lookingglass/records/{id} [post]
func RemoveRecord(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "Invalid record ID"))
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	err = service.RemoveRecord(userID.(int64), id)
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success[any](nil))
}

// ListRecords handles GET /lookingglass/records - lists looking glass records for current user
// @Summary List Looking Glass Records
// @Description List looking glass records for current user. Requires authentication.
// @Tags lookingglass
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse[[]response.LookingGlassResponse]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /lookingglass/records [get]
func ListRecords(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	records, err := service.ListRecords(userID.(int64))
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success(records))
}

// ListAllRecords handles GET /lookingglass/list - lists all looking glass records (public)
// @Summary List All Looking Glass Records
// @Description Retrieve all looking glass records (public endpoint).
// @Tags lookingglass
// @Accept json
// @Produce json
// @Success 200 {object} common.APIResponse[[]response.LookingGlassResponse]
// @Failure 500 {object} common.APIResponse[any]
// @Router /lookingglass/list [get]
func ListAllRecords(ctx *gin.Context) {
	records, err := service.ListAllRecords()
	if err != nil {
		common.DefaultErrorHandler(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, common.Success(records))
}
