package auth

import (
	"fmt"
	"time"

	"github.com/alexey-petrov/go-server/server/db"
	"github.com/alexey-petrov/go-server/server/structs"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("my_secret_key") // Replace with a strong secret key

// Define a structure for JWT claims (payload)
type Claims struct {
	UserID int64 `json:"userId"`
	Email string `json:"email"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	jwt.RegisteredClaims
}

// Generate JWT with user ID
func generateJWT(userData structs.User) (string, error) {
	// Set expiration time for the token
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the claims, which includes the user ID and standard JWT claims
	claims := &Claims{
		UserID: int64(userData.ID),
		FirstName: userData.FirstName,
		LastName: userData.LastName,
		Email: userData.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Create the token with the specified signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func Auth(user structs.User) (string, error) {
	// Example: generate a JWT for user with ID 123
	userData := db.InsertUser(user)
	token, err := generateJWT(userData)
	if err != nil {
		fmt.Println("Error generating JWT:", err)
		return "", err
	}

	fmt.Println("Generated JWT:", token)

	return token, nil
}