package service

import (
	"VPSBenchmarkBackend/internal/report/model"
	"VPSBenchmarkBackend/internal/report/request"
	"VPSBenchmarkBackend/internal/report/response"
	"VPSBenchmarkBackend/internal/report/store"
	"fmt"
)

// ListReports returns a list of all reports with pagination
func ListReports(page, pageSize int) ([]response.ReportInfoResponse, int64, error) {
	results, total, err := store.ListReports(page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}

	// Convert to response format
	responses := make([]response.ReportInfoResponse, len(results))
	for i, r := range results {
		responses[i] = response.ReportInfoResponse{
			Name: r.Title,
			Id:   r.ReportID,
			Date: r.Time,
		}
	}

	return responses, total, nil
}

// GetReportDetails returns the full details of a specific report
func GetReportDetails(reportID string) (*model.BenchmarkResult, error) {
	if reportID == "" {
		return nil, fmt.Errorf("report ID is required")
	}

	result, err := store.GetReportByID(reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report details: %w", err)
	}

	return result, nil
}

// SearchReports performs a search based on the given criteria
func SearchReports(searchReq *request.SearchRequest, page, pageSize int) ([]response.ReportInfoResponse, int64, error) {
	if searchReq == nil {
		return ListReports(page, pageSize)
	}

	// Convert request ISP params to store ISP params
	var ctISP, cuISP, cmISP *store.ISPSearchParams
	if searchReq.CTParams != nil {
		ctISP = &store.ISPSearchParams{
			BackRoute:   searchReq.CTParams.BackRoute,
			MinDownload: searchReq.CTParams.MinDownload,
			MaxDownload: searchReq.CTParams.MaxDownload,
			MinUpload:   searchReq.CTParams.MinUpload,
			MaxUpload:   searchReq.CTParams.MaxUpload,
			Latency:     searchReq.CTParams.Latency,
		}
	}
	if searchReq.CUParams != nil {
		cuISP = &store.ISPSearchParams{
			BackRoute:   searchReq.CUParams.BackRoute,
			MinDownload: searchReq.CUParams.MinDownload,
			MaxDownload: searchReq.CUParams.MaxDownload,
			MinUpload:   searchReq.CUParams.MinUpload,
			MaxUpload:   searchReq.CUParams.MaxUpload,
			Latency:     searchReq.CUParams.Latency,
		}
	}
	if searchReq.CMParams != nil {
		cmISP = &store.ISPSearchParams{
			BackRoute:   searchReq.CMParams.BackRoute,
			MinDownload: searchReq.CMParams.MinDownload,
			MaxDownload: searchReq.CMParams.MaxDownload,
			MinUpload:   searchReq.CMParams.MinUpload,
			MaxUpload:   searchReq.CMParams.MaxUpload,
			Latency:     searchReq.CMParams.Latency,
		}
	}

	results, total, err := store.SearchReports(
		searchReq.Keyword,
		searchReq.MediaUnlocks,
		searchReq.Virtualization,
		searchReq.IPv6Support,
		searchReq.DiskLevel,
		ctISP, cuISP, cmISP,
		page, pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search reports: %w", err)
	}

	// Convert to response format
	responses := make([]response.ReportInfoResponse, len(results))
	for i, r := range results {
		responses[i] = response.ReportInfoResponse{
			Name: r.Title,
			Id:   r.ReportID,
			Date: r.Time,
		}
	}

	return responses, total, nil
}

// GetAllMediaNames returns all distinct media names
func GetAllMediaNames() ([]string, error) {
	return store.GetAllMediaNames()
}

// GetAllVirtualizations returns all distinct virtualization types
func GetAllVirtualizations() ([]string, error) {
	return store.GetAllVirtualizations()
}

// GetAllBackRouteTypes returns all distinct back route types
func GetAllBackRouteTypes() ([]string, error) {
	return store.GetAllBackRouteTypes()
}
