package store

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/lookingglass/model"
	"context"

	"gorm.io/gorm"
)

const LookingGlassTableName = "looking_glass"

type LookingGlass struct {
	Id           int64               `gorm:"primaryKey"`
	ServerName   string              `gorm:"index"`
	TestURL      string
	Uploader     string
	UploaderName string
	ReviewStatus common.ReviewStatus `gorm:"default:0;index"`
}

func (LookingGlass) TableName() string {
	return LookingGlassTableName
}

var db *gorm.DB
var lookingGlassRecords gorm.Interface[LookingGlass]

func init() {
	// Register the initializer
	common.RegisterDBInitializer(InitLookingGlassStore)
}

func InitLookingGlassStore(dbPath string) error {
	db = common.GetDB()
	lookingGlassRecords = gorm.G[LookingGlass](db)
	if err := db.AutoMigrate(&LookingGlass{}); err != nil {
		return err
	}
	return nil
}

func CountUserRecords(userID string) (int64, error) {
	cnt, err := lookingGlassRecords.Where("uploader = ?", userID).Count(context.Background(), "*")
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

func AddRecord(serverName, testURL, username, userID string) (int64, error) {
	record := LookingGlass{
		ServerName:   serverName,
		TestURL:      testURL,
		Uploader:     userID,
		UploaderName: username,
		ReviewStatus: common.ReviewStatusPending,
	}
	err := lookingGlassRecords.Create(context.Background(), &record)
	if err != nil {
		return 0, err
	}
	return record.Id, nil
}

func UpdateRecord(id int64, serverName, testURL, uploader string) error {
	record, err := GetRecord(id)
	if err != nil {
		return err
	}
	if record == nil || record.Uploader != uploader {
		return gorm.ErrRecordNotFound
	}

	record.ServerName = serverName
	record.TestURL = testURL

	affected, err := lookingGlassRecords.Where("id = ? AND uploader = ?", id, uploader).Updates(context.Background(), *record)
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func UpdateRecordAsAdmin(id int64, serverName, testURL string) error {
	record, err := GetRecord(id)
	if err != nil {
		return err
	}
	if record == nil {
		return gorm.ErrRecordNotFound
	}

	record.ServerName = serverName
	record.TestURL = testURL

	affected, err := lookingGlassRecords.Where("id = ?", id).Updates(context.Background(), *record)
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func RemoveRecord(id int64, uploader string) error {
	affected, err := lookingGlassRecords.Where("id = ? AND uploader = ?", id, uploader).Delete(context.Background())
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func RemoveRecordAsAdmin(id int64) error {
	affected, err := lookingGlassRecords.Where("id = ?", id).Delete(context.Background())
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ListRecordsByUploader(uploader string) ([]LookingGlass, error) {
	records, err := lookingGlassRecords.Where("uploader = ?", uploader).Find(context.Background())
	return records, err
}

func ListAllRecords() ([]LookingGlass, error) {
	records, err := lookingGlassRecords.Where("review_status = ?", common.ReviewStatusApproved).Find(context.Background())
	return records, err
}

func GetRecord(id int64) (*LookingGlass, error) {
	records, err := lookingGlassRecords.Where("id = ?", id).Find(context.Background())
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

func RecordToModel(record LookingGlass) model.LookingGlass {
	return model.LookingGlass{
		Id:           record.Id,
		ServerName:   record.ServerName,
		TestURL:      record.TestURL,
		Uploader:     record.Uploader,
		ReviewStatus: record.ReviewStatus,
	}
}

// ListPendingRecords lists all records awaiting review
func ListPendingRecords() ([]LookingGlass, error) {
	records, err := lookingGlassRecords.Where("review_status = ?", common.ReviewStatusPending).Find(context.Background())
	return records, err
}

// UpdateReviewStatus updates the review status of a record
func UpdateReviewStatus(id int64, status common.ReviewStatus) error {
	affected, err := lookingGlassRecords.Where("id = ?", id).Update(context.Background(), "review_status", status)
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
