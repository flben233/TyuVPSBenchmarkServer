package service

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/pkg/perfmon"
	"os"
	"sync"
	"time"
)

const serverStatusSampleInterval = 250 * time.Millisecond
const serverStatusIfaceEnv = "SERVER_STATUS_IFACE"

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
	env, ok := os.LookupEnv(serverStatusIfaceEnv)
	var iface *string
	if ok {
		iface = &env
	}
	status, err := perfmon.CollectServerStatus(iface)
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
