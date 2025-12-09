package main

import (
	"VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/report"
	"VPSBenchmarkBackend/internal/report/store"
	"VPSBenchmarkBackend/internal/tool"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// TODO: 支持IPQuality
// TODO: 小鸡Ping监控（需要支持用户提交服务器）
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
	if err := store.InitReportStore(dbPath); err != nil {
		log.Fatalf("Failed to initialize report store: %v", err)
	}
	log.Println("Database initialized successfully at", dbPath)

	r := gin.Default()
	auth.RegisterRouter(config.Get().BaseURL, r)
	report.RegisterRouter(config.Get().BaseURL, r)
	tool.RegisterRouter(config.Get().BaseURL, r)
	r.Run(fmt.Sprintf(":%d", config.Get().Port))
}
