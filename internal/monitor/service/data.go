package service

import (
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/internal/monitor/store"
	"encoding/json"
	"log"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// 本文件可以查看过去一段时间内的监控数据，

func queryHosts(targets []string) map[string]float32 {
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
	results := make(map[string]float32)
	for i := 0; i < len(targets); i++ {
		stats := <-resultCh
		if stats != nil {
			results[stats.Addr] = float32(stats.AvgRtt.Milliseconds()) / 1000.0
		}
	}
	return results
}

func GetStatistics() ([]response.StatisticsResponse, error) {
	// Get all monitoring hosts
	hosts, err := store.ListAllHosts()
	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return []response.StatisticsResponse{}, nil
	}

	// Extract targets
	targets := make([]string, len(hosts))
	hostMap := make(map[string]store.MonitorHost)
	for i, host := range hosts {
		targets[i] = host.Target
		hostMap[host.Target] = host
	}

	// Query hosts
	results := queryHosts(targets)

	// Convert to response
	statistics := make([]response.StatisticsResponse, 0, len(hosts))
	for target, avgRtt := range results {
		if host, ok := hostMap[target]; ok {
			var history []float32
			if host.History != "" && host.History != "[]" {
				err := json.Unmarshal([]byte(host.History), &history)
				if err != nil {
					log.Printf("Failed to unmarshal history for %s: %v", target, err)
					history = []float32{}
				}
			}

			// Add new value to history
			history = append(history, avgRtt)

			// Keep only last 100 values
			if len(history) > 100 {
				history = history[len(history)-100:]
			}

			// Update history in database
			historyJson, err := json.Marshal(history)
			if err == nil {
				store.UpdateHostHistory(host.Id, string(historyJson))
			}

			statistics = append(statistics, response.StatisticsResponse{
				Name:     host.Name,
				Uploader: host.Uploader,
				History:  history,
			})
		}
	}

	return statistics, nil
}
