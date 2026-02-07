package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"bank-ledger-core/services"
)

type TransferHandler struct {
	transferService *services.TransferService
}

func NewTransferHandler(transferService *services.TransferService) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
	}
}

func (h *TransferHandler) TransferMoney(c *gin.Context) {
	var req services.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	response, err := h.transferService.TransferMoney(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *TransferHandler) TransferMoneyByUserIDs(c *gin.Context) {
	var req services.UserTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	response, err := h.transferService.TransferMoneyByUserIDs(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusOK, response)
}
