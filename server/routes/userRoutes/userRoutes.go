package userRoutes

import (
	"github.com/alexey-petrov/go-server/server/gormAuth"
	"github.com/alexey-petrov/go-server/server/gormJwtService"
	"github.com/alexey-petrov/go-server/server/structs"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	app.Post("api/register", func(c *fiber.Ctx) error {

		user := &gormAuth.User{}

		if err := c.BodyParser(user); err != nil {
			return err
		}

		token, refreshToken, err := gormAuth.Auth(*user)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to register",
				"detail": err.Error(),
			})
		}

		gormJwtService.SetRefreshCookie(c, refreshToken)
		gormJwtService.SetAccessTokenCookie(c, token)

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
	_, _, err := gormJwtService.HandleInvalidateTokenByJti(c)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to Invalidate JWT Refresh",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

func handleRefreshToken (c *fiber.Ctx) error {
	accessToken, err := gormJwtService.RefreshAccessToken(c)

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

	accessToken, refreshToken, err := gormAuth.Login(user.Email, user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	gormJwtService.SetRefreshCookie(c, refreshToken)
	gormJwtService.SetAccessTokenCookie(c, accessToken)

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}

// Handler function for refreshing the access token using the refresh token
func ManualResetAccessTokenHandler(c *fiber.Ctx) error {
	accessToken, refreshToken, _ := gormJwtService.ManualResetAccessToken(c)

	gormJwtService.SetRefreshCookie(c, refreshToken)
	gormJwtService.SetAccessTokenCookie(c, accessToken)

	// Return the new tokens as JSON response
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}