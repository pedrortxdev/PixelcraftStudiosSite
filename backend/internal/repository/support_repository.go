package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/pixelcraft/api/internal/models"
)

// SupportRepository handles support ticket database operations
type SupportRepository struct {
	db *sql.DB
}

// NewSupportRepository creates a new support repository
func NewSupportRepository(db *sql.DB) *SupportRepository {
	return &SupportRepository{db: db}
}

// CreateTicket creates a new support ticket
func (r *SupportRepository) CreateTicket(ctx context.Context, ticket *models.SupportTicket) error {
	query := `
		INSERT INTO support_tickets (user_id, subject, category, priority, subscription_id)
		VALUES ($1, $2, $3::ticket_category, $4, $5)
		RETURNING id, status, created_at, updated_at
	`
	if ticket.Category == "" {
		ticket.Category = models.CategoryGeneral
	}
	return r.db.QueryRowContext(ctx, query,
		ticket.UserID, ticket.Subject, string(ticket.Category), ticket.Priority, ticket.SubscriptionID,
	).Scan(&ticket.ID, &ticket.Status, &ticket.CreatedAt, &ticket.UpdatedAt)
}

// CreateMessage creates a new message in a ticket
func (r *SupportRepository) CreateMessage(ctx context.Context, msg *models.SupportMessage) error {
	query := `
		INSERT INTO support_messages (ticket_id, sender_id, content, is_staff, attachment_url, attachment_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query,
		msg.TicketID, msg.SenderID, msg.Content, msg.IsStaff, msg.AttachmentURL, msg.AttachmentType,
	).Scan(&msg.ID, &msg.CreatedAt)
	
	if err != nil {
		return err
	}
	
	// Update ticket's updated_at
	_, err = r.db.ExecContext(ctx, `UPDATE support_tickets SET updated_at = NOW() WHERE id = $1`, msg.TicketID)
	return err
}

// GetTicketByID returns a ticket by its ID
func (r *SupportRepository) GetTicketByID(ctx context.Context, id string) (*models.SupportTicket, error) {
	query := `
		SELECT t.id, t.user_id, t.subject, t.category, t.priority, t.status, t.assigned_to, 
		       t.subscription_id, t.created_at, t.updated_at, t.resolved_at, t.closed_at,
		       u.username, u.full_name, u.avatar_url,
		       s.username, s.full_name, s.avatar_url
		FROM support_tickets t
		JOIN users u ON t.user_id = u.id
		LEFT JOIN users s ON t.assigned_to = s.id
		WHERE t.id = $1
	`
	var t models.SupportTicket
	var user, staff models.User
	var assignedTo, subID *string
	var resolvedAt, closedAt *sql.NullTime
	var staffUsername, staffFullName, staffAvatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.UserID, &t.Subject, &t.Category, &t.Priority, &t.Status, &assignedTo,
		&subID, &t.CreatedAt, &t.UpdatedAt, &resolvedAt, &closedAt,
		&user.Username, &user.FullName, &user.AvatarURL,
		&staffUsername, &staffFullName, &staffAvatarURL,
	)
	if err != nil {
		return nil, err
	}

	t.AssignedTo = assignedTo
	t.SubscriptionID = subID
	if resolvedAt != nil && resolvedAt.Valid {
		t.ResolvedAt = &resolvedAt.Time
	}
	if closedAt != nil && closedAt.Valid {
		t.ClosedAt = &closedAt.Time
	}

	user.ID = t.UserID
	t.User = &user

	if t.AssignedTo != nil && staffUsername.Valid {
		staff.ID = *t.AssignedTo
		staff.Username = staffUsername.String
		staff.FullName = staffFullName.String
		staff.AvatarURL = staffAvatarURL.String
		t.AssignedStaff = &staff
	}

	return &t, nil
}

// GetTicketsByUserID returns all tickets for a user
func (r *SupportRepository) GetTicketsByUserID(ctx context.Context, userID string, page, limit int) ([]models.SupportTicket, int, error) {
	filter := models.TicketListFilter{
		UserID: &userID,
		Page:   page,
		Limit:  limit,
	}
	resp, err := r.ListTickets(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return resp.Tickets, resp.Total, nil
}


// GetMessages returns all messages for a ticket
func (r *SupportRepository) GetMessages(ctx context.Context, ticketID string) ([]models.SupportMessage, error) {
	query := `
		SELECT sm.id, sm.ticket_id, sm.sender_id, sm.content, sm.is_staff, sm.attachment_url, sm.attachment_type, sm.created_at,
		       u.username, u.full_name, u.avatar_url
		FROM support_messages sm
		JOIN users u ON sm.sender_id = u.id
		WHERE sm.ticket_id = $1
		ORDER BY sm.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.SupportMessage
	for rows.Next() {
		var m models.SupportMessage
		var sender models.User
		if err := rows.Scan(
			&m.ID, &m.TicketID, &m.SenderID, &m.Content, &m.IsStaff, &m.AttachmentURL, &m.AttachmentType, &m.CreatedAt,
			&sender.Username, &sender.FullName, &sender.AvatarURL,
		); err != nil {
			return nil, err
		}
		sender.ID = m.SenderID
		m.Sender = &sender
		messages = append(messages, m)
	}

	return messages, rows.Err()
}

