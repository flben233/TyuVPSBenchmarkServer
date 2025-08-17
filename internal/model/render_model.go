package model

type ReportInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Date string `json:"date"`
}

type SitemapItem struct {
	Loc  string
	Last string
}
