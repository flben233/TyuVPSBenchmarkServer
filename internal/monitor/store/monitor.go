package store

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/monitor/model"

	"gorm.io/gorm"
)

// Table name for monitor hosts
const MonitorTableName = "monitor_hosts"

// MonitorHost represents a monitor host record in the database
type MonitorHost struct {
	Id       int64  `gorm:"primaryKey"`
	Target   string `gorm:"index"`
	Name     string
	Uploader string
	History  string // JSON string of ping history
}

func (MonitorHost) TableName() string {
	return MonitorTableName
}

// InitMonitorStore initializes the monitor store and creates tables
func InitMonitorStore(dbPath string) error {
	db := common.GetDB()
	if err := db.AutoMigrate(&MonitorHost{}); err != nil {
		return err
	}
	return nil
}

// AddHost adds a new monitor host
func AddHost(target, name, uploader string) (int64, error) {
	db := common.GetDB()
	host := MonitorHost{
		Target:   target,
		Name:     name,
		Uploader: uploader,
		History:  "[]",
	}
	result := db.Create(&host)
	if result.Error != nil {
		return 0, result.Error
	}
	return host.Id, nil
}

// RemoveHost removes a monitor host by ID
func RemoveHost(id int64, uploader string) error {
	db := common.GetDB()
	// Allow deletion if user is the uploader
	return db.Where("id = ? AND uploader = ?", id, uploader).Delete(&MonitorHost{}).Error
}

// RemoveHostAsAdmin removes a monitor host by ID as admin
func RemoveHostAsAdmin(id int64) error {
	db := common.GetDB()
	return db.Where("id = ?", id).Delete(&MonitorHost{}).Error
}

// ListHostsByUploader lists all hosts uploaded by a specific user
func ListHostsByUploader(uploader string) ([]MonitorHost, error) {
	db := common.GetDB()
	var hosts []MonitorHost
	err := db.Where("uploader = ?", uploader).Find(&hosts).Error
	return hosts, err
}

// ListAllHosts lists all monitor hosts
func ListAllHosts() ([]MonitorHost, error) {
	db := common.GetDB()
	var hosts []MonitorHost
	err := db.Find(&hosts).Error
	return hosts, err
}

// GetHost retrieves a single host by ID
func GetHost(id int64) (*MonitorHost, error) {
	db := common.GetDB()
	var host MonitorHost
	err := db.Where("id = ?", id).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &host, err
}

// UpdateHostHistory updates the history field for a host
func UpdateHostHistory(id int64, history string) error {
	db := common.GetDB()
	return db.Model(&MonitorHost{}).Where("id = ?", id).Update("history", history).Error
}

// GetAllTargets returns all target addresses for monitoring
func GetAllTargets() ([]string, error) {
	db := common.GetDB()
	var targets []string
	err := db.Model(&MonitorHost{}).Pluck("target", &targets).Error
	return targets, err
}

// HostToModel converts a MonitorHost database record to a model
func HostToModel(host MonitorHost) model.Host {
	return model.Host{
		Id:       host.Id,
		Target:   host.Target,
		Name:     host.Name,
		Uploader: host.Uploader,
	}
}
