package service

import (
	"VPSBenchmarkBackend/internal/report/model"
	"VPSBenchmarkBackend/internal/report/parser"
	"VPSBenchmarkBackend/internal/report/store"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// generateID generates a random ID for reports
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// AddReport parses and saves a new benchmark report
func AddReport(rawHTML string) (string, error) {
	if rawHTML == "" {
		return "", errors.New("raw HTML content is required")
	}

	// Parse the report
	parsedResult := parser.MainParser(rawHTML)

	// Generate unique ID
	reportID := generateID()

	// Check if report already exists
	exists, err := store.ReportExists(reportID)
	if err != nil {
		return "", fmt.Errorf("failed to check report existence: %w", err)
	}
	if exists {
		// Retry with a new ID
		reportID = generateID()
	}

	currentTime := time.Now()
	// Create BenchmarkResult for database
	report := &model.BenchmarkResult{
		ReportID:  reportID,
		Title:     parsedResult.Title,
		Time:      parsedResult.Time,
		Link:      parsedResult.Link,
		RawHTML:   rawHTML,
		SpdTest:   model.JSONField{Data: parsedResult.SpdTest},
		ECS:       model.JSONField{Data: parsedResult.ECS},
		Media:     model.JSONField{Data: parsedResult.Media},
		BestTrace: model.JSONField{Data: parsedResult.BestTrace},
		Itdog:     model.JSONField{Data: parsedResult.Itdog},
		Disk:      model.JSONField{Data: parsedResult.Disk},
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	// Save to database
	if err := store.SaveReport(report); err != nil {
		return "", fmt.Errorf("failed to save report: %w", err)
	}

	return reportID, nil
}

// DeleteReport removes a report from the database
func DeleteReport(reportID string) error {
	if reportID == "" {
		return errors.New("report ID is required")
	}

	if err := store.DeleteReport(reportID); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}
