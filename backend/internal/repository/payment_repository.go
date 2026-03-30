package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pixelcraft/api/internal/models"

	"github.com/google/uuid"
)

// PaymentRepository handles all database operations for payments
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository creates a new PaymentRepository
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment record within a transaction
func (r *PaymentRepository) Create(ctx context.Context, tx *sql.Tx, payment *models.Payment) (uuid.UUID, error) {
	query := `
		INSERT INTO payments (id, user_id, description, amount, discount_applied, final_amount, status, is_test, payment_method, payment_metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var execTx interface {
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	var paymentID uuid.UUID
	err := execTx.QueryRowContext(ctx, query,
		payment.ID,
		payment.UserID,
		payment.Description,
		payment.Amount,
		payment.DiscountApplied,
		payment.FinalAmount,
		payment.Status,
		payment.IsTest,
		payment.PaymentMethod,
		payment.PaymentMetadata,
		payment.CreatedAt,
	).Scan(&paymentID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return paymentID, nil
}

// GetByID retrieves a payment by its ID
func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*models.Payment, error) {
	return r.GetByIDTx(ctx, nil, id)
}

// GetByIDTx retrieves a payment by its ID within a transaction with row-level lock
func (r *PaymentRepository) GetByIDTx(ctx context.Context, tx *sql.Tx, id string) (*models.Payment, error) {
	paymentID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}

	query := `
		SELECT 
			id, user_id, subscription_id, description, amount, discount_applied, 
			final_amount, status, is_test, payment_gateway_id, payment_method, 
			payment_metadata, created_at, completed_at, failed_at
		FROM payments
		WHERE id = $1
	`
	if tx != nil {
		query += " FOR UPDATE"
	}

	var p models.Payment
	var metadata sql.NullString
	
	var execTx interface {
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	err = execTx.QueryRowContext(ctx, query, paymentID).Scan(
		&p.ID, &p.UserID, &p.SubscriptionID, &p.Description, &p.Amount, &p.DiscountApplied,
		&p.FinalAmount, &p.Status, &p.IsTest, &p.PaymentGatewayID, &p.PaymentMethod,
		&metadata, &p.CreatedAt, &p.CompletedAt, &p.FailedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if metadata.Valid {
		p.PaymentMetadata = &metadata.String
	}

	return &p, nil
}

// UpdateStatus updates the status of a payment
func (r *PaymentRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, id string, status models.PaymentStatus, gatewayID *string) error {
	paymentID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid payment ID: %w", err)
	}

	query := `
		UPDATE payments 
		SET status = $1, 
		    payment_gateway_id = COALESCE($2, payment_gateway_id),
		    completed_at = CASE WHEN $1 = 'COMPLETED' THEN NOW() ELSE completed_at END,
		    failed_at = CASE WHEN $1 = 'FAILED' THEN NOW() ELSE failed_at END
		WHERE id = $3
	`

	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	_, err = execTx.ExecContext(ctx, query, status, gatewayID, paymentID)
	return err
}

// GetUserPaymentStats gets payment statistics for a user
func (r *PaymentRepository) GetUserPaymentStats(ctx context.Context, userID uuid.UUID) (*models.PaymentStats, error) {
	// Total spent from completed payments
	querySpent := `
		SELECT COALESCE(SUM(CASE WHEN status = 'COMPLETED' THEN final_amount ELSE 0 END), 0) as total_spent
		FROM payments
		WHERE user_id = $1
	`

	var stats models.PaymentStats
	if err := r.db.QueryRowContext(ctx, querySpent, userID).Scan(&stats.TotalSpent); err != nil {
		// If there's an error getting total spent, still try to get the other stats
		stats.TotalSpent = 0
	}

	// Products purchased from user_purchases table
	queryPurchases := `
		SELECT COALESCE(COUNT(*), 0) FROM user_purchases WHERE user_id = $1
	`
	if err := r.db.QueryRowContext(ctx, queryPurchases, userID).Scan(&stats.ProductsPurchased); err != nil {
		stats.ProductsPurchased = 0
	}

	// Get active subscriptions count
	subQuery := `
		SELECT COALESCE(COUNT(*), 0)
		FROM subscriptions
		WHERE user_id = $1 AND status = 'ACTIVE'
	`
	if err := r.db.QueryRowContext(ctx, subQuery, userID).Scan(&stats.ActiveSubscriptions); err != nil {
		stats.ActiveSubscriptions = 0
	}

	return &stats, nil
}

// GetRecentPayments gets recent payments for a user
func (r *PaymentRepository) GetRecentPayments(ctx context.Context, userID uuid.UUID, limit int) ([]models.PaymentInfo, error) {
	query := `
		SELECT
			COALESCE(id, '00000000-0000-0000-0000-000000000000') as id,
			COALESCE(description, 'N/A') as description,
			COALESCE(final_amount, 0) as final_amount,
			COALESCE(status, 'UNKNOWN') as status,
			COALESCE(created_at, NOW()) as created_at
		FROM payments
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return []models.PaymentInfo{}, fmt.Errorf("failed to query recent payments: %w", err)
	}
	defer rows.Close()

	var payments []models.PaymentInfo
	for rows.Next() {
		var p models.PaymentInfo
		var createdAt time.Time

		if err := rows.Scan(&p.ID, &p.Description, &p.Amount, &p.Status, &createdAt); err != nil {
			return []models.PaymentInfo{}, fmt.Errorf("failed to scan payment: %w", err)
		}

		p.CreatedAt = createdAt.Format("2006-01-02T15:04:05Z07:00")
		payments = append(payments, p)
	}

	if err = rows.Err(); err != nil {
		return payments, fmt.Errorf("rows error: %w", err)
	}

	return payments, nil
}

