package model

type MediaIndex struct {
	ID       uint `gorm:"primaryKey"`
	ReportID string
	Region   string
	Media    string
	Unlock   bool
	IPv6     bool
}

type SpeedtestIndex struct {
	ID       uint `gorm:"primaryKey"`
	ReportID string
	Spot     string
	Download float32
	Upload   float32
	Latency  float32
	ISP      string
}

type InfoIndex struct {
	ID             uint `gorm:"primaryKey"`
	ReportID       string
	IPv6Support    bool
	Virtualization string
	SeqRead        float32
	SeqWrite       float32
}

type BacktraceIndex struct {
	ID        uint `gorm:"primaryKey"`
	ReportID  string
	Spot      string
	RouteType string
	ISP       string
}
