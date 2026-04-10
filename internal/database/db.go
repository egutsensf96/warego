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

// IntialDB returns a singleton database instance
func IntialDB() (*gorm.DB, error) {
	var err error

	// sync.Once ensures the code inside runs only once, even with concurrent requests
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

		// Initialize GORM with a logger to see SQL queries during development
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			return
		}

		// Configure Connection Pool
		sqlDB, errPool := db.DB()
		if errPool != nil {
			err = errPool
			return
		}

		// Set connection pool limits for ERP performance
		sqlDB.SetMaxIdleConns(10)           // Keep 10 connections ready
		sqlDB.SetMaxOpenConns(100)          // Max 100 concurrent connections
		sqlDB.SetConnMaxLifetime(time.Hour) // Recycle connections every hour
	})

	return db, err
}
