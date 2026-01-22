package parser

import (
	"VPSBenchmarkBackend/internal/report/model"
	"strings"
)

func TyuDiskParser(textLines []string) *model.TyuDiskResult {
	startCase := "-------------------------------- TyuDiskMark --------------------------------"
	endCase := "-----------------------------------------------------------------------------"
	result := model.TyuDiskResult{Data: make([][]string, 0)}
	inBlock := false
	i := 0
	for ; i < len(textLines); i++ {
		line := textLines[i]
		if strings.Contains(line, startCase) {
			inBlock = true
			i += 5
		} else if inBlock {
			if strings.Contains(line, endCase) {
				break
			}
			data := strings.Fields(line)
			result.Data = append(result.Data, data)
		}
	}
	if i < len(textLines) {
		result.Time = strings.Replace(textLines[i+2], "北京时间: ", "", 1)
	}
	if inBlock == false {
		return nil
	}
	return &result
}
