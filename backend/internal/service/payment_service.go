package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/apierrors"
	"github.com/pixelcraft/api/internal/models"
)

// PaymentRepositoryGateway defines the interface for payment repository operations
// This allows dependency injection and testability
type PaymentRepositoryGateway interface {
	GetUserPaymentStats(ctx context.Context, userID uuid.UUID) (*models.PaymentStats, error)
	GetRecentPayments(ctx context.Context, userID uuid.UUID, limit int) ([]models.PaymentInfo, error)
	GetMonthlySpending(ctx context.Context, userID uuid.UUID, months int) ([]models.MonthlySpend, error)
	GetNextBillingSummary(ctx context.Context, userID uuid.UUID) (int64, []string, error)
}

// PaymentService handles payment-related business logic
type PaymentService struct {
	paymentRepo PaymentRepositoryGateway
}

// Validation constants for input sanitization
const (
	MinPaymentLimit      = 1
	MaxPaymentLimit      = 100
	DefaultPaymentLimit  = 10
	MinMonths            = 1
	MaxMonths            = 24
	DefaultMonths        = 6
)

// NewPaymentService creates a new PaymentService with injected repository
func NewPaymentService(paymentRepo PaymentRepositoryGateway) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
	}
}

// validateLimit ensures the limit is within acceptable bounds
func (s *PaymentService) validateLimit(limit int) (int, error) {
	if limit <= 0 {
		return 0, apierrors.ErrInvalidPaymentLimit
	}
	if limit > MaxPaymentLimit {
		return 0, fmt.Errorf("%w: cannot exceed %d (requested: %d)", apierrors.ErrPaymentLimitExceeded, MaxPaymentLimit, limit)
	}
	return limit, nil
}

// validateMonths ensures the months parameter is within acceptable bounds
func (s *PaymentService) validateMonths(months int) (int, error) {
	if months <= 0 {
		return 0, apierrors.ErrInvalidPaymentMonths
	}
	if months > MaxMonths {
		return 0, fmt.Errorf("%w: cannot exceed %d (requested: %d)", apierrors.ErrPaymentMonthsExceeded, MaxMonths, months)
	}
	return months, nil
}

// GetUserPaymentStats gets payment statistics for a user
func (s *PaymentService) GetUserPaymentStats(ctx context.Context, userID uuid.UUID) (*models.PaymentStats, error) {
	stats, err := s.paymentRepo.GetUserPaymentStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment stats: %w", err)
	}
	return stats, nil
}

// GetRecentPayments gets recent payments for a user with validated limit
func (s *PaymentService) GetRecentPayments(ctx context.Context, userID uuid.UUID, limit int) ([]models.PaymentInfo, error) {
	// Validate and sanitize limit parameter (business rule: prevent abuse)
	validatedLimit, err := s.validateLimit(limit)
	if err != nil {
		return nil, fmt.Errorf("invalid limit parameter: %w", err)
	}

	payments, err := s.paymentRepo.GetRecentPayments(ctx, userID, validatedLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent payments: %w", err)
	}
	return payments, nil
}

// GetMonthlySpending gets monthly spending for a user with validated months parameter
func (s *PaymentService) GetMonthlySpending(ctx context.Context, userID uuid.UUID, months int) ([]models.MonthlySpend, error) {
	// Validate and sanitize months parameter (business rule: prevent expensive queries)
	validatedMonths, err := s.validateMonths(months)
	if err != nil {
		return nil, fmt.Errorf("invalid months parameter: %w", err)
	}

	spending, err := s.paymentRepo.GetMonthlySpending(ctx, userID, validatedMonths)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly spending: %w", err)
	}
	return spending, nil
}

// GetNextBillingSummary gets aggregate next billing info for active subscriptions
func (s *PaymentService) GetNextBillingSummary(ctx context.Context, userID uuid.UUID) (models.NextBillingSummary, error) {
	total, dates, err := s.paymentRepo.GetNextBillingSummary(ctx, userID)
	if err != nil {
		return models.NextBillingSummary{}, fmt.Errorf("failed to get next billing summary: %w", err)
	}
	return models.NextBillingSummary{Total: total, Dates: dates}, nil
}
