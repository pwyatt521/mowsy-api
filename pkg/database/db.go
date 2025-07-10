package database

import (
	"fmt"
	"log"
	"os"

	"mowsy-api/internal/models"

	"gorm.io/gorm"
)

var DB *gorm.DB


func InitDB() error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	if host == "" || port == "" || dbname == "" || user == "" || password == "" {
		return fmt.Errorf("missing required database environment variables")
	}

	log.Printf("Database credentials loaded: %s:%s, user: %s, db: %s", host, port, user, dbname)
	return nil // Skip actual connection for now to test the endpoints
}

func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	err := DB.AutoMigrate(
		&models.User{},
		&models.Job{},
		&models.JobApplication{},
		&models.Equipment{},
		&models.EquipmentRental{},
		&models.Review{},
		&models.Payment{},
	)
	if err != nil {
		return fmt.Errorf("failed to run auto migration: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}