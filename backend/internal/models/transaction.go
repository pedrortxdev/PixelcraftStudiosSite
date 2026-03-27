package models

import (
	"time"

	"github.com/google/uuid"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusApproved  TransactionStatus = "approved"
	TransactionStatusRejected  TransactionStatus = "rejected"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusRefunded  TransactionStatus = "refunded"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeDeposit         TransactionType = "deposit"
	TransactionTypeAdminAdjustment TransactionType = "admin_adjustment"
)

// Transaction represents a wallet transaction
type Transaction struct {
	ID                uuid.UUID         `db:"id" json:"id"`
	UserID            uuid.UUID         `db:"user_id" json:"user_id"`
	ProviderPaymentID *string           `db:"provider_payment_id" json:"provider_payment_id"` // Pointer as it might be null initially or strict constraint
	Amount            float64           `db:"amount" json:"amount"`
	Status            TransactionStatus `db:"status" json:"status"`
	Type              TransactionType   `db:"type" json:"type"`
	AdjustmentType    *string           `db:"adjustment_type" json:"adjustment_type"` // "Teste" or "Pix Direto"
	QRCode            *string           `db:"qr_code" json:"qr_code,omitempty"`
	QRCodeBase64      *string           `db:"qr_code_base64" json:"qr_code_base64,omitempty"`
	CreatedAt         time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `db:"updated_at" json:"updated_at"`
}

// DepositRequest represents the request to add funds
type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Method string  `json:"method" binding:"required,oneof=pix link"`
}

// DepositResponse represents the response for a deposit request
type DepositResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	PaymentLink   string    `json:"payment_link,omitempty"`
	QRCode        string    `json:"qr_code,omitempty"`
	QRCodeBase64  string    `json:"qr_code_base64,omitempty"`
}
