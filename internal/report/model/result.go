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
	SpdTest   JSONField `json:"spdtest"`
	ECS       JSONField `json:"ecs"`
	Media     JSONField `json:"media"`
	BestTrace JSONField `json:"besttrace"`
	Itdog     JSONField `json:"itdog"`
	Disk      JSONField `json:"disk"`
	IPQuality JSONField `json:"ipquality"`
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

type IPQualityResult struct {
	Head []struct {
		IP      string `json:"IP"`
		Command string `json:"Command"`
		GitHub  string `json:"GitHub"`
		Time    string `json:"Time"`
		Version string `json:"Version"`
	} `json:"Head"`
	Info []struct {
		ASN          string `json:"ASN"`
		Organization string `json:"Organization"`
		Latitude     string `json:"Latitude"`
		Longitude    string `json:"Longitude"`
		DMS          string `json:"DMS"`
		Map          string `json:"Map"`
		TimeZone     string `json:"TimeZone"`
		City         struct {
			Name         string `json:"Name"`
			PostalCode   string `json:"PostalCode"`
			SubCode      string `json:"SubCode"`
			Subdivisions string `json:"Subdivisions"`
		} `json:"City"`
		Region struct {
			Code string `json:"Code"`
			Name string `json:"Name"`
		} `json:"Region"`
		Continent struct {
			Code string `json:"Code"`
			Name string `json:"Name"`
		} `json:"Continent"`
		RegisteredRegion struct {
			Code string `json:"Code"`
			Name string `json:"Name"`
		} `json:"RegisteredRegion"`
		Type string `json:"Type"`
	} `json:"Info"`
	Type []struct {
		Usage struct {
			IPinfo      string `json:"IPinfo"`
			Ipregistry  string `json:"ipregistry"`
			Ipapi       string `json:"ipapi"`
			AbuseIPDB   string `json:"AbuseIPDB"`
			IP2LOCATION string `json:"IP2LOCATION"`
		} `json:"Usage"`
		Company struct {
			IPinfo     string `json:"IPinfo"`
			Ipregistry string `json:"ipregistry"`
			Ipapi      string `json:"ipapi"`
		} `json:"Company"`
	} `json:"Type"`
	Score []struct {
		IP2LOCATION string `json:"IP2LOCATION"`
		SCAMALYTICS string `json:"SCAMALYTICS"`
		Ipapi       string `json:"ipapi"`
		AbuseIPDB   string `json:"AbuseIPDB"`
		IPQS        string `json:"IPQS"`
		DBIP        string `json:"DBIP"`
	} `json:"Score"`
	Factor []struct {
		CountryCode map[string]string      `json:"CountryCode"`
		Proxy       map[string]interface{} `json:"Proxy"`
		Tor         map[string]interface{} `json:"Tor"`
		VPN         map[string]interface{} `json:"VPN"`
		Server      map[string]interface{} `json:"Server"`
		Abuser      map[string]interface{} `json:"Abuser"`
		Robot       map[string]interface{} `json:"Robot"`
	} `json:"Factor"`
	Media []struct {
		TikTok           struct{ Status, Region, Type string } `json:"TikTok"`
		DisneyPlus       struct{ Status, Region, Type string } `json:"DisneyPlus"`
		Netflix          struct{ Status, Region, Type string } `json:"Netflix"`
		Youtube          struct{ Status, Region, Type string } `json:"Youtube"`
		AmazonPrimeVideo struct{ Status, Region, Type string } `json:"AmazonPrimeVideo"`
		Spotify          struct{ Status, Region, Type string } `json:"Spotify"`
		ChatGPT          struct{ Status, Region, Type string } `json:"ChatGPT"`
	} `json:"Media"`
	Mail []struct {
		Port25       bool `json:"Port25"`
		Gmail        bool `json:"Gmail"`
		Outlook      bool `json:"Outlook"`
		Yahoo        bool `json:"Yahoo"`
		Apple        bool `json:"Apple"`
		QQ           bool `json:"QQ"`
		MailRU       bool `json:"MailRU"`
		AOL          bool `json:"AOL"`
		GMX          bool `json:"GMX"`
		MailCOM      bool `json:"MailCOM"`
		N163         bool `json:"163"`
		Sohu         bool `json:"Sohu"`
		Sina         bool `json:"Sina"`
		DNSBlacklist struct {
			Total       int `json:"Total"`
			Clean       int `json:"Clean"`
			Marked      int `json:"Marked"`
			Blacklisted int `json:"Blacklisted"`
		} `json:"DNSBlacklist"`
	} `json:"Mail"`
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

func (j JSONField) GormDBDataType() string {
	return "text"
}

// MarshalJSON implements json.Marshaler
func (j JSONField) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONField) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &j.Data)
}
