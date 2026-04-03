package test

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/inspector/service"
	"fmt"
	"log"
	"testing"
)

func TestPing(t *testing.T) {
	err := config.Load("../config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// Initialize database
	dbPath := "../data/benchmark.db"
	if err := common.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully at", dbPath)
	service.PingHosts()
}
