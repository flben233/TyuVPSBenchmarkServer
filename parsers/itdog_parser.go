package parsers

import (
	"VPSBenchmarkBackend/utils"
	"github.com/PuerkitoBio/goquery"
)

type ItdogResult struct {
	Ping  string
	Route []string
}

func ItdogParser(doc *goquery.Document) ItdogResult {
	result := ItdogResult{}
	for i, n := range doc.Find("img").Nodes {
		if i == 0 {
			result.Ping = utils.GetAttr(n, "src")
		} else {
			result.Route = append(result.Route, utils.GetAttr(n, "src"))
		}
	}
	return result
}
