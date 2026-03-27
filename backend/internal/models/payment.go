package models

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

// Payment represents a financial transaction
// All monetary values are stored in cents (int64) to avoid float precision issues
type Payment struct {
	ID               uuid.UUID      `db:"id" json:"id"`
	UserID           uuid.UUID      `db:"user_id" json:"user_id"`
	SubscriptionID   *uuid.UUID     `db:"subscription_id" json:"subscription_id,omitempty"`
	Description      string         `db:"description" json:"description"`
	Amount           int64          `db:"amount" json:"amount"` // Amount in cents
	DiscountApplied  int64          `db:"discount_applied" json:"discount_applied"` // Discount in cents
	FinalAmount      int64          `db:"final_amount" json:"final_amount"` // Final amount in cents
	Status           PaymentStatus  `db:"status" json:"status"`
	IsTest           bool           `db:"is_test" json:"is_test"`
	PaymentGatewayID *string        `db:"payment_gateway_id" json:"payment_gateway_id,omitempty"`
	PaymentMethod    *string        `db:"payment_method" json:"payment_method,omitempty"`
	PaymentMetadata  *string        `db:"payment_metadata" json:"payment_metadata,omitempty"` // JSONB stored as string
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	CompletedAt      *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
	FailedAt         *time.Time     `db:"failed_at" json:"failed_at,omitempty"`
}

// CartItem represents a single item in the shopping cart
type CartItem struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

// CheckoutRequest represents the checkout payload
type CheckoutRequest struct {
	Cart       []CartItem `json:"cart" binding:"required,min=1,dive"`
	CouponCode   *string    `json:"coupon_code"`
	ReferralCode string     `json:"referral_code"`
	UseBalance   bool       `json:"use_balance"`
}

// CheckoutResponse represents the response after checkout
// All monetary values are in cents (int64) to avoid float precision issues
type CheckoutResponse struct {
	Success          bool      `json:"success"`
	PaymentID        uuid.UUID `json:"payment_id"`
	FinalAmount      int64     `json:"final_amount"` // Final amount in cents
	DiscountApplied  int64     `json:"discount_applied"` // Discount in cents
	Message          string    `json:"message"`
	// For gateway payments
	PaymentGatewayURL *string   `json:"payment_gateway_url,omitempty"`
	PaymentIntentID   *string   `json:"payment_intent_id,omitempty"`
}

// PaymentListResponse represents a paginated list of payments
type PaymentListResponse struct {
	Payments   []Payment `json:"payments"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}
