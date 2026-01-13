package common

import (
	"path/filepath"

	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var dbInitializers = make([]func(string) error, 0)

func RegisterDBInitializer(initFunc func(string) error) {
	dbInitializers = append(dbInitializers, initFunc)
}

// InitDB initializes the database connection and creates tables
func InitDB(dbPath string) error {
	// Ensure the data directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database connection
	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	// Run initializers
	for _, initFunc := range dbInitializers {
		if err := initFunc(dbPath); err != nil {
			return err
		}
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}
