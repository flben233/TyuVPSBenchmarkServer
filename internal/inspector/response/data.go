package response

import (
	"VPSBenchmarkBackend/pkg/perfmon"
	"time"
)

type HostData struct {
	Sent            float64         `json:"sent"`
	Recv            float64         `json:"recv"`
	Ping            []PingPointData `json:"ping"`
	Loss            float64         `json:"loss"`
	LatestPing      float32         `json:"latest_ping"`
	ID              string          `json:"id"`
	Target          string          `json:"target"`
	Name            string          `json:"name"`
	Tags            string          `json:"tags"` // json array of strings
	Notify          bool            `json:"notify"`
	NotifyTolerance int64           `json:"notify_tolerance"`
	LastUpdate      time.Time       `json:"last_update"`
	perfmon.ServerStatus
}

type HostListResponse struct {
	ID              string    `json:"id"`
	UserID          int64     `json:"user_id"`
	Target          string    `json:"target"`
	Name            string    `json:"name"`
	Tags            string    `json:"tags"` // json array of strings
	Notify          bool      `json:"notify"`
	NotifyTolerance int64     `json:"notify_tolerance"`
	LastUpdate      time.Time `json:"last_update"`
	perfmon.ServerStatus
}

type SettingData struct {
	NotifyURL      *string  `json:"notify_url"`
	BgURL          *string  `json:"bg_url"`
	VisitorEnabled bool     `json:"visitor_enabled"`
	AllowedHostIDs []string `json:"allowed_host_ids"`
}

type VisitorHostData struct {
	Sent       float64         `json:"sent"`
	Recv       float64         `json:"recv"`
	Ping       []PingPointData `json:"ping"`
	Loss       float64         `json:"loss"`
	LatestPing float32         `json:"latest_ping"`
	Name       string          `json:"name"`
	Tags       string          `json:"tags"`
	LastUpdate time.Time       `json:"last_update"`
	perfmon.ServerStatus
}

type VisitorPageData struct {
	OwnerName string            `json:"owner_name"`
	OwnerID   string            `json:"owner_id"`
	BgURL     *string           `json:"bg_url"`
	Hosts     []VisitorHostData `json:"hosts"`
}

type PingPointData struct {
	Latency float32   `json:"latency"` // 单位ms，为0表示丢包
	Time    time.Time `json:"time"`    // 纳秒
}
