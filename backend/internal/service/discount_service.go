package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/pixelcraft/api/internal/apierrors"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"

	"github.com/google/uuid"
)

type DiscountService struct {
	repo *repository.DiscountRepository
}

func NewDiscountService(repo *repository.DiscountRepository) *DiscountService {
	return &DiscountService{repo: repo}
}

func (s *DiscountService) ListDiscounts(ctx context.Context, includeDeleted bool) ([]models.Discount, error) {
	return s.repo.List(ctx, includeDeleted)
}

func (s *DiscountService) GetDiscount(ctx context.Context, id uuid.UUID) (*models.Discount, error) {
	discount, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if discount == nil || !discount.IsActive {
		return nil, apierrors.ErrDiscountNotFound
	}
	return discount, nil
}

// CreateDiscount creates a new discount with full validation
func (s *DiscountService) CreateDiscount(ctx context.Context, req *models.CreateDiscountRequest, adminID uuid.UUID) (*models.Discount, error) {
	// 1. Validate discount value based on type
	if err := validateDiscountValue(req.Type, req.Value); err != nil {
		return nil, err
	}

	// 2. Validate code length and format
	normalizedCode := strings.ToUpper(strings.TrimSpace(req.Code))
	if len(normalizedCode) < 3 || strings.Contains(normalizedCode, " ") {
		return nil, fmt.Errorf("%w: código deve ter no mínimo 3 caracteres e não conter espaços", apierrors.ErrInvalidInput)
	}

	// 3. Create discount
	discount := &models.Discount{
		ID:              uuid.New(),
		Code:            normalizedCode, // Normalize to uppercase
		Type:            req.Type,
		Value:           req.Value,
		IsReferral:      req.IsReferral,
		RestrictionType: req.RestrictionType,
		TargetIDs:       req.TargetIDs,
		CreatedByUserID: &adminID,
		ExpiresAt:       req.ExpiresAt,
		MaxUses:         req.MaxUses,
		CurrentUses:     0,
		IsActive:        true,
	}

	err := s.repo.Create(ctx, discount)
	if err != nil {
		return nil, err
	}

	return discount, nil
}

// UpdateDiscount updates an existing discount with validation
func (s *DiscountService) UpdateDiscount(ctx context.Context, id uuid.UUID, req *models.UpdateDiscountRequest) error {
	// 1. Check if discount exists and is active
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil || !existing.IsActive {
		return apierrors.ErrDiscountNotFound
	}

	// 2. Validate new value if provided
	if req.Value != nil || req.Type != nil {
		newType := existing.Type
		if req.Type != nil {
			newType = *req.Type
		}
		newValue := existing.Value
		if req.Value != nil {
			newValue = *req.Value
		}
		if err := validateDiscountValue(newType, newValue); err != nil {
			return err
		}
	}

	// 3. Check if code is being changed and validate format
	if req.Code != nil {
		normalized := strings.ToUpper(strings.TrimSpace(*req.Code))
		if len(normalized) < 3 || strings.Contains(normalized, " ") {
			return fmt.Errorf("%w: código deve ter no mínimo 3 caracteres e não conter espaços", apierrors.ErrInvalidInput)
		}
		req.Code = &normalized // Assign back the normalized code
	}

	// 4. Build domain-level update struct (avoids leaking DB column names)
	updates := &models.DiscountUpdate{
		Code:            req.Code,
		Type:            req.Type,
		Value:           req.Value,
		RestrictionType: req.RestrictionType,
		TargetIDs:       req.TargetIDs,
		IsActive:        req.IsActive,
		ExpiresAt:       req.ExpiresAt,
		MaxUses:         req.MaxUses,
	}

	// Normalize code to uppercase if provided already handled
	err = s.repo.Update(ctx, id, updates)
	if err != nil {
		return err
	}
	
	return nil
}

// DeleteDiscount soft deletes (deactivates) a discount
func (s *DiscountService) DeleteDiscount(ctx context.Context, id uuid.UUID) error {
	// 1. Check if discount exists and is active
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil || !existing.IsActive {
		return apierrors.ErrDiscountNotFound
	}

	// 2. Soft delete by deactivating (preserves historical data)
	return s.repo.Deactivate(ctx, id)
}

// validateDiscountValue validates discount value based on type
// Value is in cents for FIXED_AMOUNT, or percentage points for PERCENTAGE
func validateDiscountValue(discountType models.DiscountType, value int64) error {
	// Check for negative values
	if value < 0 {
		return apierrors.ErrDiscountNegativeValue
	}

	// For percentage discounts, value must be between 0 and 100
	if discountType == models.DiscountTypePercentage {
		if value > 100 {
			return apierrors.ErrDiscountInvalidPercentage
		}
		// Additional check: percentage should be > 0 to be meaningful
		if value == 0 {
			return apierrors.ErrDiscountInvalidValue
		}
	}

	// For fixed amount discounts, just ensure it's positive
	if discountType == models.DiscountTypeFixedAmount {
		if value == 0 {
			return apierrors.ErrDiscountInvalidValue
		}
	}

	return nil
}
