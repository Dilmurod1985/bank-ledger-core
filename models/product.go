package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Price       string         `gorm:"type:decimal(15,2);not null" json:"price"`
	Stock       int            `gorm:"not null;default:0" json:"stock"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Product) TableName() string {
	return "products"
}
