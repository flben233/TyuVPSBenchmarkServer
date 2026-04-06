package service

import (
	authUtil "VPSBenchmarkBackend/internal/auth/util"
	"VPSBenchmarkBackend/internal/cache"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/lookingglass/response"
	"VPSBenchmarkBackend/internal/lookingglass/store"
)

func AddRecord(userID int64, username, serverName, testURL string) (int64, error) {
	if !authUtil.IsAdmin(userID) {
		cnt, err := store.CountUserRecords(userID)
		if err != nil {
			return 0, err
		}
		if authUtil.CheckHostQuota(userID, cnt) {
			return 0, &common.LimitExceededError{Message: "Looking glass record limit reached"}
		}
	}
	record, err := store.AddRecord(serverName, testURL, username, userID)
	if err != nil {
		return 0, err
	}
	return record, cache.PurgeSouinCache(cache.LookingGlassKey)
}

func UpdateRecord(userID int64, id int64, serverName, testURL string) error {
	var err error
	if authUtil.IsAdmin(userID) {
		err = store.UpdateRecordAsAdmin(id, serverName, testURL)
	} else {
		err = store.UpdateRecord(id, serverName, testURL, userID)
	}
	if err != nil {
		return err
	}
	return cache.PurgeSouinCache(cache.LookingGlassKey)
}

func RemoveRecord(userID int64, id int64) error {
	var err error
	if authUtil.IsAdmin(userID) {
		err = store.RemoveRecordAsAdmin(id)
	} else {
		err = store.RemoveRecord(id, userID)
	}
	if err != nil {
		return err
	}
	return cache.PurgeSouinCache(cache.LookingGlassKey)
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
