package gormdb

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DBConn *gorm.DB

func InitDB() {
	var err error;

	// Load .env file
    err = godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

	dbUrl := os.Getenv("DB_URL")
	dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")

    fmt.Printf("Connecting to database postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbUrl, dbName)
	
	// Connection string (replace with your actual PostgreSQL credentials)
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbUrl, dbName)

	DBConn, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})

	if err != nil {	
		panic("Failed to connect to database!")
	}

	err = DBConn.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error

	if err != nil {
		panic("Failed to create extension!")
	}

	err = DBConn.AutoMigrate()

	if err != nil {
		panic("Failed to migrate database!")
	}
}

func GetDB() *gorm.DB {
	return DBConn
}