package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func IntialDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("SERVER"), os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("DBNAME"),
		os.Getenv("PORT"), os.Getenv("SSLMODE"), os.Getenv("TimeZone"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	return db, err
}
