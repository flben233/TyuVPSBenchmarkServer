package model

// TrafficPoint 流量消耗记录，理论上30分钟或者一个小时上报一次
type TrafficPoint struct {
	HostID int64   `lp:"tag,host_id"` // 主机ID，首先在server端生成，然后在agent安装时填写，agent上报时携带这个ID
	Recv   float32 `lp:"field,recv"`  // 单位MB
	Sent   float32 `lp:"field,sent"`  // 单位MB
	Time   int64   `lp:"field,time"`  // 纳秒
}

type PingPoint struct {
	HostID  int64   `lp:"tag,host_id"`
	Latency float32 `lp:"field,latency"` // 单位ms，为0表示丢包
	Time    int64   `lp:"field,time"`    // 纳秒
}
