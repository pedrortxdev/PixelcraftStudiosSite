package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// HistoryService agrega dados de assinaturas e produtos comprados do usuário
type HistoryService struct {
	paymentRepo *repository.PaymentRepository
	libraryRepo *repository.LibraryRepository
}

func NewHistoryService(paymentRepo *repository.PaymentRepository, libraryRepo *repository.LibraryRepository) *HistoryService {
	return &HistoryService{paymentRepo: paymentRepo, libraryRepo: libraryRepo}
}

// GetUserHistory retorna subscriptions (mínimo) e produtos comprados (mínimo)
func (s *HistoryService) GetUserHistory(ctx context.Context, userID string) (*models.HistoryResponse, error) {
	// Subscriptions mínimas
	subs, err := s.paymentRepo.GetUserSubscriptionsMinimal(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	// Biblioteca do usuário (compras + produto). Vamos reduzir para campos mínimos
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	items, err := s.libraryRepo.GetUserLibrary(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user library: %w", err)
	}

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
	invoices, err := s.paymentRepo.GetUserSubscriptionInvoices(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscription invoices: %w", err)
	}

	var paidInvoices []models.SubscriptionInvoice
	var overdueInvoices []models.SubscriptionInvoice
	var nextInvoice *models.SubscriptionInvoice

	for i, invoice := range invoices {
		switch invoice.Status {
		case "paid":
			paidInvoices = append(paidInvoices, invoice)
		case "overdue":
			overdueInvoices = append(overdueInvoices, invoice)
		case "due":
			if nextInvoice == nil || invoice.DueDate.Before(nextInvoice.DueDate) {
				nextInvoice = &invoices[i]
			}
		}
	}

	return &models.InvoiceHistoryResponse{
		PaidInvoices:    paidInvoices,
		NextInvoice:     nextInvoice,
		OverdueInvoices: overdueInvoices,
	}, nil
}