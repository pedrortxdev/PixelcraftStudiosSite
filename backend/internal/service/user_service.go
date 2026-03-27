package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

type UserService struct {
	repo          *repository.UserRepository
	encryptionKey string
}

func NewUserService(repo *repository.UserRepository, encryptionKey string) *UserService {
	return &UserService{
		repo:          repo,
		encryptionKey: encryptionKey,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *models.UpdateUserRequest) error {
	err := s.repo.UpdateUser(ctx, userID, req)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}
	
	return nil
}

func (s *UserService) GetBalance(ctx context.Context, userID string) (float64, error) {
	return s.repo.GetBalance(ctx, userID)
}

func (s *UserService) UpdateBalance(ctx context.Context, tx *sql.Tx, userID string, newBalance float64) error {
	return s.repo.UpdateBalance(ctx, tx, userID, newBalance)
}