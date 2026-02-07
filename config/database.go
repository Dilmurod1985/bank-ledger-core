package config

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"bank-ledger-core/models"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func GetDatabaseConfig() *DatabaseConfig {
	// Check if DATABASE_URL is provided (for production)
	if databaseURL := getEnv("DATABASE_URL", ""); databaseURL != "" {
		return &DatabaseConfig{
			Host:     "",
			Port:     "",
			User:     "",
			Password: "",
			DBName:   databaseURL,
			SSLMode:  "",
		}
	}
	
	// Fallback to individual env vars (for development)
	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "bank_ledger"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func InitDatabase(driver string, config *DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch driver {
	case "postgres":
		var dsn string
		// Use DATABASE_URL if provided (production)
		if config.DBName != "" && (config.Host == "" || config.User == "") {
			dsn = config.DBName // This is actually the DATABASE_URL
		} else {
			// Use individual config parameters (development)
			dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
				config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)
		}
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		dbName := getEnv("DB_PATH", "bank_ledger.db")
		db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure SQLite connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Auto-migrate all models with error handling for SQLite
	err = db.AutoMigrate(
		&models.Session{},
		&models.Account{},
		&models.Product{},
		&models.Order{},
		&models.Transfer{},
	)
	if err != nil {
		// For SQLite, this might be a migration conflict
		if driver == "sqlite" {
			return nil, fmt.Errorf("SQLite migration failed: %w. This might be due to schema conflicts. Try removing the database file and restarting.", err)
		}
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	err = createSystemAccount(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create system account: %w", err)
	}

	return db, nil
}

func createSystemAccount(db *gorm.DB) error {
	var systemAccount models.Account
	result := db.Where("user_id = ?", "0").First(&systemAccount)
	
	if result.Error == gorm.ErrRecordNotFound {
		systemAccount = models.Account{
			UserID:   "0",
			Currency: "UZS",
			Balance:  "0.00",
		}
		
		if err := db.Create(&systemAccount).Error; err != nil {
			return fmt.Errorf("failed to create system account: %w", err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("failed to check system account: %w", result.Error)
	}
	
	return nil
}
