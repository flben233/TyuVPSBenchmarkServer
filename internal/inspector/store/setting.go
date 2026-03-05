package store

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/inspector/model"
	"context"
	"fmt"
	"gorm.io/gorm"
)

var db2 *gorm.DB
var settings gorm.Interface[model.InspectorSetting]

func init() {
	common.RegisterDBInitializer(InitializeSettingStore)
}

func InitializeSettingStore(dbPath string) error {
	db2 = common.GetDB()
	if err := db2.AutoMigrate(&model.InspectorSetting{}); err != nil {
		return fmt.Errorf("failed to migrate inspector setting store: %w", err)
	}
	settings = gorm.G[model.InspectorSetting](db2)
	return nil
}

func UpsertSetting(setting *model.InspectorSetting) error {
	return db2.Save(setting).Error
}

func GetSettingByUserID(userID int64) (*model.InspectorSetting, error) {
	setting, err := settings.Where("user_id = ?", userID).First(context.Background())
	if err != nil {
		return nil, err
	}
	return &setting, nil
}
