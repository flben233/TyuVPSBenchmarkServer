package store

import (
	"VPSBenchmarkBackend/internal/report/model"
	"VPSBenchmarkBackend/internal/report/request"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// InitDB initializes the database connection and creates tables
func InitDB(dbPath string) error {
	// Ensure the data directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database connection
	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&model.BenchmarkResult{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// SaveReport saves a new benchmark report to database
func SaveReport(report *model.BenchmarkResult) error {
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.Create(report).Error
}

// GetReportByID retrieves a report by its ReportID
func GetReportByID(reportID string) (*model.BenchmarkResult, error) {
	if db == nil {
		return nil, errors.New("database not initialized")
	}
	var report model.BenchmarkResult
	if err := db.Where("report_id = ?", reportID).First(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// ListReports returns a list of all reports with basic info
func ListReports(limit, offset int) ([]model.BenchmarkResult, int64, error) {
	if db == nil {
		return nil, 0, errors.New("database not initialized")
	}
	var reports []model.BenchmarkResult
	var total int64

	// Get total count
	if err := db.Model(&model.BenchmarkResult{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := db.Select("id", "report_id", "title", "time", "link", "created_at", "updated_at").
		Order("time DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// DeleteReport deletes a report by its ReportID
func DeleteReport(reportID string) error {
	if db == nil {
		return errors.New("database not initialized")
	}
	result := db.Where("report_id = ?", reportID).Delete(&model.BenchmarkResult{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("report not found")
	}
	return nil
}

// TODO: 重写这个，现在的实现不对
// SearchReports searches for reports based on search criteria
func SearchReports(searchReq *request.SearchRequest, limit, offset int) ([]model.BenchmarkResult, int64, error) {
	if db == nil {
		return nil, 0, errors.New("database not initialized")
	}

	query := db.Model(&model.BenchmarkResult{})

	// Keyword search (search in title)
	if searchReq.Keyword != nil && *searchReq.Keyword != "" {
		query = query.Where("title LIKE ?", "%"+*searchReq.Keyword+"%")
	}

	// Media unlocks search
	if len(searchReq.MediaUnlocks) > 0 {
		for _, media := range searchReq.MediaUnlocks {
			query = query.Where("media LIKE ?", "%"+media+"%")
		}
	}

	// Virtualization search
	if searchReq.Virtualization != nil && *searchReq.Virtualization != "" {
		query = query.Where("ecs LIKE ?", "%"+*searchReq.Virtualization+"%")
	}

	// IPv6 support search
	if searchReq.IPv6Support != nil {
		if *searchReq.IPv6Support {
			query = query.Where("media LIKE ?", "%IPv6%")
		} else {
			query = query.Where("media NOT LIKE ? OR media IS NULL", "%IPv6%")
		}
	}

	// ASN specific searches (CT, CU, CM)
	if searchReq.CTParams != nil {
		query = applyASNFilters(query, searchReq.CTParams, "CT")
	}
	if searchReq.CUParams != nil {
		query = applyASNFilters(query, searchReq.CUParams, "CU")
	}
	if searchReq.CMParams != nil {
		query = applyASNFilters(query, searchReq.CMParams, "CM")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	var reports []model.BenchmarkResult
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// TODO: 重写这个，现在的实现不对
// applyASNFilters applies ASN-specific filters to the query
func applyASNFilters(query *gorm.DB, params *request.ASNSpecificRequest, asn string) *gorm.DB {
	if params == nil {
		return query
	}

	// Back route filter
	if params.BackRoute != nil && *params.BackRoute != "" {
		query = query.Where("ecs LIKE ?", "%"+*params.BackRoute+"%")
	}

	// Note: Download/Upload/Latency filters require parsing JSON data
	// For SQLite, we can use JSON functions or filter in application layer
	// Here we use basic LIKE queries as approximation
	if params.MinDownload != nil {
		// This is a simplified approach - actual implementation might need JSON parsing
		query = query.Where("spdtest LIKE ?", "%"+asn+"%")
	}
	if params.MaxDownload != nil {
		query = query.Where("spdtest LIKE ?", "%"+asn+"%")
	}
	if params.MinUpload != nil {
		query = query.Where("spdtest LIKE ?", "%"+asn+"%")
	}
	if params.MaxUpload != nil {
		query = query.Where("spdtest LIKE ?", "%"+asn+"%")
	}
	if params.Latency != nil {
		query = query.Where("spdtest LIKE ?", "%"+asn+"%")
	}

	return query
}

// ReportExists checks if a report with the given ID exists
func ReportExists(reportID string) (bool, error) {
	if db == nil {
		return false, errors.New("database not initialized")
	}
	var count int64
	err := db.Model(&model.BenchmarkResult{}).Where("report_id = ?", reportID).Count(&count).Error
	return count > 0, err
}
