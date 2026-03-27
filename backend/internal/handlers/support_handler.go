package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/middleware"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

// SupportHandler handles support ticket endpoints
type SupportHandler struct {
	supportService *service.SupportService
	roleService    *service.RoleService
	wsHub          *WSHub
}

// NewSupportHandler creates a new support handler
func NewSupportHandler(supportService *service.SupportService, roleService *service.RoleService, wsHub *WSHub) *SupportHandler {
	return &SupportHandler{
		supportService: supportService,
		roleService:    roleService,
		wsHub:          wsHub,
	}
}

// ==========================================
// CLIENT AREA ENDPOINTS
// ==========================================

// CreateTicket creates a new support ticket
// POST /api/v1/support/tickets
func (h *SupportHandler) CreateTicket(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.supportService.CreateTicket(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

// ListMyTickets returns all tickets for the current user
// GET /api/v1/support/tickets
func (h *SupportHandler) ListMyTickets(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100 // BT-031
	}

	tickets, err := h.supportService.GetUserTickets(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

// GetTicket returns details of a specific ticket
// GET /api/v1/support/tickets/:id
func (h *SupportHandler) GetTicket(c *gin.Context) {
	userID := c.GetString("user_id")
	ticketID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	isStaff := middleware.IsStaffInContext(c)

	ticket, err := h.supportService.GetTicket(c.Request.Context(), ticketID, userID, isStaff)
	if err != nil {
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "unauthorized: you do not own this ticket" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enrich with user roles for badge display
	if ticket.User != nil {
		roles, _ := h.roleService.GetUserRoles(c.Request.Context(), ticket.UserID)
		ticket.User.Roles = roles
		ticket.User.HighestRole = models.GetHighestRole(roles)
	}
	
	// Enrich messages with sender roles
	for i := range ticket.Messages {
		if ticket.Messages[i].SenderID != "" {
			roles, _ := h.roleService.GetUserRoles(c.Request.Context(), ticket.Messages[i].SenderID)
			if ticket.Messages[i].Sender != nil {
				ticket.Messages[i].Sender.Roles = roles
				ticket.Messages[i].Sender.HighestRole = models.GetHighestRole(roles)
			}
		}
	}

	c.JSON(http.StatusOK, ticket)
}

// SendMessage sends a message in a ticket
// POST /api/v1/support/tickets/:id/messages
func (h *SupportHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	ticketID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateSupportMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isStaff := middleware.IsStaffInContext(c)

	msg, err := h.supportService.SendMessage(c.Request.Context(), ticketID, userID, req.Content, isStaff)
	if err != nil {
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "unauthorized: you do not own this ticket" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enrich sender with roles
	roles, _ := h.roleService.GetUserRoles(c.Request.Context(), userID)
	// We need to fetch sender info to return full object
	// For now, simpler to just set Sender struct if service returned it, 
	// but service returns message without enriched Sender object usually. 
	// Just return ID is fine, but frontend expects Sender object for avatar/role.
	// Actually service doesn't populate Sender. We should populate it here.
	// OR frontend refetches. 
	// Let's populate minimal sender info.
	msg.Sender = &models.User{
		ID: userID,
		Roles: roles,
		HighestRole: models.GetHighestRole(roles),
		// Username/Avatar would need DB fetch but `roleService` doesn't have `GetUser` usually exposed easily here 
		// without `userService`. `SupportService` calls repo. 
		// Actually `supportRepo.CreateMessage` returns minimal info.
		// Front end usually needs to refresh or we just return ID.
		// AdminSupport.jsx uses `selectedTicket.messages` array which generally has enriched format.
	}
	// To be safe and show avatar immediately, we might need to fetch user.
	// But `h.roleService` doesn't have `GetUser`. `h` does not have `userService`.
	// However, `support_repository`'s `GetMessages` does join.
	// For `SendMessage`, we just created it.
	
	// Broadcast update
	if h.wsHub != nil {
		h.wsHub.BroadcastToRoom(ticketID, gin.H{
			"type":    "new_message",
			"message": msg,
		})
	}

	c.JSON(http.StatusCreated, msg)
}

// CloseTicket closes a ticket
// PUT /api/v1/support/tickets/:id/close
func (h *SupportHandler) CloseTicket(c *gin.Context) {
	userID := c.GetString("user_id")
	ticketID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	isStaff := middleware.IsStaffInContext(c)

	if err := h.supportService.CloseTicket(c.Request.Context(), ticketID, userID, isStaff); err != nil {
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "unauthorized: you do not own this ticket" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket closed successfully"})
}

// ==========================================
// ADMIN AREA ENDPOINTS
// ==========================================

// ListAllTickets returns all tickets for admin view
// GET /api/v1/admin/support/tickets
func (h *SupportHandler) ListAllTickets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100 // BT-031
	}

	filter := models.TicketListFilter{
		Page:  page,
		Limit: limit,
	}

	// Apply status filter
	if status := c.Query("status"); status != "" {
		s := models.TicketStatus(status)
		filter.Status = &s
	}

	// Apply category filter
	if category := c.Query("category"); category != "" {
		cat := models.TicketCategory(category)
		filter.Category = &cat
	}

	// Apply assigned filter
	if assigned := c.Query("assigned_to"); assigned != "" {
		filter.AssignedTo = &assigned
	}

	// Apply priority filter
	if priorityStr := c.Query("priority"); priorityStr != "" {
		if p, err := strconv.ParseFloat(priorityStr, 64); err == nil {
			filter.Priority = &p
		}
	}

	tickets, err := h.supportService.ListTickets(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enrich with user roles for badge display
	for i := range tickets.Tickets {
		if tickets.Tickets[i].User != nil {
			roles, _ := h.roleService.GetUserRoles(c.Request.Context(), tickets.Tickets[i].UserID)
			tickets.Tickets[i].User.Roles = roles
			tickets.Tickets[i].User.HighestRole = models.GetHighestRole(roles)
		}
	}

	c.JSON(http.StatusOK, tickets)
}

// AssignTicket assigns a ticket to a staff member
// PUT /api/v1/admin/support/tickets/:id/assign
func (h *SupportHandler) AssignTicket(c *gin.Context) {
	ticketID := c.Param("id")

	var req models.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.supportService.AssignTicket(c.Request.Context(), ticketID, req.AssignedTo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast update
	if h.wsHub != nil {
		h.wsHub.BroadcastToRoom(ticketID, gin.H{
			"type": "ticket_updated",
			"data": gin.H{"assigned_to": req.AssignedTo, "status": "IN_PROGRESS"},
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket assigned successfully"})
}

// ReleaseTicket releases a ticket (unassigns staff)
// PUT /api/v1/admin/support/tickets/:id/release
func (h *SupportHandler) ReleaseTicket(c *gin.Context) {
	ticketID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.supportService.ReleaseTicket(c.Request.Context(), ticketID, userID); err != nil {
		if err.Error() == "unauthorized: you are not assigned to this ticket" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast update
	if h.wsHub != nil {
		h.wsHub.BroadcastToRoom(ticketID, gin.H{
			"type": "ticket_updated",
			"data": gin.H{"assigned_to": nil}, // Explicitly null
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket released successfully"})
}

// UpdateStatus updates the status of a ticket
// PUT /api/v1/admin/support/tickets/:id/status
func (h *SupportHandler) UpdateStatus(c *gin.Context) {
	ticketID := c.Param("id")

	var req models.UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.supportService.UpdateTicketStatus(c.Request.Context(), ticketID, req.Status); err != nil {
		// Log the actual error for debugging
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "details": "Database error updating status"})
		return
	}

	// Broadcast update
	if h.wsHub != nil {
		h.wsHub.BroadcastToRoom(ticketID, gin.H{
			"type": "ticket_updated",
			"data": gin.H{"status": req.Status},
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// GetTicketStats returns ticket statistics for dashboard
// GET /api/v1/admin/support/stats
func (h *SupportHandler) GetTicketStats(c *gin.Context) {
	open, inProgress, resolved, err := h.supportService.GetTicketStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"open":        open,
		"in_progress": inProgress,
		"resolved":    resolved,
	})
}
