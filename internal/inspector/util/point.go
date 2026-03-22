package util

import (
	"VPSBenchmarkBackend/internal/inspector/model"
	"VPSBenchmarkBackend/internal/inspector/response"
)

func ConvertToPointVO(points []model.PingPoint) []response.PingPointData {
	pingPoints := make([]response.PingPointData, len(points))
	for i, p := range points {
		pingPoints[i] = response.PingPointData{
			Latency: p.Latency,
			Time:    p.Time,
		}
	}
	return pingPoints
}
