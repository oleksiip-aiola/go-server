package structs

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type Todo struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Done bool `json:"done"`
	Body string `json:"body"`
}

type User struct {
	ID int `json:"id"`
	Email string `json:"email"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Password string `json:"password"`
}

// Define a structure for JWT claims (payload)
type Claims struct {
	UserID int64 `json:"userId"`
	Email string `json:"email"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	jwt.RegisteredClaims
}

var JwtKey = []byte(os.Getenv("JWT_SECRET_KEY")) // Replace with a strong secret key
var JwtRefreshKey = []byte(os.Getenv("JWT_REFRESH_KEY")) // Secret for refresh token