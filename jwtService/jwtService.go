package jwtService

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alexey-petrov/go-server/config"
	"github.com/alexey-petrov/go-server/db"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
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

var ACCESS_TOKEN_EXPIRATION = 24 * time.Hour

var REFRESH_TOKEN_EXPIRATION = 7 * ACCESS_TOKEN_EXPIRATION
var ACCESS_TOKEN_EXPIRATION_DEVELOPMENT = REFRESH_TOKEN_EXPIRATION

func isDevelopment() bool {
	return os.Getenv("ENV") == "development"
}

// Store JTI in HTTP-only cookie
func SetRefreshCookie(c *fiber.Ctx, jti string) {
	publicUrl := os.Getenv("PUBLIC_URL")
	publicDomain := os.Getenv("PUBLIC_DOMAIN")

	sameSite := "Lax"
	secure := false
	domain := "/"

	if publicUrl != "" {
		secure = true
		domain = publicDomain
	}

	c.Cookie(&fiber.Cookie{
		Name:     os.Getenv("JTI_COOKIE_NAME"),             // Name of the cookie to store JTI
		Value:    jti,                                      // JTI as value
		Expires:  time.Now().Add(REFRESH_TOKEN_EXPIRATION), // Cookie expiry matches refresh token expiry
		HTTPOnly: true,                                     // HTTP-only, prevents JavaScript access
		// @TODO: Set Secure to true/Strict in production
		Secure:   secure,   // Send only over HTTPS
		SameSite: sameSite, // Prevent CSRF attacks
		Domain:   domain,
	})
}

// Store JTI in HTTP-only cookie
func SetAccessTokenCookie(c *fiber.Ctx, token string) {
	publicUrl := os.Getenv("PUBLIC_URL")
	publicDomain := os.Getenv("PUBLIC_DOMAIN")

	sameSite := "Lax"
	secure := false
	domain := "/"

	if publicUrl != "" {
		secure = true
		domain = publicDomain
	}

	expires := ACCESS_TOKEN_EXPIRATION

	if isDevelopment() {
		expires = ACCESS_TOKEN_EXPIRATION_DEVELOPMENT
	}

	c.Cookie(&fiber.Cookie{
		Name:     os.Getenv("ACCESS_TOKEN_COOKIE_NAME"), // Name of the cookie to store JTI
		Value:    token,                                 // JTI as value
		Expires:  time.Now().Add(expires),               // Cookie expiry matches refresh token expiry
		HTTPOnly: true,                                  // HTTP-only, prevents JavaScript access
		// @TODO: Set Secure to true/Strict in production
		Secure:   secure,   // Send only over HTTPS
		SameSite: sameSite, // Prevent CSRF attacks
		Domain:   domain,
	})
}

func DeleteAccessTokenCookie(c *fiber.Ctx) {
	publicDomain := os.Getenv("PUBLIC_DOMAIN")
	publicUrl := os.Getenv("PUBLIC_URL")

	domain := "/"

	if publicUrl != "" {
		domain = publicDomain
	}
	fmt.Println("Deleting cookie: ", domain)
	c.Cookie(&fiber.Cookie{
		Name:     os.Getenv("ACCESS_TOKEN_COOKIE_NAME"),
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "None",
		Domain:   domain,
	})
}

type AuthClaims struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Admin     bool   `json:"role"`
	jwt.RegisteredClaims
}
type RefreshJWTClaims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

func GenerateJWTAccessToken(userId string) (string, error) {
	// Set expiration time for the token
	expirationTime := time.Now().Add(ACCESS_TOKEN_EXPIRATION)
	userData, _ := db.GetUserById(userId)

	// Create the claims, which includes the user ID and standard JWT claims
	claims := &AuthClaims{
		ID:        userData.UserId,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Email:     userData.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "go-server",
		},
	}

	// Create the token with the specified signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	accessToken, err := token.SignedString(config.JwtKey)
	if err != nil {
		return "", err
	}
	fmt.Println("Generated JWT:", accessToken)
	return accessToken, err
}

func GenerateJWTRefreshToken(userId string) (string, time.Time, error) {
	// Set expiration time for the token
	expirationTime := time.Now().Add(REFRESH_TOKEN_EXPIRATION)

	jti, err := generateJTI()

	if err != nil {
		return "", time.Time{}, err
	}

	refreshClaims := &RefreshJWTClaims{
		ID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "go-server",
			ID:        jti, // Set JTI in the refresh token
		},
	}

	// Create the token with the specified signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	// Sign the token with the secret key
	refreshToken, err := token.SignedString(config.JwtKey)
	if err != nil {
		return "", time.Time{}, err
	}
	fmt.Println("Generated REFRESH TOKEN:", refreshToken)
	return refreshToken, expirationTime, err
}

// Generate JWT with user ID
func GenerateJWTPair(userId string) (string, error) {
	accessToken, err := GenerateJWTAccessToken(userId)
	if err != nil {
		return "", err
	}
	// Set expiration time for Refresh Token (long-lived)
	refreshToken, expirationTime, err := GenerateJWTRefreshToken(userId)

	if err != nil {
		return "", err
	}

	userData, _ := db.GetUserById(userId)

	// Store the JTI in the database
	err = db.StoreJTI(refreshToken, userData.UserId, expirationTime.Format(time.RFC3339))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func handleRefreshAccessTokenByUserId(userId string) (string, error) {
	result := db.DBConn.Model(&db.RefreshToken{}).Where("user_id = ? AND expiry > NOW() AND is_revoked=false", userId).Limit(1)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("refresh token is expired or invalid")
		}
		return "", result.Error
	}

	accessToken, err := GenerateJWTAccessToken(userId)

	return accessToken, err
}

func HandleInvalidateUserSession(userId string) error {
	if userId == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "No user id found")
	}

	err := db.DBConn.Model(&db.RefreshToken{}).Where("user_id = ?", userId).Update("is_revoked", true).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("refresh token is expired or invalid")
		}
		return err
	}

	return nil
}

func RefreshAccessToken(c *fiber.Ctx, userId string) (string, error) {

	// Validate the JTI against stored refresh tokens in your database (mock validation here)
	// In production, check if the JTI is valid and not revoked.
	accessToken, err := handleRefreshAccessTokenByUserId(userId)

	if err != nil {
		return "", c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token refresh failed"})
	}

	SetAccessTokenCookie(c, accessToken)

	return accessToken, nil
}

func VerifyToken(token string) (*jwt.Token, error) {
	// Parse the JWT token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {

		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		// Return the secret key used for signing
		return []byte(config.JwtKey), nil
	})

	if err != nil {
		// Check if the error is due to token expiration
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("access token expired")
		}
		return nil, fmt.Errorf("invalid access token")
	}

	// Check if the token is valid
	if !parsedToken.Valid {
		return nil, errors.New("invalid JWT token")
	}
	// Return the parsed token
	return parsedToken, nil
}

func RevokeJWTByUserId(userId string) error {

	err := db.RevokeJWTByUserId(userId)

	if err != nil {
		return err
	}

	return nil
}
