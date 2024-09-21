package jwtService

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func VerifyTokenProtectedRoute(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	fmt.Println(authHeader)
	if authHeader == "" {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing Authorization header",
		})

		return errors.New("authorization header missing")
	}
	fmt.Println("post auth header")
	// Extract JWT token from Authorization header
	accessTokenCookie := authHeader[len("Bearer "):]

	_, verificationError := VerifyToken(accessTokenCookie)
	if verificationError != nil {
		if verificationError.Error() == "access token expired" {
			fmt.Println(verificationError, verificationError.Error() == "access token expired")

			c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT token",
			})
			return errors.New("invalid JWT token")

		} else {
			c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT token",
			})
			return errors.New("invalid JWT token")

		}
	}

	return nil
}
