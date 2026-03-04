package response

import "VPSBenchmarkBackend/internal/inspector/model"

type HostData struct {
	Traffic []model.TrafficPoint `json:"traffic"`
	Ping    []model.PingPoint    `json:"ping"`
	model.InspectHost
}
