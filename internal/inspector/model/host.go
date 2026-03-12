package model

import (
	"VPSBenchmarkBackend/pkg/perfmon"
	"time"
)

type InspectHost struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	UserID     int64     `json:"user_id"`
	Target     string    `json:"target"`
	Name       string    `json:"name"`
	Tags       string    `json:"tags"` // json array of strings
	Notify     bool      `json:"notify"`
	LastUpdate time.Time `json:"last_update"`
	perfmon.ServerStatus
}
