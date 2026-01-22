package parser

import (
	"VPSBenchmarkBackend/internal/report/model"
	"encoding/json"
	"strings"
)

func IPQualityParser(textLines []string) *model.IPQualityResult {
	result := model.IPQualityResult{}
	inBlock := false
	inJSON := false
	jsonData := ""
	for _, line := range textLines {
		if line == "======== IPQuality ========" {
			inBlock = true
			continue
		}
		if inBlock {
			if line == "{" {
				inJSON = true
				jsonData += line + "\n"
				continue
			}
		}
		if inJSON {
			jsonData += line + "\n"
		}
		if line == "" && inJSON {
			break
		}
	}
	if jsonData != "" {
		json.NewDecoder(strings.NewReader(jsonData)).Decode(&result)
	} else {
		return nil
	}
	return &result
}
