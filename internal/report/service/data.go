package service

import (
	"VPSBenchmarkBackend/internal/report/model"
	"VPSBenchmarkBackend/internal/report/request"
	"VPSBenchmarkBackend/internal/report/response"
	"VPSBenchmarkBackend/internal/report/store"
	"errors"
)

// ListReports returns a list of all reports with pagination
func ListReports(page, pageSize int) ([]response.ReportInfoResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	reports, total, err := store.ListReports(pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response format
	var result []response.ReportInfoResponse
	for _, report := range reports {
		result = append(result, response.ReportInfoResponse{
			Name: report.Title,
			Id:   report.ReportID,
			Date: report.Time,
		})
	}

	return result, total, nil
}

// GetReportDetails returns the full details of a specific report
func GetReportDetails(reportID string) (*model.BenchmarkResult, error) {
	if reportID == "" {
		return nil, errors.New("report ID is required")
	}

	report, err := store.GetReportByID(reportID)
	if err != nil {
		return nil, err
	}

	return report, nil
}

// SearchReports performs a search based on the given criteria
func SearchReports(searchReq *request.SearchRequest, page, pageSize int) ([]response.ReportInfoResponse, int64, error) {
	if searchReq == nil {
		return nil, 0, errors.New("search request is required")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	reports, total, err := store.SearchReports(searchReq, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Convert to response format
	var result []response.ReportInfoResponse
	for _, report := range reports {
		result = append(result, response.ReportInfoResponse{
			Name: report.Title,
			Id:   report.ReportID,
			Date: report.Time,
		})
	}

	return result, total, nil
}
