package service

import (
	"context"

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

func (s *DiscountService) ListDiscounts(ctx context.Context) ([]models.Discount, error) {
	return s.repo.List(ctx)
}

func (s *DiscountService) GetDiscount(ctx context.Context, id uuid.UUID) (*models.Discount, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DiscountService) CreateDiscount(ctx context.Context, req *models.CreateDiscountRequest, adminID uuid.UUID) (*models.Discount, error) {
	discount := &models.Discount{
		ID:              uuid.New(),
		Code:            req.Code,
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

func (s *DiscountService) UpdateDiscount(ctx context.Context, id uuid.UUID, req *models.UpdateDiscountRequest) error {
	updates := make(map[string]interface{})
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Value != nil {
		updates["value"] = *req.Value
	}
	if req.RestrictionType != nil {
		updates["restriction_type"] = *req.RestrictionType
	}
	if req.TargetIDs != nil {
		updates["target_ids"] = req.TargetIDs
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = req.ExpiresAt
	}
	if req.MaxUses != nil {
		updates["max_uses"] = req.MaxUses
	}

	return s.repo.Update(ctx, id, updates)
}

func (s *DiscountService) DeleteDiscount(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
