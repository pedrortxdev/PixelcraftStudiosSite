package models

import "time"

// TicketStatus represents the status of a support ticket
type TicketStatus string

const (
	TicketOpen            TicketStatus = "OPEN"
	TicketInProgress      TicketStatus = "IN_PROGRESS"
	TicketWaitingResponse TicketStatus = "WAITING_RESPONSE"
	TicketResolved        TicketStatus = "RESOLVED"
	TicketClosed          TicketStatus = "CLOSED"
)

// TicketCategory represents the category of a support ticket
type TicketCategory string

const (
	CategoryGeneral      TicketCategory = "GENERAL"
	CategorySubscription TicketCategory = "SUBSCRIPTION"
	CategoryPayment      TicketCategory = "PAYMENT"
	CategoryTechnical    TicketCategory = "TECHNICAL"
	CategoryBilling      TicketCategory = "BILLING"
	CategoryOther        TicketCategory = "OTHER"
)

// SupportTicket represents a support ticket
type SupportTicket struct {
	ID             string         `json:"id" db:"id"`
	UserID         string         `json:"user_id" db:"user_id"`
	Subject        string         `json:"subject" db:"subject"`
	Category       TicketCategory `json:"category" db:"category"`
	Priority       float64        `json:"priority" db:"priority"` // 1-5 stars based on role
	Status         TicketStatus   `json:"status" db:"status"`
	AssignedTo     *string        `json:"assigned_to,omitempty" db:"assigned_to"`
	SubscriptionID *string        `json:"subscription_id,omitempty" db:"subscription_id"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
	ResolvedAt     *time.Time     `json:"resolved_at,omitempty" db:"resolved_at"`
	ClosedAt       *time.Time     `json:"closed_at,omitempty" db:"closed_at"`
	// Populated fields for API responses
	User           *User            `json:"user,omitempty"`
	AssignedStaff  *User            `json:"assigned_staff,omitempty"`
	Messages       []SupportMessage `json:"messages,omitempty"`
	MessageCount   int              `json:"message_count,omitempty"`
	LastMessage    *SupportMessage  `json:"last_message,omitempty"`
}

// SupportMessage represents a message within a support ticket
type SupportMessage struct {
	ID        string    `json:"id" db:"id"`
	TicketID  string    `json:"ticket_id" db:"ticket_id"`
	SenderID  string    `json:"sender_id" db:"sender_id"`
	Content   string    `json:"content" db:"content"`
	IsStaff   bool      `json:"is_staff" db:"is_staff"`
	AttachmentURL  *string   `json:"attachment_url,omitempty" db:"attachment_url"`
	AttachmentType *string   `json:"attachment_type,omitempty" db:"attachment_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// Populated field
	Sender    *User     `json:"sender,omitempty"`
}

// CreateTicketRequest represents the request to create a new support ticket
type CreateTicketRequest struct {
	Subject        string         `json:"subject" binding:"required,min=5,max=255"`
	Category       TicketCategory `json:"category" binding:"omitempty"`
	Content        string         `json:"content" binding:"required,min=10"` // Initial message
	SubscriptionID *string        `json:"subscription_id,omitempty"`
}

// CreateMessageRequest represents the request to send a message in a ticket
type CreateSupportMessageRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

// UpdateTicketStatusRequest represents the request to update ticket status
type UpdateTicketStatusRequest struct {
	Status TicketStatus `json:"status" binding:"required"`
}

// AssignTicketRequest represents the request to assign a ticket to staff
type AssignTicketRequest struct {
	AssignedTo string `json:"assigned_to" binding:"required,uuid"`
}

// TicketListFilter contains filter options for listing tickets
type TicketListFilter struct {
	Status     *TicketStatus   `json:"status,omitempty"`
	Category   *TicketCategory `json:"category,omitempty"`
	Priority   *float64        `json:"priority,omitempty"`
	AssignedTo *string         `json:"assigned_to,omitempty"`
	UserID     *string         `json:"user_id,omitempty"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
}

// TicketListResponse contains paginated ticket list
type TicketListResponse struct {
	Tickets    []SupportTicket `json:"tickets"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}
