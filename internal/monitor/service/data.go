package service

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/monitor/model"
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/internal/monitor/store"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func init() {
	interval := 120 * time.Second
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
	req, err := json.Marshal(targets)
	if err != nil {
		log.Printf("Failed to marshal targets: %v", err)
		return
	}
	resp, err := http.Post(config.Get().ExporterURL+"/monitor", "application/json", bytes.NewReader(req))
	if err != nil {
		log.Printf("Failed to get exporter data: %v", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read exporter response body: %v", err)
		return
	}
	var data map[string]float32
	err = json.Unmarshal(bodyBytes, &data)
	for addr, rtt := range data {
		host := hostMap[addr]
		history := append(host.History, rtt)
		// Keep only last 720 values
		if len(history) > 720 {
			history = history[len(history)-720:]
		}
		err := store.UpdateHostHistory(host.Id, history)
		if err != nil {
			log.Printf("Failed to update history for %s: %v", addr, err)
		}
	}
}

func GetStatistics(id string) ([]response.StatisticsResponse, error) {
	var hosts []model.MonitorHost
	var err error
	if id != "" {
		idNumber, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}
		host, err := store.GetHost(idNumber)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, *host)
	} else {
		// Get all monitoring hosts
		hosts, err = store.ListAllHosts()
		if err != nil {
			return nil, err
		}
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
