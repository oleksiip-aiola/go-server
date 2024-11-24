package userRoutes

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/oleksiip-aiola/go-server/db"
	"github.com/oleksiip-aiola/go-server/jwtService"
	"github.com/oleksiip-aiola/go-server/structs"
	"gorm.io/gorm"
)

type User struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func Auth(user User) (string, error) {
	var err error

	gormUser := db.User{}
	id, err := gormUser.CreateAdmin(user.Email, user.Password, user.FirstName, user.LastName)

	if err != nil {
		return "", err
	}

	token, err := jwtService.GenerateJWTPair(id)

	if err != nil {
		fmt.Println("Error generating JWT:", err)
		return "", err
	}

	return token, err
}

func UserRoutes(app *fiber.App) {
	fmt.Println("user routes")
	app.Post("api/register", func(c *fiber.Ctx) error {

		user := &User{}

		if err := c.BodyParser(user); err != nil {
			return err
		}

		token, err := Auth(*user)

		if err != nil {
			if err == gorm.ErrDuplicatedKey {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "User already exists",
				})
			}
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

	// // Optionally handle OPTIONS for CORS requests

	app.Post("api/refresh-token", handleRefreshToken)
	app.Post("api/verify", handleRefreshToken)
	app.Post("api/logout", handleLogout)

	app.Post("api/")
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
