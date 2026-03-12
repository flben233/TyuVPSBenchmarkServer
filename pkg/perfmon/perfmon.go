package perfmon

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"time"
)

type ServerStatus struct {
	UptimeSeconds      uint64  `json:"uptime_seconds"`
	CpuUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryTotalBytes   uint64  `json:"memory_total_bytes"`
	MemoryUsedBytes    uint64  `json:"memory_used_bytes"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	UploadMbps         float64 `json:"upload_mbps"`
	DownloadMbps       float64 `json:"download_mbps"`
	System             string  `json:"system"`
}

type netSample struct {
	uploadMbps   float64
	downloadMbps float64
	err          error
}

const serverStatusSampleInterval = 250 * time.Millisecond

func CollectServerStatus(iface *string) (ServerStatus, error) {
	uptime, err := host.Uptime()
	if err != nil {
		return ServerStatus{}, err
	}

	memStats, err := mem.VirtualMemory()
	if err != nil {
		return ServerStatus{}, err
	}

	interval := serverStatusSampleInterval

	cpuCh := make(chan struct {
		percent float64
		err     error
	}, 1)
	go func() {
		percents, err := cpu.Percent(interval, false)
		if err != nil {
			cpuCh <- struct {
				percent float64
				err     error
			}{err: err}
			return
		}
		if len(percents) == 0 {
			cpuCh <- struct {
				percent float64
				err     error
			}{percent: 0, err: nil}
			return
		}
		cpuCh <- struct {
			percent float64
			err     error
		}{percent: percents[0], err: nil}
	}()

	netCh := make(chan netSample, 1)
	go func() {
		up, down, err := sampleNetMbps(interval, iface)
		netCh <- netSample{uploadMbps: up, downloadMbps: down, err: err}
	}()

	cpuRes := <-cpuCh
	if cpuRes.err != nil {
		return ServerStatus{}, cpuRes.err
	}
	netRes := <-netCh
	if netRes.err != nil {
		return ServerStatus{}, netRes.err
	}

	info, err := host.Info()
	if err != nil {
		return ServerStatus{}, err
	}

	return ServerStatus{
		UptimeSeconds:      uptime,
		CpuUsagePercent:    cpuRes.percent,
		MemoryTotalBytes:   memStats.Total,
		MemoryUsedBytes:    memStats.Used,
		MemoryUsagePercent: memStats.UsedPercent,
		UploadMbps:         netRes.uploadMbps,
		DownloadMbps:       netRes.downloadMbps,
		System:             info.PlatformVersion,
	}, nil
}

func FindNetIOCounter(iface *string) (*net.IOCountersStat, error) {
	counters, err := net.IOCounters(true)
	if err != nil || len(counters) == 0 {
		return nil, err
	}
	for i := 0; i < len(counters) && iface != nil; i++ {
		if counters[i].Name == *iface {
			return &counters[i], nil
		}
	}
	return &counters[0], nil
}

// sampleNetMbps samples network upload/download speed in Mbps
func sampleNetMbps(interval time.Duration, iface *string) (uploadMbps float64, downloadMbps float64, err error) {
	if interval <= 0 {
		interval = serverStatusSampleInterval
	}

	counter1, err := FindNetIOCounter(iface)
	if err != nil {
		return 0, 0, err
	}

	time.Sleep(interval)

	counter2, err := FindNetIOCounter(iface)
	if err != nil {
		return 0, 0, err
	}

	sent1 := counter1.BytesSent
	recv1 := counter1.BytesRecv
	sent2 := counter2.BytesSent
	recv2 := counter2.BytesRecv

	var sentDelta uint64
	if sent2 >= sent1 {
		sentDelta = sent2 - sent1
	}
	var recvDelta uint64
	if recv2 >= recv1 {
		recvDelta = recv2 - recv1
	}

	seconds := interval.Seconds()
	if seconds <= 0 {
		return 0, 0, nil
	}

	uploadMbps = (float64(sentDelta) * 8) / seconds / 1_000_000
	downloadMbps = (float64(recvDelta) * 8) / seconds / 1_000_000
	return uploadMbps, downloadMbps, nil
}
