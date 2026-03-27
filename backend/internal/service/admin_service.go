package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"math"
	"time"
)

type AdminService struct {
	repo           *repository.AdminRepository
	txRepo         *repository.TransactionRepository
	userRepo       *repository.UserRepository
	subRepo        *repository.SubscriptionRepository
	libRepo        *repository.LibraryRepository
	roleRepo       *repository.RoleRepository
	depositService *DepositService
}

type AdminUserDetail struct {
	User          *models.User                         `json:"user"`
	Balance       float64                              `json:"balance"`
	Transactions  []models.Transaction                 `json:"transactions"`
	Subscriptions []models.Subscription                `json:"subscriptions"`
	Library       []models.UserPurchaseWithProduct     `json:"library"`
}

func NewAdminService(
	repo *repository.AdminRepository,
	txRepo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	depositService *DepositService,
	subRepo *repository.SubscriptionRepository,
	libRepo *repository.LibraryRepository,
	roleRepo *repository.RoleRepository,
) *AdminService {
	return &AdminService{
		repo:           repo,
		txRepo:         txRepo,
		userRepo:       userRepo,
		depositService: depositService,
		subRepo:        subRepo,
		libRepo:        libRepo,
		roleRepo:       roleRepo,
	}
}

func (s *AdminService) GetDashboardStats(ctx context.Context) (*repository.AnalyticsSnapshot, error) {
	return s.repo.GetAnalyticsSnapshot()
}

func (s *AdminService) GetRecentOrders(ctx context.Context) ([]repository.RecentOrder, error) {
	return s.repo.GetRecentOrders()
}

func (s *AdminService) GetTopProducts(ctx context.Context) ([]repository.TopProduct, error) {
	return s.repo.GetTopProducts()
}

// ListTransactions lists transactions with pagination and filtering
func (s *AdminService) ListTransactions(ctx context.Context, page, limit int, status string) ([]repository.TransactionWithUser, int, error) {
	return s.txRepo.ListTransactions(page, limit, status)
}

// GetMercadoPagoBalance gets the Mercado Pago account balance
func (s *AdminService) GetMercadoPagoBalance(ctx context.Context) (*MPBalanceResponse, error) {
	return s.depositService.GetAccountBalance(ctx)
}

// ListUsers lists users with pagination
func (s *AdminService) ListUsers(ctx context.Context, page, limit int, search string) ([]models.User, int, error) {
	return s.userRepo.ListAll(ctx, page, limit, search)
}

// GetUserDetail returns full details for a user
func (s *AdminService) GetUserDetail(ctx context.Context, userID string) (*AdminUserDetail, error) {
	// 1. Get User
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// 2. Get Balance (already in user struct usually, but verify)
	balance, err := s.userRepo.GetBalance(ctx, userID)
	if err != nil {
		balance = 0
	}
	user.Balance = balance // Ensure consistent

	// 3. Get Transactions
	txs, err := s.txRepo.ListByUserID(userID, 100) // Limit to 100 recent
	if err != nil {
		txs = []models.Transaction{}
	}

	// 4. Get Subscriptions
	subs, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		subs = []models.Subscription{}
	}

	// 5. Get Library
	// Parse UUID
	uid, err := uuid.Parse(userID)
	var library []models.UserPurchaseWithProduct
	if err == nil {
		library, err = s.libRepo.GetUserLibrary(ctx, uid)
		if err != nil {
			library = []models.UserPurchaseWithProduct{}
		}
	}

	// 6. Get User Roles
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		roles = []models.RoleType{}
	}
	user.Roles = roles
	if highest := models.GetHighestRole(roles); highest != nil {
		user.HighestRole = highest
	}

	return &AdminUserDetail{
		User:          user,
		Balance:       balance,
		Transactions:  txs,
		Subscriptions: subs,
		Library:       library,
	}, nil
}

