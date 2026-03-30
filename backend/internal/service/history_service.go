package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
	"golang.org/x/sync/errgroup"
)

// HistoryService aggregates user subscription and purchased product data
type HistoryService struct {
	paymentRepo *repository.PaymentRepository
	libraryRepo *repository.LibraryRepository
}

func NewHistoryService(paymentRepo *repository.PaymentRepository, libraryRepo *repository.LibraryRepository) *HistoryService {
	return &HistoryService{paymentRepo: paymentRepo, libraryRepo: libraryRepo}
}

// GetUserHistory returns minimal subscriptions and purchased products using parallel I/O
func (s *HistoryService) GetUserHistory(ctx context.Context, userID string) (*models.HistoryResponse, error) {
	// Parse userID once
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Use errgroup for parallel execution of independent I/O operations
	var subs []models.SubscriptionMini
	var products []models.ProductMini

	g, ctx := errgroup.WithContext(ctx)

	// Goroutine 1: Get minimal subscriptions
	g.Go(func() error {
		var fetchErr error
		subs, fetchErr = s.paymentRepo.GetUserSubscriptionsMinimal(ctx, uid)
		if fetchErr != nil {
			return fmt.Errorf("failed to get subscriptions: %w", fetchErr)
		}
		return nil
	})

	// Goroutine 2: Get user library minimal (optimized query)
	g.Go(func() error {
		var fetchErr error
		products, fetchErr = s.libraryRepo.GetUserLibraryMinimal(ctx, uid)
		if fetchErr != nil {
			return fmt.Errorf("failed to get user library minimal: %w", fetchErr)
		}
		return nil
	})

	// Wait for both goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	resp := &models.HistoryResponse{
		Subscriptions: subs,
		Products:      products,
	}
	return resp, nil
}

// GetUserInvoiceHistory retrieves and categorizes user invoices.
// Optimized: Single-pass categorization with NextInvoice priority (Overdue > Due)
func (s *HistoryService) GetUserInvoiceHistory(ctx context.Context, userID string) (*models.InvoiceHistoryResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	invoices, err := s.paymentRepo.GetUserSubscriptionInvoices(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscription invoices: %w", err)
	}

	capacity := len(invoices)
	paidInvoices := make([]models.SubscriptionInvoice, 0, capacity)
	overdueInvoices := make([]models.SubscriptionInvoice, 0, capacity)
	dueInvoices := make([]models.SubscriptionInvoice, 0, capacity)
	var nextInvoice *models.SubscriptionInvoice

	for i := range invoices {
		invoice := &invoices[i]
		switch invoice.Status {
		case models.InvoiceStatusPaid:
			paidInvoices = append(paidInvoices, *invoice)

		case models.InvoiceStatusOverdue:
			overdueInvoices = append(overdueInvoices, *invoice)
			// Overdue takes priority: always update nextInvoice if this is the earliest overdue
			if nextInvoice == nil || truncateToDay(invoice.DueDate).Before(truncateToDay(nextInvoice.DueDate)) {
				nextInvoice = invoice
			}

		case models.InvoiceStatusDue:
			dueInvoices = append(dueInvoices, *invoice)
			// Only update nextInvoice if:
			// 1. No nextInvoice exists yet, OR
			// 2. Current nextInvoice is Overdue (keep the overdue one), OR
			// 3. This Due invoice is earlier than the current nextInvoice (and current is also Due)
			if nextInvoice == nil {
				nextInvoice = invoice
			} else if nextInvoice.Status != models.InvoiceStatusOverdue && truncateToDay(invoice.DueDate).Before(truncateToDay(nextInvoice.DueDate)) {
				nextInvoice = invoice
			}
		}
	}

	return &models.InvoiceHistoryResponse{
		PaidInvoices:    paidInvoices,
		NextInvoice:     nextInvoice,
		OverdueInvoices: overdueInvoices,
		DueInvoices:     dueInvoices,
	}, nil
}

// truncateToDay normalizes a timestamp precisely to its local calendar day 
// ignoring time components, useful for equitable invoice "due day" comparisons.
func truncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}