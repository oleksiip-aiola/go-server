package userRoutes

import (
	"errors"

	"github.com/alexey-petrov/go-server/auth"
	"github.com/alexey-petrov/go-server/jwtService"
	"github.com/alexey-petrov/go-server/structs"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func UserRoutes(app *fiber.App) {
	app.Post("api/register", func(c *fiber.Ctx) error {

		user := &auth.User{}

		if err := c.BodyParser(user); err != nil {
			return err
		}

		token, refreshToken, err := auth.Auth(*user)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to register",
				"detail": err.Error(),
			})
		}

		jwtService.SetRefreshCookie(c, refreshToken)
		jwtService.SetAccessTokenCookie(c, token)

		return c.JSON(fiber.Map{
			"access_token": token,
		})
	})

	app.Post("api/refresh", ManualResetAccessTokenHandler)
	app.Post("api/login", handleLogin)
	app.Post("api/refresh-token", handleRefreshToken)
	app.Post("api/logout", handleLogout)
}

func handleLogout(c *fiber.Ctx) error {
	_, _, err := jwtService.HandleInvalidateTokenByJti(c)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to Invalidate JWT Refresh",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

func handleRefreshToken(c *fiber.Ctx) error {
	accessToken, err := jwtService.RefreshAccessToken(c)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}

func handleLogin(c *fiber.Ctx) error {
	user := &structs.User{}

	if err := c.BodyParser(user); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)

	if err != nil {
		return errors.New("failed to hash password when logging in")
	}

	password := string(hashedPassword)

	accessToken, refreshToken, err := auth.Login(user.Email, password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	jwtService.SetRefreshCookie(c, refreshToken)
	jwtService.SetAccessTokenCookie(c, accessToken)

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}

// Handler function for refreshing the access token using the refresh token
func ManualResetAccessTokenHandler(c *fiber.Ctx) error {
	accessToken, refreshToken, _ := jwtService.ManualResetAccessToken(c)

	jwtService.SetRefreshCookie(c, refreshToken)
	jwtService.SetAccessTokenCookie(c, accessToken)

	// Return the new tokens as JSON response
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
