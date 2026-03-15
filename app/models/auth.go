package models

import "gorm.io/gorm"

// AuthUser represents an user that might be authenticated.
type AuthUser struct {
	ID        uint
	Email     string
	LoggedIn  bool
	FirstName string
	Role      string
}

// Check should return true if the user is authenticated.
// See handlers/auth.go.
func (user AuthUser) Check() bool {
	return user.ID > 0 && user.LoggedIn
}

func (user AuthUser) checkRole() string {
	return user.Role
}

type User struct {
	gorm.Model
	Email           string
	FirstName       string
	LastName        string
	Role            string
	PasswordHash    string
	EmailVerifiedAt string
	CreatedAt       string
	UpdatedAt       string
}
