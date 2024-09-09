package jwtService

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/alexey-petrov/go-server/server/db"
	"github.com/alexey-petrov/go-server/server/structs"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Generate random JTI (JWT ID)
func generateJTI() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}


// Generate JWT with user ID
func GenerateJWT(userId int64) (string, string, error) {
	// Set expiration time for the token
	expirationTime := time.Now().Add(24 * time.Hour)
	userData, _ := db.GetUserByID(userId)

	// Create the claims, which includes the user ID and standard JWT claims
	claims := &structs.Claims{
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
	accessToken, err := token.SignedString(structs.JwtKey)
	if err != nil {
		return "", "", err
	}

	// Set expiration time for Refresh Token (long-lived)
	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour) // 7 days
	jti, err := generateJTI()                            // Generate JTI
	if err != nil {
		return "", "", err
	}
	refreshClaims := &structs.Claims{
		UserID: int64(userData.ID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExp),
			ID:        jti, // Set JTI in the refresh token
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(structs.JwtRefreshKey)
	if err != nil {
		return "", "", err
	}

	// Store the JTI in the database
	err = db.StoreJTI(refreshToken, userData.ID, refreshTokenExp.Format(time.RFC3339))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Verify the refresh token and JTI
func VerifyRefreshToken(tokenString string) (*structs.Claims, error) {
	database := db.ConnectDB()
	fmt.Println(":", tokenString)

	token, err := jwt.ParseWithClaims(tokenString, &structs.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return structs.JwtKey, nil
	})
	
	fmt.Println(token.Valid, err)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	// Check if the JTI exists in the database and is not revoked
	var isRevoked bool
	
	claims, ok := token.Claims.(*structs.Claims)
	fmt.Println(claims, ok)
	err = database.QueryRow("SELECT is_revoked FROM refresh_tokens WHERE user_id = $1 AND expiry > NOW()", claims.UserID).Scan(&isRevoked)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refresh token is expired or invalid")
		}
		return nil, err
	}

	if isRevoked {
		return nil, fmt.Errorf("refresh token is revoked")
	}

	defer db.CloseDB()

	return claims, nil
}

// Refresh the access token using the refresh token
func RefreshAccessToken(c *fiber.Ctx) (string, string, error) {
	database := db.ConnectDB()
	// Extract the refresh token from the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return "", "" , fiber.NewError(fiber.StatusUnauthorized, "Authorization header missing")
	}

	// Extract the token from the Bearer prefix
	tokenString := authHeader[len("Bearer "):]

	// Verify the refresh token
	claims, err := VerifyRefreshToken(tokenString)
	if err != nil {
		return "", "" , fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	// Generate new access token and refresh token
	accessToken, refreshToken, err := GenerateJWT(claims.UserID)

	if err != nil {
		return "", "" , fiber.NewError(fiber.StatusInternalServerError, "Error generating tokens")
	}


	//@TODO: Fix revoke
	// Revoke the old refresh token by marking it as revoked in the database
	_, err = database.Exec("UPDATE refresh_tokens SET is_revoked = true WHERE jti = $1", tokenString)

	if err != nil {
		return "", "" , fiber.NewError(fiber.StatusInternalServerError, "Error revoking old refresh token")
	}

	defer db.CloseDB()
fmt.Println("tokens", accessToken, refreshToken)
	return accessToken, refreshToken, nil
}