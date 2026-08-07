package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a system user
type User struct {
	Base
	Email       string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password    string     `gorm:"size:255;not null" json:"-"`
	Name        string     `gorm:"size:255;not null" json:"name"`
	Status      string     `gorm:"size:20;default:active;index" json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	Roles       []Role     `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// HashPassword hashes the password using bcrypt
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword compares a hashed password with a plain password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
