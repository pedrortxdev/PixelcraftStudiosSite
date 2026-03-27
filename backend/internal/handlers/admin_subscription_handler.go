package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

// AdminSubscriptionHandler handles admin requests for subscriptions
type AdminSubscriptionHandler struct {
	service        *service.SubscriptionService
	messageService *service.MessageService
}

// NewAdminSubscriptionHandler creates a new AdminSubscriptionHandler
func NewAdminSubscriptionHandler(service *service.SubscriptionService, messageService *service.MessageService) *AdminSubscriptionHandler {
	return &AdminSubscriptionHandler{
		service:        service,
		messageService: messageService,
	}
}

// GetSubscriptionDetails handles GET /api/v1/admin/subscriptions/:id
func (h *AdminSubscriptionHandler) GetSubscriptionDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	sub, logs, err := h.service.GetSubscriptionDetails(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscription details"})
		return
	}
	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": sub,
		"logs":         logs,
	})
}

// UpdateSubscription handles PUT /api/v1/admin/subscriptions/:id
func (h *AdminSubscriptionHandler) UpdateSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateSubscription(c.Request.Context(), id, req)
	if err != nil {
		fmt.Printf("Error updating subscription %s: %v\n", idStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription updated successfully"})
}

// CreateSubscriptionLog handles POST /api/v1/admin/subscriptions/:id/logs
func (h *AdminSubscriptionHandler) CreateSubscriptionLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var req models.AddProjectLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get admin ID from context (optional, if we want to track who added it)
	// adminID, _ := c.Get("user_id")
	// For now pass nil or implement retrieval if needed

	err = h.service.AddProjectLog(c.Request.Context(), id, req.Message, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add log"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Log added successfully"})
}

// GetActiveSubscriptions handles GET /api/v1/admin/subscriptions/active
func (h *AdminSubscriptionHandler) GetActiveSubscriptions(c *gin.Context) {
	subscriptions, err := h.service.GetActiveSubscriptions(c.Request.Context())
	if err != nil {
		fmt.Printf("Error fetching active subscriptions: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch active subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// GetSubscriptionChat handles GET /api/v1/admin/subscriptions/:id/chat
func (h *AdminSubscriptionHandler) GetSubscriptionChat(c *gin.Context) {
	subIDStr := c.Param("id")
	adminIDStr := c.GetString("user_id")

	// Parse UUIDs at the boundary
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID format"})
		return
	}

	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin ID format"})
		return
	}

	// Pass isAdmin = true to bypass ownership check
	messages, err := h.messageService.GetChatHistory(c.Request.Context(), subID, adminID, true, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat history"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// SendSubscriptionMessage handles POST /api/v1/admin/subscriptions/:id/chat
func (h *AdminSubscriptionHandler) SendSubscriptionMessage(c *gin.Context) {
	subIDStr := c.Param("id")
	adminIDStr := c.GetString("user_id")

	// Parse UUIDs at the boundary
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID format"})
		return
	}

	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin ID format"})
		return
	}

	var req models.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pass isAdmin = true to indicate this is an admin message
	msg, err := h.messageService.SendMessage(c.Request.Context(), subID, adminID, req.Content, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

// CreatePlan handles POST /api/v1/admin/plans
func (h *AdminSubscriptionHandler) CreatePlan(c *gin.Context) {
	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if plan.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than zero"})
		return
	}
	if plan.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if err := h.service.CreatePlan(c.Request.Context(), &plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Plan created successfully", "id": plan.ID})
}

// UpdatePlan handles PUT /api/v1/admin/plans/:id
func (h *AdminSubscriptionHandler) UpdatePlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan.ID = id

	if plan.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than zero"})
		return
	}
	if plan.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	if err := h.service.UpdatePlan(c.Request.Context(), &plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan updated successfully"})
}

// DeletePlan handles DELETE /api/v1/admin/plans/:id
func (h *AdminSubscriptionHandler) DeletePlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		return
	}

	if err := h.service.DeletePlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plan deleted successfully"})
}
