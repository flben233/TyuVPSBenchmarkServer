package service

import (
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/lookingglass/response"
	"VPSBenchmarkBackend/internal/lookingglass/store"
)

type RecordLimitError struct{}

func (e *RecordLimitError) Error() string {
	return "Looking glass record limit reached"
}

func AddRecord(userID, username, serverName, testURL string, isAdmin bool) (int64, error) {
	if !isAdmin {
		cnt, err := store.CountUserRecords(userID)
		if err != nil {
			return 0, err
		}
		limit := config.Get().MaxHostsPerUser
		if cnt >= int64(limit) {
			return 0, &RecordLimitError{}
		}
	}
	return store.AddRecord(serverName, testURL, username, userID)
}

func UpdateRecord(userID string, id int64, serverName, testURL string, isAdmin bool) error {
	if isAdmin {
		return store.UpdateRecordAsAdmin(id, serverName, testURL)
	}
	return store.UpdateRecord(id, serverName, testURL, userID)
}

func RemoveRecord(userID string, id int64, isAdmin bool) error {
	if isAdmin {
		return store.RemoveRecordAsAdmin(id)
	}
	return store.RemoveRecord(id, userID)
}

func ListRecords(userID string, isAdmin bool) ([]response.LookingGlassResponse, error) {
	var records []store.LookingGlass
	var err error

	if isAdmin {
		records, err = store.ListAllRecords()
	} else {
		records, err = store.ListRecordsByUploader(userID)
	}

	if err != nil {
		return nil, err
	}

	result := make([]response.LookingGlassResponse, len(records))
	for i, record := range records {
		result[i] = response.LookingGlassResponse{
			Id:           record.Id,
			ServerName:   record.ServerName,
			TestURL:      record.TestURL,
			UploaderName: record.UploaderName,
			ReviewStatus: int(record.ReviewStatus),
		}
	}
	return result, nil
}

func ListAllRecords() ([]response.LookingGlassResponse, error) {
	records, err := store.ListAllRecords()
	if err != nil {
		return nil, err
	}

	result := make([]response.LookingGlassResponse, len(records))
	for i, record := range records {
		result[i] = response.LookingGlassResponse{
			Id:           record.Id,
			ServerName:   record.ServerName,
			TestURL:      record.TestURL,
			UploaderName: record.UploaderName,
		}
	}
	return result, nil
}
