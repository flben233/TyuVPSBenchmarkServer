package main

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// TODO: 对所有用户上传内容强制人工审核，审核通过才可公开（目前计划中的只有monitor和looking glass，数据库增加一个字段区分未审核/通过/拒绝即可）
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

	r.Run(fmt.Sprintf(":%d", config.Get().Port))
}
