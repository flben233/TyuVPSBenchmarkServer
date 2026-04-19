package model

import "time"

type WebsshSync struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	EncryptedData string    `gorm:"type:text;not null" json:"encrypted_data"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (WebsshSync) TableName() string {
	return "webssh_sync"
}
