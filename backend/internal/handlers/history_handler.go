package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/service"
)

// HistoryHandler lida com requests do histórico do usuário
type HistoryHandler struct {
	historyService *service.HistoryService
}

func NewHistoryHandler(historyService *service.HistoryService) *HistoryHandler {
	return &HistoryHandler{historyService: historyService}
}

// GetMyHistory godoc
// @Summary Get user's subscriptions and purchased products (minimal fields)
// @Tags history
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /history [get]
func (h *HistoryHandler) GetMyHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	resp, err := h.historyService.GetUserHistory(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMyInvoices godoc
// @Summary Get user's invoice history
// @Tags history
// @Produce json
// @Success 200 {object} models.InvoiceHistoryResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /history/invoices [get]
func (h *HistoryHandler) GetMyInvoices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	resp, err := h.historyService.GetUserInvoiceHistory(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}