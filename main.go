package main

import (
	"VPSBenchmarkBackend/config"
	"VPSBenchmarkBackend/renderer"
	"fmt"
	"net/http"
	"path/filepath"
)

func main() {
	err := config.Load("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// Start the scheduler
	renderer.RegularlyRenderIndex(60000)
	renderer.RegularlyRenderReports(60000)

	// Set up HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Get().StaticsDir, "index.html"))
	})
	http.Handle("/reports/", http.StripPrefix("/reports/", http.FileServer(http.Dir(config.Get().OutputDir))))

	port := ":8080"
	fmt.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
