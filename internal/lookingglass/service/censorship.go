package service

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/lookingglass/response"
	"VPSBenchmarkBackend/internal/lookingglass/store"
)

// ListPendingRecords lists all records awaiting review (admin only)
func ListPendingRecords() ([]response.LookingGlassResponse, error) {
	records, err := store.ListPendingRecords()
	if err != nil {
		return nil, err
	}

	result := make([]response.LookingGlassResponse, len(records))
	for i, record := range records {
		result[i] = recordToResponse(record)
	}
	return result, nil
}

// ApproveRecord approves a record for public display (admin only)
func ApproveRecord(id int64) error {
	return store.UpdateReviewStatus(id, common.ReviewStatusApproved)
}

// RejectRecord rejects a record (admin only)
func RejectRecord(id int64) error {
	return store.UpdateReviewStatus(id, common.ReviewStatusRejected)
}

func recordToResponse(record store.LookingGlass) response.LookingGlassResponse {
	return response.LookingGlassResponse{
		Id:           record.Id,
		ServerName:   record.ServerName,
		TestURL:      record.TestURL,
		UploaderName: record.UploaderName,
		ReviewStatus: int(record.ReviewStatus),
	}
}
