package db

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DBConn *gorm.DB
var shardDBs []*gorm.DB
var shardWriteQueue = make(chan User, 100) // Channel for the task queue

func InitDB() {
	var err error

	// Load .env file

	dbUrl := os.Getenv("DB_URL")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	prodUrl := os.Getenv("POSTGRES_PROD_URL")
	prodShard1Url := os.Getenv("POSTGRES_PROD_SHARD_1_URL")
	prodShard2Url := os.Getenv("POSTGRES_PROD_SHARD_2_URL")

	if prodUrl != "" {
		fmt.Printf("Connecting to database %s", prodUrl)
		DBConn, err = gorm.Open(postgres.Open(prodUrl), &gorm.Config{TranslateError: true})

		shardDBs = []*gorm.DB{
			initShardDB(prodShard1Url),
			initShardDB(prodShard2Url),
			// Add more shards here
		}
	} else {
		fmt.Printf("Connecting to database postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbUrl, dbName)
		connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbUrl, dbName)
		DBConn, err = gorm.Open(postgres.Open(connStr), &gorm.Config{TranslateError: true})

		shardDBs = []*gorm.DB{
			initShardDB(fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbUrl, "shard1")),
			initShardDB(fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbUrl, "shard2")),
			// Add more shards here
		}
	}

	go shardWorker()

	// Connection string (replace with your actual PostgreSQL credentials)

	if err != nil {
		panic("Failed to connect to database!")
	}

	err = DBConn.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error

	if err != nil {
		panic("Failed to create extension!")
	}

	err = DBConn.AutoMigrate(&User{}, &SearchSettings{}, &CrawledUrl{}, &SearchIndex{}, &MoodScore{}, &RefreshToken{})

	if err != nil {
		fmt.Println("Failed to migrate database!")
		panic(err)
	}
	for _, shard := range shardDBs {
		err = shard.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
		if err != nil {
			panic("Failed to create extension!")
		}
		err = shard.AutoMigrate(&User{}, &SearchSettings{}, &CrawledUrl{}, &SearchIndex{}, &MoodScore{}, &RefreshToken{})
		if err != nil {
			fmt.Println("Failed to migrate shard database!")
			panic(err)
		}
	}

}

// Determine shard by using hash of UserId
func determineShardByUserID(userID string) int {
	uuid, err := uuid.Parse(userID)
	if err != nil {
		fmt.Println("Failed to parse UUID", userID)
	}
	hash := sha256.Sum256(uuid[:])

	shardID := binary.BigEndian.Uint32(hash[:4]) % uint32(len(shardDBs))
	return int(shardID)
}

// Queue a task for writing to the shard
func QueueShardWrite(user User) {
	select {
	case shardWriteQueue <- user: // Enqueue the user to the shard write queue
		fmt.Printf("User with ID %s enqueued for shard write\n", user.UserId)
	default:
		fmt.Println("Shard write queue is full. Task could not be enqueued.")
	}
}

// Background worker that processes shard write tasks asynchronously
func shardWorker() {
	for user := range shardWriteQueue {
		writeToShard(user)
	}
}

// Write user to the appropriate shard
func writeToShard(user User) {
	shardID := determineShardByUserID(user.UserId)
	shardDB := shardDBs[shardID]
	// Write user data to the selected shard
	if err := shardDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&user).Error; err != nil {
		log.Printf("Failed to write to shard DB %d: %v", shardID, err)
	} else {
		// @TODO Add exact shard name
		fmt.Printf("User with ID %s written to shard of index %d with name \n", user.UserId, shardID)
	}
}

// Read from the appropriate shard based on UserId
func ReadFromShard(userID string) (User, error) {
	shardID := determineShardByUserID(userID)
	shardDB := shardDBs[shardID]

	var user User

	err := shardDB.Model(&User{}).Where("user_id = ?", userID).First(&user).Error

	if err != nil {
		if err.Error() == "record not found" {
			if err := DBConn.Model(&User{}).Where("user_id = ?", userID).First(&user).Error; err != nil {
				return User{}, err
			}
			return user, nil
		}
		return User{}, err
	}

	return user, nil
}

func initShardDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		log.Fatalf("Failed to connect to shard DB: %v", err)
	}
	return db
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
    	jti UUID NOT NULL,
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

func CreateUserMoodRecordsTable() {
	query := `
	CREATE TABLE IF NOT EXISTS user_mood_records (
    	id UUID NOT NULL,
    	user_id TEXT NOT NULL,
    	year INTEGER NOT NULL,
    	month INTEGER NOT NULL,
    	day INTEGER NOT NULL,
    	mood_id INTEGER NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL
	);`

	if err := DBConn.Exec(query).Error; err != nil {
		log.Fatal(err)
	}

	fmt.Println("Search settings Table created successfully!")
}

func GetDB() *gorm.DB {
	return DBConn
}
