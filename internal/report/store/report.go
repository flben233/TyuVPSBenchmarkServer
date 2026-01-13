package store

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/report/model"
	"context"
	"fmt"

	"gorm.io/gorm"
)

var db *gorm.DB
var (
	benchmarkResults gorm.Interface[model.BenchmarkResult]
	mediaIndex       gorm.Interface[model.MediaIndex]
	speedtestIndex   gorm.Interface[model.SpeedtestIndex]
	ecsIndex         gorm.Interface[model.InfoIndex]
	backtraceIndex   gorm.Interface[model.BacktraceIndex]
)

func init() {
	// Register the initializer
	common.RegisterDBInitializer(InitReportStore)
}

// InitReportStore initializes the tables
func InitReportStore(dbPath string) error {
	db = common.GetDB()
	// Auto migrate the schema
	if err := db.AutoMigrate(&model.BenchmarkResult{}, &model.MediaIndex{}, &model.SpeedtestIndex{}, &model.InfoIndex{}, &model.BacktraceIndex{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	benchmarkResults = gorm.G[model.BenchmarkResult](db)
	mediaIndex = gorm.G[model.MediaIndex](db)
	speedtestIndex = gorm.G[model.SpeedtestIndex](db)
	ecsIndex = gorm.G[model.InfoIndex](db)
	backtraceIndex = gorm.G[model.BacktraceIndex](db)
	return nil
}

// SaveReport saves a new benchmark report to database
func SaveReport(report *model.BenchmarkResult, mi []model.MediaIndex, si []model.SpeedtestIndex, ei *model.InfoIndex, bi []model.BacktraceIndex) error {
	ctx := context.Background()
	err := benchmarkResults.Create(ctx, report)
	for _, m := range mi {
		if err == nil {
			err = mediaIndex.Create(ctx, &m)
		}
	}
	for _, s := range si {
		if err == nil {
			err = speedtestIndex.Create(ctx, &s)
		}
	}
	if err == nil {
		err = ecsIndex.Create(ctx, ei)
	}
	for _, b := range bi {
		if err == nil {
			err = backtraceIndex.Create(ctx, &b)
		}
	}
	return err
}

// ReportExists checks if a report with the given ReportID exists
func ReportExists(reportID string) (bool, error) {
	count, err := benchmarkResults.Where("report_id = ?", reportID).Count(context.Background(), "*")
	return count > 0, err
}

// DeleteReport deletes a report and all related index data by ReportID
func DeleteReport(reportID string) error {
	ctx := context.Background()

	// Delete related index data first
	if _, err := mediaIndex.Where("report_id = ?", reportID).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete media index: %w", err)
	}
	if _, err := speedtestIndex.Where("report_id = ?", reportID).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete speedtest index: %w", err)
	}
	if _, err := ecsIndex.Where("report_id = ?", reportID).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete info index: %w", err)
	}
	if _, err := backtraceIndex.Where("report_id = ?", reportID).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete backtrace index: %w", err)
	}

	// Delete the main report
	if _, err := benchmarkResults.Where("report_id = ?", reportID).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}

// ListReports returns a paginated list of reports
func ListReports(page, pageSize int) ([]model.BenchmarkResult, int64, error) {
	ctx := context.Background()

	// Get total count
	total, err := benchmarkResults.Count(ctx, "*")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get paginated results
	results, err := benchmarkResults.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}

	return results, total, nil
}

// GetReportByID returns a report by its ReportID
func GetReportByID(reportID string) (*model.BenchmarkResult, error) {
	ctx := context.Background()
	result, err := benchmarkResults.Where("report_id = ?", reportID).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	return &result, nil
}

