package handler

import (
	"encoding/json"
	"net/http"

	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/store"

	"github.com/gin-gonic/gin"
)

type WhitelistResponse struct {
	Commands []string `json:"commands"`
}

type SaveWhitelistRequest struct {
	Commands []string `json:"commands" binding:"required"`
}

// HandleGetWhitelist returns the user's persistent command whitelist.
// @Summary Get command whitelist
// @Description Retrieve the user's persistent command whitelist for the LLM agent.
// @Tags webssh
// @Accept json
// @Produce json
// @Success 200 {object} common.APIResponse[WhitelistResponse]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /webssh/llm/whitelist/get [get]
func HandleGetWhitelist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	record, err := store.GetWhitelist(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "Failed to retrieve whitelist"))
		return
	}

	var commands []string
	if record != nil && record.Commands != "" {
		_ = json.Unmarshal([]byte(record.Commands), &commands)
	}
	if commands == nil {
		commands = []string{}
	}

	c.JSON(http.StatusOK, common.Success(WhitelistResponse{Commands: commands}))
}

// HandleSaveWhitelist saves the user's persistent command whitelist.
// @Summary Save command whitelist
// @Description Save or update the user's persistent command whitelist for the LLM agent.
// @Tags webssh
// @Accept json
// @Produce json
// @Param request body SaveWhitelistRequest true "Whitelist to save"
// @Success 200 {object} common.APIResponse[map[string]bool]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 401 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /webssh/llm/whitelist/save [post]
func HandleSaveWhitelist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	var req SaveWhitelistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "Invalid request: "+err.Error()))
		return
	}

	commandsJSON, err := json.Marshal(req.Commands)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "Failed to marshal commands"))
		return
	}

	if err := store.UpsertWhitelist(userID.(int64), string(commandsJSON)); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "Failed to save whitelist"))
		return
	}

	c.JSON(http.StatusOK, common.Success(map[string]bool{"success": true}))
}
