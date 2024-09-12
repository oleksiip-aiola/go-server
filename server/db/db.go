// db/db.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/alexey-petrov/go-server/server/structs"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

var DB *sql.DB // Global DB variable to be used across the application

// ConnectDB initializes the connection to the PostgreSQL database
func ConnectDB() *sql.DB {
	// Load .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    // Retrieve environment variables
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")

    fmt.Printf("Connecting to database '%s' as user '%s' with password '%s'.\n", dbName, dbUser, dbPassword)
	// Connection string (replace with your actual PostgreSQL credentials)
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)

	// Open the connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open a DB connection:", err)
	}

	// Verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping DB:", err)
	}

	fmt.Println("Connected to the database successfully")
	DB = db
	return DB
}

// CloseDB closes the DB connection
func CloseDB() {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			log.Fatal("Failed to close DB connection:", err)
		}
	}

	fmt.Println("Closed the database connection")
}

func CreateTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		password TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE
	)`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Table created successfully!")
}

func CreateJTITable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS refresh_tokens (
    	token_id SERIAL PRIMARY KEY,
    	user_id INT NOT NULL,
    	jti TEXT NOT NULL UNIQUE,
    	expiry TIMESTAMPTZ NOT NULL,
    	is_revoked BOOLEAN DEFAULT FALSE
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Table created successfully!")
}

func InsertUser(user structs.User) (structs.User, error) {
	ConnectDB()

	query := `INSERT INTO users (email, first_name, last_name, password) VALUES ($1, $2, $3, $4) RETURNING user_id`
	// Insert the user into the database
	_, err := DB.Exec(query, user.Email, user.FirstName, user.LastName, user.Password)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			// Check if the error is due to a unique constraint violation (code 23505)
			if pqErr.Code == "23505" {
				// Return a custom error message
				return structs.User{}, fmt.Errorf("this email is already in use")
			}
		}
		return structs.User{}, err
	}

	// Get the ID of the newly inserted user
	var id int
	err = DB.QueryRow("SELECT user_id FROM users ORDER BY user_id DESC LIMIT 1").Scan(&id)
	if err != nil {
		return structs.User{}, err
	}

	// Fetch the user data from the database
	userData := structs.User{}
	err = DB.QueryRow("SELECT user_id, email, first_name, last_name FROM users WHERE user_id = $1", id).Scan(&userData.ID, &userData.Email, &userData.FirstName, &userData.LastName)
	if err != nil {
		return structs.User{}, err
	}

	defer CloseDB()

	// Return the user data
	return userData, err

}

func GetUserByID(userID int64) (structs.User, error) {
	ConnectDB()
	query := `SELECT user_id, email, first_name, last_name FROM users WHERE user_id = $1`
	row := DB.QueryRow(query, userID)

	user := structs.User{}
	err := row.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("user not found")
		}
		log.Fatal(err)
		return user, err
	}

	defer CloseDB()

	return user, nil
}

func GetUserByEmailPassword(email, password string) (structs.User, error) {
	ConnectDB()
	query := `SELECT user_id, email, first_name, last_name FROM users WHERE email = $1 AND password = $2`
	row := DB.QueryRow(query, email, password)

	user := structs.User{}
	err := row.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("user not found")
		}
		log.Fatal(err)
		return user, err
	}

	defer CloseDB()

	return user, nil
}

func GetUsers(db *sql.DB) {
	rows, err := db.Query("SELECT user_id, name, email FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		err = rows.Scan(&id, &name, &email)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User: %d | Name: %s | Email: %s\n", id, name, email)
	}

	// Check for errors from iterating over rows.
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func StoreJTI(jti string, userID int, refreshTokenExp string) error {
	ConnectDB()
	query := `INSERT INTO refresh_tokens (user_id, jti, expiry, is_revoked) VALUES ($1, $2, $3, $4)`
	_, err := DB.Exec(query, userID, jti, refreshTokenExp, false)
	if err != nil {
		log.Fatal(err)
		return err
	}

	defer CloseDB()

	return nil
}

func RevokeJWTByUserId(userId int64) error{
	ConnectDB()

	_, err := DB.Exec("UPDATE refresh_tokens SET is_revoked = true WHERE user_id = $1", userId)

	if err != nil {
		log.Fatal(err)
		return err
	}

	defer CloseDB()

	fmt.Println("Revoked JWT for user ID:", userId)

	return nil
}