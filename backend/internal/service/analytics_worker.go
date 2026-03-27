package service

import (
	"database/sql"
	"log"
	"time"
)

type AnalyticsWorker struct {
	db *sql.DB
}

func NewAnalyticsWorker(db *sql.DB) *AnalyticsWorker {
	return &AnalyticsWorker{db: db}
}

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

func (w *AnalyticsWorker) RefreshNow() {
	w.calculateAndSaveStats()
}

func (w *AnalyticsWorker) calculateAndSaveStats() {
	log.Println("🔄 Analytics Worker: Calculating growth stats and totals...")

	// 1. Calculate Growth Percentages (Excluding test data)
	revenueGrowth, err := w.calculateRevenueGrowth()
	if err != nil {
		log.Printf("⚠️ Analytics Worker Error (Revenue Growth): %v", err)
	}

	usersGrowth, err := w.calculateGrowth("users", "1", "created_at", "")
	if err != nil {
		log.Printf("⚠️ Analytics Worker Error (Users Growth): %v", err)
	}

	salesGrowth, err := w.calculateGrowth("payments", "1", "created_at", "AND is_test = FALSE AND status = 'COMPLETED'")
	if err != nil {
		log.Printf("⚠️ Analytics Worker Error (Sales Growth): %v", err)
	}

	// 2. Calculate Real-Time Totals (Resync)
	var totalRevenue float64
	var totalUsers, activeProducts, totalSales, usersGrowthCount int

	// Sum completed transactions (real money in): type 'deposit' OR type 'admin_adjustment' AND adjustment_type = 'Pix Direto'
	err = w.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) 
		FROM transactions 
		WHERE status IN ('completed', 'approved') 
		AND (type = 'deposit' OR (type = 'admin_adjustment' AND adjustment_type = 'Pix Direto'))
	`).Scan(&totalRevenue)
	if err != nil {
		log.Printf("⚠️ Analytics Worker Error (Total Revenue from Transactions): %v", err)
	}

	err = w.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		log.Printf("⚠️ Analytics Worker Error (Total Users): %v", err)
	}

	// Calculate user growth in the last 30 days (+X users)
	err = w.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '30 days'").Scan(&usersGrowthCount)
	if err != nil {
		log.Printf("⚠️ Analytics Worker Error (Users Growth Count): %v", err)
	}

	err = w.db.QueryRow("SELECT COUNT(*) FROM products WHERE is_active = true").Scan(&activeProducts)
	if err != nil {
		// Fallback to counting all products if is_active column doesn't exist or other error
		w.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&activeProducts)
	}

	// Sales exclude test payments
	err = w.db.QueryRow("SELECT COUNT(*) FROM payments WHERE status = 'COMPLETED' AND is_test = FALSE").Scan(&totalSales)
	if err != nil {
		log.Printf("⚠️ Analytics Worker Error (Total Sales): %v", err)
	}

	// 3. Update Snapshot with BOTH growth and totals
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
	_, err = w.db.Exec(query, totalRevenue, totalUsers, activeProducts, totalSales, revenueGrowth, usersGrowth, salesGrowth, usersGrowthCount)
	if err != nil {
		log.Printf("❌ Analytics Worker Failed to Update Snapshot: %v", err)
		return
	}

	log.Println("✅ Analytics Worker: Stats updated successfully (Resynced totals and growth)")
}

// calculateGrowth compares current month sum/count vs last month
func (w *AnalyticsWorker) calculateGrowth(table, column, dateColumn, extraFilter string) (float64, error) {
	// Current Month
	queryCurrent := `
		SELECT COALESCE(SUM(` + column + `), 0)
		FROM ` + table + `
		WHERE ` + dateColumn + ` >= date_trunc('month', CURRENT_DATE)
		` + extraFilter + `
	`
	
	// Last Month
	queryLast := `
		SELECT COALESCE(SUM(` + column + `), 0)
		FROM ` + table + `
		WHERE ` + dateColumn + ` >= date_trunc('month', CURRENT_DATE - INTERVAL '1 month')
		  AND ` + dateColumn + ` < date_trunc('month', CURRENT_DATE)
		` + extraFilter + `
	`

	var current, last float64
	if err := w.db.QueryRow(queryCurrent).Scan(&current); err != nil {
		return 0, err
	}
	if err := w.db.QueryRow(queryLast).Scan(&last); err != nil {
		return 0, err
	}

	if last == 0 {
		if current > 0 {
			return 100.0, nil // 100% growth if started from 0
		}
		return 0.0, nil
	}

	return ((current - last) / last) * 100.0, nil
}

// calculateRevenueGrowth uses the same logic as total revenue (real cash in)
func (w *AnalyticsWorker) calculateRevenueGrowth() (float64, error) {
	filter := "AND status IN ('completed', 'approved') AND (type = 'deposit' OR (type = 'admin_adjustment' AND adjustment_type = 'Pix Direto'))"
	
	queryCurrent := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE created_at >= date_trunc('month', CURRENT_DATE)
		` + filter + `
	`
	
	queryLast := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE created_at >= date_trunc('month', CURRENT_DATE - INTERVAL '1 month')
		  AND created_at < date_trunc('month', CURRENT_DATE)
		` + filter + `
	`

	var current, last float64
	if err := w.db.QueryRow(queryCurrent).Scan(&current); err != nil {
		return 0, err
	}
	if err := w.db.QueryRow(queryLast).Scan(&last); err != nil {
		return 0, err
	}

	if last == 0 {
		if current > 0 {
			return 100.0, nil
		}
		return 0.0, nil
	}

	return ((current - last) / last) * 100.0, nil
}
