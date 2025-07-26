package parsers

import (
	"VPSBenchmarkBackend/utils"
	"github.com/PuerkitoBio/goquery"
	"html/template"
)

type ItdogResult struct {
	Ping  template.URL
	Route []template.URL
}

func ItdogParser(doc *goquery.Document) ItdogResult {
	result := ItdogResult{}
	for i, n := range doc.Find("img").Nodes {
		if i == 0 {
			result.Ping = template.URL(utils.GetAttr(n, "src"))
		} else {
			result.Route = append(result.Route, template.URL(utils.GetAttr(n, "src")))
		}
	}
	return result
}
