package models

import (
	"time"

	"github.com/google/uuid"
)

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypePercentage  DiscountType = "PERCENTAGE"
	DiscountTypeFixedAmount DiscountType = "FIXED_AMOUNT"
)

// RestrictionType represents what the discount applies to
type RestrictionType string

const (
	RestrictionAll          RestrictionType = "ALL"
	RestrictionItemCategory RestrictionType = "ITEM_CATEGORY" // Specific product categories
	RestrictionGame         RestrictionType = "GAME"          // Specific game
	RestrictionProduct      RestrictionType = "PRODUCT"       // Specific products
)

// Discount represents a coupon or referral code
type Discount struct {
	ID                uuid.UUID       `db:"id" json:"id"`
	Code              string          `db:"code" json:"code"`
	Type              DiscountType    `db:"type" json:"type"`
	Value             float64         `db:"value" json:"value"`
	IsReferral        bool            `db:"is_referral" json:"is_referral"`
	RestrictionType   RestrictionType `db:"restriction_type" json:"restriction_type"`
	TargetIDs         []uuid.UUID     `db:"target_ids" json:"target_ids,omitempty"`
	CreatedByUserID   *uuid.UUID      `db:"created_by_user_id" json:"created_by_user_id,omitempty"`
	ExpiresAt         *time.Time      `db:"expires_at" json:"expires_at,omitempty"`
	MaxUses           *int            `db:"max_uses" json:"max_uses,omitempty"`
	CurrentUses       int             `db:"current_uses" json:"current_uses"`
	IsActive          bool            `db:"is_active" json:"is_active"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
}

// CreateDiscountRequest represents the request to create a discount
type CreateDiscountRequest struct {
	Code            string          `json:"code" binding:"required,min=3,max=50"`
	Type            DiscountType    `json:"type" binding:"required"`
	Value           float64         `json:"value" binding:"required,min=0"`
	IsReferral      bool            `json:"is_referral"`
	RestrictionType RestrictionType `json:"restriction_type" binding:"required"`
	TargetIDs       []uuid.UUID     `json:"target_ids"`
	ExpiresAt       *time.Time      `json:"expires_at"`
	MaxUses         *int            `json:"max_uses" binding:"omitempty,min=1"`
}

// UpdateDiscountRequest represents the request to update a discount
type UpdateDiscountRequest struct {
	Code            *string          `json:"code"`
	Type            *DiscountType    `json:"type"`
	Value           *float64         `json:"value"`
	RestrictionType *RestrictionType `json:"restriction_type"`
	TargetIDs       []uuid.UUID      `json:"target_ids"`
	IsActive        *bool            `json:"is_active"`
	ExpiresAt       *time.Time       `json:"expires_at"`
	MaxUses         *int             `json:"max_uses"`
}

// DiscountUpdate represents domain-level discount updates (avoids leaking DB column names)
type DiscountUpdate struct {
	Code            *string
	Type            *DiscountType
	Value           *float64
	RestrictionType *RestrictionType
	TargetIDs       []uuid.UUID
	IsActive        *bool
	ExpiresAt       *time.Time
	MaxUses         *int
}

// ValidateDiscountRequest represents the request to validate a discount code
type ValidateDiscountRequest struct {
	Code      string     `json:"code" binding:"required"`
	Amount    float64    `json:"amount" binding:"required,min=0"`
	CartItems []CartItem `json:"cart_items"`
}

// ValidateDiscountResponse represents the response with discount calculation
type ValidateDiscountResponse struct {
	IsValid        bool    `json:"is_valid"`
	DiscountAmount float64 `json:"discount_amount"`
	FinalAmount    float64 `json:"final_amount"`
	Message        string  `json:"message,omitempty"`
}
