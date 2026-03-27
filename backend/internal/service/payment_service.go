package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// PaymentService handles payment-related business logic
type PaymentService struct {
	paymentRepo *repository.PaymentRepository
}

// NewPaymentService creates a new PaymentService
func NewPaymentService(db *sql.DB) *PaymentService {
	return &PaymentService{
		paymentRepo: repository.NewPaymentRepository(db),
	}
}

// GetUserPaymentStats gets payment statistics for a user
func (s *PaymentService) GetUserPaymentStats(ctx context.Context, userID string) (*models.PaymentStats, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	stats, err := s.paymentRepo.GetUserPaymentStats(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment stats: %w", err)
	}

	return stats, nil
}

// GetRecentPayments gets recent payments for a user
func (s *PaymentService) GetRecentPayments(ctx context.Context, userID string, limit int) ([]models.PaymentInfo, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	payments, err := s.paymentRepo.GetRecentPayments(ctx, uid, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent payments: %w", err)
	}

	return payments, nil
}

// GetMonthlySpending gets monthly spending for a user
func (s *PaymentService) GetMonthlySpending(ctx context.Context, userID string, months int) ([]models.MonthlySpend, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	spending, err := s.paymentRepo.GetMonthlySpending(ctx, uid, months)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly spending: %w", err)
	}

	return spending, nil
}

// GetNextBillingSummary gets aggregate next billing info for active subscriptions
func (s *PaymentService) GetNextBillingSummary(ctx context.Context, userID string) (models.NextBillingSummary, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return models.NextBillingSummary{}, fmt.Errorf("invalid user ID: %w", err)
	}
	total, dates, err := s.paymentRepo.GetNextBillingSummary(ctx, uid)
	if err != nil {
		return models.NextBillingSummary{}, fmt.Errorf("failed to get next billing summary: %w", err)
	}
	return models.NextBillingSummary{Total: total, Dates: dates}, nil
}