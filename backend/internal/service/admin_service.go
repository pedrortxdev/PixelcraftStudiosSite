package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AdminService orchestrates admin operations by delegating to domain-specific services
// Follows SRP: this service coordinates, domain services handle business logic
type AdminService struct {
	repo             *repository.AdminRepository
	balanceService   *BalanceService
	userQueryService *UserQueryService
	depositService   *DepositService
	db               *sqlx.DB
}

// AdminUserDetail represents aggregated user data for admin view
type AdminUserDetail struct {
	User          *models.User                         `json:"user"`
	Balance       int64                                `json:"balance"` // Balance in cents
	Transactions  []models.Transaction                 `json:"transactions"`
	Subscriptions []models.Subscription                `json:"subscriptions"`
	Library       []models.UserPurchaseWithProduct     `json:"library"`
}

// NewAdminService creates a new AdminService with proper dependencies
func NewAdminService(
	repo *repository.AdminRepository,
	balanceService *BalanceService,
	userQueryService *UserQueryService,
	depositService *DepositService,
	db *sqlx.DB,
) *AdminService {
	return &AdminService{
		repo:             repo,
		balanceService:   balanceService,
		userQueryService: userQueryService,
		depositService:   depositService,
		db:               db,
	}
}

// GetDashboardStats gets analytics snapshot for dashboard
func (s *AdminService) GetDashboardStats(ctx context.Context) (*repository.AnalyticsSnapshot, error) {
	return s.repo.GetAnalyticsSnapshot()
}

// GetRecentOrders gets recent orders for dashboard
func (s *AdminService) GetRecentOrders(ctx context.Context) ([]repository.RecentOrder, error) {
	return s.repo.GetRecentOrders()
}

// GetTopProducts gets top products for dashboard
func (s *AdminService) GetTopProducts(ctx context.Context) ([]repository.TopProduct, error) {
	return s.repo.GetTopProducts()
}

// ListTransactions lists transactions with pagination and filtering
func (s *AdminService) ListTransactions(ctx context.Context, page, limit int, status string) ([]repository.TransactionWithUser, int, error) {
	// Delegate to deposit service's repository
	return s.depositService.repo.ListTransactions(page, limit, status)
}

// GetMercadoPagoBalance gets the Mercado Pago account balance
func (s *AdminService) GetMercadoPagoBalance(ctx context.Context) (*MPBalanceResponse, error) {
	return s.depositService.GetAccountBalance(ctx)
}

// ListUsers lists users with pagination
func (s *AdminService) ListUsers(ctx context.Context, page, limit int, search string) ([]models.User, int, error) {
	// Delegate to user query service's repository
	return s.userQueryService.userRepo.ListAll(ctx, page, limit, search)
}

