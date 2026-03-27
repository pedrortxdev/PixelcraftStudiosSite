package models

// PaymentStats represents payment statistics for a user
type PaymentStats struct {
	TotalSpent          float64 `json:"total_spent"`
	ProductsPurchased   int     `json:"products_purchased"`
	ActiveSubscriptions int     `json:"active_subscriptions"`
}

// PaymentInfo represents payment information for dashboard
type PaymentInfo struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

// MonthlySpend represents monthly spending data
type MonthlySpend struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

// NextBillingSummary represents aggregate of upcoming billing for subscriptions
type NextBillingSummary struct {
	Total  float64  `json:"total_next_billing"`
	Dates  []string `json:"next_billing_dates"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	Balance             float64            `json:"balance"`
	TotalSpent          float64            `json:"total_spent"`
	ProductsPurchased   int                `json:"products_purchased"`
	ActiveSubscriptions int                `json:"active_subscriptions"`
	RecentPayments      []PaymentInfo      `json:"recent_payments"`
	MonthlySpending     []MonthlySpend     `json:"monthly_spending"`
	NextBilling         NextBillingSummary `json:"next_billing"`
}