package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"bank-ledger-core/handlers"
	"bank-ledger-core/middleware"
	"bank-ledger-core/services"
)

func SetupRoutes(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Initialize services and handlers
	transferService := services.NewTransferService(db)
	orderService := services.NewOrderService(db, transferService)
	historyService := services.NewHistoryService(db)
	
	// Handlers
	accountHandler := handlers.NewAccountHandler(db)
	productHandler := handlers.NewProductHandler(db)
	orderHandler := handlers.NewOrderHandler(orderService)
	transferHandler := handlers.NewTransferHandler(transferService)
	historyHandler := handlers.NewHistoryHandler(historyService)
	authHandler := handlers.NewAuthHandler(db)

	api := r.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(db))
		{
			accounts := protected.Group("/accounts")
			{
				accounts.POST("", accountHandler.CreateAccount)
				accounts.GET("", accountHandler.GetAccounts)
				accounts.GET("/:id", accountHandler.GetAccount)
			}

			products := protected.Group("/products")
			{
				products.GET("", productHandler.GetProducts)
				products.GET("/:id", productHandler.GetProduct)
				products.POST("", productHandler.CreateProduct)
			}

			orders := protected.Group("/orders")
			{
				orders.POST("", orderHandler.CreateOrder)
				orders.GET("", orderHandler.GetOrders)
				orders.GET("/:id", orderHandler.GetOrder)
			}

			transfers := protected.Group("/transfers")
			{
				transfers.POST("/money", transferHandler.TransferMoney)
				transfers.POST("/money/users", transferHandler.TransferMoneyByUserIDs)
			}

			// Separate route for account history to avoid conflicts
			protected.GET("/users/:user_id/history", historyHandler.GetAccountHistory)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return r
}