// GetMonthlySpending gets monthly spending for a user
func (r *PaymentRepository) GetMonthlySpending(ctx context.Context, userID uuid.UUID, months int) ([]models.MonthlySpend, error) {
	// Use parameterized query to avoid SQL injection
	query := `
		SELECT
			COALESCE(TO_CHAR(created_at, 'YYYY-MM'), 'N/A') as month,
			COALESCE(SUM(CASE WHEN status = 'COMPLETED' THEN final_amount ELSE 0 END), 0) as amount
		FROM payments
		WHERE user_id = $1
			AND created_at >= NOW() - ($2::text || ' months')::interval
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY month DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, months)
	if err != nil {
		return []models.MonthlySpend{}, fmt.Errorf("failed to query monthly spending: %w", err)
	}
	defer rows.Close()

	var spending []models.MonthlySpend
	for rows.Next() {
		var s models.MonthlySpend
		if err := rows.Scan(&s.Month, &s.Amount); err != nil {
			return []models.MonthlySpend{}, fmt.Errorf("failed to scan monthly spending: %w", err)
		}
		spending = append(spending, s)
	}

	if err = rows.Err(); err != nil {
		return spending, fmt.Errorf("rows error: %w", err)
	}

	return spending, nil
}

// GetNextBillingSummary returns total next billing amount and list of dates for active subscriptions
func (r *PaymentRepository) GetNextBillingSummary(ctx context.Context, userID uuid.UUID) (int64, []string, error) {
	// Sum of price_per_month for active subscriptions
	var total int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(price_per_month), 0)
		FROM subscriptions
		WHERE user_id = $1 AND status = 'ACTIVE'
	`, userID).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("failed to sum next billing: %w", err)
	}

	// Collect next billing dates for active subscriptions
	rows, err := r.db.QueryContext(ctx, `
		SELECT next_billing_date
		FROM subscriptions
		WHERE user_id = $1 AND status = 'ACTIVE'
		ORDER BY next_billing_date ASC
	`, userID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to query next billing dates: %w", err)
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return 0, nil, fmt.Errorf("failed to scan next billing date: %w", err)
		}
		dates = append(dates, d.Format("2006-01-02"))
	}
	if err = rows.Err(); err != nil {
		return 0, nil, fmt.Errorf("rows error: %w", err)
	}

	return total, dates, nil
}

// GetUserSubscriptionsMinimal returns list of subscriptions with minimal fields for a user
func (r *PaymentRepository) GetUserSubscriptionsMinimal(ctx context.Context, userID uuid.UUID) ([]models.SubscriptionMini, error) {
	query := `
		SELECT id, plan_name, price_per_month, created_at
		FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []models.SubscriptionMini
	for rows.Next() {
		var s models.SubscriptionMini
		if err := rows.Scan(&s.ID, &s.PlanName, &s.PricePerMonth, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subs = append(subs, s)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return subs, nil
}

// GetUserSubscriptionInvoices retrieves all subscription invoices for a user
func (r *PaymentRepository) GetUserSubscriptionInvoices(ctx context.Context, userID uuid.UUID) ([]models.SubscriptionInvoice, error) {
	query := `
		SELECT id, plan_name, price_per_month, next_billing_date, status
		FROM subscriptions
		WHERE user_id = $1 AND status NOT IN ('CANCELED')
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscription invoices: %w", err)
	}
	defer rows.Close()

	var invoices []models.SubscriptionInvoice
	for rows.Next() {
		var s models.SubscriptionInvoice
		var subID uuid.UUID
		var nextBillingDate time.Time
		var status string

		if err := rows.Scan(&subID, &s.PlanName, &s.Amount, &nextBillingDate, &status); err != nil {
			return nil, fmt.Errorf("failed to scan subscription invoice: %w", err)
		}

		s.SubscriptionID = subID
		s.DueDate = nextBillingDate

		// Determine invoice status
		now := time.Now()
		if status == "ACTIVE" {
			if now.After(nextBillingDate) {
				s.Status = models.InvoiceStatusOverdue
			} else {
				s.Status = models.InvoiceStatusDue
			}
		} else {
			s.Status = models.InvoiceStatusPaid // Assuming other statuses are paid
		}

		invoices = append(invoices, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return invoices, nil
}