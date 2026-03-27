package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// UserService handles business logic for users
type UserService struct {
	repo          *repository.UserRepository
	encryptionKey string
}

// NewUserService creates a new UserService
func NewUserService(repo *repository.UserRepository, encryptionKey string) *UserService {
	return &UserService{
		repo:          repo,
		encryptionKey: encryptionKey,
	}
}

// GetProfile retrieves a user's profile
func (s *UserService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	// Get user data from repository
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	
	return user, nil
}

// UpdateProfile updates a user's profile
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *models.UpdateUserRequest) error {
	// Update user data in repository
	err := s.repo.UpdateUser(ctx, userID, req)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}
	
	return nil
}

// GetBalance retrieves the balance for a user
func (s *UserService) GetBalance(ctx context.Context, userID string) (float64, error) {
	return s.repo.GetBalance(ctx, userID)
}

// UpdateBalance updates the balance for a user
func (s *UserService) UpdateBalance(ctx context.Context, tx *sql.Tx, userID string, newBalance float64) error {
	return s.repo.UpdateBalance(ctx, tx, userID, newBalance)
}