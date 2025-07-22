package parsers

import (
	"strconv"
	"strings"
)

type SpeedtestResult struct {
	Spot     string
	Download float32
	Upload   float32
	Latency  float32
	Jitter   float32
}

type SpeedtestResults struct {
	Results []SpeedtestResult
	Time    string
}

func SpeedtestParser(textLines []string) []SpeedtestResults {
	startCases := []string{
		"大陆三网+教育网 IPv4 多线程测速，v",
		"大陆三网+教育网 IPv4 单线程测速，v",
		"各大洲 IPv4 八线程测速，v"}
	endCase := "系统时间："
	finalResults := make([]SpeedtestResults, 0)
	for _, startCase := range startCases {
		inBlock, i := false, 0
		results := make([]SpeedtestResult, 0)
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
				data := strings.Fields(line)
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
				down, _ := strconv.ParseFloat(data[spdIdx], 32)
				up, _ := strconv.ParseFloat(data[spdIdx+2], 32)
				lat, jit := 0.0, 0.0
				if len(data) < spdIdx+7 {
					lat, _ = strconv.ParseFloat(data[spdIdx+4], 32)
					jit, _ = strconv.ParseFloat(data[spdIdx+5], 32)
				} else {
					lat, _ = strconv.ParseFloat(data[spdIdx+4], 32)
					jit, _ = strconv.ParseFloat(data[spdIdx+6], 32)
				}
				results = append(results, SpeedtestResult{spot, float32(down), float32(up), float32(lat), float32(jit)})
			} else {
				head++
			}
		}
		head += i + 2
		time = strings.Replace(textLines[head-1], "北京时间：", "", 1)
		finalResults = append(finalResults, SpeedtestResults{results, time})
	}
	return finalResults
}
