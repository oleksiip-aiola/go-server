package db

import (
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	UserId    string     `gorm:"type:uuid;default:uuid_generate_v4()" json:"userId"`
	Email     string     `gorm:"unique" json:"email"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Password  string     `json:"-"`
	IsAdmin   bool       `gorm:"default:false" json:"isAdmin"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (u *User) CreateAdmin(email string, password string, firstName string, lastName string) (string, error) {
	user := User{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		IsAdmin:   true,
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)

	if err != nil {
		return "", errors.New("failed to hash password")
	}

	user.Password = string(hashedPassword)

	if err := DBConn.Create(&user).Error; err != nil {
		return "", err
	}

	return user.UserId, nil
}

func (u *User) LoginAsAdmin(email string, password string) (*User, error) {
	fmt.Println(email)
	if err := DBConn.Where("email = ? AND is_admin = ?", email, true).First(&u).Error; err != nil {
		return nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("password is incorrect")
	}

	return u, nil
}

func RevokeJWTByUserId(userId string) error {

	err := DBConn.Model(&RefreshToken{}).Where("user_id = ?", userId).Update("is_revoked", true).Error

	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
		return err
	}

	fmt.Println("Revoked JWT for user ID:", userId)

	return nil
}

func GetUserById(id string) (*User, error) {
	user := &User{}

	if err := DBConn.Model(&User{}).Where("user_id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

type RefreshToken struct {
	ID        string `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID    string `json:"userID"`
	JTI       string `json:"jti"`
	Expiry    string `json:"expiry"`
	IsRevoked bool   `json:"isRevoked"`
}

func StoreJTI(jti string, userID string, refreshTokenExp string) error {
	refreshToken := RefreshToken{
		UserID:    userID,
		JTI:       jti,
		Expiry:    refreshTokenExp,
		IsRevoked: false,
	}

	if err := DBConn.Table("refresh_tokens").Create(&refreshToken).Error; err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func CheckIfRefreshTokenIsRevokedByUserId(userId string) (string, error) {
	var refreshToken RefreshToken

	err := DBConn.Table("refresh_tokens").Where("user_id = ? AND expiry > NOW()", userId).Order("token_id DESC").Limit(1).Select("is_revoked, jti").Scan(&refreshToken).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("refresh token is expired or invalid")
		}
		return "", err
	}

	isRevoked := refreshToken.IsRevoked

	if isRevoked {
		return "", fmt.Errorf("refresh token is revoked")
	}

	return refreshToken.JTI, nil
}
