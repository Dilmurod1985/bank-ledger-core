package models

import (
	"time"

	"gorm.io/gorm"
)

type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "pending"
	TransferStatusCompleted TransferStatus = "completed"
	TransferStatusFailed    TransferStatus = "failed"
)

type Transfer struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	FromAccountID uint          `gorm:"not null;index" json:"from_account_id"`
	ToAccountID   uint          `gorm:"not null;index" json:"to_account_id"`
	Amount        string        `gorm:"type:decimal(15,2);not null" json:"amount"`
	Status        TransferStatus `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	FromAccount Account `gorm:"foreignKey:FromAccountID" json:"from_account,omitempty"`
	ToAccount   Account `gorm:"foreignKey:ToAccountID" json:"to_account,omitempty"`
}

func (Transfer) TableName() string {
	return "transfers"
}
