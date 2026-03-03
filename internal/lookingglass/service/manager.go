package service

import (
	authUtil "VPSBenchmarkBackend/internal/auth/util"
	"VPSBenchmarkBackend/internal/lookingglass/response"
	"VPSBenchmarkBackend/internal/lookingglass/store"
)

type RecordLimitError struct{}

func (e *RecordLimitError) Error() string {
	return "Looking glass record limit reached"
}

func AddRecord(userID int64, username, serverName, testURL string) (int64, error) {
	if authUtil.IsAdmin(userID) {
		cnt, err := store.CountUserRecords(userID)
		if err != nil {
			return 0, err
		}
		if authUtil.CheckHostQuota(userID, cnt) {
			return 0, &RecordLimitError{}
		}
	}
	return store.AddRecord(serverName, testURL, username, userID)
}

func UpdateRecord(userID int64, id int64, serverName, testURL string) error {
	if authUtil.IsAdmin(userID) {
		return store.UpdateRecordAsAdmin(id, serverName, testURL)
	}
	return store.UpdateRecord(id, serverName, testURL, userID)
}

func RemoveRecord(userID int64, id int64) error {
	if authUtil.IsAdmin(userID) {
		return store.RemoveRecordAsAdmin(id)
	}
	return store.RemoveRecord(id, userID)
}

func ListRecords(userID int64) ([]response.LookingGlassResponse, error) {
	var records []store.LookingGlass
	var err error

	if authUtil.IsAdmin(userID) {
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
