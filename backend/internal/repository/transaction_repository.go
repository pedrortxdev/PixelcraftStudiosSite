package repository

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pixelcraft/api/internal/models"
)

// TransactionRepository handles database operations for transactions
type TransactionRepository struct {
	db *sqlx.DB
}

// TransactionWithUser extends Transaction with User details
type TransactionWithUser struct {
	models.Transaction
	UserEmail string `db:"user_email" json:"user_email"`
	UserName  string `db:"user_name" json:"user_name"`
}

// NewTransactionRepository creates a new TransactionRepository
func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create creates a new transaction
func (r *TransactionRepository) Create(tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, provider_payment_id, amount, status, type, adjustment_type, created_at, updated_at)
		VALUES (:id, :user_id, :provider_payment_id, :amount, :status, :type, :adjustment_type, :created_at, :updated_at)
	`
	_, err := r.db.NamedExec(query, tx)
	return err
}

// CreateTx creates a new transaction within an existing transaction
func (r *TransactionRepository) CreateTx(tx *sqlx.Tx, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, provider_payment_id, amount, status, type, adjustment_type, created_at, updated_at)
		VALUES (:id, :user_id, :provider_payment_id, :amount, :status, :type, :adjustment_type, :created_at, :updated_at)
	`
	_, err := tx.NamedExec(query, transaction)
	return err
}

// GetByProviderPaymentID retrieves a transaction by its provider payment ID
func (r *TransactionRepository) GetByProviderPaymentID(paymentID string) (*models.Transaction, error) {
	var tx models.Transaction
	query := `SELECT * FROM transactions WHERE provider_payment_id = $1`
	err := r.db.Get(&tx, query, paymentID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tx, err
}

// GetByID retrieves a transaction by its ID
func (r *TransactionRepository) GetByID(id string) (*models.Transaction, error) {
	var tx models.Transaction
	query := `SELECT * FROM transactions WHERE id = $1`
	err := r.db.Get(&tx, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tx, err
}

// ListTransactions retrieves transactions with pagination and optional filtering
func (r *TransactionRepository) ListTransactions(page, limit int, status string) ([]TransactionWithUser, int, error) {
	offset := (page - 1) * limit
	
	baseQuery := `
		SELECT t.*, u.email as user_email, COALESCE(u.username, '') as user_name
		FROM transactions t
		JOIN users u ON t.user_id = u.id
	`
	countQuery := `SELECT COUNT(*) FROM transactions`
	
	var args []interface{}
	
	if status != "" {
		baseQuery += ` WHERE t.status = $1`
		countQuery += ` WHERE status = $1`
		args = append(args, status)
	}
	
	baseQuery += ` ORDER BY t.created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	
	// Get total count
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	args = append(args, limit, offset)
	
	var txs []TransactionWithUser
	err = r.db.Select(&txs, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	return txs, total, nil
}

// ListByUserID retrieves transactions for a specific user
func (r *TransactionRepository) ListByUserID(userID string, limit int) ([]models.Transaction, error) {
	var txs []models.Transaction
	query := `SELECT * FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`
	err := r.db.Select(&txs, query, userID, limit)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

// UpdateStatus updates the status of a transaction
func (r *TransactionRepository) UpdateStatus(id string, status models.TransactionStatus) error {
	query := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

// CompleteDeposit updates the transaction status to completed and increments the user's balance transactionally
// Amount is in cents (int64) to avoid float precision issues
func (r *TransactionRepository) CompleteDeposit(transactionID string, amount int64) error {
	// Start a transaction
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 0. Lock transaction and check status (Pessimistic Locking to avoid race conditions)
	var currentStatus string
	err = tx.Get(&currentStatus, "SELECT status FROM transactions WHERE id = $1 FOR UPDATE", transactionID)
	if err != nil {
		return fmt.Errorf("failed to lock transaction: %w", err)
	}

	if currentStatus == string(models.TransactionStatusCompleted) {
		// Already completed, nothing to do
		return nil
	}

	// 1. Update Transaction Status
	queryTx := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(queryTx, models.TransactionStatusCompleted, transactionID)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// 2. Update User Balance
	queryUser := `
		UPDATE users
		SET balance = balance + $1, updated_at = NOW()
		WHERE id = (SELECT user_id FROM transactions WHERE id = $2)
	`

	result, err := tx.Exec(queryUser, amount, transactionID)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or transaction invalid")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RefundDeposit updates the transaction status to refunded and decrements the user's balance transactionally
// Amount is in cents (int64) to avoid float precision issues
func (r *TransactionRepository) RefundDeposit(transactionID string, amount int64) error {
	// Start a transaction
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 0. Lock transaction and check status
	var currentStatus string
	err = tx.Get(&currentStatus, "SELECT status FROM transactions WHERE id = $1 FOR UPDATE", transactionID)
	if err != nil {
		return fmt.Errorf("failed to lock transaction for refund: %w", err)
	}

	if currentStatus == string(models.TransactionStatusRefunded) {
		// Already refunded
		return nil
	}

	// 1. Update Transaction Status
	queryTx := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(queryTx, models.TransactionStatusRefunded, transactionID)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// 2. Decrement User Balance
	// Allow negative balance if chargeback occurs
	queryUser := `
		UPDATE users
		SET balance = balance - $1, updated_at = NOW()
		WHERE id = (SELECT user_id FROM transactions WHERE id = $2)
	`

	result, err := tx.Exec(queryUser, amount, transactionID)
	if err != nil {
		return fmt.Errorf("failed to decrement user balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or transaction invalid")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}