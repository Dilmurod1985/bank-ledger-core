package middleware

import (
	"net/http"

	"bank-ledger-core/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "No session found",
			})
			c.Abort()
			return
		}

		var session models.Session
		if err := db.Where("id = ?", sessionID).First(&session).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid session",
			})
			c.Abort()
			return
		}

		if session.IsExpired() {
			db.Where("id = ?", sessionID).Delete(&models.Session{})
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Session expired",
			})
			c.Abort()
			return
		}

		// Set user_id in context
		c.Set("user_id", session.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}
