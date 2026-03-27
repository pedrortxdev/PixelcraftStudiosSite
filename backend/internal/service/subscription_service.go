package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// SubscriptionService handles business logic for subscriptions
type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

// NewSubscriptionService creates a new SubscriptionService
func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

// ListActivePlans retrieves all active plans
func (s *SubscriptionService) ListActivePlans(ctx context.Context) ([]models.Plan, error) {
	return s.repo.ListActivePlans(ctx)
}

// GetUserSubscriptions retrieves subscriptions for a specific user
func (s *SubscriptionService) GetUserSubscriptions(ctx context.Context, userID string) ([]models.Subscription, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// GetSubscriptionDetails retrieves full details of a subscription
func (s *SubscriptionService) GetSubscriptionDetails(ctx context.Context, id uuid.UUID) (*models.Subscription, []models.ProjectLog, error) {
	// 1. Pega a assinatura (O Repo já faz JOIN com Plans para trazer o nome real)
	sub, err := s.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if sub == nil {
		return nil, nil, nil
	}

	// 2. Pega os logs (Diário de Bordo)
	logs, err := s.repo.GetLogs(ctx, id)
	if err != nil {
		// Não vamos falhar tudo se os logs falharem, apenas logar o erro se fosse production
		// Mas aqui vamos retornar nil logs
		logs = []models.ProjectLog{} 
	}

	return sub, logs, nil
}

// UpdateSubscription updates subscription status and stage
func (s *SubscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req models.UpdateSubscriptionRequest) error {
	sub, err := s.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return nil // Or error not found
	}

	if req.Status != nil {
		sub.Status = *req.Status
	}
	if req.ProjectStage != nil {
		sub.ProjectStage = *req.ProjectStage
	}
	if req.NextBillingDate != nil {
		sub.NextBillingDate = *req.NextBillingDate
	}

	err = s.repo.UpdateSubscription(ctx, sub)
	if err != nil {
		fmt.Printf("Service Error updating subscription %s: %v\n", id, err)
		return err
	}
	return nil
}

// AddProjectLog adds a log to a subscription
func (s *SubscriptionService) AddProjectLog(ctx context.Context, subscriptionID uuid.UUID, message string, createdBy *uuid.UUID) error {
	log := &models.ProjectLog{
		ID:             uuid.New(),
		SubscriptionID: subscriptionID,
		Message:        message,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
	}
	return s.repo.AddLog(ctx, log)
}

// GetActiveSubscriptions retrieves all active subscriptions for admin
func (s *SubscriptionService) GetActiveSubscriptions(ctx context.Context) ([]models.ActiveSubscriptionDTO, error) {
	return s.repo.ListActiveSubscriptions(ctx)
}

// CreatePlan creates a new plan
func (s *SubscriptionService) CreatePlan(ctx context.Context, plan *models.Plan) error {
	plan.ID = uuid.New()
	return s.repo.CreatePlan(ctx, plan)
}

// UpdatePlan updates an existing plan
func (s *SubscriptionService) UpdatePlan(ctx context.Context, plan *models.Plan) error {
	return s.repo.UpdatePlan(ctx, plan)
}

// DeletePlan soft deletes a plan
func (s *SubscriptionService) DeletePlan(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePlan(ctx, id)
}