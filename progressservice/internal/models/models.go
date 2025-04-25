package models

import "time"

type PlanItem struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:255"  json:"title"`
	ResourceURL string    `gorm:"size:255"  json:"resource"`
	StartAt     time.Time `json:"start_at"`
	DueAt       time.Time `json:"due_at"`
	CreatedAt   time.Time `json:"-"`
}

type Progress struct {
	ID         uint64     `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index"       json:"-"`
	PlanItemID uint64     `gorm:"index"       json:"-"`
	Done       bool       `gorm:"default:false" json:"done"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}
