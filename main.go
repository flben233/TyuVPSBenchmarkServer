package main

import (
	"VPSBenchmarkBackend/internal/auth"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/monitor"
	"VPSBenchmarkBackend/internal/report"
	"VPSBenchmarkBackend/internal/report/store"
	"VPSBenchmarkBackend/internal/tool"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

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
	if err := monitor.InitMonitorStore(dbPath); err != nil {
		log.Fatalf("Failed to initialize monitor store: %v", err)
	}
	log.Println("Database initialized successfully at", dbPath)

	r := gin.Default()
	auth.RegisterRouter(config.Get().BaseURL, r)
	report.RegisterRouter(config.Get().BaseURL, r)
	monitor.RegisterRouter(config.Get().BaseURL, r)
	tool.RegisterRouter(config.Get().BaseURL, r)
	r.Run(fmt.Sprintf(":%d", config.Get().Port))
}
