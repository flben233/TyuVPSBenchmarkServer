package parser

import (
	"VPSBenchmarkBackend/model"
	"strings"
)

func BestTraceParser(textLines []string) []model.BestTraceResult {
	start := "关于软件卸载，因为nexttrace是绿色版单文件"
	inBlock := false
	results := make([]model.BestTraceResult, 0)
	for i, j := 0, 0; j < len(textLines); j++ {
		if !strings.Contains(textLines[j], start) && !inBlock {
			continue
		} else if !inBlock {
			inBlock = true
			i = j + 2
			j++
			continue
		}
		if strings.Contains(textLines[j], "-----------------------------------------------------------------") {
			results = append(results, model.BestTraceResult{Region: textLines[i], Route: strings.Join(textLines[i+1:j], "\n")})
			i = j + 1
		}
	}
	return results
}
