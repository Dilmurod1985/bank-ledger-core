package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"bank-ledger-core/models"
)

type AccountHandler struct {
	db *gorm.DB
}

func NewAccountHandler(db *gorm.DB) *AccountHandler {
	return &AccountHandler{db: db}
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var account models.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.db.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create account",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, account)
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	idStr := c.Param("id")
	var account models.Account
	var err error

	// Try to parse as numeric ID first
	id, parseErr := strconv.ParseUint(idStr, 10, 32)
	if parseErr == nil {
		// Search by numeric ID
		err = h.db.First(&account, id).Error
	} else {
		// Search by user_id (string)
		err = h.db.Where("user_id = ?", idStr).First(&account).Error
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Account not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch account",
		})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
	var accounts []models.Account
	if err := h.db.Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve accounts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, accounts)
}
