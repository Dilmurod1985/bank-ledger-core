package services

import (
	"errors"
	"fmt"
	"math/big"

	"gorm.io/gorm"
	"bank-ledger-core/models"
)

type OrderService struct {
	db               *gorm.DB
	transferService  *TransferService
}

func NewOrderService(db *gorm.DB, transferService *TransferService) *OrderService {
	return &OrderService{
		db:              db,
		transferService: transferService,
	}
}

type CreateOrderRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	ProductID uint   `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type CreateOrderResponse struct {
	OrderID uint   `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (s *OrderService) CreateOrder(req CreateOrderRequest) (*CreateOrderResponse, error) {
	var result *CreateOrderResponse

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, req.ProductID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("product not found")
			}
			return fmt.Errorf("failed to lock product: %w", err)
		}

		if product.Stock < req.Quantity {
			return errors.New("out of stock")
		}

		productPrice, err := parseDecimal(product.Price)
		if err != nil {
			return fmt.Errorf("invalid product price: %w", err)
		}

		quantity := new(big.Float).SetInt64(int64(req.Quantity))
		totalAmount := new(big.Float).Mul(productPrice, quantity)

		var userAccount models.Account
		if err := tx.Where("user_id = ?", req.UserID).First(&userAccount).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("user account not found")
			}
			return fmt.Errorf("failed to find user account: %w", err)
		}

		var systemAccount models.Account
		if err := tx.Where("user_id = ?", "0").First(&systemAccount).Error; err != nil {
			return fmt.Errorf("failed to find system account: %w", err)
		}

		// Perform money transfer within the same transaction
		fromBalance, err := parseDecimal(userAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid sender balance: %w", err)
		}

		if fromBalance.Cmp(totalAmount) < 0 {
			return errors.New("insufficient funds")
		}

		newFromBalance := new(big.Float).Sub(fromBalance, totalAmount)
		newToBalance, err := parseDecimal(systemAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid recipient balance: %w", err)
		}
		newToBalance = new(big.Float).Add(newToBalance, totalAmount)

		// Update user balance
		if err := tx.Model(&userAccount).Update("balance", newFromBalance.String()).Error; err != nil {
			return fmt.Errorf("failed to update sender balance: %w", err)
		}

		// Update system account balance
		if err := tx.Model(&systemAccount).Update("balance", newToBalance.String()).Error; err != nil {
			return fmt.Errorf("failed to update recipient balance: %w", err)
		}

		// Create transfer record
		transfer := models.Transfer{
			FromAccountID: userAccount.ID,
			ToAccountID:   systemAccount.ID,
			Amount:        totalAmount.String(),
			Status:        models.TransferStatusCompleted,
		}

		if err := tx.Create(&transfer).Error; err != nil {
			return fmt.Errorf("failed to create transfer record: %w", err)
		}

		// Update product stock
		newStock := product.Stock - req.Quantity
		if err := tx.Model(&product).Update("stock", newStock).Error; err != nil {
			return fmt.Errorf("failed to update product stock: %w", err)
		}

		// Create order
		order := models.Order{
			UserID:    req.UserID,
			ProductID: req.ProductID,
			Amount:    totalAmount.String(),
			Quantity:  req.Quantity,
			Status:    models.OrderStatusPaid,
		}

		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		result = &CreateOrderResponse{
			OrderID: order.ID,
			Status:  string(order.Status),
			Message: "Order created successfully",
		}

		return nil
	})

	if err != nil {
		return &CreateOrderResponse{
			Status:  "failed",
			Message: err.Error(),
		}, err
	}

	return result, nil
}

func (s *OrderService) GetOrdersByUserID(userID string) ([]models.Order, error) {
	var orders []models.Order
	if err := s.db.Where("user_id = ?", userID).Preload("Product").Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve orders: %w", err)
	}
	return orders, nil
}

func (s *OrderService) GetOrderByID(orderID uint) (*models.Order, error) {
	var order models.Order
	if err := s.db.Preload("Product").First(&order, orderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}
