package main

import (
	"log"
	"os"

	"bank-ledger-core/config"
	"bank-ledger-core/routes"
)

func main() {
	dbDriver := getEnv("DB_DRIVER", "postgres")
	dbConfig := config.GetDatabaseConfig()

	db, err := config.InitDatabase(dbDriver, dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	router := routes.SetupRoutes(db)
	port := getEnv("PORT", "8080")

	// Serve static files
	router.Static("/static", "./static")
	router.StaticFile("/", "./static/index.html")
	router.StaticFile("/login", "./static/login.html")

	log.Printf("Server starting on port %s", port)
	log.Printf("Web interface available at http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