// SearchReports searches reports based on various criteria
func SearchReports(keyword *string, mediaUnlocks []string, virtualization *string, ipv6Support *bool, diskLevel *int,
	ctISP, cuISP, cmISP *ISPSearchParams, page, pageSize int) ([]model.BenchmarkResult, int64, error) {

	// Start with base query
	query := db.Model(&model.BenchmarkResult{})

	// Keyword search on title
	if keyword != nil && *keyword != "" {
		query = query.Where("title LIKE ?", "%"+*keyword+"%")
	}

	// Collect report IDs that match all criteria
	var reportIDs []string
	needFilter := false

	// Media unlock filter
	if len(mediaUnlocks) > 0 {
		needFilter = true
		var mediaReportIDs []string
		for _, media := range mediaUnlocks {
			var ids []string
			db.Model(&model.MediaIndex{}).Where("media = ? AND unlock = TRUE", media).Distinct().Pluck("report_id", &ids)
			if len(mediaReportIDs) == 0 {
				mediaReportIDs = ids
			} else {
				// Intersection
				mediaReportIDs = intersect(mediaReportIDs, ids)
			}
		}
		reportIDs = mediaReportIDs
	}

	// Virtualization filter
	if virtualization != nil && *virtualization != "" {
		needFilter = true
		var virtIDs []string
		db.Model(&model.InfoIndex{}).Where("virtualization = ?", *virtualization).Pluck("report_id", &virtIDs)
		if len(reportIDs) == 0 {
			reportIDs = virtIDs
		} else {
			reportIDs = intersect(reportIDs, virtIDs)
		}
	}

	// IPv6 support filter
	if ipv6Support != nil {
		needFilter = true
		var ipv6IDs []string
		db.Model(&model.InfoIndex{}).Where("ipv6_support = ?", *ipv6Support).Pluck("report_id", &ipv6IDs)
		if len(reportIDs) == 0 {
			reportIDs = ipv6IDs
		} else {
			reportIDs = intersect(reportIDs, ipv6IDs)
		}
	}

	// Disk level filter (0-5: <100, 100-200, 200-400, 400-600, 600-1000, >1000 MB/s average)
	if diskLevel != nil {
		needFilter = true
		var diskIDs []string
		diskQuery := db.Model(&model.InfoIndex{})
		avgExpr := "(seq_read + seq_write) / 2"
		switch *diskLevel {
		case 0:
			diskQuery = diskQuery.Where(avgExpr+" < ?", 100)
		case 1:
			diskQuery = diskQuery.Where(avgExpr+" >= ? AND "+avgExpr+" < ?", 100, 200)
		case 2:
			diskQuery = diskQuery.Where(avgExpr+" >= ? AND "+avgExpr+" < ?", 200, 400)
		case 3:
			diskQuery = diskQuery.Where(avgExpr+" >= ? AND "+avgExpr+" < ?", 400, 600)
		case 4:
			diskQuery = diskQuery.Where(avgExpr+" >= ? AND "+avgExpr+" < ?", 600, 1000)
		case 5:
			diskQuery = diskQuery.Where(avgExpr+" >= ?", 1000)
		}
		diskQuery.Pluck("report_id", &diskIDs)
		if len(reportIDs) == 0 {
			reportIDs = diskIDs
		} else {
			reportIDs = intersect(reportIDs, diskIDs)
		}
	}

	// ISP speed filters
	ispFilters := []struct {
		isp    string
		params *ISPSearchParams
	}{
		{model.ISPChinaTelecom, ctISP},
		{model.ISPChinaUnicom, cuISP},
		{model.ISPChinaMobile, cmISP},
	}

	for _, f := range ispFilters {
		if f.params != nil {
			needFilter = true
			var ispIDs []string

			// Filter by speedtest index
			hasSpeedFilter := f.params.MinDownload != nil || f.params.MaxDownload != nil ||
				f.params.MinUpload != nil || f.params.MaxUpload != nil || f.params.Latency != nil
			if hasSpeedFilter {
				ispQuery := db.Model(&model.SpeedtestIndex{}).Where("isp = ?", f.isp)
				if f.params.MinDownload != nil {
					ispQuery = ispQuery.Where("download >= ?", *f.params.MinDownload)
				}
				if f.params.MaxDownload != nil {
					ispQuery = ispQuery.Where("download <= ?", *f.params.MaxDownload)
				}
				if f.params.MinUpload != nil {
					ispQuery = ispQuery.Where("upload >= ?", *f.params.MinUpload)
				}
				if f.params.MaxUpload != nil {
					ispQuery = ispQuery.Where("upload <= ?", *f.params.MaxUpload)
				}
				if f.params.Latency != nil {
					ispQuery = ispQuery.Where("latency <= ?", *f.params.Latency)
				}
				ispQuery.Distinct().Pluck("report_id", &ispIDs)
				if len(reportIDs) == 0 {
					reportIDs = ispIDs
				} else {
					reportIDs = intersect(reportIDs, ispIDs)
				}
			}

			// Filter by backtrace index for back route
			if f.params.BackRoute != nil && *f.params.BackRoute != "" {
				var backtraceIDs []string
				db.Model(&model.BacktraceIndex{}).Where("isp = ? AND route_type LIKE ?", f.isp, "%"+*f.params.BackRoute+"%").Distinct().Pluck("report_id", &backtraceIDs)
				if len(reportIDs) == 0 {
					reportIDs = backtraceIDs
				} else {
					reportIDs = intersect(reportIDs, backtraceIDs)
				}
			}
		}
	}

	// Apply report ID filter if any index-based filters were used
	if needFilter {
		if len(reportIDs) == 0 {
			return []model.BenchmarkResult{}, 0, nil
		}
		query = query.Where("report_id IN ?", reportIDs)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Calculate offset and get paginated results
	offset := (page - 1) * pageSize
	var results []model.BenchmarkResult
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search reports: %w", err)
	}

	return results, total, nil
}

// ISPSearchParams holds search parameters for ISP-specific queries
type ISPSearchParams struct {
	BackRoute   *string
	MinDownload *float32
	MaxDownload *float32
	MinUpload   *float32
	MaxUpload   *float32
	Latency     *float32
}

// intersect returns the intersection of two string slices
func intersect(a, b []string) []string {
	m := make(map[string]bool)
	for _, v := range a {
		m[v] = true
	}
	var result []string
	for _, v := range b {
		if m[v] {
			result = append(result, v)
		}
	}
	return result
}

// GetAllMediaNames returns all distinct media names from MediaIndex
func GetAllMediaNames() ([]string, error) {
	var mediaNames []string
	if err := db.Model(&model.MediaIndex{}).Distinct().Pluck("media", &mediaNames).Error; err != nil {
		return nil, fmt.Errorf("failed to get media names: %w", err)
	}
	return mediaNames, nil
}

// GetAllVirtualizations returns all distinct virtualization types from InfoIndex
func GetAllVirtualizations() ([]string, error) {
	var virtualizations []string
	if err := db.Model(&model.InfoIndex{}).Where("virtualization != ''").Distinct().Pluck("virtualization", &virtualizations).Error; err != nil {
		return nil, fmt.Errorf("failed to get virtualizations: %w", err)
	}
	return virtualizations, nil
}

// GetAllBackRouteTypes returns all distinct route types from BacktraceIndex
func GetAllBackRouteTypes() ([]string, error) {
	var routeTypes []string
	if err := db.Model(&model.BacktraceIndex{}).Where("route_type != ''").Distinct().Pluck("route_type", &routeTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to get back route types: %w", err)
	}
	return routeTypes, nil
}
