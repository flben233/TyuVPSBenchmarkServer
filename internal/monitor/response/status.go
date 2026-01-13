package response

type ServerStatusResponse struct {
	UptimeSeconds      uint64  `json:"uptime_seconds"`
	CpuUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryTotalBytes   uint64  `json:"memory_total_bytes"`
	MemoryUsedBytes    uint64  `json:"memory_used_bytes"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	UploadMbps         float64 `json:"upload_mbps"`
	DownloadMbps       float64 `json:"download_mbps"`
}
