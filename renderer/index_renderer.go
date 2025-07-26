package renderer

import (
	"VPSBenchmarkBackend/config"
	"VPSBenchmarkBackend/parsers"
	"VPSBenchmarkBackend/utils"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ReportInfo struct {
	Name string
	Path string
	Date string
}

func RenderIndex(reportsCache map[string]ReportInfo) {
	tmpl, err := template.ParseFiles("templates/index.gohtml")
	if err != nil {
		log.Println(err)
		return
	}
	file, _ := os.OpenFile(filepath.Join(config.Get().StaticsDir, "index.html"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	reports := make([]ReportInfo, 0, len(reportsCache))
	for _, v := range reportsCache {
		reports = append(reports, v)
	}
	err = tmpl.Execute(file, reports)
	if err != nil {
		log.Printf("Failed to render template: %+v", err)
		return
	}
	log.Println("Index page rendered successfully with", len(reports), "reports")
}

func RegularlyRenderIndex(interval int) chan bool {
	path := config.Get().InputDir
	resultsCache := make(map[string]ReportInfo)
	return utils.SetInterval(func() {
		files, err := os.ReadDir(path)
		modified := false
		if err != nil {
			log.Printf("Error reading directory: %+v", err)
			return
		}
		// Check deleted files
		for fileName := range resultsCache {
			deleted := true
			for _, file := range files {
				if file.Name() == fileName {
					deleted = false
					break
				}
			}
			if deleted {
				delete(resultsCache, fileName)
				modified = true
			}
		}
		// Check new or modified files
		for _, file := range files {
			if _, exists := resultsCache[file.Name()]; exists || file.IsDir() || !strings.HasSuffix(file.Name(), ".html") {
				continue
			}
			inputFile := filepath.Join(path, file.Name())
			textLines, err := os.ReadFile(inputFile)
			if err != nil {
				log.Printf("Error reading file %s: %+v", inputFile, err)
				continue
			}
			results := parsers.MainParser(string(textLines))
			resultsCache[file.Name()] = ReportInfo{
				Name: file.Name(),
				Path: "/reports/" + file.Name(),
				Date: results.Time}
			modified = true
		}
		if modified {
			log.Println("Rendering index page with updated reports")
			RenderIndex(resultsCache)
		}
	}, interval)
}
