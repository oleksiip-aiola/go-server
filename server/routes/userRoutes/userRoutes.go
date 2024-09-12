package userRoutes

import (
	"github.com/alexey-petrov/go-server/server/auth"
	"github.com/alexey-petrov/go-server/server/jwtService"
	"github.com/alexey-petrov/go-server/server/structs"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	app.Post("api/register", func(c *fiber.Ctx) error {

		user := &structs.User{}

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

		return c.JSON(fiber.Map{
			"token": token,
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

func handleRefreshToken (c *fiber.Ctx) error {
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

	accessToken, refreshToken, err := auth.Login(user.Email, user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	jwtService.SetRefreshCookie(c, refreshToken)

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}

// Handler function for refreshing the access token using the refresh token
func ManualResetAccessTokenHandler(c *fiber.Ctx) error {
	accessToken, refreshToken, _ := jwtService.ManualResetAccessToken(c)

	jwtService.SetRefreshCookie(c, refreshToken)
	// Return the new tokens as JSON response
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}