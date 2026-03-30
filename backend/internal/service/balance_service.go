package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// BalanceService handles atomic wallet balance operations
// This service ensures ACID compliance for all balance-related transactions
type BalanceService struct {
	txRepo *repository.TransactionRepository
	userRepo *repository.UserRepository
	db *sql.DB
}

// NewBalanceService creates a new BalanceService
func NewBalanceService(
	txRepo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	db *sql.DB,
) *BalanceService {
	return &BalanceService{
		txRepo: txRepo,
		userRepo: userRepo,
		db: db,
	}
}

// AdminAdjustmentInput represents the input for an admin balance adjustment
type AdminAdjustmentInput struct {
	UserID          string
	AdminID         string
	NewBalance      int64
	AdjustmentType  *string
	Reason          string
}

// AdminAdjustmentResult contains the result of an adjustment operation
type AdminAdjustmentResult struct {
	OldBalance      int64
	NewBalance      int64
	Difference      int64
	TransactionID   uuid.UUID
}

// AdminAdjustment performs an atomic balance adjustment with full audit trail
// Uses FOR UPDATE row locking to prevent race conditions
func (s *BalanceService) AdminAdjustment(ctx context.Context, input AdminAdjustmentInput) (*AdminAdjustmentResult, error) {
	// Parse and validate user ID upfront
	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	adminUUID, err := uuid.Parse(input.AdminID)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID format: %w", err)
	}

	// Start transaction with proper isolation
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock the user row and get current balance (atomic read)
	var oldBalance int64
	query := `SELECT balance FROM users WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, query, input.UserID).Scan(&oldBalance); err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	// Calculate difference
	balanceDiff := input.NewBalance - oldBalance

	// Atomically update balance using incremental update (prevents race conditions)
	updateQuery := `UPDATE users SET balance = balance + $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateQuery, balanceDiff, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// Create transaction record
	txID := uuid.New()
	providerPaymentID := fmt.Sprintf("admin-adjustment-%s", adminUUID.String()[:8])
	
	// Record absolute value for amount
	absDiff := balanceDiff
	if absDiff < 0 {
		absDiff = -absDiff
	}

	transaction := &models.Transaction{
		ID:                txID,
		UserID:            userUUID,
		ProviderPaymentID: &providerPaymentID,
		Amount:            absDiff,
		Status:            models.TransactionStatusCompleted,
		Type:              models.TransactionTypeAdminAdjustment,
		AdjustmentType:    input.AdjustmentType,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.txRepo.Create(transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &AdminAdjustmentResult{
		OldBalance:    oldBalance,
		NewBalance:    input.NewBalance,
		Difference:    balanceDiff,
		TransactionID: txID,
	}, nil
}

// RefundDepositInput represents the input for refunding a deposit
type RefundDepositInput struct {
	TransactionID   string
	AdminID         string
	Reason          string
}

// RefundDeposit performs an atomic refund of a deposit transaction
// Includes balance check and deduction within the same transaction
func (s *BalanceService) RefundDeposit(ctx context.Context, input RefundDepositInput) error {
	// Parse and validate IDs upfront
	txUUID, err := uuid.Parse(input.TransactionID)
	if err != nil {
		return fmt.Errorf("invalid transaction ID format: %w", err)
	}

	adminUUID, err := uuid.Parse(input.AdminID)
	if err != nil {
		return fmt.Errorf("invalid admin ID format: %w", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock and get transaction with FOR UPDATE
	var transaction models.Transaction
	query := `SELECT id, user_id, provider_payment_id, amount, status, type, adjustment_type, qr_code, qr_code_base64, created_at, updated_at FROM transactions WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, query, txUUID).Scan(
		&transaction.ID, &transaction.UserID, &transaction.ProviderPaymentID, &transaction.Amount,
		&transaction.Status, &transaction.Type, &transaction.AdjustmentType,
		&transaction.QRCode, &transaction.QRCodeBase64, &transaction.CreatedAt, &transaction.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("transaction not found")
		}
		return fmt.Errorf("failed to lock transaction: %w", err)
	}

	// Validate transaction status
	if transaction.Status != models.TransactionStatusCompleted {
		return fmt.Errorf("transaction is not completed, cannot refund (status: %s)", transaction.Status)
	}

	if transaction.Type != models.TransactionTypeDeposit {
		return fmt.Errorf("transaction is not a deposit, cannot refund (type: %s)", transaction.Type)
	}

	if transaction.ProviderPaymentID == nil {
		return fmt.Errorf("transaction has no provider payment ID")
	}

	// Lock and check user balance
	userID := transaction.UserID.String()
	var balance int64
	balanceQuery := `SELECT balance FROM users WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, balanceQuery, userID).Scan(&balance); err != nil {
		return fmt.Errorf("failed to get user balance: %w", err)
	}

	if balance < transaction.Amount {
		return fmt.Errorf("insufficient user balance to refund (Current: %d cents, Required: %d cents)", balance, transaction.Amount)
	}

	// Deduct balance atomically
	deductQuery := `UPDATE users SET balance = balance - $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, deductQuery, transaction.Amount, userID); err != nil {
		return fmt.Errorf("failed to deduct balance: %w", err)
	}

	// Update transaction status to refunded
	updateQuery := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateQuery, models.TransactionStatusRefunded, txUUID); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Create audit transaction record for the refund
	refundTxID := uuid.New()
	providerPaymentID := fmt.Sprintf("admin-refund-%s", adminUUID.String()[:8])

	refundTransaction := &models.Transaction{
		ID:                refundTxID,
		UserID:            transaction.UserID,
		ProviderPaymentID: &providerPaymentID,
		Amount:            transaction.Amount,
		Status:            models.TransactionStatusCompleted,
		Type:              models.TransactionTypeAdminAdjustment,
		AdjustmentType:    &[]string{"Refund"}[0],
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.txRepo.Create(refundTransaction); err != nil {
		return fmt.Errorf("failed to create refund transaction record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
