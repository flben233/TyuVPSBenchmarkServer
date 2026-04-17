package model

import "time"

type WebsshCommandWhitelist struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	Commands  string    `gorm:"type:text;not null" json:"commands"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WebsshCommandWhitelist) TableName() string {
	return "webssh_command_whitelist"
}
