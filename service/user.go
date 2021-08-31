package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User contains user's info
type User struct {
	Username     string
	PasswordHash string
	Role         string
}

// NewUser create a new user
func NewUser(username string, password string, role string) (*User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %v", err)
	}

	user := &User{
		Username:     username,
		PasswordHash: string(passwordHash),
		Role:         role,
	}

	return user, nil
}

// IsCorrectPassword check if provided password is correct
func (user *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// Clone returns a clone of user
func (user *User) Clone() *User {
	return &User{
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
	}
}