// UpdateTicketStatus updates the status of a ticket
func (r *SupportRepository) UpdateTicketStatus(ctx context.Context, ticketID string, status models.TicketStatus) error {
	query := `
		UPDATE support_tickets 
		SET status = $2::ticket_status,
		    resolved_at = CASE WHEN $2 = 'RESOLVED' THEN NOW() ELSE resolved_at END,
		    closed_at = CASE WHEN $2 = 'CLOSED' THEN NOW() ELSE closed_at END,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, ticketID, string(status))
	return err
}

// AssignTicket assigns a ticket to a staff member
func (r *SupportRepository) AssignTicket(ctx context.Context, ticketID, staffID string) error {
	query := `UPDATE support_tickets SET assigned_to = $2, status = 'IN_PROGRESS'::ticket_status, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, ticketID, staffID)
	return err
}

// CloseTicket closes a ticket
func (r *SupportRepository) CloseTicket(ctx context.Context, ticketID string) error {
	return r.UpdateTicketStatus(ctx, ticketID, models.TicketClosed)
}

// ReleaseTicket releases a ticket (unassigns staff)
func (r *SupportRepository) ReleaseTicket(ctx context.Context, ticketID string) error {
	query := `UPDATE support_tickets SET assigned_to = NULL, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, ticketID)
	return err
}

// GetTicketStats returns ticket statistics for dashboard
func (r *SupportRepository) GetTicketStats(ctx context.Context) (open, inProgress, resolved int, err error) {
	// BT-042: Sincronizar contagem com a listagem (JOIN com users)
	// Isso evita tickets "fantasmas" de usuários deletados ou inconsistentes.
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE t.status = 'OPEN'),
			COUNT(*) FILTER (WHERE t.status = 'IN_PROGRESS'),
			COUNT(*) FILTER (WHERE t.status = 'RESOLVED' AND t.resolved_at > NOW() - INTERVAL '7 days')
		FROM support_tickets t
		JOIN users u ON t.user_id = u.id
	`
	err = r.db.QueryRowContext(ctx, query).Scan(&open, &inProgress, &resolved)
	return
}

// GetRecentTickets returns the N most recent tickets
func (r *SupportRepository) GetRecentTickets(ctx context.Context, limit int) ([]models.SupportTicket, error) {
	filter := models.TicketListFilter{Page: 1, Limit: limit}
	resp, err := r.ListTickets(ctx, filter)
	if err != nil {
		return nil, err
	}
	return resp.Tickets, nil
}
// ListTickets lists tickets with pagination and filtering
func (r *SupportRepository) ListTickets(ctx context.Context, filter models.TicketListFilter) (*models.TicketListResponse, error) {
	query := `
		SELECT t.id, t.user_id, t.subject, t.category, t.priority, t.status, t.assigned_to, 
		       t.subscription_id, t.created_at, t.updated_at, t.resolved_at, t.closed_at,
		       u.username, u.full_name, u.avatar_url,
		       s.username, s.full_name, s.avatar_url
		FROM support_tickets t
		JOIN users u ON t.user_id = u.id
		LEFT JOIN users s ON t.assigned_to = s.id
	`
	
	countQuery := `SELECT COUNT(*) FROM support_tickets t`
	
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("t.status::text = $%d", argIndex))
		args = append(args, string(*filter.Status))
		argIndex++
	}
	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("t.category::text = $%d", argIndex))
		args = append(args, string(*filter.Category))
		argIndex++
	}
	if filter.AssignedTo != nil {
		conditions = append(conditions, fmt.Sprintf("t.assigned_to = $%d", argIndex))
		args = append(args, *filter.AssignedTo)
		argIndex++
	}
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("t.user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.Priority != nil {
		conditions = append(conditions, fmt.Sprintf("t.priority = $%d", argIndex))
		args = append(args, *filter.Priority)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	// Capture count query args before adding limit/offset
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	// Order by updated_at desc
	query += " ORDER BY t.updated_at DESC"

	// Pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
		argIndex++ // Incremented for consistency even if last
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []models.SupportTicket
	for rows.Next() {
		var t models.SupportTicket
		var user, staff models.User
		var assignedTo, subID *string
		var resolvedAt, closedAt *sql.NullTime
		var staffUsername, staffFullName, staffAvatarURL sql.NullString
		
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Subject, &t.Category, &t.Priority, &t.Status, &assignedTo,
			&subID, &t.CreatedAt, &t.UpdatedAt, &resolvedAt, &closedAt,
			&user.Username, &user.FullName, &user.AvatarURL,
			&staffUsername, &staffFullName, &staffAvatarURL,
		); err != nil {
			return nil, err
		}
		
		t.AssignedTo = assignedTo
		t.SubscriptionID = subID
		if resolvedAt != nil && resolvedAt.Valid {
			t.ResolvedAt = &resolvedAt.Time
		}
		if closedAt != nil && closedAt.Valid {
			t.ClosedAt = &closedAt.Time
		}
		
		user.ID = t.UserID
		t.User = &user
		
		if t.AssignedTo != nil && staffUsername.Valid {
			staff.ID = *t.AssignedTo
			staff.Username = staffUsername.String
			staff.FullName = staffFullName.String
			staff.AvatarURL = staffAvatarURL.String
			t.AssignedStaff = &staff
		}
		
		tickets = append(tickets, t)
	}

	totalPages := 0
	if filter.Limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(filter.Limit)))
	}

	return &models.TicketListResponse{
		Tickets:    tickets,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}
