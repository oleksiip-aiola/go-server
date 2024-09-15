package gormJwtService

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alexey-petrov/go-server/server/db"
	"github.com/alexey-petrov/go-server/server/gormdb"
	"github.com/alexey-petrov/go-server/server/structs"
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

// Store JTI in HTTP-only cookie
func SetRefreshCookie(c *fiber.Ctx, jti string) {
	fmt.Println("set jti cookie", jti)
	c.Cookie(&fiber.Cookie{
		Name:     os.Getenv("JTI_COOKIE_NAME"),             // Name of the cookie to store JTI
		Value:    jti,                       // JTI as value
		Expires:  time.Now().Add(REFRESH_TOKEN_EXPIRATION), // Cookie expiry matches refresh token expiry
		HTTPOnly: true,                      // HTTP-only, prevents JavaScript access
		// @TODO: Set Secure to true/Strict in production
		Secure:   false,                      // Send only over HTTPS
		SameSite: "Lax",                  // Prevent CSRF attacks
	})
}

// Store JTI in HTTP-only cookie
func SetAccessTokenCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     os.Getenv("ACCESS_TOKEN_COOKIE_NAME"),             // Name of the cookie to store JTI
		Value:    token,                       // JTI as value
		Expires:  time.Now().Add(ACCESS_TOKEN_EXPIRATION), // Cookie expiry matches refresh token expiry
		HTTPOnly: true,                      // HTTP-only, prevents JavaScript access
		// @TODO: Set Secure to true/Strict in production
		Secure:   false,                      // Send only over HTTPS
		SameSite: "Lax",                  // Prevent CSRF attacks
	})
}

