package models

import "time"

/* ---------- роли ---------- */

type Role string

const (
	RoleUser   Role = "USER"
	RoleMentor Role = "MENTOR"
	RoleAdmin  Role = "ADMIN"
)

/* ---------- модели ---------- */

type User struct {
	ID            uint   `gorm:"primaryKey"`
	Email         string `gorm:"uniqueIndex"`
	EmailVerified bool   `gorm:"default:false"`
	PasswordHash  string `gorm:"column:password_hash"`
	FirstName     string `gorm:"size:60"`
	LastName      string `gorm:"size:60"`
	AvatarURL     string `gorm:"size:255"`
	Role          Role   `gorm:"type:varchar(10)"`
	CreatedAt     time.Time
}

type RefreshToken struct {
	ID        uint   `gorm:"primaryKey"`
	Token     string `gorm:"uniqueIndex;size:512"`
	UserID    uint   `gorm:"index"`
	ExpiresAt time.Time
	Revoked   bool `gorm:"default:false"`
	CreatedAt time.Time
}
