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
	// DefaultChatHistoryOffset defines the default offset for pagination (legacy)
	DefaultChatHistoryOffset = 0
)

// Allowed subscription statuses for chat access
var AllowedChatSubscriptionStatuses = []models.SubscriptionStatus{
	models.SubscriptionStatusActive,
	models.SubscriptionStatusCanceled, // Allow chat for canceled subscriptions (grace period)
}

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
// Returns the subscription owner ID if valid, error otherwise.
// OPTIMIZATION: Only returns owner ID to reduce data transfer
func (s *MessageService) validateSubscriptionAccess(ctx context.Context, subID uuid.UUID, userID uuid.UUID, isAdmin bool) (uuid.UUID, error) {
	// OPTIMIZATION: Use lightweight validation query instead of fetching full subscription
	ownerID, status, err := s.subscriptionRepo.GetSubscriptionOwnerAndStatus(ctx, subID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}
	if ownerID == uuid.Nil {
		return uuid.Nil, errors.New("subscription not found")
	}

	// SECURITY: Check subscription status allows chat interaction
	// Prevents "zombie" chats from canceled/suspended subscriptions
	statusAllowed := false
	for _, allowedStatus := range AllowedChatSubscriptionStatuses {
		if status == allowedStatus {
			statusAllowed = true
			break
		}
	}
	if !statusAllowed {
		return uuid.Nil, fmt.Errorf("subscription chat is disabled: status=%s", status)
	}

	// Non-admin users must own the subscription
	if !isAdmin && ownerID != userID {
		return uuid.Nil, errors.New("unauthorized: you do not own this subscription")
	}

	// Return owner ID for audit trail (admin messages should use actual admin user ID)
	return ownerID, nil
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
// FIX 4: Admin user ID is stored correctly for audit trail
func (s *MessageService) SendMessage(ctx context.Context, subID uuid.UUID, userID uuid.UUID, content string, isAdmin bool) (*models.Message, error) {
	// Validate message content (DRY: business rule)
	if err := s.validateMessageContent(content); err != nil {
		return nil, err
	}

	// Validate subscription access (DRY: security check - handles both admin and user)
	if _, err := s.validateSubscriptionAccess(ctx, subID, userID, isAdmin); err != nil {
		return nil, err
	}

	// FIX 3: Save trimmed content to avoid inconsistencies and save space
	trimmedContent := strings.TrimSpace(content)

	msg := &models.Message{
		ID:             uuid.New(),
		SubscriptionID: subID,
		UserID:         userID, // FIX 4: Always store actual user ID (admin or user)
		Content:        trimmedContent, // FIX 3: Save trimmed content
		IsAdmin:        isAdmin,
		CreatedAt:      time.Now(),
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return msg, nil
}

// GetChatHistoryParams defines pagination parameters for chat history
// FIX 1: Added cursor-based pagination support
type GetChatHistoryParams struct {
	Limit      int
	Offset     int     // Legacy offset-based pagination
	CursorID   *string // Cursor-based pagination (message ID)
	CursorTime *time.Time // Cursor timestamp for validation
}

// GetChatHistory retrieves paginated chat history for a subscription
// FIX 1: Supports both cursor-based and offset-based pagination
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

	// FIX 1: Use cursor-based pagination if cursor is provided (preferred for real-time chats)
	if params.CursorID != nil {
		cursorUUID, err := uuid.Parse(*params.CursorID)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor ID: %w", err)
		}
		return s.messageRepo.GetBySubscriptionIDCursor(ctx, subID, params.Limit, cursorUUID)
	}

	// Fallback to offset-based pagination (legacy)
	return s.messageRepo.GetBySubscriptionIDPaginated(ctx, subID, params.Limit, params.Offset)
}
