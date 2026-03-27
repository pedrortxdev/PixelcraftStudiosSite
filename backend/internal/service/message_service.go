package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

type MessageService struct {
	messageRepo      *repository.MessageRepository
	subscriptionRepo *repository.SubscriptionRepository
}

func NewMessageService(msgRepo *repository.MessageRepository, subRepo *repository.SubscriptionRepository) *MessageService {
	return &MessageService{
		messageRepo:      msgRepo,
		subscriptionRepo: subRepo,
	}
}

func (s *MessageService) SendMessage(ctx context.Context, subIDStr, userIDStr, content string, isAdmin bool) (*models.Message, error) {
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		return nil, errors.New("invalid subscription ID")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Security Check: If not admin, verify subscription ownership
	if !isAdmin {
		sub, err := s.subscriptionRepo.GetSubscriptionByID(ctx, subID)
		if err != nil {
			return nil, err
		}
		if sub == nil {
			return nil, errors.New("subscription not found")
		}
		if sub.UserID != userID {
			return nil, errors.New("unauthorized: you do not own this subscription")
		}
	}

	msg := &models.Message{
		ID:             uuid.New(),
		SubscriptionID: subID,
		UserID:         userID,
		Content:        content,
		IsAdmin:        isAdmin,
		CreatedAt:      time.Now(),
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *MessageService) GetChatHistory(ctx context.Context, subIDStr, userIDStr string, isAdmin bool) ([]models.Message, error) {
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		return nil, errors.New("invalid subscription ID")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Security Check: If not admin, verify subscription ownership
	if !isAdmin {
		sub, err := s.subscriptionRepo.GetSubscriptionByID(ctx, subID)
		if err != nil {
			return nil, err
		}
		if sub == nil {
			return nil, errors.New("subscription not found")
		}
		if sub.UserID != userID {
			return nil, errors.New("unauthorized: you do not own this subscription")
		}
	}

	return s.messageRepo.GetBySubscriptionID(ctx, subID)
}
