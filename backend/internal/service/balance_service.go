package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// BalanceService handles atomic wallet balance operations
// This service ensures ACID compliance for all balance-related transactions
type BalanceService struct {
	txRepo *repository.TransactionRepository
	userRepo *repository.UserRepository
	db *sqlx.DB
}

// NewBalanceService creates a new BalanceService
func NewBalanceService(
	txRepo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	db *sqlx.DB,
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
	// Start transaction with proper isolation
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := s.AdminAdjustmentTx(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// AdminAdjustmentTx performs the adjustment within an existing transaction
func (s *BalanceService) AdminAdjustmentTx(ctx context.Context, tx *sqlx.Tx, input AdminAdjustmentInput) (*AdminAdjustmentResult, error) {
	// Parse and validate IDs upfront
	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	adminUUID, err := uuid.Parse(input.AdminID)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID format: %w", err)
	}

	// 1. Lock the user row (The "Owner" entity always comes first in the lock hierarchy)
	var oldBalance int64
	query := `SELECT balance FROM users WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, query, userUUID).Scan(&oldBalance); err != nil {
		return nil, fmt.Errorf("failed to lock user balance: %w", err)
	}

	// 2. Update balance using the absolute new value (we already have the lock)
	updateQuery := `UPDATE users SET balance = $1, updated_at = NOW() WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateQuery, input.NewBalance, userUUID); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// 3. Create audit trail
	balanceDiff := input.NewBalance - oldBalance
	txID := uuid.New()
	
	// Ensure providerPaymentID is UNIQUE per operation by combining Admin ID and Transaction ID
	providerPaymentID := fmt.Sprintf("adm-adj-%s-%s", adminUUID.String()[:4], txID.String()[:8])
	
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

	if err := s.txRepo.CreateTx(tx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
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
	// Start transaction
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.RefundDepositTx(ctx, tx, input); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RefundDepositTx performs the refund within an existing transaction
func (s *BalanceService) RefundDepositTx(ctx context.Context, tx *sqlx.Tx, input RefundDepositInput) error {
	// 1. Validate inputs immediately (Fail-Fast)
	txUUID, err := uuid.Parse(input.TransactionID)
	if err != nil {
		return fmt.Errorf("invalid transaction ID format: %w", err)
	}

	adminUUID, err := uuid.Parse(input.AdminID)
	if err != nil {
		return fmt.Errorf("invalid admin ID format: %w", err)
	}

	// 2. Identify the owner for hierarchical locking (no lock yet)
	var userID string
	var txType models.TransactionType
	
	getInfoQuery := `SELECT user_id, type FROM transactions WHERE id = $1`
	if err := tx.QueryRowContext(ctx, getInfoQuery, txUUID).Scan(&userID, &txType); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("transaction not found")
		}
		return fmt.Errorf("failed to fetch transaction info: %w", err)
	}

	if txType != models.TransactionTypeDeposit {
		return fmt.Errorf("transaction is not a deposit (type: %s)", txType)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID in transaction record: %w", err)
	}

	// 3. LOCK HIERARCHY START: Always lock User (Owner) before Transaction (Child)
	var userBalance int64
	userLockQuery := `SELECT balance FROM users WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, userLockQuery, userUUID).Scan(&userBalance); err != nil {
		return fmt.Errorf("failed to lock user for refund: %w", err)
	}

	// 4. LOCK HIERARCHY CONTINUED: Lock the specific Transaction
	// Now we fetch all critical values AGAIN inside the lock (Double-Checked Locking)
	var amount int64
	var dbStatus models.TransactionStatus
	txLockQuery := `SELECT amount, status FROM transactions WHERE id = $1 FOR UPDATE`
	if err := tx.QueryRowContext(ctx, txLockQuery, txUUID).Scan(&amount, &dbStatus); err != nil {
		return fmt.Errorf("failed to lock transaction for refund: %w", err)
	}

	// 5. Critical Verifications (Inside Locks)
	if dbStatus != models.TransactionStatusCompleted {
		return fmt.Errorf("transaction cannot be refunded (status: %s)", dbStatus)
	}

	if userBalance < amount {
		return fmt.Errorf("insufficient balance for refund (Current: %d, Required: %d)", userBalance, amount)
	}

	// 6. ATOMIC UPDATES
	// Use absolute values since we have the lock (Consistent with AdminAdjustmentTx)
	newBalance := userBalance - amount
	updateUserQuery := `UPDATE users SET balance = $1, updated_at = NOW() WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateUserQuery, newBalance, userUUID); err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	// Update original transaction status
	updateTxQuery := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateTxQuery, models.TransactionStatusRefunded, txUUID); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// 7. AUDIT TRAIL
	refundTxID := uuid.New()
	
	// Ensure providerPaymentID is UNIQUE per operation by combining Admin ID and Transaction ID
	providerPaymentID := fmt.Sprintf("adm-ref-%s-%s", adminUUID.String()[:4], refundTxID.String()[:8])
	adjType := "Refund"

	refundTx := &models.Transaction{
		ID:                refundTxID,
		UserID:            userUUID,
		ProviderPaymentID: &providerPaymentID,
		Amount:            amount,
		Status:            models.TransactionStatusCompleted,
		Type:              models.TransactionTypeAdminAdjustment,
		AdjustmentType:    &adjType,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.txRepo.CreateTx(tx, refundTx); err != nil {
		return fmt.Errorf("failed to create refund audit record: %w", err)
	}

	return nil
}
