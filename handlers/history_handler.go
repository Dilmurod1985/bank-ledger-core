package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"bank-ledger-core/services"
)

type HistoryHandler struct {
	historyService *services.HistoryService
}

func NewHistoryHandler(historyService *services.HistoryService) *HistoryHandler {
	return &HistoryHandler{
		historyService: historyService,
	}
}

func (h *HistoryHandler) GetAccountHistory(c *gin.Context) {
	userID := c.Param("user_id")
	
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	response, err := h.historyService.GetAccountHistory(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
