package config

import (
	"testing"
	"github.com/glebarez/go-sqlite"
	"gorm.io/gorm"
)

func TestSQLiteDriver(t *testing.T) {
	dialector := sqlite.New(sqlite.Config{DSN: "test.db"})
	if dialector == nil {
		t.Error("SQLite driver not working")
	}
	
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		t.Errorf("Failed to get underlying DB: %v", err)
	}
	sqlDB.Close()
}
