package model

type User struct {
	ID      int64  `gorm:"primaryKey;autoIncrement:false" json:"id"` // Github ID
	Name    string `json:"name"`
	Login   string `json:"login"`
	GroupID uint32 `json:"group_id"`
}

type UserGroup struct {
	ID           uint32 `gorm:"primaryKey" json:"id"`
	Name         string `json:"name"`
	IsAdmin      bool   `json:"is_admin"`
	InspectorNum uint32 `json:"inspector_num"`
	MaxHostNum   uint32 `json:"max_host_num"`
}
