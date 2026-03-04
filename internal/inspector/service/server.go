package service

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/inspector/model"
	"VPSBenchmarkBackend/internal/inspector/response"
	"VPSBenchmarkBackend/internal/inspector/store"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var putRecord = make(map[int64]time.Time)

const (
	putInterval  = 30 * time.Second
	putMaxLength = 1
)

func CreateHost(userID int64, target, name, tags string) (int64, error) {
	host := &model.InspectHost{
		UserID: userID,
		Target: target,
		Name:   name,
		Tags:   tags,
	}
	if err := store.CreateHost(host); err != nil {
		return 0, err
	}
	return host.ID, nil
}

func UpdateHost(userID int64, hostID int64, name, tags, target string) error {
	// 校验主机属于当前用户
	ids := store.GetHostIDByUser(userID)
	found := false
	for _, id := range ids {
		if id == hostID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("host %d not found or not owned by user", hostID)
	}

	host, err := store.GetHostByID(hostID)
	if err != nil {
		return fmt.Errorf("failed to get host by ID %d: %w", hostID, err)
	}
	host.Name = name
	host.Tags = tags

	store.UpdateHost(host)
	return nil
}

func DeleteHost(userID int64, hostID int64) error {
	// 校验主机属于当前用户
	ids := store.GetHostIDByUser(userID)
	found := false
	for _, id := range ids {
		if id == hostID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("host %d not found or not owned by user", hostID)
	}
	return store.DeleteHost(hostID)
}

func ListHosts(userID int64) ([]model.InspectHost, error) {
	return store.ListHostsByUser(userID)
}

func PutData(userID int64, trafficData []model.TrafficPoint, hostInfo common.ServerStatus, hostID int64) error {
	// 校验数据，理论上ID都是一样的
	ids := store.GetHostIDByUser(userID)
	idSet := make(map[int64]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	for _, point := range trafficData {
		if _, ok := idSet[point.HostID]; !ok {
			return fmt.Errorf("invalid host ID %d in traffic data", point.HostID)
		}
	}
	if len(trafficData) > putMaxLength {
		return fmt.Errorf("too many traffic points, max length is %d", putMaxLength)
	}
	if lastPut, ok := putRecord[hostID]; ok && time.Since(lastPut) < putInterval {
		return fmt.Errorf("put data too frequently, please wait for %v", putInterval-time.Since(lastPut))
	}
	putRecord[hostID] = time.Now()

	host, err := store.GetHostByID(hostID)
	if err != nil {
		return fmt.Errorf("failed to get host by ID %d: %w", hostID, err)
	}
	lat, err := queryHost(host.Target)
	if err != nil {
		return fmt.Errorf("failed to query host %s: %w", host.Target, err)
	}
	pingData := []model.PingPoint{{
		HostID:  hostID,
		Latency: lat,
		Time:    trafficData[0].Time,
	}}

	host = &model.InspectHost{
		ID:           host.ID,
		UserID:       host.UserID,
		Target:       host.Target,
		Name:         host.Name,
		Tags:         host.Tags,
		ServerStatus: hostInfo,
	}

	store.UpdateHost(host)

	// 保存数据
	if err := store.SaveTrafficPoints(trafficData); err != nil {
		return err
	}
	if err := store.SavePingPoints(pingData); err != nil {
		return err
	}
	return nil
}

func QueryData(userID int64, start, end int64, interval string) ([]*response.HostData, error) {
	hosts, err := store.ListHostsByUser(userID)
	if err != nil {
		return nil, err
	}
	data := make([]*response.HostData, len(hosts))
	for i, host := range hosts {
		pingPoints, err := store.QueryPingPoints(host.ID, start, end, interval)
		if err != nil {
			return nil, err
		}
		recv, sent, err := store.QueryTrafficSum(host.ID, start, end, interval)
		if err != nil {
			return nil, err
		}
		data[i] = &response.HostData{
			Ping:        pingPoints,
			Sent:        sent,
			Recv:        recv,
			InspectHost: host,
		}
	}
	return data, nil
}

func queryHost(target string) (float32, error) {
	req, err := json.Marshal([]string{target})
	if err != nil {
		log.Printf("Failed to marshal targets: %v", err)
		return 0, err
	}
	resp, err := http.Post(config.Get().ExporterURL+"/monitor", "application/json", bytes.NewReader(req))
	if err != nil {
		log.Printf("Failed to get exporter data: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read exporter response body: %v", err)
		return 0, err
	}
	var data map[string]float32
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Printf("Failed to unmarshal exporter response: %v", err)
		return 0, err
	}
	return data[target], nil
}
