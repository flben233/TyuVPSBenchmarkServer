package parser

import (
	"VPSBenchmarkBackend/internal/report/model"
	"strconv"
	"strings"
)

func SpeedtestParser(textLines []string) []model.SpeedtestResults {
	startCases := []string{
		"大陆三网+教育网 IPv4 多线程测速，v",
		"大陆三网+教育网 IPv4 单线程测速，v",
		"各大洲 IPv4 八线程测速，v"}
	endCase := "系统时间："
	finalResults := make([]model.SpeedtestResults, 0)
	for _, startCase := range startCases {
		inBlock, i := false, 0
		results := make([]model.SpeedtestResult, 0)
		time := ""
		head := 0
		for _, line := range textLines[head:] {
			if strings.Contains(line, startCase) {
				inBlock = true
			} else if inBlock {
				i++
				if i < 3 || strings.Contains(line, "测速次数过多") || strings.Contains(line, "-----------") {
					continue
				} else if strings.Contains(line, endCase) {
					break
				}
				data := strings.Fields(strings.ReplaceAll(line, "失败", "失败 Mbps"))
				spdIdx := 0
				spot := ""
				for j, d := range data {
					if strings.Contains(d, "Mbps") {
						spdIdx = j - 1
						for k := 0; k <= j-2; k++ {
							spot += data[k]
						}
						break
					}
				}
				if len(data) < spdIdx+3 {
					continue
				}
				down, err := strconv.ParseFloat(data[spdIdx], 32)
				if err != nil {
					continue
				}
				up, err := strconv.ParseFloat(data[spdIdx+2], 32)
				if err != nil {
					continue
				}
				lat, jit := 0.0, 0.0
				if len(data) == spdIdx+6 {
					lat, _ = strconv.ParseFloat(data[spdIdx+4], 32)
					jit, _ = strconv.ParseFloat(data[spdIdx+5], 32)
				} else if len(data) >= spdIdx+7 {
					lat, _ = strconv.ParseFloat(data[spdIdx+4], 32)
					jit, _ = strconv.ParseFloat(data[spdIdx+6], 32)
				}
				results = append(results, model.SpeedtestResult{Spot: spot, Download: float32(down), Upload: float32(up), Latency: float32(lat), Jitter: float32(jit)})
			} else {
				head++
			}
		}
		head += i + 2
		if head >= len(textLines) {
			time = ""
		} else {
			time = strings.Replace(textLines[head-1], "北京时间: ", "", 1)
		}
		finalResults = append(finalResults, model.SpeedtestResults{Results: results, Time: time})
	}
	return finalResults
}
