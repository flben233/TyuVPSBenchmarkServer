package store

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/monitor/model"
	"context"

	"gorm.io/gorm"
)

// Table name for monitor hosts
const MonitorTableName = "monitor_hosts"

// MonitorHost represents a monitor host record in the database
type MonitorHost struct {
	Id           int64  `gorm:"primaryKey"`
	Target       string `gorm:"index"`
	Name         string
	Uploader     string
	UploaderName string
	History      string // JSON string of ping history
}

func (MonitorHost) TableName() string {
	return MonitorTableName
}

var db *gorm.DB
var monitorHosts gorm.Interface[MonitorHost]

func init() {
	// Register the initializer
	common.RegisterDBInitializer(InitMonitorStore)
}

// InitMonitorStore initializes the monitor store and creates tables
func InitMonitorStore(dbPath string) error {
	db = common.GetDB()
	monitorHosts = gorm.G[MonitorHost](db)
	if err := db.AutoMigrate(&MonitorHost{}); err != nil {
		return err
	}
	return nil
}

func CountUserHosts(userID string) (int64, error) {
	cnt, err := monitorHosts.Where("user_id = ?", userID).Count(context.Background(), "*")
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

// AddHost adds a new monitor host
func AddHost(target, name, username, userID string) (int64, error) {
	host := MonitorHost{
		Target:       target,
		Name:         name,
		Uploader:     userID,
		UploaderName: username,
		History:      "[]",
	}
	err := monitorHosts.Create(context.Background(), &host)
	if err != nil {
		return 0, err
	}
	return host.Id, nil
}

// RemoveHost removes a monitor host by ID
func RemoveHost(id int64, uploader string) error {
	affected, err := monitorHosts.Where("id = ? AND uploader = ?", id, uploader).Delete(context.Background())
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// RemoveHostAsAdmin removes a monitor host by ID as admin
func RemoveHostAsAdmin(id int64) error {
	affected, err := monitorHosts.Where("id = ?", id).Delete(context.Background())
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListHostsByUploader lists all hosts uploaded by a specific user
func ListHostsByUploader(uploader string) ([]MonitorHost, error) {
	hosts, err := monitorHosts.Where("uploader = ?", uploader).Find(context.Background())
	return hosts, err
}

// ListAllHosts lists all monitor hosts
func ListAllHosts() ([]MonitorHost, error) {
	hosts, err := monitorHosts.Find(context.Background())
	return hosts, err
}

// GetHost retrieves a single host by ID
func GetHost(id int64) (*MonitorHost, error) {
	hosts, err := monitorHosts.Where("id = ?", id).Find(context.Background())
	if err != nil {
		return nil, err
	}
	if len(hosts) == 0 {
		return nil, nil
	}
	return &hosts[0], nil
}

// UpdateHostHistory updates the history field for a host
func UpdateHostHistory(id int64, history string) error {
	affected, err := monitorHosts.Where("id = ?", id).Update(context.Background(), "history", history)
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetAllTargets returns all target addresses for monitoring
func GetAllTargets() ([]string, error) {
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
