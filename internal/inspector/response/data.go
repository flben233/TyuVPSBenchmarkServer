package response

import "VPSBenchmarkBackend/internal/inspector/model"

type HostData struct {
	Sent float32           `json:"sent"`
	Recv float32           `json:"recv"`
	Ping []model.PingPoint `json:"ping"`
	model.InspectHost
}

type SettingData struct {
	NotifyURL *string `json:"notify_url"`
	BgURL     *string `json:"bg_url"`
}
