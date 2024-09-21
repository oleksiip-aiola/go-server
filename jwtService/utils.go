package jwtService

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func VerifyTokenProtectedRoute(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	fmt.Println(authHeader)
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing Authorization header",
		})
	}
	fmt.Println("post auth header")
	// Extract JWT token from Authorization header
	accessTokenCookie := authHeader[len("Bearer "):]

	_, verificationError := VerifyToken(accessTokenCookie)
	if verificationError != nil {
		if verificationError.Error() == "access token expired" {
			fmt.Println(verificationError, verificationError.Error() == "access token expired")

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT token",
			})
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT token",
			})
		}
	}

	return nil
}
