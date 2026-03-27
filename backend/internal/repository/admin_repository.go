package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// LogAction logs an administrative action for auditing (BT-012)
func (r *AdminRepository) LogAction(ctx context.Context, adminID string, action string, details string) error {
	query := `
		INSERT INTO admin_audit_logs (admin_id, action, details)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, adminID, action, details)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}

// AnalyticsSnapshot represents the data from admin_analytics_snapshot table
type AnalyticsSnapshot struct {
	TotalRevenue     float64   `json:"totalRevenue"`
	TotalUsers       int       `json:"totalUsers"`
	ActiveProducts   int       `json:"activeProducts"`
	TotalSales       int       `json:"totalSales"`
	RevenueGrowth    float64   `json:"revenueGrowthPct"`
	UsersGrowth      float64   `json:"usersGrowthPct"`
	UsersGrowthCount int       `json:"usersGrowthCount"` // New field for "+X users"
	ProductsStatus   string    `json:"productsStatus"`
	SalesGrowth      float64   `json:"salesGrowthPct"`
	LastUpdated      time.Time `json:"lastUpdated"`
}

// RecentOrder represents a recent order for the dashboard
type RecentOrder struct {
	ID          string   `json:"id"`
	UserName    string   `json:"userName"`
	ProductName string   `json:"productName"`
	Value       float64  `json:"value"`
	Status      string   `json:"status"`
	Items       []string `json:"items"` // List of all product names
}

// TopProduct represents a top selling product
type TopProduct struct {
	ProductName  string  `json:"productName"`
	SalesCount   int     `json:"salesCount"`
	TotalRevenue float64 `json:"totalRevenue"`
}

// GetAnalyticsSnapshot returns the latest analytics snapshot
func (r *AdminRepository) GetAnalyticsSnapshot() (*AnalyticsSnapshot, error) {
	query := `
        SELECT
            COALESCE(total_revenue, 0)::FLOAT8 as total_revenue,
            COALESCE(total_users, 0)::INT as total_users,
            COALESCE(active_products, 0)::INT as active_products,
            COALESCE(total_sales, 0)::INT as total_sales,
            COALESCE(revenue_growth_pct, 0)::FLOAT8 as revenue_growth_pct,
            COALESCE(users_growth_pct, 0)::FLOAT8 as users_growth_pct,
            COALESCE(users_growth_count, 0)::INT as users_growth_count,
            COALESCE(products_status, '{}')::TEXT as products_status,
            COALESCE(sales_growth_pct, 0)::FLOAT8 as sales_growth_pct,
            COALESCE(last_updated, NOW()) as last_updated
        FROM admin_analytics_snapshot
        ORDER BY last_updated DESC
        LIMIT 1
    `

	var s AnalyticsSnapshot
	err := r.db.QueryRow(query).Scan(
		&s.TotalRevenue, &s.TotalUsers, &s.ActiveProducts, &s.TotalSales,
		&s.RevenueGrowth, &s.UsersGrowth, &s.UsersGrowthCount, &s.ProductsStatus, &s.SalesGrowth, &s.LastUpdated,
	)
	if err != nil {
		log.Printf("Admin Stats: Snapshots access issue (rows missing or schema error): %v. Returning default values.", err)
		return &AnalyticsSnapshot{
			TotalRevenue:   0,
			TotalUsers:     0,
			ActiveProducts: 0,
			TotalSales:     0,
			RevenueGrowth:  0,
			UsersGrowth:    0,
			ProductsStatus: "Estável",
			SalesGrowth:    0,
			LastUpdated:    time.Now(),
		}, nil
	}

	return &s, nil
}

// GetRecentOrders returns the 5 most recent orders
func (r *AdminRepository) GetRecentOrders() ([]RecentOrder, error) {
	// Joining payments -> library -> products to get all items
	query := `
        SELECT
            pay.id::TEXT,
            COALESCE(u.full_name, 'Unknown User') as full_name,
            COALESCE(pay.final_amount, 0) as final_amount,
            COALESCE(pay.status::TEXT, 'pending') as status,
            COALESCE(array_agg(p.name) FILTER (WHERE p.name IS NOT NULL), ARRAY[]::TEXT[]) as items
        FROM payments pay
        JOIN users u ON pay.user_id = u.id
        LEFT JOIN library l ON pay.id = l.payment_id
        LEFT JOIN products p ON l.product_id = p.id
        GROUP BY pay.id, u.full_name, pay.created_at
        ORDER BY pay.created_at DESC
        LIMIT 5
    `

	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("Erro Admin Stats: Falha ao buscar ordens recentes - %v", err)
		return []RecentOrder{}, nil
	}
	defer rows.Close()

	var orders []RecentOrder
	for rows.Next() {
		var o RecentOrder
		var items pq.StringArray
		if err := rows.Scan(&o.ID, &o.UserName, &o.Value, &o.Status, &items); err != nil {
			log.Printf("Erro Admin Stats: Falha ao escanear ordem recente - %v", err)
			continue
		}
		o.Items = items
		
		// Set primary product name logic
		if len(o.Items) > 1 {
			o.ProductName = "Multiple Items"
		} else if len(o.Items) == 1 {
			o.ProductName = o.Items[0]
		} else {
			o.ProductName = "No Items"
		}
		
		orders = append(orders, o)
	}

	return orders, nil
}

// GetTopProducts returns the top 3 best-selling products
func (r *AdminRepository) GetTopProducts() ([]TopProduct, error) {
	query := `
        SELECT
            COALESCE(p.name, 'Unknown Product')::TEXT as name,
            COALESCE(COUNT(l.id), 0)::INT as sales_count,
            COALESCE(SUM(pay.final_amount), 0)::FLOAT8 as total_revenue
        FROM library l
        JOIN products p ON l.product_id = p.id
        JOIN payments pay ON l.payment_id = pay.id
        WHERE pay.status::TEXT IN ('COMPLETED', 'completed')
        GROUP BY p.name
        ORDER BY sales_count DESC
        LIMIT 3
    `

	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("Erro Admin Stats: Falha ao buscar produtos top - %v", err)
		return []TopProduct{}, nil
	}
	defer rows.Close()

	var products []TopProduct
	for rows.Next() {
		var p TopProduct
		if err := rows.Scan(&p.ProductName, &p.SalesCount, &p.TotalRevenue); err != nil {
			log.Printf("Erro Admin Stats: Falha ao escanear produto top - %v", err)
			continue
		}
		products = append(products, p)
	}

	return products, nil
}
