package parser

import (
	"VPSBenchmarkBackend/internal/report/model"
	"encoding/json"
	"regexp"
	"strings"
)

var ansiRegex = regexp.MustCompile("\\x1b\\[?[0-9;]*[a-zA-Z]")

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
		jsonData = ansiRegex.ReplaceAllString(jsonData, "")
		err := json.NewDecoder(strings.NewReader(jsonData)).Decode(&result)
		if err != nil {
			return nil
		}
	} else {
		return nil
	}
	return &result
}
