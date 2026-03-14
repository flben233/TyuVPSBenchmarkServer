package request

import (
	"VPSBenchmarkBackend/internal/inspector/model"
	"VPSBenchmarkBackend/pkg/perfmon"
)

type CreateHostRequest struct {
	Target          string `json:"target" binding:"required"`
	Name            string `json:"name" binding:"required"`
	Tags            string `json:"tags"`
	Notify          bool   `json:"notify"`
	NotifyTolerance int64  `json:"notify_tolerance"`
}

type UpdateHostRequest struct {
	Name            string `json:"name"`
	Tags            string `json:"tags"`
	Target          string `json:"target"`
	Notify          bool   `json:"notify"`
	NotifyTolerance int64  `json:"notify_tolerance"`
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
	NotifyURL *string `json:"notify_url"`
	BgURL     *string `json:"bg_url"`
}

type TestNotifyRequest struct {
	NotifyURL string `json:"notify_url" binding:"required"`
}
