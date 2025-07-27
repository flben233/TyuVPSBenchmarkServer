package parser

import (
	"VPSBenchmarkBackend/model"
	"regexp"
	"strings"
)

func MediaParser(textLines []string) model.MediaResults {
	starts := []string{"** 正在测试 IPv4 解锁情况", "** 正在测试 IPv6 解锁情况"}
	results := model.MediaResults{}
	i := 0
	for idx, start := range starts {
		result := make([]model.MediaBlock, 0)
		inBlock := false
		for ; i < len(textLines); i++ {
			if !strings.Contains(textLines[i], start) && !inBlock {
				continue
			} else if !inBlock {
				inBlock = true
			}
			if idx+1 < len(starts) && strings.Contains(textLines[i], starts[idx+1]) ||
				strings.Contains(textLines[i], "当前主机不支持") {
				break
			} else if strings.Contains(textLines[i], "==[") {
				l, r := processMBlk(textLines[i:], result)
				result = r
				i += l
			}
		}
		switch idx {
		case 0:
			results.IPv4 = result
		case 1:
			results.IPv6 = result
		}
	}
	return results
}

func processMBlk(textLines []string, results []model.MediaBlock) (int, []model.MediaBlock) {
	region, i := "", 0
	result := make([]model.MediaPair, 0)
	leftRe, _ := regexp.Compile("=+\\[ ")
	rightRe, _ := regexp.Compile(" ]=+")
	for ; i < len(textLines); i++ {
		if i == 0 {
			region = leftRe.ReplaceAllString(rightRe.ReplaceAllString(textLines[i], ""), "")
			continue
		}
		parts := strings.Split(textLines[i], ":\t")
		if len(parts) == 2 {
			result = append(result, model.MediaPair{
				Media:  parts[0],
				Unlock: strings.TrimSpace(parts[1]),
			})
		}
		if strings.Contains(textLines[i], "=====================") {
			break
		}
	}
	return i, append(results, model.MediaBlock{
		Region:  region,
		Results: result,
	})
}
