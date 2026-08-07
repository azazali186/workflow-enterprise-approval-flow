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

// UserResponse is the safe representation of a user (no sensitive data)
type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	Roles       []Role     `json:"roles,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts a User to a safe UserResponse
func (u *User) ToResponse() *UserResponse {
	roles := make([]Role, len(u.Roles))
	copy(roles, u.Roles)

	return &UserResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		Name:        u.Name,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		Roles:       roles,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// IsAdmin checks if the user has the admin role
func (u *User) IsAdmin() bool {
	for _, role := range u.Roles {
		if role.Name == "admin" {
			return true
		}
	}
	return false
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// GetRoleNames returns a list of role names
func (u *User) GetRoleNames() []string {
	roleNames := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		roleNames[i] = role.Name
	}
	return roleNames
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == "active"
}

// User status constants
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusLocked   = "locked"
	UserStatusPending  = "pending"
)
