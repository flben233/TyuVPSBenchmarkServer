package parser

import (
	"VPSBenchmarkBackend/internal/model"
	"VPSBenchmarkBackend/internal/utils"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

func MainParser(html string) model.BenchmarkResult {
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
	return model.BenchmarkResult{SpeedtestParser(textLines), ECSParser(textLines), MediaParser(textLines), BestTraceParser(textLines), ItdogParser(doc), title, time, link}
}
