package service

import (
	"context"
	"fmt"

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
) *AdminService {
	return &AdminService{
		repo:             repo,
		balanceService:   balanceService,
		userQueryService: userQueryService,
		depositService:   depositService,
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

// GetTopProducts gets top selling products for dashboard
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
// Validates UUID early and returns errors for critical failures (no silent fallbacks)
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
// Uses atomic balance operations to prevent race conditions
func (s *AdminService) UpdateUser(ctx context.Context, userID string, adminID string, updates map[string]interface{}) error {
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

		// Use atomic balance adjustment (handles transaction + audit internally)
		input := AdminAdjustmentInput{
			UserID:          userID,
			AdminID:         adminID,
			NewBalance:      newBalance,
			AdjustmentType:  adjType,
			Reason:          "Admin adjustment via UpdateUser",
		}

		_, err := s.balanceService.AdminAdjustment(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to adjust balance: %w", err)
		}

		// Remove balance and adjustment_type from updates map
		// (already handled by balance service)
		delete(updates, "balance")
		delete(updates, "adjustment_type")
		delete(updates, "adjustment_reason")
	}

	// If no other fields to update, we're done
	if len(updates) == 0 {
		return nil
	}

	// Update other user fields (email, name, etc.)
	// Note: balance field is already removed from updates map
	return s.userQueryService.userRepo.UpdateUserAdmin(ctx, userID, updates)
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

	if tx.ProviderPaymentID == nil {
		return fmt.Errorf("transaction has no provider payment ID")
	}

	// 2. For deposit refunds, use BalanceService which handles atomic refund
	if tx.Type == models.TransactionTypeDeposit {
		// BalanceService.RefundDeposit handles:
		// - Balance check with proper int64 formatting
		// - Atomic balance deduction
		// - Transaction status update
		// All within a single ACID transaction
		input := RefundDepositInput{
			TransactionID: transactionID,
			AdminID:       adminID,
			Reason:        "Admin-initiated refund",
		}

		if err := s.balanceService.RefundDeposit(ctx, input); err != nil {
			// Error message now correctly uses %d for int64 values
			return err
		}
		return nil
	}

	// 3. For other transaction types (e.g., purchase refunds)
	// Note: This would typically credit balance back to user
	// For now, just update status (extend as needed)
	if err := s.depositService.repo.UpdateStatus(transactionID, models.TransactionStatusRefunded); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// RefundTransactionMP calls Mercado Pago refund and updates local DB
// This is a lower-level method for external refund coordination
func (s *AdminService) RefundTransactionMP(ctx context.Context, transactionID string, adminID string) error {
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

	if tx.ProviderPaymentID == nil {
		return fmt.Errorf("transaction has no provider payment ID")
	}

	// 2. Call MP Refund first (external API)
	if err := s.depositService.RefundPayment(ctx, *tx.ProviderPaymentID); err != nil {
		return fmt.Errorf("failed to refund in Mercado Pago: %w", err)
	}

	// 3. Update local DB atomically
	if tx.Type == models.TransactionTypeDeposit {
		input := RefundDepositInput{
			TransactionID: transactionID,
			AdminID:       adminID,
			Reason:        "MP refund completed",
		}
		// This will fail if user doesn't have sufficient balance
		// which is the correct safety behavior
		if err := s.balanceService.RefundDeposit(ctx, input); err != nil {
			return fmt.Errorf("refunded in MP but failed to update local DB: %w", err)
		}
	} else {
		if err := s.depositService.repo.UpdateStatus(transactionID, models.TransactionStatusRefunded); err != nil {
			return fmt.Errorf("refunded in MP but failed to update status: %w", err)
		}
	}

	return nil
}