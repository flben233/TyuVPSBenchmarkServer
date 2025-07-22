package main

import (
	"VPSBenchmarkBackend/parsers"
	"encoding/json"
	"os"
)

func main() {
	textLines, _ := os.ReadFile("output_1753019899.html")
	results := parsers.MainParser(string(textLines))
	j, _ := json.Marshal(results)
	err := os.WriteFile("output.json", j, 0644)
	if err != nil {
		return
	}
	//fmt.Println(results)
}
