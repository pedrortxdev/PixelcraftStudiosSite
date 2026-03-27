package models

import (
	"time"
	"github.com/google/uuid"
)

// SubscriptionMini representa os campos mínimos de uma assinatura para o histórico
type SubscriptionMini struct {
	ID            uuid.UUID `json:"id"`
	PlanName      string    `json:"plan_name"`
	PricePerMonth float64   `json:"price_per_month"`
	CreatedAt     time.Time `json:"created_at"`
}

// ProductMini representa os campos mínimos de um produto comprado para o histórico
type ProductMini struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Price float64   `json:"price"`
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
	Amount float64   `json:"amount"`
	Date   time.Time `json:"date"`
}

// SubscriptionInvoice represents a subscription invoice with detailed information
type SubscriptionInvoice struct {
	SubscriptionID uuid.UUID   `json:"subscription_id"`
	PlanName       string      `json:"plan_name"`
	Amount         float64     `json:"amount"`
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