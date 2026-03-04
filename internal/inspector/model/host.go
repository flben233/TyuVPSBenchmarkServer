package model

type InspectHost struct {
	ID             int64   `gorm:"primaryKey" json:"id"`
	UserID         int64   `json:"user_id"`
	Target         string  `json:"target"`
	Name           string  `json:"name"`
	Tags           string  `json:"tags"`            // json array of strings
	ProcessorUsage float32 `json:"processor_usage"` // late init
	MemoryUsage    float32 `json:"memory_usage"`    // late init
	DiskUsage      float32 `json:"disk_usage"`      // late init
	System         string  `json:"system"`          // late init
	Uptime         int64   `json:"uptime"`          // late init, seconds
	Upload         int64   `json:"upload"`          // late init, bytes
	Download       int64   `json:"download"`        // late init, bytes
}
