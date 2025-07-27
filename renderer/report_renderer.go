package renderer

import (
    "VPSBenchmarkBackend/config"
    "VPSBenchmarkBackend/parsers"
    "VPSBenchmarkBackend/utils"
    "fmt"
    "html/template"
    "log"
    "os"
    "path/filepath"
    "strings"
)

func RenderReports(filename string, results parsers.BenchmarkResult) {
    outputFilePath := config.Get().OutputDir + string(filepath.Separator) + filename
    tmpl, err := template.New("report.gohtml").Funcs(
        map[string]any{"contains": strings.Contains}).ParseFiles(
        "templates" + string(filepath.Separator) + "report.gohtml")
    if err != nil {
        fmt.Println("Error parsing template:", err)
        return
    }
    if _, err := os.Stat(outputFilePath); os.IsNotExist(err) {
        file, _ := os.OpenFile(outputFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
        err = tmpl.Execute(file, results)
        if err != nil {
            log.Println(err)
        }
    }
}

func RegularlyRenderReports(interval int) chan bool {
    path := config.Get().InputDir
    files, err := os.ReadDir(path)
    if err != nil {
        fmt.Printf("Error reading directory: %+v", err)
        return nil
    }
    fileSet := make(map[string]struct{})
    return utils.SetInterval(func() {
        for _, file := range files {
            inputFile := path + string(filepath.Separator) + file.Name()
            if _, exists := fileSet[inputFile]; exists || file.IsDir() {
                continue
            }
            textLines, _ := os.ReadFile(inputFile)
            results := parsers.MainParser(string(textLines))
            RenderReports(file.Name(), results)
            fileSet[inputFile] = struct{}{}
        }
    }, interval)
}
