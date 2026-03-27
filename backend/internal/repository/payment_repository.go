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

// GetUserPaymentStats gets payment statistics for a user
func (r *PaymentRepository) GetUserPaymentStats(ctx context.Context, userIDStr string) (*models.PaymentStats, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Return default stats on error
		defaultStats := &models.PaymentStats{
			TotalSpent:          0,
			ProductsPurchased:   0,
			ActiveSubscriptions: 0,
		}
		return defaultStats, fmt.Errorf("invalid user ID: %w", err)
	}

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
func (r *PaymentRepository) GetRecentPayments(ctx context.Context, userIDStr string, limit int) ([]models.PaymentInfo, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return []models.PaymentInfo{}, fmt.Errorf("invalid user ID: %w", err)
	}

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
func (r *PaymentRepository) GetMonthlySpending(ctx context.Context, userIDStr string, months int) ([]models.MonthlySpend, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return []models.MonthlySpend{}, fmt.Errorf("invalid user ID: %w", err)
	}

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
func (r *PaymentRepository) GetNextBillingSummary(ctx context.Context, userIDStr string) (float64, []string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Sum of price_per_month for active subscriptions
	var total float64
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
func (r *PaymentRepository) GetUserSubscriptionsMinimal(ctx context.Context, userIDStr string) ([]models.SubscriptionMini, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

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
func (r *PaymentRepository) GetUserSubscriptionInvoices(ctx context.Context, userIDStr string) ([]models.SubscriptionInvoice, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

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
				s.Status = "overdue"
			} else {
				s.Status = "due"
			}
		} else {
			s.Status = "paid" // Assuming other statuses are paid
		}

		invoices = append(invoices, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return invoices, nil
}