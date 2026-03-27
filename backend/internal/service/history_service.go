package service

import (
	"context"
	"fmt"

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
	var items []models.UserPurchaseWithProduct

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

	// Goroutine 2: Get user library (purchases + product)
	g.Go(func() error {
		var fetchErr error
		items, fetchErr = s.libraryRepo.GetUserLibrary(ctx, uid)
		if fetchErr != nil {
			return fmt.Errorf("failed to get user library: %w", fetchErr)
		}
		return nil
	})

	// Wait for both goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Transform library items to minimal product format
	products := make([]models.ProductMini, 0, len(items))
	for _, it := range items {
		products = append(products, models.ProductMini{
			ID:    it.Product.ID,
			Name:  it.Product.Name,
			Price: it.Product.Price,
			Type:  it.Product.Type,
		})
	}

	resp := &models.HistoryResponse{
		Subscriptions: subs,
		Products:      products,
	}
	return resp, nil
}

// GetUserInvoiceHistory retrieves and categorizes user invoices.
func (s *HistoryService) GetUserInvoiceHistory(ctx context.Context, userID string) (*models.InvoiceHistoryResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	invoices, err := s.paymentRepo.GetUserSubscriptionInvoices(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscription invoices: %w", err)
	}

	var paidInvoices []models.SubscriptionInvoice
	var overdueInvoices []models.SubscriptionInvoice
	var dueInvoices []models.SubscriptionInvoice

	for _, invoice := range invoices {
		switch invoice.Status {
		case models.InvoiceStatusPaid:
			paidInvoices = append(paidInvoices, invoice)
		case models.InvoiceStatusOverdue:
			overdueInvoices = append(overdueInvoices, invoice)
		case models.InvoiceStatusDue:
			dueInvoices = append(dueInvoices, invoice)
		}
	}

	// Select the earliest due invoice as next invoice
	var nextInvoice *models.SubscriptionInvoice
	for i, invoice := range dueInvoices {
		if nextInvoice == nil || invoice.DueDate.Before(nextInvoice.DueDate) {
			nextInvoice = &dueInvoices[i]
		}
	}

	return &models.InvoiceHistoryResponse{
		PaidInvoices:    paidInvoices,
		NextInvoice:     nextInvoice,
		OverdueInvoices: overdueInvoices,
		DueInvoices:     dueInvoices,
	}, nil
}