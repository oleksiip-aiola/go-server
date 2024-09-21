package userRoutes

import (
	"fmt"

	"github.com/alexey-petrov/go-server/auth"
	"github.com/alexey-petrov/go-server/jwtService"
	"github.com/alexey-petrov/go-server/structs"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	app.Post("api/register", func(c *fiber.Ctx) error {

		user := &auth.User{}

		if err := c.BodyParser(user); err != nil {
			return err
		}

		token, err := auth.Auth(*user)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to register",
				"detail": err.Error(),
			})
		}

		jwtService.SetAccessTokenCookie(c, token)

		return c.JSON(fiber.Map{
			"access_token": token,
		})
	})

	app.Post("api/login", handleLogin)
	app.Post("api/refresh-token", handleRefreshToken)
	app.Post("api/logout", handleLogout)
}

type LogoutStruct struct {
	ID string `json:"id"`
}

func handleLogout(c *fiber.Ctx) error {
	fmt.Println("LOGOUT")

	user := &LogoutStruct{}
	if err := c.BodyParser(user); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("userId", user)
	err := jwtService.HandleInvalidateUserSession(user.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to Invalidate User session",
		})
	}

	jwtService.DeleteAccessTokenCookie(c)

	return c.JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

func handleRefreshToken(c *fiber.Ctx) error {
	fmt.Println("REFRESH TOKEN")
	user := &structs.User{}

	if err := c.BodyParser(user); err != nil {
		return err
	}
	accessToken, err := jwtService.RefreshAccessToken(c, user.ID)

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
	fmt.Println("LOGIN")
	if err := c.BodyParser(user); err != nil {
		return err
	}

	accessToken, err := auth.Login(user.Email, user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	jwtService.SetAccessTokenCookie(c, accessToken)

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}
