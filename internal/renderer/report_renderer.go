package renderer

import (
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/model"
	"VPSBenchmarkBackend/internal/parser"
	"VPSBenchmarkBackend/internal/utils"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func RenderReports(filename string, result model.BenchmarkResult) {
	outputFilePath := config.Get().OutputDir + string(filepath.Separator) + filename
	tmpl, err := template.New("report.gohtml").Funcs(
		map[string]any{"contains": strings.Contains}).ParseFiles(
		"templates" + string(filepath.Separator) + "report.gohtml")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	if !utils.FileExists(outputFilePath) {
		file, _ := os.OpenFile(outputFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		defer file.Close()
		if result.Title == "" {
			result.Title = strings.Split(filename, ".")[0]
		}
		err = tmpl.Execute(file, result)
		if err != nil {
			log.Println(err)
		}
	}
}

func RegularlyRenderReports(interval int) chan bool {
	path := config.Get().InputDir
	fileSet := make(map[string]struct{})
	return utils.SetInterval(func() {
		files, err := os.ReadDir(path)
		if err != nil {
			fmt.Printf("Error reading directory: %+v", err)
		}
		for _, file := range files {
			inputFile := path + string(filepath.Separator) + file.Name()
			if _, exists := fileSet[inputFile]; exists || file.IsDir() {
				continue
			}
			textLines, _ := os.ReadFile(inputFile)
			results := parser.MainParser(string(textLines))
			RenderReports(file.Name(), results)
			fileSet[inputFile] = struct{}{}
		}
	}, interval)
}
