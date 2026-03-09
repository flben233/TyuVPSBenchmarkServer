package response

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/inspector/model"
)

type HostData struct {
	Sent       float64           `json:"sent"`
	Recv       float64           `json:"recv"`
	Ping       []model.PingPoint `json:"ping"`
	LatestPing float32           `json:"latest_ping"`
	ID         string            `json:"id"`
	Target     string            `json:"target"`
	Name       string            `json:"name"`
	Tags       string            `json:"tags"` // json array of strings
	Notify     bool              `json:"notify"`
	common.ServerStatus
}

type HostListResponse struct {
	ID     string `json:"id"`
	UserID int64  `json:"user_id"`
	Target string `json:"target"`
	Name   string `json:"name"`
	Tags   string `json:"tags"` // json array of strings
	Notify bool   `json:"notify"`
	common.ServerStatus
}

type SettingData struct {
	NotifyURL *string `json:"notify_url"`
	BgURL     *string `json:"bg_url"`
}
