package service

import (
	"VPSBenchmarkBackend/internal/monitor/response"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	gnet "github.com/shirou/gopsutil/v4/net"
)

const serverStatusSampleInterval = 250 * time.Millisecond

type netSample struct {
	uploadMbps   float64
	downloadMbps float64
	err          error
}

func sampleNetMbps(interval time.Duration) (uploadMbps float64, downloadMbps float64, err error) {
	if interval <= 0 {
		interval = serverStatusSampleInterval
	}

	counters1, err := gnet.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}
	if len(counters1) == 0 {
		return 0, 0, nil
	}

	time.Sleep(interval)

	counters2, err := gnet.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}
	if len(counters2) == 0 {
		return 0, 0, nil
	}

	sent1 := counters1[0].BytesSent
	recv1 := counters1[0].BytesRecv
	sent2 := counters2[0].BytesSent
	recv2 := counters2[0].BytesRecv

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

func GetServerStatus() (response.ServerStatusResponse, error) {
	uptime, err := host.Uptime()
	if err != nil {
		return response.ServerStatusResponse{}, err
	}

	memStats, err := mem.VirtualMemory()
	if err != nil {
		return response.ServerStatusResponse{}, err
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
		up, down, err := sampleNetMbps(interval)
		netCh <- netSample{uploadMbps: up, downloadMbps: down, err: err}
	}()

	cpuRes := <-cpuCh
	if cpuRes.err != nil {
		return response.ServerStatusResponse{}, cpuRes.err
	}
	netRes := <-netCh
	if netRes.err != nil {
		return response.ServerStatusResponse{}, netRes.err
	}

	return response.ServerStatusResponse{
		UptimeSeconds:      uptime,
		CpuUsagePercent:    cpuRes.percent,
		MemoryTotalBytes:   memStats.Total,
		MemoryUsedBytes:    memStats.Used,
		MemoryUsagePercent: memStats.UsedPercent,
		UploadMbps:         netRes.uploadMbps,
		DownloadMbps:       netRes.downloadMbps,
	}, nil
}
