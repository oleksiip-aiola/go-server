package jwtService

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
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

// Store JTI in HTTP-only cookie
func SetRefreshCookie(c *fiber.Ctx, jti string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_jti",             // Name of the cookie to store JTI
		Value:    jti,                       // JTI as value
		Expires:  time.Now().Add(7 * 24 * time.Hour), // Cookie expiry matches refresh token expiry
		HTTPOnly: true,                      // HTTP-only, prevents JavaScript access
		// @TODO: Set Secure to true/Strict in production
		Secure:   false,                      // Send only over HTTPS
		SameSite: "Lax",                  // Prevent CSRF attacks
	})
}

func generateJwtAccessToken(userId int64) (string, structs.User, error) {
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
		return "", structs.User{}, err
	}
	fmt.Println("Generated JWT:", accessToken)
	return accessToken, userData, err
}
// Generate JWT with user ID
func GenerateJWT(userId int64) (string, string, error) {
	
	accessToken, userData, err := generateJwtAccessToken(userId)
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

	fmt.Println("Generated JWT:", accessToken)
	fmt.Println("Generated REFRESH:", refreshToken)

	// Store the JTI in the database
	err = db.StoreJTI(refreshToken, userData.ID, refreshTokenExp.Format(time.RFC3339))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Verify the refresh token and JTI
func VerifyRefreshToken(tokenString string) (*structs.Claims, string, error) {
	database := db.ConnectDB()

	token, err := jwt.ParseWithClaims(tokenString, &structs.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return structs.JwtKey, nil
	})
	
	if err != nil || !token.Valid {
		return nil, "", fmt.Errorf("invalid token: %v", err)
	}
	// Check if the JTI exists in the database and is not revoked
	var isRevoked bool
	var jti string
	
	claims, _ := token.Claims.(*structs.Claims)

	err = database.QueryRow("SELECT is_revoked, jti FROM refresh_tokens WHERE user_id = $1 AND expiry > NOW() ORDER BY TOKEN_ID DESC LIMIT 1", claims.UserID).Scan(&isRevoked, &jti)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", fmt.Errorf("refresh token is expired or invalid")
		}
		return nil, "", err
	}

	if isRevoked {
		return nil, "", fmt.Errorf("refresh token is revoked")
	}

	defer db.CloseDB()

	return claims, jti, nil
}

func handleVerifyRefreshToken(c *fiber.Ctx) (*structs.Claims, string, error) {
	// Extract the refresh token from the Authorization header
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return &structs.Claims{}, "" , fiber.NewError(fiber.StatusUnauthorized, "Authorization header missing")
	}

	// Extract the token from the Bearer prefix
	tokenString := authHeader[len("Bearer "):]

	// Verify the refresh token
	claims, jti, err := VerifyRefreshToken(tokenString)
	if err != nil {
		return &structs.Claims{}, "" , fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	defer db.CloseDB()

	return claims, jti, nil
}

func handleRefreshTokenByJti(c *fiber.Ctx) (string, string, error) {
	// Extract the JTI from the cookie
	jti := c.Cookies("refresh_jti")
	if jti == "" {
		return "", "" , fiber.NewError(fiber.StatusUnauthorized, "No refresh token JTI found")
	}

	var userId int64
	database := db.ConnectDB()
	// Validate the JTI against stored refresh tokens in your database (mock validation here)
	// In production, check if the JTI is valid and not revoked.
	err := database.QueryRow("SELECT user_id FROM refresh_tokens WHERE jti = $1 AND expiry > NOW() ORDER BY TOKEN_ID DESC LIMIT 1", jti).Scan(&userId)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("refresh token is expired or invalid")
		}
		return "", "", err
	}

	accessToken, refreshToken, err := GenerateJWT(userId)

	defer db.CloseDB()

	return accessToken, refreshToken, err
}

func HandleInvalidateTokenByJti(c *fiber.Ctx) (string, string, error) {
	// Extract the JTI from the cookie
	jti := c.Cookies("refresh_jti")
	if jti == "" {
		return "", "" , fiber.NewError(fiber.StatusUnauthorized, "No refresh token JTI found")
	}

	database := db.ConnectDB()
	// Validate the JTI against stored refresh tokens in your database (mock validation here)
	// In production, check if the JTI is valid and not revoked.
	_, err := database.Exec("UPDATE refresh_tokens SET is_revoked = true WHERE jti = $1", jti)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("refresh token is expired or invalid")
		}
		return "", "", err
	}

	c.Cookie(&fiber.Cookie{
		Name:    "refresh_jti",
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

	database := db.ConnectDB()

	// Generate new access token and refresh token
	accessToken, refreshToken, err := GenerateJWT(claims.UserID)

	if err != nil {
		return "", "" , fiber.NewError(fiber.StatusInternalServerError, "Error generating tokens")
	}

	// Revoke the old refresh token by marking it as revoked in the database
	_, err = database.Exec("UPDATE refresh_tokens SET is_revoked = true WHERE jti = $1", jti)

	if err != nil {
		return "", "" , fiber.NewError(fiber.StatusInternalServerError, "Error revoking old refresh token")
	}

	defer db.CloseDB()

	return accessToken, refreshToken, nil
}

func RefreshAccessToken(c *fiber.Ctx) (string, error) {
	// Extract the JTI from the cookie
	jti := c.Cookies("refresh_jti")
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
	userId, ok := claims["userId"].(float64)
	
	userIdInt := int64(userId)

	
	if !ok {
		return nil, errors.New("invalid user ID in JWT token")
	}

	_, err = db.GetUserByID(userIdInt)

	if err != nil {  
		return nil, err
	}

	// Return the claims if everything is valid
	return claims, nil
}

func refreshAccessToken(refreshTokenString string) (string, error) {
	claims := &structs.Claims{}

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

	// Generate a new access token
	newAccessToken, _, err := generateJwtAccessToken(claims.UserID)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}