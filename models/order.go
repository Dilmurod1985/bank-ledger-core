package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "pending"
	OrderStatusPaid    OrderStatus = "paid"
	OrderStatusFailed  OrderStatus = "failed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	UserID    string      `gorm:"not null;index" json:"user_id"`
	ProductID uint        `gorm:"not null;index" json:"product_id"`
	Amount    string      `gorm:"type:decimal(15,2);not null" json:"amount"`
	Quantity  int         `gorm:"not null" json:"quantity"`
	Status    OrderStatus `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}
