package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// AnalyticsWorker handles periodic calculation and caching of dashboard statistics
// Optimized with timezone-aware queries, transactional consistency, and proper null handling
type AnalyticsWorker struct {
	db *sql.DB
}

// NewAnalyticsWorker creates a new AnalyticsWorker
func NewAnalyticsWorker(db *sql.DB) *AnalyticsWorker {
	return &AnalyticsWorker{db: db}
}

// Start begins the background worker
// Runs immediately on start, then every hour
func (w *AnalyticsWorker) Start() {
	// Run immediately on start
	w.calculateAndSaveStats()

	// Then run every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		w.calculateAndSaveStats()
	}
}

// RefreshNow triggers an immediate recalculation
func (w *AnalyticsWorker) RefreshNow() {
	w.calculateAndSaveStats()
}

// AnalyticsData holds all calculated analytics data
// Growth fields use pointers to distinguish between 0% growth and "N/A" (nil)
type AnalyticsData struct {
	TotalRevenue     float64  `json:"totalRevenue"`
	TotalUsers       int      `json:"totalUsers"`
	ActiveProducts   int      `json:"activeProducts"`
	TotalSales       int      `json:"totalSales"`
	RevenueGrowth    *float64 `json:"revenueGrowthPct"` // nil = N/A (mathematically undefined)
	UsersGrowth      *float64 `json:"usersGrowthPct"`   // nil = N/A
	SalesGrowth      *float64 `json:"salesGrowthPct"`   // nil = N/A
	UsersGrowthCount int      `json:"usersGrowthCount"`
}

// calculateAndSaveStats calculates all statistics in a single transaction and saves snapshot
func (w *AnalyticsWorker) calculateAndSaveStats() {
	log.Println("🔄 Analytics Worker: Calculating stats...")

	// Use context with timeout to prevent hanging queries
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data, err := w.calculateAllStats(ctx)
	if err != nil {
		log.Printf("⚠️ Analytics Worker Error: %v", err)
		return
	}

	if err := w.saveSnapshot(ctx, data); err != nil {
		log.Printf("❌ Analytics Worker Failed to Save Snapshot: %v", err)
		return
	}

	log.Println("✅ Analytics Worker: Stats updated successfully")
}

