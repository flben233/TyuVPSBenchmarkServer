package model

import "VPSBenchmarkBackend/internal/common"

type LookingGlass struct {
	ServerName   string              `json:"server_name"`
	TestURL      string              `json:"test_url"`
	Id           int64               `json:"id"`
	Uploader     string              `json:"uploader"`
	ReviewStatus common.ReviewStatus `json:"review_status"`
}
