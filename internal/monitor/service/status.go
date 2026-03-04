package service

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/monitor/response"
	"sync"
	"time"
)

const serverStatusSampleInterval = 250 * time.Millisecond

var (
	cachedServerStatus response.ServerStatusResponse
	statusMutex        sync.RWMutex
)

// init registers the background job to update server status cache periodically
func init() {
	common.RegisterCronJob(serverStatusSampleInterval, updateServerStatusCache)
}

// updateServerStatusCache collects server status and updates the cache
func updateServerStatusCache() {
	status, err := common.CollectServerStatus()
	if err == nil {
		statusMutex.Lock()
		cachedServerStatus = response.ServerStatusResponse{
			UptimeSeconds:      status.UptimeSeconds,
			CpuUsagePercent:    status.CpuUsagePercent,
			MemoryTotalBytes:   status.MemoryTotalBytes,
			MemoryUsedBytes:    status.MemoryUsedBytes,
			MemoryUsagePercent: status.MemoryUsagePercent,
			UploadMbps:         status.UploadMbps,
			DownloadMbps:       status.DownloadMbps,
		}
		statusMutex.Unlock()
	}
}

// GetServerStatus returns the cached server status
func GetServerStatus() (response.ServerStatusResponse, error) {
	statusMutex.RLock()
	defer statusMutex.RUnlock()
	return cachedServerStatus, nil
}
