package store

import (
	"errors"

	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/webssh/model"

	"gorm.io/gorm"
)

func init() {
	common.RegisterDBInitializer(InitWhitelistStore)
}

func InitWhitelistStore(dbPath string) error {
	database := common.GetDB()
	if err := database.AutoMigrate(&model.WebsshCommandWhitelist{}); err != nil {
		return err
	}
	return nil
}

func GetWhitelist(userID int64) (*model.WebsshCommandWhitelist, error) {
	database := common.GetDB()
	var record model.WebsshCommandWhitelist
	if err := database.Where("user_id = ?", userID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func UpsertWhitelist(userID int64, commands string) error {
	database := common.GetDB()
	record := model.WebsshCommandWhitelist{
		UserID:   userID,
		Commands: commands,
	}
	result := database.Where("user_id = ?", userID).
		Assign(model.WebsshCommandWhitelist{Commands: commands}).
		FirstOrCreate(&record)
	return result.Error
}
