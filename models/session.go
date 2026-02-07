package models

import (
	"time"
)

type Session struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"not null;index" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Session) TableName() string {
	return "sessions"
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
