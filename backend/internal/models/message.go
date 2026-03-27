package models

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message in a subscription
type Message struct {
	ID             uuid.UUID `json:"id"`
	SubscriptionID uuid.UUID `json:"subscriptionId"`
	UserID         uuid.UUID `json:"userId"`
	Content        string    `json:"content"`
	IsAdmin        bool      `json:"isAdmin"`
	CreatedAt      time.Time `json:"createdAt"`
}

type CreateMessageRequest struct {
	Content string `json:"content" binding:"required"`
}
