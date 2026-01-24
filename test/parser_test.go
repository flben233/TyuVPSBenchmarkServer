package test

import (
	"VPSBenchmarkBackend/internal/report/parser"
	"encoding/json"
	"os"
	"testing"
)

func TestMainParser(t *testing.T) {
	data, err := os.ReadFile("testdata/report_example.html")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}
	html := string(data)
	result := parser.MainParser(html)
	if len(result.SpdTest) != 3 {
		str, _ := json.MarshalIndent(result.SpdTest, "", "  ")
		t.Errorf("Expected Speedtest results, got: %s", str)
	}
	if result.ECS == nil {
		t.Errorf("Expected ECS result, got nil")
	}
	if result.Media == nil {
		t.Errorf("Expected Media result, got nil")
	}
	if len(result.BestTrace) == 0 {
		t.Errorf("Expected BestTrace results, got none")
	}
	if result.Itdog == nil {
		t.Errorf("Expected Itdog result, got nil")
	}
	if result.Disk == nil {
		t.Errorf("Expected Disk result, got nil")
	}
	if result.IPQuality == nil {
		t.Errorf("Expected IPQuality result, got nil")
	}
}
