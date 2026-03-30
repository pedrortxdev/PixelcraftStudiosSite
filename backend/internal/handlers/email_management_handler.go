package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/apierrors"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

type EmailManagementHandler struct {
	emailService *service.EmailService
}

func NewEmailManagementHandler(emailService *service.EmailService) *EmailManagementHandler {
	return &EmailManagementHandler{
		emailService: emailService,
	}
}

// SendEmail sends an email and logs it
func (h *EmailManagementHandler) SendEmail(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	var req struct {
		To      string `json:"to" binding:"required,email"`
		Subject string `json:"subject" binding:"required"`
		Body    string `json:"body" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Send email
	if err := h.emailService.SendEmail(c.Request.Context(), req.To, req.Subject, req.Body); err != nil {
		// Log failure
		h.emailService.LogEmail(c.Request.Context(), &models.EmailLog{
			FromEmail:    h.emailService.GetFromEmail(),
			ToEmail:      req.To,
			Subject:      req.Subject,
			Body:         req.Body,
			Status:       "failed",
			ErrorMessage: stringPtr(err.Error()),
			SentBy:       &userIDStr,
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email", "details": err.Error()})
		return
	}

	// Log success
	h.emailService.LogEmail(c.Request.Context(), &models.EmailLog{
		FromEmail: h.emailService.GetFromEmail(),
		ToEmail:   req.To,
		Subject:   req.Subject,
		Body:      req.Body,
		Status:    "sent",
		SentBy:    &userIDStr,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}

// GetEmailLogs returns email logs with proper validation
func (h *EmailManagementHandler) GetEmailLogs(c *gin.Context) {
	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Filters
	filters := map[string]string{
		"from":    c.Query("from"),
		"to":      c.Query("to"),
		"status":  c.Query("status"),
		"sent_by": c.Query("sent_by"),
	}

	logs, total, err := h.emailService.GetEmailLogs(c.Request.Context(), page, limit, filters)
	if err != nil {
		// PROPER ERROR HANDLING: Use errors.Is() for sentinel errors
		if errors.Is(err, apierrors.ErrInvalidPaginationLimit) || errors.Is(err, apierrors.ErrInvalidPaginationPage) {
			c.JSON(http.StatusBadRequest, apierrors.Convert(err))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetEmailLogByID returns a specific email log
func (h *EmailManagementHandler) GetEmailLogByID(c *gin.Context) {
	id := c.Param("id")

	log, err := h.emailService.GetEmailLogByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email log not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// ResendEmail resends an email from history
func (h *EmailManagementHandler) ResendEmail(c *gin.Context) {
	id := c.Param("id")

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	// Get original email
	originalLog, err := h.emailService.GetEmailLogByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
		return
	}

	// Resend
	if err := h.emailService.SendEmail(c.Request.Context(), originalLog.ToEmail, originalLog.Subject, originalLog.Body); err != nil {
		// Log failure
		h.emailService.LogEmail(c.Request.Context(), &models.EmailLog{
			FromEmail:    h.emailService.GetFromEmail(),
			ToEmail:      originalLog.ToEmail,
			Subject:      fmt.Sprintf("[RESEND] %s", originalLog.Subject),
			Body:         originalLog.Body,
			Status:       "failed",
			ErrorMessage: stringPtr(err.Error()),
			SentBy:       &userIDStr,
			Metadata: map[string]interface{}{
				"original_id": id,
			},
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend email", "details": err.Error()})
		return
	}

	// Log success
	h.emailService.LogEmail(c.Request.Context(), &models.EmailLog{
		FromEmail: h.emailService.GetFromEmail(),
		ToEmail:   originalLog.ToEmail,
		Subject:   fmt.Sprintf("[RESEND] %s", originalLog.Subject),
		Body:      originalLog.Body,
		Status:    "sent",
		SentBy:    &userIDStr,
		Metadata: map[string]interface{}{
			"original_id": id,
		},
	})

	c.JSON(http.StatusOK, gin.H{"message": "Email resent successfully"})
}

// GetSMTPConfig returns the current SMTP configuration
func (h *EmailManagementHandler) GetSMTPConfig(c *gin.Context) {
	config := h.emailService.GetSMTPConfig(c.Request.Context())
	c.JSON(http.StatusOK, config)
}

// UpdateSMTPConfig updates the SMTP configuration
func (h *EmailManagementHandler) UpdateSMTPConfig(c *gin.Context) {
	var req service.EmailConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	if err := h.emailService.UpdateSMTPConfig(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SMTP configuration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SMTP configuration updated successfully"})
}

// TestSMTPConnection tests the SMTP connection with provided credentials
func (h *EmailManagementHandler) TestSMTPConnection(c *gin.Context) {
	var req service.EmailConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	if err := h.emailService.TestSMTPConnection(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SMTP connection test failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SMTP connection test successful"})
}

func stringPtr(s string) *string {
	return &s
}
