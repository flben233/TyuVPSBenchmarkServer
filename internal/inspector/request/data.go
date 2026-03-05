package request

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/inspector/model"
)

type CreateHostRequest struct {
	Target string `json:"target" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Tags   string `json:"tags"`
	Notify bool   `json:"notify" binding:"required"`
}

type UpdateHostRequest struct {
	Name   string `json:"name"`
	Tags   string `json:"tags"`
	Target string `json:"target"`
}

type PutDataRequest struct {
	HostID   int64 `json:"host_id" binding:"required"`
	HostInfo common.ServerStatus
	Traffic  []model.TrafficPoint `json:"traffic"`
}

type QueryDataRequest struct {
	Start    int64  `form:"start" binding:"required"`
	End      int64  `form:"end" binding:"required"`
	Interval string `form:"interval" binding:"required"`
}
