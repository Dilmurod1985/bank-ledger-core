package services

import (
	"errors"
	"fmt"
	"math/big"

	"gorm.io/gorm"
	"bank-ledger-core/models"
)

type TransferService struct {
	db *gorm.DB
}

func NewTransferService(db *gorm.DB) *TransferService {
	return &TransferService{db: db}
}

type TransferRequest struct {
	FromAccountID uint   `json:"from_account_id" binding:"required"`
	ToAccountID   uint   `json:"to_account_id" binding:"required"`
	Amount        string `json:"amount" binding:"required"`
}

type UserTransferRequest struct {
	FromUserID string `json:"from_user_id" binding:"required"`
	ToUserID   string `json:"to_user_id" binding:"required"`
	Amount     string `json:"amount" binding:"required"`
}

type TransferResponse struct {
	TransferID uint   `json:"transfer_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

func (s *TransferService) TransferMoney(req TransferRequest) (*TransferResponse, error) {
	if req.FromAccountID == req.ToAccountID {
		return nil, errors.New("cannot transfer to the same account")
	}

	var result *TransferResponse

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var fromAccount, toAccount models.Account

		if err := tx.First(&fromAccount, req.FromAccountID).Error; err != nil {
			return fmt.Errorf("sender account not found: %w", err)
		}

		if err := tx.First(&toAccount, req.ToAccountID).Error; err != nil {
			return fmt.Errorf("recipient account not found: %w", err)
		}

		if fromAccount.Currency != toAccount.Currency {
			return errors.New("currency mismatch between accounts")
		}

		fromBalance, err := parseDecimal(fromAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid sender balance: %w", err)
		}

		amount, err := parseDecimal(req.Amount)
		if err != nil {
			return fmt.Errorf("invalid transfer amount: %w", err)
		}

		if fromBalance.Cmp(amount) < 0 {
			return errors.New("insufficient funds")
		}

		newFromBalance := new(big.Float).Sub(fromBalance, amount)
		newToBalance, err := parseDecimal(toAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid recipient balance: %w", err)
		}
		newToBalance = new(big.Float).Add(newToBalance, amount)

		if err := tx.Model(&fromAccount).Update("balance", newFromBalance.String()).Error; err != nil {
			return fmt.Errorf("failed to update sender balance: %w", err)
		}

		if err := tx.Model(&toAccount).Update("balance", newToBalance.String()).Error; err != nil {
			return fmt.Errorf("failed to update recipient balance: %w", err)
		}

		transfer := models.Transfer{
			FromAccountID: req.FromAccountID,
			ToAccountID:   req.ToAccountID,
			Amount:        req.Amount,
			Status:        models.TransferStatusCompleted,
		}

		if err := tx.Create(&transfer).Error; err != nil {
			return fmt.Errorf("failed to create transfer record: %w", err)
		}

		result = &TransferResponse{
			TransferID: transfer.ID,
			Status:     string(transfer.Status),
			Message:    "Transfer completed successfully",
		}

		return nil
	})

	if err != nil {
		return &TransferResponse{
			Status:  "failed",
			Message: err.Error(),
		}, err
	}

	return result, nil
}

func (s *TransferService) TransferMoneyByUserIDs(req UserTransferRequest) (*TransferResponse, error) {
	if req.FromUserID == req.ToUserID {
		return nil, errors.New("cannot transfer to the same user")
	}

	var result *TransferResponse

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Find from account
		var fromAccount models.Account
		if err := tx.Where("user_id = ?", req.FromUserID).First(&fromAccount).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("sender account not found")
			}
			return fmt.Errorf("failed to find sender account: %w", err)
		}

		// Find to account
		var toAccount models.Account
		if err := tx.Where("user_id = ?", req.ToUserID).First(&toAccount).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("recipient account not found")
			}
			return fmt.Errorf("failed to find recipient account: %w", err)
		}

		// Check currency match
		if fromAccount.Currency != toAccount.Currency {
			return errors.New("currency mismatch between accounts")
		}

		// Parse amounts
		fromBalance, err := parseDecimal(fromAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid sender balance: %w", err)
		}

		amount, err := parseDecimal(req.Amount)
		if err != nil {
			return fmt.Errorf("invalid transfer amount: %w", err)
		}

		// Check sufficient funds
		if fromBalance.Cmp(amount) < 0 {
			return errors.New("insufficient funds")
		}

		// Calculate new balances
		newFromBalance := new(big.Float).Sub(fromBalance, amount)
		newToBalance, err := parseDecimal(toAccount.Balance)
		if err != nil {
			return fmt.Errorf("invalid recipient balance: %w", err)
		}
		newToBalance = new(big.Float).Add(newToBalance, amount)

		// Update balances
		if err := tx.Model(&fromAccount).Update("balance", newFromBalance.String()).Error; err != nil {
			return fmt.Errorf("failed to update sender balance: %w", err)
		}

		if err := tx.Model(&toAccount).Update("balance", newToBalance.String()).Error; err != nil {
			return fmt.Errorf("failed to update recipient balance: %w", err)
		}

		// Create transfer record
		transfer := models.Transfer{
			FromAccountID: fromAccount.ID,
			ToAccountID:   toAccount.ID,
			Amount:        amount.String(),
			Status:        models.TransferStatusCompleted,
		}

		if err := tx.Create(&transfer).Error; err != nil {
			return fmt.Errorf("failed to create transfer record: %w", err)
		}

		result = &TransferResponse{
			TransferID: transfer.ID,
			Status:     string(transfer.Status),
			Message:    "Transfer completed successfully",
		}

		return nil
	})

	if err != nil {
		return &TransferResponse{
			Status:  "failed",
			Message: err.Error(),
		}, err
	}

	return result, nil
}
