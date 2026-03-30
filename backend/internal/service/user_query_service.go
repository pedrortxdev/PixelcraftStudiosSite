package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// UserQueryService handles read operations for user data aggregation
// This service is optimized for fetching user details with related data
type UserQueryService struct {
	userRepo *repository.UserRepository
	txRepo   *repository.TransactionRepository
	subRepo  *repository.SubscriptionRepository
	libRepo  *repository.LibraryRepository
	roleRepo *repository.RoleRepository
}

// NewUserQueryService creates a new UserQueryService
func NewUserQueryService(
	userRepo *repository.UserRepository,
	txRepo *repository.TransactionRepository,
	subRepo *repository.SubscriptionRepository,
	libRepo *repository.LibraryRepository,
	roleRepo *repository.RoleRepository,
) *UserQueryService {
	return &UserQueryService{
		userRepo: userRepo,
		txRepo:   txRepo,
		subRepo:  subRepo,
		libRepo:  libRepo,
		roleRepo: roleRepo,
	}
}

// GetUserDetailResult contains all aggregated data for a user
type GetUserDetailResult struct {
	User          *models.User                     `json:"user"`
	Balance       int64                            `json:"balance"`
	Transactions  []models.Transaction             `json:"transactions"`
	Subscriptions []models.Subscription            `json:"subscriptions"`
	Library       []models.UserPurchaseWithProduct `json:"library"`
}

// GetUserDetailOpt fetches complete user details with all related data
// Returns error immediately if any critical data fetch fails (no silent fallbacks)
func (s *UserQueryService) GetUserDetailOpt(ctx context.Context, userID string) (*GetUserDetailResult, error) {
	// Step 1: Validate UUID format FIRST before any DB queries
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Step 2: Get user (critical - must succeed)
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Step 3: Get balance (critical - must succeed, no silent fallback to 0)
	balance, err := s.userRepo.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}
	user.Balance = balance

	// Step 4: Get transactions (non-critical - return empty on error)
	transactions, err := s.txRepo.ListByUserID(userID, 100)
	if err != nil {
		// Log error but don't fail the whole request
		transactions = []models.Transaction{}
	}

	// Step 5: Get subscriptions (non-critical - return empty on error)
	subscriptions, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		subscriptions = []models.Subscription{}
	}

	// Step 6: Get library (non-critical - return empty on error)
	library, err := s.libRepo.GetUserLibrary(ctx, userUUID)
	if err != nil {
		library = []models.UserPurchaseWithProduct{}
	}

	// Step 7: Get user roles (non-critical - return empty on error)
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		roles = []models.RoleType{}
	}
	user.Roles = roles
	
	if highest := models.GetHighestRole(roles); highest != nil {
		user.HighestRole = highest
	}

	return &GetUserDetailResult{
		User:          user,
		Balance:       balance,
		Transactions:  transactions,
		Subscriptions: subscriptions,
		Library:       library,
	}, nil
}

// GetUserBalance fetches only the user's balance
// Use this when you don't need full user details
func (s *UserQueryService) GetUserBalance(ctx context.Context, userID string) (int64, error) {
	return s.userRepo.GetBalance(ctx, userID)
}

// GetUserWithRoles fetches user with roles populated
func (s *UserQueryService) GetUserWithRoles(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		roles = []models.RoleType{}
	}
	
	user.Roles = roles
	if highest := models.GetHighestRole(roles); highest != nil {
		user.HighestRole = highest
	}

	return user, nil
}
