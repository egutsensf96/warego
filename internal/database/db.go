package database

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// GetDB is a helper to retrieve the initialized database instance.
// It ensures IntialDB is called if the instance is nil.
func GetDB() *gorm.DB {
	if db == nil {
		_, err := IntialDB()
		if err != nil {
			log.Fatalf("Critical: Could not initialize database: %v", err)
		}
	}
	return db
}

// IntialDB returns a singleton database instance
func IntialDB() (*gorm.DB, error) {
	var err error

	once.Do(func() {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			os.Getenv("SERVER"),
			os.Getenv("USERDB"),
			os.Getenv("PASSWD"),
			os.Getenv("DBNAME"),
			os.Getenv("PORTDB"),
			os.Getenv("SSLMODE"),
			os.Getenv("TIMEZONE"),
		)

		// Set Log Level based on environment (Silent for Prod, Info for Dev)
		logLevel := logger.Info
		if os.Getenv("GIN_MODE") == "release" {
			logLevel = logger.Silent
		}

		// Initialize GORM
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
		})

		if err != nil {
			return
		}

		// Configure Connection Pool
		sqlDB, errPool := db.DB()
		if errPool != nil {
			err = errPool
			return
		}

		// Optimized connection pool for ERP workloads
		sqlDB.SetMaxIdleConns(10)           // Recommended: Keep idle connections ready
		sqlDB.SetMaxOpenConns(100)          // Prevents database exhaustion
		sqlDB.SetConnMaxLifetime(time.Hour) // Prevents stale connection errors
	})

	return db, err
}
