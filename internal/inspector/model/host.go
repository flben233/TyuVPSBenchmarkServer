package model

import "VPSBenchmarkBackend/internal/common"

type InspectHost struct {
	ID     int64  `gorm:"primaryKey" json:"id"`
	UserID int64  `json:"user_id"`
	Target string `json:"target"`
	Name   string `json:"name"`
	Tags   string `json:"tags"` // json array of strings
	common.ServerStatus
}
