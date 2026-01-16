package service

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/monitor/model"
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/internal/monitor/store"
	"log"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

func init() {
	interval := 60 * time.Second
	common.RegisterCronJob(interval, queryHosts)
}

// TODO: 也记录路由追踪历史
func queryHosts() {
	// Get all monitoring hosts
	hosts, err := store.ListAllHosts()
	if err != nil || len(hosts) == 0 {
		return
	}
	// Extract targets
	targets := make([]string, len(hosts))
	hostMap := make(map[string]*model.MonitorHost)
	for i, host := range hosts {
		targets[i] = host.Target
		hostMap[host.Target] = &hosts[i]
	}
	// Query hosts
	resultCh := make(chan *probing.Statistics, len(targets))
	for _, target := range targets {
		go func(t string) {
			pinger, err := probing.NewPinger(t)
			if err != nil {
				log.Printf("Failed to create pinger for %s: %v", t, err)
				resultCh <- nil
				return
			}
			pinger.SetPrivileged(true)
			pinger.Count = 1
			pinger.Timeout = 1000 * time.Millisecond
			err = pinger.Run()
			if err != nil {
				log.Printf("Failed to ping %s: %v", t, err)
				resultCh <- nil
				return
			}
			resultCh <- pinger.Statistics()
		}(target)
	}
	for i := 0; i < len(targets); i++ {
		stats := <-resultCh
		if stats != nil {
			host := hostMap[stats.Addr]
			history := append(host.History, float32(stats.AvgRtt.Milliseconds()))
			// Keep only last 100 values
			if len(history) > 100 {
				history = history[len(history)-100:]
			}
			err := store.UpdateHostHistory(host.Id, history)
			if err != nil {
				log.Printf("Failed to update history for %s: %v", stats.Addr, err)
			}
		}
	}
}

func GetStatistics() ([]response.StatisticsResponse, error) {
	// Get all monitoring hosts
	hosts, err := store.ListAllHosts()
	if err != nil {
		return nil, err
	}

	// Convert to response
	statistics := make([]response.StatisticsResponse, 0, len(hosts))
	for _, host := range hosts {
		statistics = append(statistics, response.StatisticsResponse{
			Name:     host.Name,
			Uploader: host.UploaderName,
			History:  host.History,
		})
	}

	return statistics, nil
}
