package service

import (
	"context"
	"errors"
	"log"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// SupportService handles support ticket business logic
type SupportService struct {
	supportRepo *repository.SupportRepository
	roleService *RoleService
}

// NewSupportService creates a new support service
func NewSupportService(supportRepo *repository.SupportRepository, roleService *RoleService) *SupportService {
	return &SupportService{
		supportRepo: supportRepo,
		roleService: roleService,
	}
}

// CreateTicket creates a new support ticket with priority based on user role
func (s *SupportService) CreateTicket(ctx context.Context, userID string, req models.CreateTicketRequest) (*models.SupportTicket, error) {
	// Get user's support priority based on their roles and ticket category
	priority, err := s.roleService.GetSupportPriority(ctx, userID, req.Category)
	if err != nil {
		priority = 1.0 // Default priority if error
	}
	
	ticket := &models.SupportTicket{
		UserID:         userID,
		Subject:        req.Subject,
		Category:       req.Category,
		Priority:       priority,
		SubscriptionID: req.SubscriptionID,
	}
	
	// Create the ticket
	if err := s.supportRepo.CreateTicket(ctx, ticket); err != nil {
		return nil, err
	}
	
	// Create the initial message
	msg := &models.SupportMessage{
		TicketID: ticket.ID,
		SenderID: userID,
		Content:  req.Content,
		IsStaff:  false,
	}
	
	if err := s.supportRepo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	
	ticket.Messages = []models.SupportMessage{*msg}
	return ticket, nil
}

// GetTicket returns a ticket by ID (with authorization check)
func (s *SupportService) GetTicket(ctx context.Context, ticketID, userID string, isStaff bool) (*models.SupportTicket, error) {
	ticket, err := s.supportRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}
	
	// Authorization: user can only view their own tickets, staff can view all
	if !isStaff && ticket.UserID != userID {
		return nil, errors.New("unauthorized: you do not own this ticket")
	}
	
	// Load messages
	messages, err := s.supportRepo.GetMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	ticket.Messages = messages
	
	return ticket, nil
}

// GetUserTickets returns all tickets for a user
func (s *SupportService) GetUserTickets(ctx context.Context, userID string, page, limit int) (*models.TicketListResponse, error) {
	tickets, total, err := s.supportRepo.GetTicketsByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}
	
	totalPages := (total + limit - 1) / limit
	
	return &models.TicketListResponse{
		Tickets:    tickets,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// ListTickets returns tickets for admin view with filters
func (s *SupportService) ListTickets(ctx context.Context, filter models.TicketListFilter) (*models.TicketListResponse, error) {
	return s.supportRepo.ListTickets(ctx, filter)
}

// SendMessage sends a message in a ticket
func (s *SupportService) SendMessage(ctx context.Context, ticketID, senderID, content string, isStaff bool) (*models.SupportMessage, error) {
	// Verify ticket exists and user has access
	ticket, err := s.supportRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}
	
	// Authorization: user can only message in their own tickets, staff can message in all
	if !isStaff && ticket.UserID != senderID {
		return nil, errors.New("unauthorized: you do not own this ticket")
	}

	// Staff Claim Logic enforcement
	if isStaff {
		if ticket.AssignedTo == nil {
			return nil, errors.New("access denied: you must claim this ticket before replying")
		}
		if *ticket.AssignedTo != senderID {
			return nil, errors.New("access denied: this ticket is assigned to another staff member")
		}
	}
	
	// Create message
	msg := &models.SupportMessage{
		TicketID: ticketID,
		SenderID: senderID,
		Content:  content,
		IsStaff:  isStaff,
	}
	
	if err := s.supportRepo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	
	// Update ticket status based on who sent the message
	if isStaff {
		// Staff replied, set status to waiting for customer response
		if err := s.supportRepo.UpdateTicketStatus(ctx, ticketID, models.TicketWaitingResponse); err != nil {
			log.Printf("Failed to update ticket status to waiting response: %v", err)
		}
	} else {
		// Customer replied, set status to open if it was waiting
		if ticket.Status == models.TicketWaitingResponse {
			if err := s.supportRepo.UpdateTicketStatus(ctx, ticketID, models.TicketOpen); err != nil {
				log.Printf("Failed to update ticket status to open: %v", err)
			}
		}
	}
	
	return msg, nil
}

// UpdateTicketStatus updates the status of a ticket (staff only)
func (s *SupportService) UpdateTicketStatus(ctx context.Context, ticketID string, status models.TicketStatus) error {
	return s.supportRepo.UpdateTicketStatus(ctx, ticketID, status)
}

// AssignTicket assigns a ticket to a staff member
func (s *SupportService) AssignTicket(ctx context.Context, ticketID, staffID string) error {
	if err := s.supportRepo.AssignTicket(ctx, ticketID, staffID); err != nil {
		return err
	}
	
	// Create system log message
	msg := &models.SupportMessage{
		TicketID: ticketID,
		SenderID: staffID,
		Content:  "🤖 **Sistema**: Atendimento reivindicado.",
		IsStaff:  true,
	}
	return s.supportRepo.CreateMessage(ctx, msg)
}

// ReleaseTicket releases a ticket (unassigns staff)
func (s *SupportService) ReleaseTicket(ctx context.Context, ticketID, staffID string) error {
	ticket, err := s.supportRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}

	// Verify if the staff releasing is actually the one assigned
	if ticket.AssignedTo == nil || *ticket.AssignedTo != staffID {
		return errors.New("unauthorized: you are not assigned to this ticket")
	}

	if err := s.supportRepo.ReleaseTicket(ctx, ticketID); err != nil {
		return err
	}

	// Create system log message
	msg := &models.SupportMessage{
		TicketID: ticketID,
		SenderID: staffID,
		Content:  "🤖 **Sistema**: Atendimento liberado.",
		IsStaff:  true,
	}
	return s.supportRepo.CreateMessage(ctx, msg)
}

// CloseTicket closes a ticket (can be done by owner or staff)
func (s *SupportService) CloseTicket(ctx context.Context, ticketID, userID string, isStaff bool) error {
	ticket, err := s.supportRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}
	
	// Authorization: user can only close their own tickets, staff can close all
	if !isStaff && ticket.UserID != userID {
		return errors.New("unauthorized: you do not own this ticket")
	}
	
	return s.supportRepo.CloseTicket(ctx, ticketID)
}

// GetTicketStats returns ticket statistics for dashboard
func (s *SupportService) GetTicketStats(ctx context.Context) (open, inProgress, resolved int, err error) {
	return s.supportRepo.GetTicketStats(ctx)
}
