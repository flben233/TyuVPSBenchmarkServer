package renderer

import (
	"VPSBenchmarkBackend/config"
	"VPSBenchmarkBackend/model"
	"VPSBenchmarkBackend/parser"
	"VPSBenchmarkBackend/repo"
	"VPSBenchmarkBackend/utils"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func RenderIndex(reportsCache map[string]model.ReportInfo) {
	tmpl, err := template.ParseFiles("templates/index.gohtml")
	if err != nil {
		log.Println(err)
		return
	}
	outputDir := config.Get().StaticsDir
	file, _ := os.OpenFile(filepath.Join(outputDir, "index.html"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer file.Close()
	reports := make([]model.ReportInfo, 0, len(reportsCache))
	for _, v := range reportsCache {
		reports = append(reports, v)
	}
	sort.Slice(reports, func(i, j int) bool {
		time1, _ := time.Parse("2006-01-02 15:04:05", reports[i].Date)
		time2, _ := time.Parse("2006-01-02 15:04:05", reports[j].Date)
		return time1.After(time2)
	})
	err = tmpl.Execute(file, reports)
	if err != nil {
		log.Printf("Failed to render template: %+v", err)
		return
	}
	log.Println("Index page rendered successfully with", len(reports), "reports")
}

func RegularlyRenderIndex(interval int) chan bool {
	path := config.Get().InputDir
	resultsCache := make(map[string]model.ReportInfo)
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
				err := repo.CascadeDeleteReport(resultsCache[fileName].Name)
				if err != nil {
					log.Printf("Error deleting report in database %s: %+v", fileName, err)
					return
				}
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
			results := parser.MainParser(string(textLines))
			info := model.ReportInfo{
				Name: results.Title,
				Path: "/reports/" + file.Name(),
				Date: results.Time}
			resultsCache[file.Name()] = info
			err = repo.CascadeInsertReport(info, results.SpdTest[0].Results, results.ECS.Trace)
			if err != nil {
				log.Printf("Error inserting report into database %s: %+v", file.Name(), err)
				return
			}
			modified = true
		}
		if modified {
			log.Println("Rendering index page with updated reports")
			RenderIndex(resultsCache)
		}
	}, interval)
}
