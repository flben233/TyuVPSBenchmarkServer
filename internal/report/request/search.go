package request

// nil表示不限制
type SearchRequest struct {
	Keyword        *string             `json:"name"`          // 关键词，模糊匹配
	MediaUnlocks   []string            `json:"media_unlocks"` // 媒体解锁名称列表，TIKTOK需要特殊处理
	CTParams       *ASNSpecificRequest `json:"ct_params"`
	CUParams       *ASNSpecificRequest `json:"cu_params"`
	CMParams       *ASNSpecificRequest `json:"cm_params"`
	Virtualization *string             `json:"virtualization"` // 虚拟化技术名称
	IPv6Support    *bool               `json:"ipv6_support"`   // 是否支持IPv6，
	DiskLevel      *int                `json:"disk_level"`     // 磁盘等级，0-5对应读写平均100MB/s以下到1000MB/s以上
}

type ASNSpecificRequest struct {
	BackRoute   *string  `json:"back_route"`   // 回程线路名称
	MinDownload *float32 `json:"min_download"` // 单位 Mbps
	MaxDownload *float32 `json:"max_download"`
	MinUpload   *float32 `json:"min_upload"`
	MaxUpload   *float32 `json:"max_upload"`
	Latency     *float32 `json:"latency"` // 单位 ms
}
