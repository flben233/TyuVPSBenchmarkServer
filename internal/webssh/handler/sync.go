package handler

import (
	"net/http"

	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/store"

	"github.com/gin-gonic/gin"
)

type UploadRequest struct {
	EncryptedData string `json:"encrypted_data" binding:"required"`
}

type DownloadResponse struct {
	EncryptedData string `json:"encrypted_data"`
	UpdatedAt     string `json:"updated_at"`
}

func HandleUpload(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Error(common.InvalidParamCode, "Invalid request: "+err.Error()))
		return
	}

	if err := store.UpsertSyncData(userID.(int64), req.EncryptedData); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "Failed to save data"))
		return
	}

	c.JSON(http.StatusOK, common.Success(map[string]bool{"success": true}))
}

func HandleDownload(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	sync, err := store.GetSyncData(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "Failed to retrieve data"))
		return
	}

	if sync == nil {
		c.JSON(http.StatusOK, common.Success(DownloadResponse{
			EncryptedData: "",
			UpdatedAt:     "",
		}))
		return
	}

	c.JSON(http.StatusOK, common.Success(DownloadResponse{
		EncryptedData: sync.EncryptedData,
		UpdatedAt:     sync.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}))
}

func HandleReset(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, common.Error(common.BadRequestCode, "User not authenticated"))
		return
	}

	if err := store.DeleteSyncData(userID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, "Failed to reset data"))
		return
	}

	c.JSON(http.StatusOK, common.Success(map[string]bool{"success": true}))
}
