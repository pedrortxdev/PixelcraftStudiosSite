package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

const (
	// MaxMessageLength defines the maximum allowed message length (10000 runes/characters)
	MaxMessageLength = 10000
	// MinMessageLength defines the minimum allowed message length
	MinMessageLength = 1
	// DefaultChatHistoryLimit defines the default number of messages to return
	DefaultChatHistoryLimit = 100
	// DefaultChatHistoryOffset defines the default offset for pagination
	DefaultChatHistoryOffset = 0
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

// validateSubscriptionAccess verifies that the user has access to the subscription.
// Returns the subscription if valid, error otherwise.
func (s *MessageService) validateSubscriptionAccess(ctx context.Context, subID uuid.UUID, userID uuid.UUID, isAdmin bool) (*models.Subscription, error) {
	// Admin bypasses ownership check but still needs valid subscription
	sub, err := s.subscriptionRepo.GetSubscriptionByID(ctx, subID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}
	if sub == nil {
		return nil, errors.New("subscription not found")
	}

	// Non-admin users must own the subscription
	if !isAdmin && sub.UserID != userID {
		return nil, errors.New("unauthorized: you do not own this subscription")
	}

	return sub, nil
}

// validateMessageContent ensures the message content is valid
func (s *MessageService) validateMessageContent(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return errors.New("message content cannot be empty")
	}
	// Use RuneCountInString to count actual characters (runes), not bytes.
	// This correctly handles multi-byte characters like emojis (🧟‍♂️ = 4 bytes),
	// accented characters (ã, é, ñ), and CJK characters.
	if utf8.RuneCountInString(content) > MaxMessageLength {
		return fmt.Errorf("message content exceeds maximum length of %d characters", MaxMessageLength)
	}
	return nil
}

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	Content string
}

// SendMessage sends a message to a subscription chat
func (s *MessageService) SendMessage(ctx context.Context, subID uuid.UUID, userID uuid.UUID, content string, isAdmin bool) (*models.Message, error) {
	// Validate message content (DRY: business rule)
	if err := s.validateMessageContent(content); err != nil {
		return nil, err
	}

	// Validate subscription access (DRY: security check - handles both admin and user)
	if _, err := s.validateSubscriptionAccess(ctx, subID, userID, isAdmin); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return msg, nil
}

// GetChatHistoryParams defines pagination parameters for chat history
type GetChatHistoryParams struct {
	Limit  int
	Offset int
}

// GetChatHistory retrieves paginated chat history for a subscription
func (s *MessageService) GetChatHistory(ctx context.Context, subID uuid.UUID, userID uuid.UUID, isAdmin bool, params *GetChatHistoryParams) ([]models.Message, error) {
	// Set defaults if params not provided
	if params == nil {
		params = &GetChatHistoryParams{
			Limit:  DefaultChatHistoryLimit,
			Offset: DefaultChatHistoryOffset,
		}
	}

	// Validate subscription access (DRY: security check - handles both admin and user)
	if _, err := s.validateSubscriptionAccess(ctx, subID, userID, isAdmin); err != nil {
		return nil, err
	}

	// Get paginated messages to prevent OOM on large chat histories
	return s.messageRepo.GetBySubscriptionIDPaginated(ctx, subID, params.Limit, params.Offset)
}
