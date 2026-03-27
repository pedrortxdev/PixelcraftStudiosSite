package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

type EmailManagementHandler struct {
	emailService *service.EmailService
	permService  *service.PermissionService
}

func NewEmailManagementHandler(emailService *service.EmailService, permService *service.PermissionService) *EmailManagementHandler {
	return &EmailManagementHandler{
		emailService: emailService,
		permService:  permService,
	}
}

// SendEmail envia um email e registra no log
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

	// Enviar email
	if err := h.emailService.SendEmail(c.Request.Context(), req.To, req.Subject, req.Body); err != nil {
		// Registrar falha no log
		h.permService.LogEmail(&models.EmailLog{
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

	// Registrar sucesso no log
	h.permService.LogEmail(&models.EmailLog{
		FromEmail: h.emailService.GetFromEmail(),
		ToEmail:   req.To,
		Subject:   req.Subject,
		Body:      req.Body,
		Status:    "sent",
		SentBy:    &userIDStr,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}

// GetEmailLogs retorna o histórico de emails
func (h *EmailManagementHandler) GetEmailLogs(c *gin.Context) {
	// Paginação
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100 // BT-031
	}

	// Filtros
	filters := map[string]string{
		"from":    c.Query("from"),
		"to":      c.Query("to"),
		"status":  c.Query("status"),
		"sent_by": c.Query("sent_by"),
	}

	logs, total, err := h.permService.GetEmailLogs(page, limit, filters)
	if err != nil {
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

// GetEmailLog retorna um email específico
func (h *EmailManagementHandler) GetEmailLog(c *gin.Context) {
	id := c.Param("id")

	log, err := h.permService.GetEmailLogByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// ResendEmail reenvia um email do histórico
func (h *EmailManagementHandler) ResendEmail(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	id := c.Param("id")

	// Buscar email original
	originalLog, err := h.permService.GetEmailLogByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
		return
	}

	// Reenviar
	if err := h.emailService.SendEmail(c.Request.Context(), originalLog.ToEmail, originalLog.Subject, originalLog.Body); err != nil {
		// Registrar falha
		h.permService.LogEmail(&models.EmailLog{
			FromEmail:    h.emailService.GetFromEmail(),
			ToEmail:      originalLog.ToEmail,
			Subject:      fmt.Sprintf("[RESEND] %s", originalLog.Subject),
			Body:         originalLog.Body,
			Status:       "failed",
			ErrorMessage: stringPtr(err.Error()),
			SentBy:       &userIDStr,
			Metadata: map[string]interface{}{
				"resend_of": originalLog.ID,
			},
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend email", "details": err.Error()})
		return
	}

	// Registrar sucesso
	h.permService.LogEmail(&models.EmailLog{
		FromEmail: h.emailService.GetFromEmail(),
		ToEmail:   originalLog.ToEmail,
		Subject:   fmt.Sprintf("[RESEND] %s", originalLog.Subject),
		Body:      originalLog.Body,
		Status:    "sent",
		SentBy:    &userIDStr,
		Metadata: map[string]interface{}{
			"resend_of": originalLog.ID,
		},
	})

	c.JSON(http.StatusOK, gin.H{"message": "Email resent successfully"})
}

// GetEmailStats retorna estatísticas de emails
func (h *EmailManagementHandler) GetEmailStats(c *gin.Context) {
	// Aqui você pode adicionar queries para estatísticas
	// Por enquanto, vou retornar um placeholder
	c.JSON(http.StatusOK, gin.H{
		"total_sent":   0,
		"total_failed": 0,
		"today_sent":   0,
		"today_failed": 0,
	})
}

func stringPtr(s string) *string {
	return &s
}
