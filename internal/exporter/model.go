package exporter

import "time"

const (
	PingRoute    = "exporter_ping"
	TracertRoute = "exporter_tracert"
	ProbePing    = "ping"
	ProbeTCP     = "tcp"
	ProbeHTTP    = "http"
)

type PingResp struct {
	HostID      int64
	Lat         float32
	Time        time.Time
	MonitorType string
}

type PingReq struct {
	HostID      int64
	Target      string
	MonitorType string
}

type TracertReq struct {
	Mode   string // "tcp" or "icmp"
	Target string
	Port   uint64 // Only for TCP mode
}

type TracertResp struct {
	Result map[string]interface{}
}
