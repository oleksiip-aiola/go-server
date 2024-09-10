package auth

import (
	"fmt"

	"github.com/alexey-petrov/go-server/server/db"
	jwtService "github.com/alexey-petrov/go-server/server/jwt"
	"github.com/alexey-petrov/go-server/server/structs"
)




func Auth(user structs.User) (string, error) {
	// Example: generate a JWT for user with ID 123
	var err error

	userData, err := db.InsertUser(user)

	if err != nil {
		return "", err
	}

	token, refreshToken, err := jwtService.GenerateJWT(int64(userData.ID))

	if err != nil {
		fmt.Println("Error generating JWT:", err)
		return "", err
	}

	fmt.Println("Generated JWT:", token)
	fmt.Println("Generated REFRESH:", refreshToken)

	return token, err
}

func Login(email string, password string) (string, error) {
	// Example: generate a JWT for user with ID 123
	userData, _ := db.GetUserByEmailPassword(email, password)
	
	db.RevokeJWTByUserId(int64(userData.ID))

	token, refreshToken, err := jwtService.GenerateJWT(int64(userData.ID))
	if err != nil {
		fmt.Println("Error generating JWT:", err)
		return "", err
	}

	fmt.Println("Generated JWT:", token)
	fmt.Println("Generated REFRESH:", refreshToken)

	return token, nil
}