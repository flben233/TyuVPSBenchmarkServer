package service

import (
	"VPSBenchmarkBackend/internal/inspector/model"
	"VPSBenchmarkBackend/internal/inspector/response"
	"VPSBenchmarkBackend/internal/inspector/store"
	"fmt"
)

func PutData(userID int64, trafficData []model.TrafficPoint, pingData []model.PingPoint) error {
	// 校验数据
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
	for _, point := range pingData {
		if _, ok := idSet[point.HostID]; !ok {
			return fmt.Errorf("invalid host ID %d in ping data", point.HostID)
		}
	}

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
		trafficPoints, err := store.QueryTrafficPoints(host.ID, start, end, interval)
		if err != nil {
			return nil, err
		}
		data[i] = &response.HostData{
			Ping:        pingPoints,
			Traffic:     trafficPoints,
			InspectHost: host,
		}
	}
	return data, nil
}
