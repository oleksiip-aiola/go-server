package db

import (
	"errors"
	"fmt"
	"log"
	"time"

	"connectrpc.com/connect"
	"github.com/go-webauthn/webauthn/webauthn"
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

	// WebAuthn-specific fields
	CredentialID        []byte `gorm:"type:bytea" json:"credentialID"`        // WebAuthn Credential ID
	PublicKey           []byte `gorm:"type:bytea" json:"publicKey"`           // Public Key used for authentication
	AuthenticatorAAGUID []byte `gorm:"type:bytea" json:"authenticatorAAGUID"` // Authenticator AAGUID
	SignCount           uint32 `json:"signCount"`                             // Sign counter to prevent replay attacks
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
	fmt.Println(user)
	if err := DBConn.Create(&user).Error; err != nil {
		return "", err
	}

	QueueShardWrite(user)

	return user.UserId, nil
}

func (u *User) CreateWebAuthnAdmin(webAuthnUser *User) (string, error) {
	webAuthnUser.IsAdmin = true
	// fmt.Println(webAuthnUser, &webAuthnUser)
	var copyuser *User
	var publickey struct {
		PublicKey []byte
	}
	fmt.Println("user key", webAuthnUser.PublicKey)
	DBConn.Where("user_id = ?", "214966e6-2708-41b8-8aa1-d7b42c702aea").Select("public_key").Scan(&publickey)
	fmt.Println("db key", publickey)
	fmt.Println(copyuser)
	if err := DBConn.Create(&webAuthnUser).Error; err != nil {
		return "", err
	}

	QueueShardWrite(*webAuthnUser)

	return webAuthnUser.UserId, nil
}

func (u *User) LoginAsAdmin(email string, password string) (*User, error) {

	if err := DBConn.Where("email = ? AND is_admin = ?", email, true).First(&u).Error; err != nil {
		fmt.Println(err)
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not found"))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		fmt.Println(err)

		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("password is incorrect"))
	}

	return u, nil
}

func (u *User) LoginAsWebAuthAdmin(userId string) (*User, error) {
	if err := DBConn.Where("user_id = ? AND is_admin = ?", userId, true).First(&u).Error; err != nil {
		fmt.Println(err)
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("user not found"))
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

func GetUserById(id string) (User, error) {
	user, err := ReadFromShard(id)

	return user, err
}

type RefreshToken struct {
	ID          string `gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID      string `json:"userID"`
	JTI         string `json:"jti"`
	AccessToken string `json:"accessToken"`
	Expiry      string `json:"expiry"`
	IsRevoked   bool   `json:"isRevoked"`
}

func StoreJTI(jti string, userID string, refreshTokenExp string, accessToken string) error {
	refreshToken := RefreshToken{
		UserID:      userID,
		JTI:         jti,
		Expiry:      refreshTokenExp,
		IsRevoked:   false,
		AccessToken: accessToken,
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

// Implement WebAuthn User interface for the User struct
func (u *User) WebAuthnID() []byte {
	return []byte(u.UserId) // Use UUID as the WebAuthn ID
}

func (u *User) WebAuthnName() string {
	return u.Email // Use the email address as the WebAuthn name
}

func (u *User) WebAuthnDisplayName() string {
	return u.FirstName + " " + u.LastName // Full name for display
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return []webauthn.Credential{
		{
			ID:        u.CredentialID,
			PublicKey: u.PublicKey,
		},
	}
}
