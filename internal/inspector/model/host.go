package model

import (
	"VPSBenchmarkBackend/pkg/perfmon"
	"time"
)

type InspectHost struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	UserID          int64     `json:"user_id"`
	Target          string    `json:"target"`
	MonitorType     string    `json:"monitor_type" gorm:"default:ping"`
	Name            string    `json:"name"`
	Tags            string    `json:"tags"` // json array of strings
	Notify          bool      `json:"notify"`
	NotifyTolerance int64     `json:"notify_tolerance"` // 0: 立刻通知, >0: 当异常值超过该值才通知
	LastUpdate      time.Time `json:"last_update"`
	perfmon.ServerStatus
}
