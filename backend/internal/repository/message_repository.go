package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *models.Message) error {
	query := `
		INSERT INTO messages (id, subscription_id, user_id, content, is_admin, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID, msg.SubscriptionID, msg.UserID, msg.Content, msg.IsAdmin, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

func (r *MessageRepository) GetBySubscriptionID(ctx context.Context, subID uuid.UUID) ([]models.Message, error) {
	query := `
		SELECT id, subscription_id, user_id, content, is_admin, created_at
		FROM messages
		WHERE subscription_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, subID)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		err := rows.Scan(
			&m.ID, &m.SubscriptionID, &m.UserID, &m.Content, &m.IsAdmin, &m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}

	if messages == nil {
		messages = []models.Message{}
	}
	return messages, nil
}

// GetBySubscriptionIDPaginated retrieves paginated messages for a subscription
// Uses offset-based pagination (legacy - can cause inconsistencies in real-time chats)
func (r *MessageRepository) GetBySubscriptionIDPaginated(ctx context.Context, subID uuid.UUID, limit, offset int) ([]models.Message, error) {
	query := `
		SELECT id, subscription_id, user_id, content, is_admin, created_at
		FROM messages
		WHERE subscription_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, subID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		err := rows.Scan(
			&m.ID, &m.SubscriptionID, &m.UserID, &m.Content, &m.IsAdmin, &m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}

	if messages == nil {
		messages = []models.Message{}
	}
	return messages, nil
}

// GetBySubscriptionIDCursor retrieves messages using cursor-based pagination
// FIX 1: Cursor-based pagination prevents inconsistencies in real-time chats
// Returns messages OLDER than the cursor (for "load more" functionality)
func (r *MessageRepository) GetBySubscriptionIDCursor(ctx context.Context, subID uuid.UUID, limit int, cursorID uuid.UUID) ([]models.Message, error) {
	query := `
		SELECT id, subscription_id, user_id, content, is_admin, created_at
		FROM messages
		WHERE subscription_id = $1 AND id < $2
		ORDER BY created_at DESC
		LIMIT $3
	`
	rows, err := r.db.QueryContext(ctx, query, subID, cursorID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		err := rows.Scan(
			&m.ID, &m.SubscriptionID, &m.UserID, &m.Content, &m.IsAdmin, &m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}

	// Reverse to maintain chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	if messages == nil {
		messages = []models.Message{}
	}
	return messages, nil
}

// GetSubscriptionOwnerAndStatus returns only the owner ID and status for a subscription
// OPTIMIZATION: Lightweight query instead of fetching full subscription object
func (r *MessageRepository) GetSubscriptionOwnerAndStatus(ctx context.Context, subID uuid.UUID) (uuid.UUID, models.SubscriptionStatus, error) {
	query := `
		SELECT user_id, status
		FROM subscriptions
		WHERE id = $1
	`

	var userID uuid.UUID
	var status models.SubscriptionStatus

	err := r.db.QueryRowContext(ctx, query, subID).Scan(&userID, &status)
	if err == sql.ErrNoRows {
		return uuid.Nil, "", nil
	}
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to query subscription: %w", err)
	}

	return userID, status, nil
}
