package parser

import (
	"VPSBenchmarkBackend/internal/report/model"
	"VPSBenchmarkBackend/internal/report/utils"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParsedResult is a temporary structure for parsed data
type ParsedResult struct {
	SpdTest   []model.SpeedtestResults
	ECS       model.ECSResult
	Media     model.MediaResults
	BestTrace []model.BestTraceResult
	Itdog     model.ItdogResult
	Disk      model.TyuDiskResult
	IPQuality model.IPQualityResult
	Title     string
	Time      string
	Link      string
}

func MainParser(html string) ParsedResult {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	textLines := strings.Split(doc.Text(), "\n")
	title, time, link := "", "", ""
	for _, n := range doc.Find("meta").Nodes {
		name := utils.GetAttr(n, "name")
		content := utils.GetAttr(n, "content")
		switch name {
		case "title":
			title = content
		case "time":
			time = content
		case "link":
			link = content
		}
	}
	return ParsedResult{
		SpdTest:   SpeedtestParser(textLines),
		ECS:       ECSParser(textLines),
		Media:     MediaParser(textLines),
		BestTrace: BestTraceParser(textLines),
		Itdog:     ItdogParser(doc),
		Disk:      TyuDiskParser(textLines),
		IPQuality: IPQualityParser(textLines),
		Title:     title,
		Time:      time,
		Link:      link,
	}
}
