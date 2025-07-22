package parsers

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
)

type BenchmarkResult struct {
	SpdTest   []SpeedtestResults
	ECS       ECSResult
	Media     MediaResults
	BestTrace []BestTraceResult
	Itdog     ItdogResult
}

func MainParser(html string) BenchmarkResult {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	textLines := strings.Split(doc.Text(), "\n")
	return BenchmarkResult{SpeedtestParser(textLines), ECSParser(textLines), MediaParser(textLines), BestTraceParser(textLines), ItdogParser(doc)}
}
