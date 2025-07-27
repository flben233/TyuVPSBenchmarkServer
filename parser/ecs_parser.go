package parser

import (
	"VPSBenchmarkBackend/model"
	"regexp"
	"strconv"
	"strings"
)

func ECSParser(textLines []string) model.ECSResult {
	blocks := []string{"基础信息查询", "CPU测试", "内存测试", "磁盘dd读写测试", "TikTok解锁", "IP质量检测", "邮件端口检测", "三网回程", "回程路由--", "----------------"}
	blkIdx := -1
	result := model.ECSResult{}
	tmpTypes := make(map[string]string)
	for i, j := 0, 0; blkIdx < len(blocks); j++ {
		if !strings.Contains(textLines[j], blocks[0]) && blkIdx == -1 {
			continue
		} else if blkIdx == -1 {
			i = j + 1
			blkIdx++
		}
		if blkIdx+1 < len(blocks) && strings.Contains(textLines[j], blocks[blkIdx+1]) {
			switch blkIdx {
			case 0:
				result.Info = infoParser(textLines[i:j])
			case 1:
				result.Cpu = cpuParser(textLines[i:j])
			case 2:
				result.Mem = memParser(textLines[i:j])
			case 3:
				result.Disk = diskParser(textLines[i:j])
			case 4:
				result.Tiktok = tiktokParser(textLines[i:j])
			case 5:
				result.IPQuality = strings.Join(textLines[i:j], "\n")
			case 6:
				result.Mail = mailParser(textLines[i:j])
			case 7:
				tmpTypes = typeParser(textLines[i:j])
			case 8:
				result.Trace = model.TraceResult{
					Types:     tmpTypes,
					BackRoute: strings.Join(textLines[i:j], "\n"),
				}
			}
			blkIdx++
			i = j + 1
		}
		if strings.Contains(textLines[j], "总共花费") {
			result.Time = strings.TrimSpace(strings.Split(textLines[j+1], " :")[1])
			break
		}
	}
	return result
}

func infoParser(textLines []string) map[string]string {
	info := make(map[string]string)
	for _, line := range textLines {
		parts := strings.Split(line, ":")
		val := strings.Join(parts[1:], "")
		info[strings.TrimSpace(parts[0])] = strings.TrimSpace(val)
	}
	return info
}

func cpuParser(textLines []string) model.CPUResult {
	result := model.CPUResult{}
	for _, line := range textLines {
		if strings.Contains(line, "(单核)得分") {
			parts := strings.Fields(line)
			i64, _ := strconv.ParseInt(parts[2], 10, 32)
			result.Single = int32(i64)
		} else if strings.Contains(line, "(多核)得分") {
			parts := strings.Fields(line)
			i64, _ := strconv.ParseInt(parts[2], 10, 32)
			result.Multi = int32(i64)
		}
	}
	return result
}

func memParser(textLines []string) model.MemResult {
	result := model.MemResult{}
	for _, line := range textLines {
		if strings.Contains(line, "单线程读测试") {
			parts := strings.Fields(line)
			f64, _ := strconv.ParseFloat(parts[1], 32)
			result.Read = float32(f64)
		} else if strings.Contains(line, "单线程写测试") {
			parts := strings.Fields(line)
			f64, _ := strconv.ParseFloat(parts[1], 32)
			result.Write = float32(f64)
		}
	}
	return result
}

func diskParser(textLines []string) model.DiskResult {
	result := model.DiskResult{}
	re, _ := regexp.Compile("\\t+")
	for _, line := range textLines {
		if strings.Contains(line, "1GB-1M Block") {
			parts := re.Split(line, 3)
			result.SeqWrite = parts[1]
			result.SeqRead = parts[2]
		}
	}
	return result
}

func tiktokParser(textLines []string) string {
	re, _ := regexp.Compile("\\t+")
	return re.Split(textLines[0], 2)[1]
}

func mailParser(textLines []string) map[string][]bool {
	mail := make(map[string][]bool)
	for _, line := range textLines[1:] {
		parts := strings.Fields(line)
		ports := make([]bool, 6)
		for i, part := range parts[1:] {
			if part == "✔" {
				ports[i] = true
			}
		}
		mail[parts[0]] = ports
	}
	return mail
}

func typeParser(textLines []string) map[string]string {
	types := make(map[string]string)
	for _, line := range textLines {
		parts := strings.Fields(line)
		if len(parts) == 3 {
			types[parts[0]] = parts[2] // Handle cases with fewer parts
		} else if len(parts) >= 4 {
			types[parts[0]] = parts[2] + " " + parts[3]
		}
	}
	return types
}
