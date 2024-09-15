package gormAuth

import (
	"fmt"

	// Add this line
	"github.com/alexey-petrov/go-server/server/gormJwtService"
	"github.com/alexey-petrov/go-server/server/gormdb"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
}

func Auth(user User) (string, string, error) {
	var err error

	gormUser := gormdb.User{}
	id, err := gormUser.CreateAdmin(user.Email, user.Password, user.FirstName, user.LastName)

	if err != nil {
		return "", "", err
	}
fmt.Println(id)
	token, refreshToken, err := gormJwtService.GenerateJWT(id)

	if err != nil {
		fmt.Println("Error generating JWT:", err)
		return "", "",err
	}

	return token, refreshToken, err
}

func Login(email string, password string) (string, string, error) {
	// Example: generate a JWT for user with ID 123
	user := gormdb.User{}
	userData, _ := user.LoginAsAdmin(email, password)
	
	gormdb.RevokeJWTByUserId(userData.UserId)

	token, refreshToken, err := gormJwtService.GenerateJWT(userData.UserId)
	if err != nil {
		fmt.Println("Error generating JWT:", err)
		return "", "", err
	}

	return token, refreshToken, nil
}