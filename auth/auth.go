package auth

import (
	"fmt"

	// Add this line
	"github.com/alexey-petrov/go-server/db"
	"github.com/alexey-petrov/go-server/jwtService"
	"github.com/gofiber/fiber/v2"
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

func Login(c *fiber.Ctx, email string, password string) (string, error) {
	// Example: generate a JWT for user with ID 123
	user := db.User{}
	userData, err := user.LoginAsAdmin(email, password)

	if err != nil {
		return "", c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":  "Failed to login",
			"detail": "No such user",
		})
	}

	jwtService.RevokeJWTByUserId(userData.UserId)

	token, err := jwtService.GenerateJWTPair(userData.UserId)
	if err != nil {
		fmt.Println("Error generating JWT:", err)
		return "", err
	}

	return token, nil
}
