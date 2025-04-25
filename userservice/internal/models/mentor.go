package models

import "time"

type Mentor struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FirstName   string    `gorm:"size:60"    json:"first_name"`
	LastName    string    `gorm:"size:60"    json:"last_name"`
	About       string    `gorm:"size:255"   json:"about"`
	Track       string    `gorm:"size:100"   json:"track"`
	Education   string    `gorm:"size:100"   json:"education"`
	ExperienceY uint      `json:"experience_years"`
	Age         uint      `json:"age"`
	Gender      string    `gorm:"size:10"    json:"gender"`
	AvatarURL   string    `gorm:"size:255"   json:"avatar_url"`
	ContactURL  string    `gorm:"size:255"   json:"contact_url"`
	CreatedAt   time.Time `json:"created_at"`
}