// calculateAllStats fetches all statistics using optimized aggregated queries
// Uses REPEATABLE READ transaction to ensure snapshot consistency
// Uses AT TIME ZONE 'America/Sao_Paulo' for correct Brazil timezone handling
func (w *AnalyticsWorker) calculateAllStats(ctx context.Context) (*AnalyticsData, error) {
	data := &AnalyticsData{}

	// Start transaction with REPEATABLE READ isolation
	// This ensures all 4 queries see the same point-in-time snapshot of the database
	tx, err := w.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Query 1: Revenue metrics (current month, last month, total)
	// Uses FILTER for conditional aggregation in single query
	// Uses AT TIME ZONE to ensure Brazil timezone consistency
	revenueQuery := `
		SELECT
			COALESCE(SUM(amount) FILTER (
				WHERE created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Sao_Paulo' >= date_trunc('month', CURRENT_DATE)
				AND status IN ('completed', 'approved')
				AND (type = 'deposit' OR (type = 'admin_adjustment' AND adjustment_type = 'Pix Direto'))
			), 0) AS current_month,
			COALESCE(SUM(amount) FILTER (
				WHERE created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Sao_Paulo' >= date_trunc('month', CURRENT_DATE - INTERVAL '1 month')
				AND created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Sao_Paulo' < date_trunc('month', CURRENT_DATE)
				AND status IN ('completed', 'approved')
				AND (type = 'deposit' OR (type = 'admin_adjustment' AND adjustment_type = 'Pix Direto'))
			), 0) AS last_month,
			COALESCE(SUM(amount) FILTER (
				WHERE status IN ('completed', 'approved')
				AND (type = 'deposit' OR (type = 'admin_adjustment' AND adjustment_type = 'Pix Direto'))
			), 0) AS total
		FROM transactions
	`

	var currentRevenue, lastRevenue float64
	err = tx.QueryRowContext(ctx, revenueQuery).Scan(&currentRevenue, &lastRevenue, &data.TotalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate revenue: %w", err)
	}

	data.RevenueGrowth = calculateGrowth(currentRevenue, lastRevenue)

	// Query 2: User metrics (current month, last month, total, last 30 days)
	userQuery := `
		SELECT
			COUNT(*) FILTER (
				WHERE created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Sao_Paulo' >= date_trunc('month', CURRENT_DATE)
			) AS current_month,
			COUNT(*) FILTER (
				WHERE created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Sao_Paulo' >= date_trunc('month', CURRENT_DATE - INTERVAL '1 month')
				AND created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Sao_Paulo' < date_trunc('month', CURRENT_DATE)
			) AS last_month,
			COUNT(*) AS total,
			COUNT(*) FILTER (
				WHERE created_at >= NOW() - INTERVAL '30 days'
			) AS last_30_days
		FROM users
	`

	var currentUsers, lastUsers int
	err = tx.QueryRowContext(ctx, userQuery).Scan(&currentUsers, &lastUsers, &data.TotalUsers, &data.UsersGrowthCount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate users: %w", err)
	}

	data.UsersGrowth = calculateGrowth(float64(currentUsers), float64(lastUsers))

	// Query 3: Sales metrics (current month, last month, total)
	// Excludes test payments
	salesQuery := `
		SELECT
			COUNT(*) FILTER (
				WHERE created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Sao_Paulo' >= date_trunc('month', CURRENT_DATE)
				AND status = 'COMPLETED'
				AND is_test = FALSE
			) AS current_month,
			COUNT(*) FILTER (
				WHERE created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Sao_Paulo' >= date_trunc('month', CURRENT_DATE - INTERVAL '1 month')
				AND created_at AT TIME ZONE 'UTC' AT TIME ZONE 'America/Sao_Paulo' < date_trunc('month', CURRENT_DATE)
				AND status = 'COMPLETED'
				AND is_test = FALSE
			) AS last_month,
			COUNT(*) FILTER (
				WHERE status = 'COMPLETED'
				AND is_test = FALSE
			) AS total
		FROM payments
	`

	var currentSales, lastSales int
	err = tx.QueryRowContext(ctx, salesQuery).Scan(&currentSales, &lastSales, &data.TotalSales)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate sales: %w", err)
	}

	data.SalesGrowth = calculateGrowth(float64(currentSales), float64(lastSales))

	// Query 4: Active products count
	productQuery := `
		SELECT COUNT(*)
		FROM products
		WHERE is_active = true
	`

	err = tx.QueryRowContext(ctx, productQuery).Scan(&data.ActiveProducts)
	if err != nil {
		// Fallback: count all products if is_active column doesn't exist
		err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&data.ActiveProducts)
		if err != nil {
			return nil, fmt.Errorf("failed to count products: %w", err)
		}
	}

	// Note: No need to commit explicitly for read-only transactions
	// The deferred Rollback() is a no-op after successful execution

	return data, nil
}

// saveSnapshot saves the calculated statistics to the database
func (w *AnalyticsWorker) saveSnapshot(ctx context.Context, data *AnalyticsData) error {
	query := `
		UPDATE admin_analytics_snapshot
		SET
			total_revenue = $1,
			total_users = $2,
			active_products = $3,
			total_sales = $4,
			revenue_growth_pct = $5,
			users_growth_pct = $6,
			sales_growth_pct = $7,
			users_growth_count = $8,
			last_updated = NOW()
		WHERE id = 1
	`

	_, err := w.db.ExecContext(ctx, query,
		data.TotalRevenue,
		data.TotalUsers,
		data.ActiveProducts,
		data.TotalSales,
		data.RevenueGrowth,    // nil will be stored as NULL
		data.UsersGrowth,      // nil will be stored as NULL
		data.SalesGrowth,      // nil will be stored as NULL
		data.UsersGrowthCount,
	)

	return err
}

// calculateGrowth calculates percentage growth between two values
// Returns nil when growth is mathematically undefined (last == 0 and current > 0)
// This allows the frontend to display "N/A" or "New" instead of misleading percentages
func calculateGrowth(current, last float64) *float64 {
	if last == 0 {
		if current > 0 {
			// Growth from zero is mathematically undefined (infinite)
			// Return nil to signal "N/A" to the frontend
			return nil
		}
		// No change from zero (0 -> 0)
		zero := 0.0
		return &zero
	}

	growth := ((current - last) / last) * 100.0
	return &growth
}
