package request

import (
	"VPSBenchmarkBackend/internal/inspector/model"
	"VPSBenchmarkBackend/pkg/perfmon"
)

type CreateHostRequest struct {
	Target               string  `json:"target" binding:"required"`
	MonitorType          string  `json:"monitor_type"`
	Name                 string  `json:"name" binding:"required"`
	Tags                 string  `json:"tags"`
	Notify               bool    `json:"notify"`
	NotifyTolerance      int64   `json:"notify_tolerance"`
	TrafficSettlementDay int     `json:"traffic_settlement_day"`
	MonthlyTrafficLimit  float64 `json:"monthly_traffic_limit"`
}

type UpdateHostRequest struct {
	Name                 string  `json:"name"`
	Tags                 string  `json:"tags"`
	Target               string  `json:"target"`
	MonitorType          string  `json:"monitor_type"`
	Notify               bool    `json:"notify"`
	NotifyTolerance      int64   `json:"notify_tolerance"`
	TrafficSettlementDay int     `json:"traffic_settlement_day"`
	MonthlyTrafficLimit  float64 `json:"monthly_traffic_limit"`
}

type PutDataRequest struct {
	HostID   int64 `json:"host_id" binding:"required"`
	HostInfo perfmon.ServerStatus
	Traffic  []model.TrafficPoint `json:"traffic"`
}

type QueryDataRequest struct {
	Start    int64  `form:"start" binding:"required"`
	End      int64  `form:"end" binding:"required"`
	Interval string `form:"interval" binding:"required"`
}

type UpdateInspectorSettingRequest struct {
	NotifyURL      *string  `json:"notify_url"`
	BgURL          *string  `json:"bg_url"`
	VisitorEnabled bool     `json:"visitor_enabled"`
	AllowedHostIDs []string `json:"allowed_host_ids"`
}

type TestNotifyRequest struct {
	NotifyURL string `json:"notify_url" binding:"required"`
}

type VisitorPageRequest struct {
	Start    int64  `form:"start" binding:"required"`
	End      int64  `form:"end" binding:"required"`
	Interval string `form:"interval" binding:"required"`
}
