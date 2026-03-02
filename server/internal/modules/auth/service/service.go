// Package service provides authentication and authorization functionality.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/agentteams/server/internal/modules/auth/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	// ErrUserNotFound is returned when user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists is returned when user already exists.
	ErrUserExists = errors.New("user already exists")
	// ErrInvalidCredentials is returned when credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrAccountLocked is returned when account is locked.
	ErrAccountLocked = errors.New("account is locked")
)

// Repository handles user data persistence.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new user.
func (r *Repository) Create(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUserExists
		}
		return result.Error
	}
	return nil
}

// GetByID gets a user by ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByUsername gets a user by username.
func (r *Repository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByEmail gets a user by email.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// List lists users with pagination.
func (r *Repository) List(ctx context.Context, page, pageSize int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.User{}).Where("deleted_at IS NULL")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return users, total, nil
}

// Update updates a user.
func (r *Repository) Update(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete soft deletes a user.
func (r *Repository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateLastLogin updates the last login time.
func (r *Repository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("last_login_at", now)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Service provides user business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new user service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateUser creates a new user with hashed password.
func (s *Service) CreateUser(ctx context.Context, username, password, email, role string) (*domain.User, error) {
	// Check if username exists
	if _, err := s.repo.GetByUsername(ctx, username); err == nil {
		return nil, ErrUserExists
	}

	// Check if email exists (if provided)
	if email != "" {
		if _, err := s.repo.GetByEmail(ctx, email); err == nil {
			return nil, ErrUserExists
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		Role:         role,
		Status:       domain.StatusActive,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Authenticate authenticates a user by username and password.
func (s *Service) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if account is locked
	if user.Status == domain.StatusLocked {
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login time
	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	return user, nil
}

// GetUser gets a user by ID.
func (s *Service) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// ListUsers lists users with pagination.
func (s *Service) ListUsers(ctx context.Context, page, pageSize int) ([]domain.User, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// UpdateUser updates a user.
func (s *Service) UpdateUser(ctx context.Context, id string, updates map[string]interface{}) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if role, ok := updates["role"].(string); ok {
		user.Role = role
	}
	if status, ok := updates["status"].(string); ok {
		user.Status = status
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword changes a user's password.
func (s *Service) ChangePassword(ctx context.Context, id, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return s.repo.Update(ctx, user)
}

// DeleteUser deletes a user.
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
