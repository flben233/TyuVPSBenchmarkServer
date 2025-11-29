package main

import (
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/report"
	"VPSBenchmarkBackend/internal/report/store"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// TODO: 统一的返回体结构
// TODO: 鉴权
// TODO: Traceroute、小鸡Ping监控、WHOIS、IP查询、ping测试工具集成
// TODO: 支持IPQuality
// TODO: 短链接
func main() {
	err := config.Load("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// Initialize database
	dbPath := "./data/benchmark.db"
	if err := store.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully at", dbPath)

	r := gin.Default()
	report.RegisterRouter(config.Get().BaseURL, r)
	r.Run(fmt.Sprintf(":%d", config.Get().Port))
}
