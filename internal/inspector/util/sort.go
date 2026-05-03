package util

import (
	"VPSBenchmarkBackend/internal/inspector/model"
	"math"
	"sort"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

const defaultCustomOrder = math.MaxInt32

var collator = collate.New(language.Chinese)

// SortHosts 对主机列表进行排序
// 规则：
// 1. 按 custom_order 升序排序
// 2. custom_order 为 2147483647 的主机按照名称排序（英文按字母，中文按拼音）
// 3. 名称相同时按 ID 排序
func SortHosts(hosts []model.InspectHost) {
	sort.Slice(hosts, func(i, j int) bool {
		orderI := hosts[i].CustomOrder
		orderJ := hosts[j].CustomOrder

		// 两者都有自定义排序值
		if orderI != defaultCustomOrder && orderJ != defaultCustomOrder {
			return orderI < orderJ
		}

		// 一个有自定义排序值，一个没有
		if orderI != defaultCustomOrder {
			return true
		}
		if orderJ != defaultCustomOrder {
			return false
		}

		// 两者都没有自定义排序值，按名称排序
		cmp := collator.CompareString(hosts[i].Name, hosts[j].Name)
		if cmp != 0 {
			return cmp < 0
		}

		// 名称相同时按 ID 排序
		return hosts[i].ID < hosts[j].ID
	})
}
