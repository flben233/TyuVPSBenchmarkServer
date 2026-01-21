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
type monitorHost struct {
	Id           int64  `gorm:"primaryKey"`
	Target       string `gorm:"index"`
	Name         string
	Uploader     string
	UploaderName string
	History      common.JSONField[[]float32]
	ReviewStatus common.ReviewStatus `gorm:"default:0;index"`
}

func (monitorHost) TableName() string {
	return MonitorTableName
}

var db *gorm.DB
var monitorHosts gorm.Interface[monitorHost]

func init() {
	// Register the initializer
	common.RegisterDBInitializer(InitMonitorStore)
}

// InitMonitorStore initializes the monitor store and creates tables
func InitMonitorStore(dbPath string) error {
	db = common.GetDB()
	monitorHosts = gorm.G[monitorHost](db)
	if err := db.AutoMigrate(&monitorHost{}); err != nil {
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

func toMonitorHost(host monitorHost) model.MonitorHost {
	return model.MonitorHost{
		Id:           host.Id,
		Target:       host.Target,
		Name:         host.Name,
		Uploader:     host.Uploader,
		UploaderName: host.UploaderName,
		History:      *host.History.GetValue(),
		ReviewStatus: host.ReviewStatus,
	}
}

// AddHost adds a new monitor host
func AddHost(target, name, username, userID string) (int64, error) {
	host := monitorHost{
		Target:       target,
		Name:         name,
		Uploader:     userID,
		UploaderName: username,
		History:      *common.NewJSONField([]float32{}),
		ReviewStatus: common.ReviewStatusPending,
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
func ListHostsByUploader(uploader string) ([]model.MonitorHost, error) {
	hosts, err := monitorHosts.Where("uploader = ?", uploader).Find(context.Background())
	hostModels := make([]model.MonitorHost, len(hosts))
	for i, host := range hosts {
		hostModels[i] = toMonitorHost(host)
	}
	return hostModels, err
}

// ListAllHosts lists all approved monitor hosts
func ListAllHosts() ([]model.MonitorHost, error) {
	hosts, err := monitorHosts.Where("review_status = ?", common.ReviewStatusApproved).Order("name desc").Find(context.Background())
	hostModels := make([]model.MonitorHost, len(hosts))
	for i, host := range hosts {
		hostModels[i] = toMonitorHost(host)
	}
	return hostModels, err
}

// GetHost retrieves a single host by ID
func GetHost(id int64) (*model.MonitorHost, error) {
	hosts, err := monitorHosts.Where("id = ?", id).Find(context.Background())
	if err != nil {
		return nil, err
	}
	if len(hosts) == 0 {
		return nil, nil
	}
	hostModel := toMonitorHost(hosts[0])
	return &hostModel, nil
}

// UpdateHostHistory updates the history field for a host
func UpdateHostHistory(id int64, history []float32) error {
	affected, err := monitorHosts.Where("id = ?", id).Update(context.Background(), "history", common.NewJSONField(history))
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetAllTargets returns all approved target addresses for monitoring
func GetAllTargets() ([]string, error) {
	var targets []string
	err := db.Model(&monitorHost{}).Where("review_status = ?", common.ReviewStatusApproved).Pluck("target", &targets).Error
	return targets, err
}

// ListPendingHosts lists all hosts awaiting review
func ListPendingHosts() ([]model.MonitorHost, error) {
	hosts, err := monitorHosts.Where("review_status = ?", common.ReviewStatusPending).Find(context.Background())
	hostModels := make([]model.MonitorHost, len(hosts))
	for i, host := range hosts {
		hostModels[i] = toMonitorHost(host)
	}
	return hostModels, err
}

// UpdateReviewStatus updates the review status of a host
func UpdateReviewStatus(id int64, status common.ReviewStatus) error {
	affected, err := monitorHosts.Where("id = ?", id).Update(context.Background(), "review_status", status)
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