// UpdateUser updates user details
func (s *AdminService) UpdateUser(ctx context.Context, userID string, adminID string, updates map[string]interface{}) error {
	// Check for balance change
	var balanceChanged bool
	var balanceDiff float64
	var oldBal, newBal float64

	if val, ok := updates["balance"]; ok {
		newBalOK, ok := val.(float64)
		if ok {
			user, err := s.userRepo.GetUserByID(ctx, userID)
			if err == nil {
				if user.Balance != newBalOK {
					balanceChanged = true
					balanceDiff = newBalOK - user.Balance
					oldBal = user.Balance
					newBal = newBalOK
				}
			}
		}
	}

	// Extract and remove adjustment_type as it's not a field in the users table
	var adjType *string
	if val, ok := updates["adjustment_type"]; ok {
		if sVal, ok := val.(string); ok {
			adjType = &sVal
		}
		delete(updates, "adjustment_type")
	}

	// Perform the update (includes balance update in DB)
	err := s.userRepo.UpdateUserAdmin(ctx, userID, updates)
	if err != nil {
		return err
	}

	// Create transaction log if balance changed
	if balanceChanged {
		txID := uuid.New()
		paymentID := fmt.Sprintf("admin-%s", adminID)
		
		tx := &models.Transaction{
			ID:                txID,
			UserID:            uuid.MustParse(userID),
			ProviderPaymentID: &paymentID,
			Amount:            math.Abs(balanceDiff), // Record absolute amount
			Status:            models.TransactionStatusCompleted,
			Type:              models.TransactionTypeAdminAdjustment,
			AdjustmentType:    adjType,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		
		s.txRepo.Create(tx)

		// Audit Log BT-012
		s.repo.LogAction(ctx, adminID, "balance_update", fmt.Sprintf("User %s balance changed from %.2f to %.2f (diff: %.2f, type: %v)", userID, oldBal, newBal, balanceDiff, adjType))
	}

	return nil
}

// UpdateUserPassword resets user password
func (s *AdminService) UpdateUserPassword(ctx context.Context, userID, newPassword string) error {
	// Hash password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.userRepo.UpdatePassword(ctx, userID, string(hashedBytes))
}

// RefundTransaction handles the full refund logic
func (s *AdminService) RefundTransaction(ctx context.Context, transactionID string) error {
	// 1. Get Transaction
	tx, err := s.txRepo.GetByID(transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("transaction not found")
	}

	if tx.Status != models.TransactionStatusCompleted {
		return fmt.Errorf("transaction is not completed, cannot refund")
	}

	if tx.ProviderPaymentID == nil {
		return fmt.Errorf("transaction has no provider payment ID")
	}

	// 2. Check User Balance (Safety Check)
	if tx.Type == models.TransactionTypeDeposit {
		balance, err := s.userRepo.GetBalance(ctx, tx.UserID.String())
		if err != nil {
			return fmt.Errorf("failed to get user balance: %w", err)
		}
		if balance < tx.Amount {
			return fmt.Errorf("insufficient user balance to refund (Current: %.2f, Required: %.2f)", balance, tx.Amount)
		}
	}

	// 3. Call MP Refund
	if err := s.depositService.RefundPayment(ctx, *tx.ProviderPaymentID); err != nil {
		return fmt.Errorf("failed to refund in Mercado Pago: %w", err)
	}

	// 4. Update DB (Deduct balance and update status)
	if tx.Type == models.TransactionTypeDeposit {
		// Use RefundDeposit from repository which is transactional
		if err := s.txRepo.RefundDeposit(transactionID, tx.Amount); err != nil {
			return fmt.Errorf("refunded in MP but failed to update local DB: %w", err)
		}
	} else {
		// For other types (e.g. Purchase), strictly speaking we should refund ONLY if it was a purchase reversal.
		// If it's a purchase refund, usually it credits balance back to user.
		// But here `RefundTransaction` naming implies "Reverse the transaction".
		// Since we only really process Deposits via MP here (according to context), the logic holds.
		if err := s.txRepo.UpdateStatus(transactionID, models.TransactionStatusRefunded); err != nil {
			return fmt.Errorf("refunded in MP but failed to update status: %w", err)
		}
	}

	return nil
}