// GetUserDetail returns full details for a user
func (s *AdminService) GetUserDetail(ctx context.Context, userID string) (*AdminUserDetail, error) {
	// Delegate to UserQueryService which handles all the aggregation logic
	result, err := s.userQueryService.GetUserDetailOpt(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &AdminUserDetail{
		User:          result.User,
		Balance:       result.Balance,
		Transactions:  result.Transactions,
		Subscriptions: result.Subscriptions,
		Library:       result.Library,
	}, nil
}

// UpdateUser updates user details
// Uses a single DB transaction to ensure atomicity across balance and profile updates
func (s *AdminService) UpdateUser(ctx context.Context, userID string, adminID string, updates map[string]interface{}) error {
	// Start transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if balance is being updated
	if newBalanceVal, ok := updates["balance"]; ok {
		// JSON numbers are float64, convert to int64
		newBalanceFloat, ok := newBalanceVal.(float64)
		if !ok {
			return fmt.Errorf("invalid balance value type")
		}
		newBalance := int64(newBalanceFloat)

		// Extract adjustment_type for audit
		var adjType *string
		if val, ok := updates["adjustment_type"]; ok {
			if sVal, ok := val.(string); ok {
				adjType = &sVal
			}
		}

		// Use atomic balance adjustment within the transaction
		input := AdminAdjustmentInput{
			UserID:          userID,
			AdminID:         adminID,
			NewBalance:      newBalance,
			AdjustmentType:  adjType,
			Reason:          "Admin adjustment via UpdateUser",
		}

		_, err := s.balanceService.AdminAdjustmentTx(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("failed to adjust balance: %w", err)
		}

		// Remove balance and adjustment_type from updates map
		delete(updates, "balance")
		delete(updates, "adjustment_type")
		delete(updates, "adjustment_reason")
	}

	// If there are other fields to update (email, name, etc.)
	if len(updates) > 0 {
		// Pass the transaction to the repository
		if err := s.userQueryService.userRepo.UpdateUserAdminTx(ctx, tx.Tx, userID, updates); err != nil {
			return fmt.Errorf("failed to update user profile: %w", err)
		}
	}

	// Commit everything atomically
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit update: %w", err)
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
	return s.userQueryService.userRepo.UpdatePassword(ctx, userID, string(hashedBytes))
}

// RefundTransaction handles the full refund logic
func (s *AdminService) RefundTransaction(ctx context.Context, transactionID string, adminID string) error {
	// 1. Get Transaction
	tx, err := s.depositService.repo.GetByID(transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		return fmt.Errorf("transaction not found")
	}

	if tx.Status != models.TransactionStatusCompleted {
		return fmt.Errorf("transaction is not completed, cannot refund")
	}

	// 2. For deposit refunds, use BalanceService which handles atomic refund
	if tx.Type == models.TransactionTypeDeposit {
		input := RefundDepositInput{
			TransactionID: transactionID,
			AdminID:       adminID,
			Reason:        "Admin-initiated refund",
		}

		if err := s.balanceService.RefundDeposit(ctx, input); err != nil {
			return err
		}
		return nil
	}

	// 3. For other transaction types
	if err := s.depositService.repo.UpdateStatus(transactionID, models.TransactionStatusRefunded); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// RefundTransactionMP calls Mercado Pago refund and updates local DB safely
// Uses a "Deduct-then-API-then-Commit" strategy to prevent "Infinite Money" glitches
func (s *AdminService) RefundTransactionMP(ctx context.Context, transactionID string, adminID string) error {
	// 1. Get Transaction
	txData, err := s.depositService.repo.GetByID(transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if txData == nil {
		return fmt.Errorf("transaction not found")
	}

	if txData.Status != models.TransactionStatusCompleted {
		return fmt.Errorf("transaction is not completed, cannot refund")
	}

	if txData.ProviderPaymentID == nil {
		return fmt.Errorf("transaction has no provider payment ID")
	}

	// 2. Start local DB transaction to "reserve" the balance deduction
	dbTx, err := s.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer dbTx.Rollback()

	// 3. Perform local balance deduction first (atomic check & update)
	if txData.Type == models.TransactionTypeDeposit {
		input := RefundDepositInput{
			TransactionID: transactionID,
			AdminID:       adminID,
			Reason:        "MP refund in progress",
		}
		// This will fail and rollback if user has insufficient balance
		if err := s.balanceService.RefundDepositTx(ctx, dbTx, input); err != nil {
			return fmt.Errorf("failed to deduct balance locally: %w", err)
		}
	} else {
		if err := s.depositService.repo.UpdateStatus(transactionID, models.TransactionStatusRefunded); err != nil {
			return fmt.Errorf("failed to update status locally: %w", err)
		}
	}

	// 4. Call external API (Mercado Pago) while holding the DB transaction open
	// This ensures that if MP fails, we rollback the balance deduction.
	// If MP succeeds, we commit.
	if err := s.depositService.RefundPayment(ctx, *txData.ProviderPaymentID); err != nil {
		return fmt.Errorf("failed to refund in Mercado Pago: %w (local changes rolled back)", err)
	}

	// 5. Commit local DB changes only after MP success
	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("MP refund succeeded but failed to commit local DB: %w", err)
	}

	return nil
}