package model

import "time"

// TrafficPoint 流量消耗记录，理论上30分钟或者一个小时上报一次
type TrafficPoint struct {
	HostID int64     `json:"host_id"` // 主机ID，首先在server端生成，然后在agent安装时填写，agent上报时携带这个ID
	Recv   float32   `json:"recv"`    // 单位MB
	Sent   float32   `json:"sent"`    // 单位MB
	Time   time.Time `json:"time"`    // 纳秒
}

type PingPoint struct {
	HostID  int64     `json:"host_id"`
	Latency float32   `json:"latency"` // 单位ms，为0表示丢包
	Time    time.Time `json:"time"`    // 纳秒
}
