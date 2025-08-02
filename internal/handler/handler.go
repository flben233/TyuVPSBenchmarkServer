package handler

import (
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/model"
	"VPSBenchmarkBackend/internal/repo"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(config.Get().StaticsDir, "index.html"))
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(config.Get().StaticsDir, "search.html"))
}

func SearchAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	keyword := r.URL.Query().Get("keyword")
	speedStr := r.URL.Query().Get("speed")
	routeType := r.URL.Query().Get("routeType")

	var results []model.ReportInfo
	var err error
	speed := float32(0)
	if speedStr != "" {
		speed1, convErr := strconv.ParseFloat(speedStr, 32)
		speed = float32(speed1)
		if convErr != nil {
			http.Error(w, "Invalid speed parameter", http.StatusBadRequest)
			return
		}
	}
	results, err = repo.FindReportsByConditions(keyword, speed, routeType)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}