type AuthClaims struct {
	ID        string   `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Admin    bool   `json:"role"`
	jwt.RegisteredClaims
}
type RefreshJWTClaims struct {
	ID        string   `json:"id"`
	jwt.RegisteredClaims
}

func generateJwtAccessToken(userId string) (string, error) {
	// Set expiration time for the token
	expirationTime := time.Now().Add(ACCESS_TOKEN_EXPIRATION)
	userData, _ := gormdb.GetUserById(userId)

	// Create the claims, which includes the user ID and standard JWT claims
	claims := &AuthClaims{
		ID: userData.UserId,
		FirstName: userData.FirstName,
		LastName: userData.LastName,
		Email: userData.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:   "go-server",
		},
	}

	// Create the token with the specified signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	accessToken, err := token.SignedString(structs.JwtKey)
	if err != nil {
		return "", err
	}
	fmt.Println("Generated JWT:", accessToken)
	return accessToken, err
}

func generateJwtRefreshToken(userId string) (string, time.Time, error) {
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
			Issuer:   "go-server",
			ID:        jti, // Set JTI in the refresh token
		},
	}

	// Create the token with the specified signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	// Sign the token with the secret key
	refreshToken, err := token.SignedString(structs.JwtKey)
	if err != nil {
		return "", time.Time{}, err
	}
	fmt.Println("Generated REFRESH TOKEN:", refreshToken)
	return refreshToken, expirationTime, err
}

// Generate JWT with user ID
func GenerateJWT(userId string) (string, string, error) {
	
	accessToken, err := generateJwtAccessToken(userId)
	if err != nil {
		return "", "", err
	}
	// Set expiration time for Refresh Token (long-lived)
	refreshToken, expirationTime, err := generateJwtRefreshToken(userId)

	if err != nil {
		return "", "", err
	}

	userData, _ := gormdb.GetUserById(userId)

	// Store the JTI in the database
	err = gormdb.StoreJTI(refreshToken, userData.UserId, expirationTime.Format(time.RFC3339))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Verify the refresh token and JTI
func VerifyRefreshToken(tokenString string) (*AuthClaims, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return structs.JwtKey, nil
	})
	
	if err != nil || !token.Valid {
		return nil, "", fmt.Errorf("invalid token: %v", err)
	}
	// Check if the JTI exists in the database and is not revoked
	var jti string
	
	claims, _ := token.Claims.(*AuthClaims)

	jti, err = gormdb.CheckIfRefreshTokenIsRevokedByUserId(claims.ID)

	if err != nil {
		return nil, "", err
	}

	return claims, jti, nil
}

func handleVerifyRefreshToken(c *fiber.Ctx) (*AuthClaims, string, error) {
	// Extract the refresh token from the Authorization header
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return &AuthClaims{}, "" , fiber.NewError(fiber.StatusUnauthorized, "Authorization header missing")
	}

	// Extract the token from the Bearer prefix
	tokenString := authHeader[len("Bearer "):]

	// Verify the refresh token
	claims, jti, err := VerifyRefreshToken(tokenString)
	if err != nil {
		return &AuthClaims{}, "" , fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	defer db.CloseDB()

	return claims, jti, nil
}

func handleRefreshTokenByJti(c *fiber.Ctx) (string, string, error) {
	// Extract the JTI from the cookie
	jti := c.Cookies(os.Getenv("JTI_COOKIE_NAME"))
	if jti == "" {
		return "", "" , fiber.NewError(fiber.StatusUnauthorized, "No refresh token JTI found")
	}
	
	// Validate the JTI against stored refresh tokens in your database (mock validation here)
	// In production, check if the JTI is valid and not revoked.
	var refreshJwtToken struct {
		UserID string
	}
fmt.Println("jti", jti)
	result := gormdb.DBConn.Model(&gormdb.RefreshToken{}).Where("jti = ? AND expiry > NOW() AND is_revoked=false", jti).Limit(1).Scan(&refreshJwtToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", "", fmt.Errorf("refresh token is expired or invalid")
		}
		return "", "", result.Error
	}

	fmt.Println(refreshJwtToken)
	userId := refreshJwtToken.UserID

	accessToken, refreshToken, err := GenerateJWT(userId)

	return accessToken, refreshToken, err
}

func HandleInvalidateTokenByJti(c *fiber.Ctx) (string, string, error) {
	// Extract the JTI from the cookie
	jti := c.Cookies(os.Getenv("JTI_COOKIE_NAME"))
	if jti == "" {
		return "", "" , fiber.NewError(fiber.StatusUnauthorized, "No refresh token JTI found")
	}

	
	// Validate the JTI against stored refresh tokens in your database (mock validation here)
	// In production, check if the JTI is valid and not revoked.
	err := gormdb.DBConn.Model(&gormdb.RefreshToken{}).Where("jti = ?", jti).Update("is_revoked", true).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", fmt.Errorf("refresh token is expired or invalid")
		}
		return "", "", err
	}

	c.Cookie(&fiber.Cookie{
		Name:    os.Getenv("JTI_COOKIE_NAME"),
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour), // Set the expiry time to a past date
	})

	defer db.CloseDB()

	return "", "", nil
}

// Refresh the access token using the refresh token
func ManualResetAccessToken(c *fiber.Ctx) (string, string, error) {
	claims, jti, err := handleVerifyRefreshToken(c)

	if err != nil {
		return "", "" , c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	

	// Generate new access token and refresh token
	accessToken, refreshToken, err := GenerateJWT(claims.ID)

	if err != nil {
		return "", "" , fiber.NewError(fiber.StatusInternalServerError, "Error generating tokens")
	}

	// Revoke the old refresh token by marking it as revoked in the database
	err = gormdb.DBConn.Model(&gormdb.RefreshToken{}).Where("jti = ?", jti).Update("is_revoked", true).Error

	if err != nil {
		return "", "", fiber.NewError(fiber.StatusInternalServerError, "Error revoking old refresh token")
	}

	return accessToken, refreshToken, nil
}

func RefreshAccessToken(c *fiber.Ctx) (string, error) {
	// Extract the JTI from the cookie
	jti := c.Cookies(os.Getenv("JTI_COOKIE_NAME"))
	if jti == "" {
		return "", c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No refresh token JTI found"})
	}

	// Validate the JTI against stored refresh tokens in your database (mock validation here)
	// In production, check if the JTI is valid and not revoked.
	accessToken, refreshToken, err := handleRefreshTokenByJti(c)

	if err != nil {
		return "",c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token refresh failed"})
	}

	SetRefreshCookie(c, refreshToken)
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
		return []byte(structs.JwtKey), nil
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
func VerifyAndParseToken(token string, jti string) (map[string]interface{}, error) {
	fmt.Println(token, jti)
	// Verify the JWT token
	verifiedToken, err := VerifyToken(token)

	if err != nil {
		if err.Error() == "access token expired" {
			// Access token has expired, validate the refresh token
			newAccessToken, refreshErr := refreshAccessToken(jti)
			if refreshErr != nil {
				fmt.Println("Error refreshing access token:", refreshErr)
			}

			fmt.Println("New access token:", verifiedToken)

			verifiedToken, _ = VerifyToken(newAccessToken)
		} else {
			fmt.Println("Access token validation error:", err)
		}
	} else {
		fmt.Println("Access token is valid")
	}

	// Extract the claims from the verified token
	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid JWT claims")
	}

	// Get the user ID from the claims
	userId, ok := claims["ID"].(string)
	
	if !ok {
		return nil, errors.New("invalid user ID in JWT token")
	}

	_, err = gormdb.GetUserById(userId)

	if err != nil {  
		return nil, err
	}

	// Return the claims if everything is valid
	return claims, nil
}

func refreshAccessToken(refreshTokenString string) (string, error) {
	claims := &AuthClaims{}

	// Parse and validate the refresh token
	token, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return structs.JwtRefreshKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid refresh token")
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}
	fmt.Println("claims", claims)
	// Generate a new access token
	newAccessToken, err := generateJwtAccessToken(claims.ID)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}