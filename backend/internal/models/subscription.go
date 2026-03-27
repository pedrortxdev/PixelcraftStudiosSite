package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "ACTIVE"
	SubscriptionStatusCanceled  SubscriptionStatus = "CANCELED"
	SubscriptionStatusPastDue   SubscriptionStatus = "PAST_DUE"
	SubscriptionStatusCompleted SubscriptionStatus = "COMPLETED"
)

// Plan represents a subscription plan available for purchase
// Price is stored in cents (int64) to avoid float precision issues
type Plan struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"` // Price in cents (e.g., 1000 = R$ 10.00)
	ImageURL    *string   `json:"imageUrl,omitempty"`
	IsActive    bool      `json:"isActive"`
	Features    []string  `json:"features,omitempty"` // Array of strings
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Subscription represents a user's subscription/project
// All monetary values are stored in cents (int64) to avoid float precision issues
type Subscription struct {
	ID              uuid.UUID          `json:"id"`
	UserID          uuid.UUID          `json:"userId"`
	PlanID          *uuid.UUID         `json:"planId"`
	PlanName        string             `json:"planName"`      // Kept for backward compatibility
	PricePerMonth   int64              `json:"pricePerMonth"` // Current plan price in cents
	AgreedPrice     *int64             `json:"agreedPrice"`   // Price at purchase time in cents
	Status          SubscriptionStatus `json:"status"`
	ProjectStage    string             `json:"projectStage"` // e.g., 'Planejamento', 'Desenvolvimento'
	StartedAt       time.Time          `json:"startedAt"`
	NextBillingDate time.Time          `json:"nextBillingDate"`
	CanceledAt      *time.Time         `json:"canceledAt,omitempty"`
	PlanMetadata    *string            `json:"planMetadata,omitempty"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`

	// Populated for responses
	Plan *Plan        `json:"plan,omitempty"`
	User *User        `json:"user,omitempty"`
	Logs []ProjectLog `json:"logs"`
}

// ProjectLog represents a log entry for a subscription project
type ProjectLog struct {
	ID             uuid.UUID  `json:"id"`
	SubscriptionID uuid.UUID  `json:"subscriptionId"`
	Message        string     `json:"message"`
	CreatedBy      *uuid.UUID `json:"createdBy,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type CreateSubscriptionRequest struct {
	PlanID uuid.UUID `json:"planId" binding:"required"`
}

type UpdateSubscriptionRequest struct {
	Status          *SubscriptionStatus `json:"status"`
	ProjectStage    *string             `json:"projectStage"`
	NextBillingDate *time.Time          `json:"nextBillingDate"`
}

type AddProjectLogRequest struct {
	Message string `json:"message" binding:"required"`
}

// ActiveSubscriptionDTO represents the data for the active subscriptions list
// Price is stored in cents (int64) to avoid float precision issues
type ActiveSubscriptionDTO struct {
	ID              string  `json:"id"`
	UserID          string  `json:"userId"`
	UserName        string  `json:"userName"`
	UserEmail       string  `json:"userEmail"`
	PlanName        string  `json:"planName"`
	Price           int64   `json:"price"` // Price in cents
	Status          string  `json:"status"`
	ProjectStage    string  `json:"projectStage"`
	NextBillingDate string  `json:"nextBillingDate"`
}
