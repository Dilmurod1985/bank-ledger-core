package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"bank-ledger-core/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

type RegisterRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Currency string `json:"currency" binding:"required"`
}

type LoginRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message  string `json:"message"`
	UserID   string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Check if user already exists
	var existingAccount models.Account
	if err := h.db.Where("user_id = ?", req.UserID).First(&existingAccount).Error; err == nil {
		c.JSON(http.StatusConflict, AuthResponse{
			Success: false,
			Message: "User already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to hash password",
		})
		return
	}

	// Create account
	account := models.Account{
		UserID:       req.UserID,
		PasswordHash: string(hashedPassword),
		Currency:     req.Currency,
		Balance:      "100000.00", // Default balance
	}

	if err := h.db.Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to create account",
		})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Success: true,
		Message: "Account created successfully",
		UserID:  account.UserID,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Find user
	var account models.Account
	if err := h.db.Where("user_id = ?", req.UserID).First(&account).Error; err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Create session
	sessionID, err := generateSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to create session",
		})
		return
	}

	session := models.Session{
		ID:        sessionID,
		UserID:    account.UserID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours
	}

	if err := h.db.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to create session",
		})
		return
	}

	// Set cookie
	c.SetCookie("session_id", sessionID, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, AuthResponse{
		Success:   true,
		Message:   "Login successful",
		UserID:    account.UserID,
		SessionID: sessionID,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		h.db.Where("id = ?", sessionID).Delete(&models.Session{})
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Logout successful",
	})
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
