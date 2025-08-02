package main

import (
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/handler"
	"VPSBenchmarkBackend/internal/renderer"
	"VPSBenchmarkBackend/internal/repo"
	"fmt"
	"log"
	"net/http"
)

func main() {
	err := config.Load("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// Initialize the database
	repo.InitDatabase()

	// Start the scheduler
	renderer.RegularlyRenderIndex(60000)
	renderer.RegularlyRenderReports(60000)

	// Set up HTTP server
	http.HandleFunc("/", handler.IndexHandler)

	http.HandleFunc("/search", handler.SearchHandler)

	http.HandleFunc("/api/search", handler.SearchAPIHandler)

	http.Handle("/reports/", http.StripPrefix("/reports/", http.FileServer(http.Dir(config.Get().OutputDir))))

	port := ":" + fmt.Sprintf("%d", config.Get().Port)
	log.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Printf("Server failed to start: %v\n", err)
	}
}
