package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "time/tzdata" // Embed timezone data for Docker/Alpine compatibility
)

// AnalyticsWorker handles periodic calculation and caching of dashboard statistics
// Uses SARGable queries with Go-calculated time boundaries for optimal index usage
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

// TimeBoundaries holds pre-calculated time boundaries in UTC for database queries
// Uses [start, end) interval pattern for precise boundary handling
type TimeBoundaries struct {
	CurrentMonthStart time.Time // Start of current month in UTC
	LastMonthStart    time.Time // Start of last month in UTC
	ThirtyDaysAgo     time.Time // 30 days ago from now in UTC
}

// calculateTimeBoundaries computes month boundaries in America/Sao_Paulo timezone,
// then converts to UTC for SARGable database queries
// Uses embedded tzdata for Docker/Alpine compatibility
func calculateTimeBoundaries(now time.Time) TimeBoundaries {
	// Load Brazil timezone (embedded via _ "time/tzdata")
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		// This should never happen with tzdata embedded, but fallback gracefully
		log.Printf("⚠️ Failed to load America/Sao_Paulo timezone, using UTC: %v", err)
		loc = time.UTC
	}

	// Convert now to Brazil time
	brazilNow := now.In(loc)

	// Calculate current month start in Brazil time (e.g., 2026-03-01 00:00:00 -0300)
	currentMonthStart := time.Date(brazilNow.Year(), brazilNow.Month(), 1, 0, 0, 0, 0, loc)

	// Calculate last month start in Brazil time
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)

	// Calculate 30 days ago in Brazil time
	thirtyDaysAgo := brazilNow.AddDate(0, 0, -30)

	// Convert all boundaries to UTC for database queries
	// This ensures indexes on created_at (stored as UTC) can be used efficiently
	return TimeBoundaries{
		CurrentMonthStart: currentMonthStart.UTC(),
		LastMonthStart:    lastMonthStart.UTC(),
		ThirtyDaysAgo:     thirtyDaysAgo.UTC(),
	}
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

// calculateAllStats fetches all statistics using SARGable queries
// Time boundaries are calculated in Go and passed as parameters for optimal index usage
// Uses REPEATABLE READ transaction to ensure snapshot consistency
// Uses [start, end) interval pattern for precise date range handling
func (w *AnalyticsWorker) calculateAllStats(ctx context.Context) (*AnalyticsData, error) {
	data := &AnalyticsData{}

	// Calculate time boundaries in Go (SARGable approach)
	boundaries := calculateTimeBoundaries(time.Now())

	// Start transaction with REPEATABLE READ isolation
	// This ensures all 4 queries see the same point-in-time snapshot of the database
	tx, err := w.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Safety net for panics or early returns

	// Query 1: Revenue metrics (current month, last month, total)
	// Uses FILTER for conditional aggregation in single query
	// SARGable: created_at >= $1 (no functions on column side)
	// Uses [start, end) pattern: >= LastMonthStart AND < CurrentMonthStart
	revenueQuery := `
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE created_at >= $1 AND status IN ('completed', 'approved') AND (type = 'deposit' OR (type = 'admin_adjustment' AND adjustment_type = 'Pix Direto'))), 0) AS current_month,
			COALESCE(SUM(amount) FILTER (WHERE created_at >= $2 AND created_at < $1 AND status IN ('completed', 'approved') AND (type = 'deposit' OR (type = 'admin_adjustment' AND adjustment_type = 'Pix Direto'))), 0) AS last_month,
			COALESCE(SUM(amount) FILTER (WHERE status IN ('completed', 'approved') AND (type = 'deposit' OR (type = 'admin_adjustment' AND adjustment_type = 'Pix Direto'))), 0) AS total
		FROM transactions
	`

	var currentRevenue, lastRevenue float64
	err = tx.QueryRowContext(ctx, revenueQuery,
		boundaries.CurrentMonthStart,
		boundaries.LastMonthStart,
	).Scan(&currentRevenue, &lastRevenue, &data.TotalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate revenue: %w", err)
	}

	data.RevenueGrowth = calculateGrowth(currentRevenue, lastRevenue)

	// Query 2: User metrics (current month, last month, total, last 30 days)
	userQuery := `
		SELECT
			COUNT(*) FILTER (WHERE created_at >= $1) AS current_month,
			COUNT(*) FILTER (WHERE created_at >= $2 AND created_at < $1) AS last_month,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE created_at >= $3) AS last_30_days
		FROM users
	`

	var currentUsers, lastUsers int
	err = tx.QueryRowContext(ctx, userQuery,
		boundaries.CurrentMonthStart,
		boundaries.LastMonthStart,
		boundaries.ThirtyDaysAgo,
	).Scan(&currentUsers, &lastUsers, &data.TotalUsers, &data.UsersGrowthCount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate users: %w", err)
	}

	data.UsersGrowth = calculateGrowth(float64(currentUsers), float64(lastUsers))

	// Query 3: Sales metrics (current month, last month, total)
	// Excludes test payments
	salesQuery := `
		SELECT
			COUNT(*) FILTER (WHERE created_at >= $1 AND status = 'COMPLETED' AND is_test = FALSE) AS current_month,
			COUNT(*) FILTER (WHERE created_at >= $2 AND created_at < $1 AND status = 'COMPLETED' AND is_test = FALSE) AS last_month,
			COUNT(*) FILTER (WHERE status = 'COMPLETED' AND is_test = FALSE) AS total
		FROM payments
	`

	var currentSales, lastSales int
	err = tx.QueryRowContext(ctx, salesQuery,
		boundaries.CurrentMonthStart,
		boundaries.LastMonthStart,
	).Scan(&currentSales, &lastSales, &data.TotalSales)
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

	// Explicitly commit read-only transaction
	// This signals to the database that the transaction completed successfully
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return data, nil
}

// saveSnapshot saves the calculated statistics to the database using UPSERT
// Creates the initial row if it doesn't exist, updates if it does
func (w *AnalyticsWorker) saveSnapshot(ctx context.Context, data *AnalyticsData) error {
	// UPSERT pattern: INSERT ... ON CONFLICT DO UPDATE
	// Ensures the snapshot row exists even on fresh databases
	query := `
		INSERT INTO admin_analytics_snapshot (
			id,
			total_revenue,
			total_users,
			active_products,
			total_sales,
			revenue_growth_pct,
			users_growth_pct,
			sales_growth_pct,
			users_growth_count,
			last_updated
		) VALUES (
			1, $1, $2, $3, $4, $5, $6, $7, $8, NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			total_revenue = $1,
			total_users = $2,
			active_products = $3,
			total_sales = $4,
			revenue_growth_pct = $5,
			users_growth_pct = $6,
			sales_growth_pct = $7,
			users_growth_count = $8,
			last_updated = NOW()
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
