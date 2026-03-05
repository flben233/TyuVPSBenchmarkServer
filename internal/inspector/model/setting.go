package model

type InspectorSetting struct {
	ID        int64   `gorm:"primaryKey" json:"id"`
	UserID    int64   `json:"user_id"`
	NotifyURL *string `json:"notify_url"`
	BgURL     *string `json:"bg_url"`
}
