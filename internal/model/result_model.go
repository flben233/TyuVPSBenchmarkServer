package model

import "html/template"

type BestTraceResult struct {
	Region string
	Route  string
}

type CPUResult struct {
	Single int32
	Multi  int32
}

type MemResult struct {
	Read  float32
	Write float32
}

type DiskResult struct {
	SeqRead  string
	SeqWrite string
}

type TraceResult struct {
	Types     map[string]string
	BackRoute string
}

type ECSResult struct {
	Info      map[string]string
	Cpu       CPUResult
	Mem       MemResult
	Disk      DiskResult
	Tiktok    string
	IPQuality string
	Mail      map[string][]bool // bool数组定义：SMTP SMTPS POP3 POP3S IMAP MAPS
	Trace     TraceResult
	Time      string
}

type ItdogResult struct {
	Ping  template.URL
	Route []template.URL
}

type BenchmarkResult struct {
	SpdTest   []SpeedtestResults
	ECS       ECSResult
	Media     MediaResults
	BestTrace []BestTraceResult
	Itdog     ItdogResult
	Title     string
	Time      string
	Link      string
}

type SpeedtestResult struct {
	Spot     string
	Download float32
	Upload   float32
	Latency  float32
	Jitter   float32
}

type SpeedtestResults struct {
	Results []SpeedtestResult
	Time    string
}

type MediaPair struct {
	Media  string
	Unlock string
}

type MediaBlock struct {
	Region  string
	Results []MediaPair
}

type MediaResults struct {
	IPv4 []MediaBlock
	IPv6 []MediaBlock
}
