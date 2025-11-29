package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type BestTraceResult struct {
	Region string `json:"region"`
	Route  string `json:"route"`
}

type CPUResult struct {
	Single int32 `json:"single"`
	Multi  int32 `json:"multi"`
}

type MemResult struct {
	Read  float32 `json:"read"`
	Write float32 `json:"write"`
}

type DiskResult struct {
	SeqRead  string `json:"seq_read"`
	SeqWrite string `json:"seq_write"`
}

type TraceResult struct {
	Types     map[string]string `json:"types"`
	BackRoute string            `json:"back_route"`
}

type ECSResult struct {
	Info      map[string]string `json:"info"`
	Cpu       CPUResult         `json:"cpu"`
	Mem       MemResult         `json:"mem"`
	Disk      DiskResult        `json:"disk"`
	Tiktok    string            `json:"tiktok"`
	IPQuality string            `json:"ip_quality"`
	Mail      map[string][]bool `json:"mail"` // bool数组定义：SMTP SMTPS POP3 POP3S IMAP MAPS
	Trace     TraceResult       `json:"trace"`
	Time      string            `json:"time"`
}

type ItdogResult struct {
	Ping  string   `json:"ping"`
	Route []string `json:"route"`
}

// BenchmarkResult is the main model stored in database
type BenchmarkResult struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	ReportID  string    `gorm:"uniqueIndex;size:255" json:"id"`
	Title     string    `gorm:"size:500" json:"title"`
	Time      string    `gorm:"size:100" json:"time"`
	Link      string    `gorm:"size:1000" json:"link"`
	RawHTML   string    `gorm:"type:text" json:"-"`
	SpdTest   JSONField `gorm:"type:text" json:"spdtest"`
	ECS       JSONField `gorm:"type:text" json:"ecs"`
	Media     JSONField `gorm:"type:text" json:"media"`
	BestTrace JSONField `gorm:"type:text" json:"besttrace"`
	Itdog     JSONField `gorm:"type:text" json:"itdog"`
	Disk      JSONField `gorm:"type:text" json:"disk"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SpeedtestResult struct {
	Spot     string  `json:"spot"`
	Download float32 `json:"download"`
	Upload   float32 `json:"upload"`
	Latency  float32 `json:"latency"`
	Jitter   float32 `json:"jitter"`
}

type SpeedtestResults struct {
	Results []SpeedtestResult `json:"results"`
	Time    string            `json:"time"`
}

type MediaPair struct {
	Media  string `json:"media"`
	Unlock string `json:"unlock"`
}

type MediaBlock struct {
	Region  string      `json:"region"`
	Results []MediaPair `json:"results"`
}

type MediaResults struct {
	IPv4 []MediaBlock `json:"ipv4"`
	IPv6 []MediaBlock `json:"ipv6"`
}

type TyuDiskResult struct {
	Data [][]string `json:"data"`
	Time string     `json:"time"`
}

// JSONField is a custom type for storing JSON data in SQLite
type JSONField struct {
	Data interface{}
}

// Scan implements the sql.Scanner interface
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		j.Data = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		bytes = []byte(value.(string))
	}
	return json.Unmarshal(bytes, &j.Data)
}

// Value implements the driver.Valuer interface
func (j JSONField) Value() (driver.Value, error) {
	if j.Data == nil {
		return nil, nil
	}
	return json.Marshal(j.Data)
}

// MarshalJSON implements json.Marshaler
func (j JSONField) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONField) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.Data)
}
