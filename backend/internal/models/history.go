package models

import (
	"time"
	"github.com/google/uuid"
)

// SubscriptionMini representa os campos mínimos de uma assinatura para o histórico
type SubscriptionMini struct {
	ID            uuid.UUID `json:"id"`
	PlanName      string    `json:"plan_name"`
	PricePerMonth int64     `json:"price_per_month"` // Price in cents
	CreatedAt     time.Time `json:"created_at"`
}

// ProductMini representa os campos mínimos de um produto comprado para o histórico
type ProductMini struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Price int64     `json:"price"` // Price in cents
	Type  ProductType `json:"type"`
}

// HistoryResponse agrega subscriptions e produtos comprados do usuário
type HistoryResponse struct {
	Subscriptions []SubscriptionMini `json:"subscriptions"`
	Products      []ProductMini      `json:"products"`
}

// Invoice representa uma fatura simplificada
type Invoice struct {
	ID     uuid.UUID `json:"id"`
	Amount int64     `json:"amount"` // Amount in cents
	Date   time.Time `json:"date"`
}

// SubscriptionInvoice represents a subscription invoice with detailed information
type SubscriptionInvoice struct {
	SubscriptionID uuid.UUID   `json:"subscription_id"`
	PlanName       string      `json:"plan_name"`
	Amount         int64       `json:"amount"` // Amount in cents
	DueDate        time.Time   `json:"due_date"`
	Status         InvoiceStatus `json:"status"` // e.g., "paid", "due", "overdue"
}

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

// InvoiceHistoryResponse is the response for invoice history queries
type InvoiceHistoryResponse struct {
	PaidInvoices      []SubscriptionInvoice `json:"paid_invoices"`
	NextInvoice       *SubscriptionInvoice  `json:"next_invoice"`
	OverdueInvoices   []SubscriptionInvoice `json:"overdue_invoices"`
	DueInvoices       []SubscriptionInvoice `json:"due_invoices"`
}