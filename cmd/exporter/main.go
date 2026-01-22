package main

import (
	ms "VPSBenchmarkBackend/internal/monitor/service"
	ts "VPSBenchmarkBackend/internal/tool/service"

	"github.com/gin-gonic/gin"
)

func QueryHosts(ctx *gin.Context) {
	var targets []string
	if err := ctx.BindJSON(&targets); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	results := ms.ExportQueryHosts(targets)
	ctx.JSON(200, results)
}

func Tracert(ctx *gin.Context) {
	var req ts.TracertRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	results := ts.ExportTracert(&req)
	ctx.String(200, results)
}

func main() {
	base := "/exporter"
	r := gin.Default()
	r.POST(base+"/monitor", QueryHosts)
	r.POST(base+"/tracert", Tracert)
	r.Run(":20831")
}
