package db

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
	var err error

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

	err = DBConn.AutoMigrate(&User{}, &SearchSettings{}, &CrawledUrl{}, &SearchIndex{})

	if err != nil {
		fmt.Println("Failed to migrate database!")
		panic(err)
	}
}

func CreateTable() {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		password TEXT NOT NULL,
		is_admin BOOLEAN NOT NULL,
    	updated_at TIMESTAMPTZ NOT NULL,
    	created_at TIMESTAMPTZ NOT NULL,
		email TEXT NOT NULL UNIQUE
	)`

	if err := DBConn.Exec(query).Error; err != nil {
		log.Fatal(err)
	}

	fmt.Println("Users Table created successfully!")
}

func CreateJTITable() {
	query := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
    	id SERIAL PRIMARY KEY,
    	user_id TEXT NOT NULL,
    	jti TEXT NOT NULL UNIQUE,
    	expiry TIMESTAMPTZ NOT NULL,
    	is_revoked BOOLEAN DEFAULT FALSE
	);`

	if err := DBConn.Exec(query).Error; err != nil {
		log.Fatal(err)
	}

	fmt.Println("Refresh Tokens Table created successfully!")
}

func CreateSearchSettingsTable() {
	query := `
	CREATE TABLE IF NOT EXISTS search_settings (
    	id SERIAL PRIMARY KEY,
    	amount INTEGER NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL,
    	search_on BOOLEAN DEFAULT FALSE,
    	add_new BOOLEAN DEFAULT FALSE
	);`

	if err := DBConn.Exec(query).Error; err != nil {
		log.Fatal(err)
	}

	fmt.Println("Search settings Table created successfully!")
}

func GetDB() *gorm.DB {
	return DBConn
}
