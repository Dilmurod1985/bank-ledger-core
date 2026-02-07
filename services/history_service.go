package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"bank-ledger-core/models"
)

type HistoryService struct {
	db *gorm.DB
}

func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

type HistoryItem struct {
	Date      time.Time `json:"date"`
	Type      string    `json:"type"`      // "Расход" или "Доход"
	Amount    string    `json:"amount"`
	Counterparty string  `json:"counterparty"` // user_id контрагента
	Reference string    `json:"reference"`     // ID операции (transfer_id или order_id)
}

type HistoryResponse struct {
	UserID    string        `json:"user_id"`
	History   []HistoryItem `json:"history"`
	Balance   string        `json:"balance"`
	Currency  string        `json:"currency"`
}

func (s *HistoryService) GetAccountHistory(userID string) (*HistoryResponse, error) {
	// Get user account
	var account models.Account
	if err := s.db.Where("user_id = ?", userID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var history []HistoryItem

	// Get transfers where user is sender (Расход)
	var sentTransfers []models.Transfer
	if err := s.db.Preload("ToAccount").Where("from_account_id = ?", account.ID).Find(&sentTransfers).Error; err != nil {
		return nil, fmt.Errorf("failed to get sent transfers: %w", err)
	}

	for _, transfer := range sentTransfers {
		history = append(history, HistoryItem{
			Date:         transfer.CreatedAt,
			Type:         "Расход",
			Amount:       transfer.Amount,
			Counterparty: transfer.ToAccount.UserID,
			Reference:    fmt.Sprintf("transfer_%d", transfer.ID),
		})
	}

	// Get transfers where user is recipient (Доход)
	var receivedTransfers []models.Transfer
	if err := s.db.Preload("FromAccount").Where("to_account_id = ?", account.ID).Find(&receivedTransfers).Error; err != nil {
		return nil, fmt.Errorf("failed to get received transfers: %w", err)
	}

	for _, transfer := range receivedTransfers {
		history = append(history, HistoryItem{
			Date:         transfer.CreatedAt,
			Type:         "Доход",
			Amount:       transfer.Amount,
			Counterparty: transfer.FromAccount.UserID,
			Reference:    fmt.Sprintf("transfer_%d", transfer.ID),
		})
	}

	// Get orders where user is buyer (Расход)
	var orders []models.Order
	if err := s.db.Preload("Product").Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	for _, order := range orders {
		history = append(history, HistoryItem{
			Date:         order.CreatedAt,
			Type:         "Расход",
			Amount:       order.Amount,
			Counterparty: "marketplace", // Системный аккаунт маркетплейса
			Reference:    fmt.Sprintf("order_%d", order.ID),
		})
	}

	// Sort by date (newest first)
	for i := 0; i < len(history)-1; i++ {
		for j := i + 1; j < len(history); j++ {
			if history[i].Date.Before(history[j].Date) {
				history[i], history[j] = history[j], history[i]
			}
		}
	}

	return &HistoryResponse{
		UserID:   userID,
		History:  history,
		Balance:  account.Balance,
		Currency: account.Currency,
	}, nil
}
