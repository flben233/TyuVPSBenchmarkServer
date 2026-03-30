package exporter

import "time"

const (
	PingSentTopic    = "exporter_ping_sent"
	PingRecvTopic    = "exporter_ping_recv"
	TracertSentTopic = "exporter_tracert_sent"
	TracertRecvTopic = "exporter_tracert_recv"
)

type PingResp struct {
	HostID int64
	Lat    float32
	Time   time.Time
}

type PingReq struct {
	HostID int64
	Target string
}

type TracertReq struct {
	Mode   string // "tcp" or "icmp"
	Target string
	Port   uint64 // Only for TCP mode
}

type TracertResp struct {
	Result map[string]interface{}
}
