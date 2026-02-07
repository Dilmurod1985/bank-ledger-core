package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"bank-ledger-core/services"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req services.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	result, err := h.orderService.CreateOrder(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Message,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"order_id": result.OrderID,
		"status":   result.Status,
		"message":  result.Message,
	})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id query parameter is required",
		})
		return
	}

	orders, err := h.orderService.GetOrdersByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve orders",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	order, err := h.orderService.GetOrderByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Order not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve order",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, order)
}
