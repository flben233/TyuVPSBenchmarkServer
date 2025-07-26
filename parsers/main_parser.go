package parsers

import (
	"VPSBenchmarkBackend/utils"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

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

func MainParser(html string) BenchmarkResult {
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
	return BenchmarkResult{SpeedtestParser(textLines), ECSParser(textLines), MediaParser(textLines), BestTraceParser(textLines), ItdogParser(doc), title, time, link}
}
