package model

import "VPSBenchmarkBackend/internal/common"

type Host struct {
	Target   string `json:"target"`
	Name     string `json:"name"`
	Id       int64  `json:"id"`
	Uploader string `json:"uploader"`
}

type MonitorHost struct {
	Id           int64
	Target       string
	Name         string
	Uploader     string
	UploaderName string
	History      []float32
	ReviewStatus common.ReviewStatus
}
