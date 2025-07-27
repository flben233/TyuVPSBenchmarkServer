package parser

import (
	"VPSBenchmarkBackend/model"
	"VPSBenchmarkBackend/utils"
	"github.com/PuerkitoBio/goquery"
	"html/template"
)

func ItdogParser(doc *goquery.Document) model.ItdogResult {
	result := model.ItdogResult{}
	for i, n := range doc.Find("img").Nodes {
		if i == 0 {
			result.Ping = template.URL(utils.GetAttr(n, "src"))
		} else {
			result.Route = append(result.Route, template.URL(utils.GetAttr(n, "src")))
		}
	}
	return result
}
