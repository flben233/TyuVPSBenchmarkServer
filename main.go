package main

import (
	"VPSBenchmarkBackend/internal/auth"
	_ "VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	_ "VPSBenchmarkBackend/internal/lookingglass"
	_ "VPSBenchmarkBackend/internal/monitor"
	_ "VPSBenchmarkBackend/internal/report"
	_ "VPSBenchmarkBackend/internal/tool"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// TODO: 优化SQL查询性能
// TODO: 评测记录可以关联到一个监控
// TODO: 修正线路类型
// @title Lolicon VPS API
// @BasePath /api
func main() {
	err := config.Load("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// Initialize database
	dbPath := "./data/benchmark.db"
	if err := common.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully at", dbPath)

	// Start background cron jobs
	common.RunCronJobs()
	log.Println("Background cron jobs started")

	r := gin.Default()
	r.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Use(auth.GetAllowCORSMiddleware())
	common.InitRouter("/api", r)

	r.Run(fmt.Sprintf(":%d", config.Get().Port))
}
