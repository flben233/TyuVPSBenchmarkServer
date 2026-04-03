package service

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/exporter"
	"VPSBenchmarkBackend/internal/monitor/model"
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/internal/monitor/store"
	"VPSBenchmarkBackend/internal/mq"
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
	"time"
)

const pingSource = "monitor_ping"

func init() {
	interval := 120 * time.Second
	common.RegisterCronJob(interval, queryHosts)
	mq.LateSubscribe(pingSource, context.Background(), handlePingResult)
}

// TODO: 也记录路由追踪历史
func queryHosts() {
	// Get all monitoring hosts
	hosts, err := store.ListAllHosts()
	if err != nil || len(hosts) == 0 {
		return
	}
	for _, host := range hosts {
		err = mq.PublishJSON(exporter.PingRoute, pingSource, exporter.PingReq{
			HostID:      host.Id,
			Target:      host.Target,
			MonitorType: exporter.ProbePing,
		})
		if err != nil {
			log.Printf("Failed to enqueue monitor ping for host %d: %v", host.Id, err)
		}
	}
}

func handlePingResult(msg *amqp.Delivery) error {
	var resp exporter.PingResp
	if err := json.Unmarshal(msg.Body, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal monitor ping message: %w", err)
	}
	host, err := store.GetHost(resp.HostID)
	if err != nil {
		return err
	}
	if host == nil {
		return nil
	}
	history := append(host.History, resp.Lat)
	if len(history) > 720 {
		history = history[len(history)-720:]
	}
	if err := store.UpdateHostHistory(host.Id, history); err != nil {
		return fmt.Errorf("failed to update history for host %d: %w", host.Id, err)
	}
	return nil
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
