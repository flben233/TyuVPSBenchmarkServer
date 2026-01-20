package parser

import (
	"VPSBenchmarkBackend/internal/report/model"
	"VPSBenchmarkBackend/internal/report/utils"
	"github.com/PuerkitoBio/goquery"
)

func ItdogParser(doc *goquery.Document) model.ItdogResult {
	result := model.ItdogResult{}
	for i, n := range doc.Find("img").Nodes {
		if i == 0 {
			result.Ping = utils.GetAttr(n, "src")
		} else {
			result.Route = append(result.Route, utils.GetAttr(n, "src"))
		}
	}
	return result
}
