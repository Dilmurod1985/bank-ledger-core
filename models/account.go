package models

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       string         `gorm:"not null;index" json:"user_id"`
	PasswordHash string         `gorm:"not null" json:"-"`
	Currency     string         `gorm:"not null;size:3" json:"currency"`
	Balance      string         `gorm:"type:decimal(15,2);not null;default:0.00" json:"balance"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Account) TableName() string {
	return "accounts"
}
