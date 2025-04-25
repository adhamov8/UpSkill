package models

import (
	"time"
)

type Profile struct {
	ID        uint      `gorm:"primaryKey"         json:"id"`
	UserID    uint      `gorm:"uniqueIndex"        json:"user_id"`
	Track     string    `gorm:"size:100"           json:"track"`
	Goal      string    `gorm:"size:255"           json:"goal"`
	AvatarURL string    `gorm:"size:255"           json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Progress struct {
	ID         uint64 `gorm:"primaryKey"`
	UserID     uint64
	PlanItemID uint64
	Done       bool
	FinishedAt *time.Time
}
