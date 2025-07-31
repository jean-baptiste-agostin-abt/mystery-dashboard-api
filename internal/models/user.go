package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TenantID  string         `json:"tenant_id" gorm:"type:varchar(36);not null;index:idx_tenant_email"`
	Email     string         `json:"email" gorm:"type:varchar(255);not null;index:idx_tenant_email;uniqueIndex:idx_tenant_email_unique"`
	Password  string         `json:"-" gorm:"type:varchar(255);not null"` // Never include in JSON responses
	FirstName string         `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName  string         `json:"last_name" gorm:"type:varchar(100);not null"`
	Role      string         `json:"role" gorm:"type:varchar(50);not null;default:'viewer'"`
	Status    string         `json:"status" gorm:"type:varchar(50);not null;default:'active'"`
	LastLogin *time.Time     `json:"last_login,omitempty" gorm:"type:timestamp"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// UserRole defines user roles
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleEditor    UserRole = "editor"
	RoleViewer    UserRole = "viewer"
	RolePublisher UserRole = "publisher"
)

// UserStatus defines user statuses
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Role      string `json:"role" validate:"required,oneof=admin editor viewer publisher"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      *User     `json:"user"`
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(user *User) error
	GetByID(tenantID, id string) (*User, error)
	GetByEmail(tenantID, email string) (*User, error)
	Update(user *User) error
	Delete(tenantID, id string) error
	List(tenantID string, limit, offset int) ([]*User, error)
	UpdateLastLogin(tenantID, id string) error
}

// UserService handles business logic for users
type UserService struct {
	repo UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(tenantID string, req *CreateUserRequest) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		TenantID:  tenantID,
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		Status:    string(StatusActive),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(tenantID, id string) (*User, error) {
	return s.repo.GetByID(tenantID, id)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(tenantID, email string) (*User, error) {
	return s.repo.GetByEmail(tenantID, email)
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(user *User) error {
	user.UpdatedAt = time.Now()
	return s.repo.Update(user)
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(tenantID, id string) error {
	return s.repo.Delete(tenantID, id)
}

// ListUsers retrieves a list of users with pagination
func (s *UserService) ListUsers(tenantID string, limit, offset int) ([]*User, error) {
	return s.repo.List(tenantID, limit, offset)
}

// AuthenticateUser authenticates a user with email and password
func (s *UserService) AuthenticateUser(tenantID string, req *LoginRequest) (*User, error) {
	user, err := s.repo.GetByEmail(tenantID, req.Email)
	if err != nil {
		return nil, err
	}

	if user.Status != string(StatusActive) {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(tenantID, user.ID); err != nil {
		// Log error but don't fail authentication
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(tenantID, userID, newPassword string) error {
	user, err := s.repo.GetByID(tenantID, userID)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.repo.Update(user)
}

// HasPermission checks if a user has a specific permission
func (u *User) HasPermission(permission string) bool {
	switch UserRole(u.Role) {
	case RoleAdmin:
		return true // Admin has all permissions
	case RoleEditor:
		return permission == "read" || permission == "write" || permission == "edit"
	case RolePublisher:
		return permission == "read" || permission == "publish"
	case RoleViewer:
		return permission == "read"
	default:
		return false
	}
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == string(StatusActive) && !u.DeletedAt.Valid
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}
