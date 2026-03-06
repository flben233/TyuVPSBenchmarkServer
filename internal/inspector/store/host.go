package store

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/inspector/model"
	"context"
	"fmt"
	"gorm.io/gorm"
)

var db *gorm.DB
var hosts gorm.Interface[model.InspectHost]

func init() {
	common.RegisterDBInitializer(InitializeHostStore)
}

func InitializeHostStore(dbPath string) error {
	db = common.GetDB()
	if err := db.AutoMigrate(&model.InspectHost{}); err != nil {
		return fmt.Errorf("failed to migrate host store: %w", err)
	}
	hosts = gorm.G[model.InspectHost](db)
	return nil
}

func CreateHost(host *model.InspectHost) error {
	return hosts.Create(context.Background(), host)
}

func CountUserHosts(userID int64) (int64, error) {
	count, err := hosts.Where("user_id = ?", userID).Count(context.Background(), "*")
	return count, err
}

func GetHostByID(id int64) (*model.InspectHost, error) {
	host, err := hosts.Where("id = ?", id).First(context.Background())
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func UpdateHost(host *model.InspectHost) {
	db.Save(host)
}

func DeleteHost(id int64) error {
	_, err := hosts.Where("id = ?", id).Delete(context.Background())
	return err
}

func ListHostsByUser(userID int64) ([]model.InspectHost, error) {
	return hosts.Where("user_id = ?", userID).Find(context.Background())
}

func ListAllHost() ([]model.InspectHost, error) {
	return hosts.Find(context.Background())
}

func GetHostIDByUser(userID int64) []int64 {
	var ids []int64
	db.Model(&model.InspectHost{}).Where("user_id = ?", userID).Pluck("id", &ids)
	return ids
}
