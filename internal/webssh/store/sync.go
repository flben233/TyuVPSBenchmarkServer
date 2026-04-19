package store

import (
	"context"
	"errors"

	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/model"

	"gorm.io/gorm"
)

var db *gorm.DB
var syncRecords gorm.Interface[model.WebsshSync]

func init() {
	common.RegisterDBInitializer(InitSyncStore)
}

func InitSyncStore(dbPath string) error {
	db = common.GetDB()
	syncRecords = gorm.G[model.WebsshSync](db)
	if err := db.AutoMigrate(&model.WebsshSync{}); err != nil {
		return err
	}
	return nil
}

func UpsertSyncData(userID int64, encryptedData string) error {
	record := model.WebsshSync{
		UserID:        userID,
		EncryptedData: encryptedData,
	}
	result := db.Where("user_id = ?", userID).Assign(model.WebsshSync{EncryptedData: encryptedData}).FirstOrCreate(&record)
	return result.Error
}

func GetSyncData(userID int64) (*model.WebsshSync, error) {
	var sync model.WebsshSync
	if err := db.Where("user_id = ?", userID).First(&sync).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sync, nil
}

func DeleteSyncData(userID int64) error {
	_, err := syncRecords.Where("user_id = ?", userID).Delete(context.Background())
	return err
}
