package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/apierrors"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	userService    *service.UserService
	paymentService *service.PaymentService
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(userService *service.UserService, paymentService *service.PaymentService) *DashboardHandler {
	return &DashboardHandler{
		userService:    userService,
		paymentService: paymentService,
	}
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get user dashboard statistics including balance, spending, and recent payments
// @Tags dashboard
// @Accept json
// @Produce json
// @Success 200 {object} DashboardStats
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /dashboard/stats [get]
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error occurred"})
		}
	}()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Type assertion - userID is string from JWT claims
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Parse UUID at the boundary (Controller responsibility)
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Get user profile (which includes balance)
	user, err := h.userService.GetProfile(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados do usuário"})
		return
	}

	// Get payment statistics
	stats, err := h.paymentService.GetUserPaymentStats(c.Request.Context(), userUUID)
	if err != nil {
		// Log the error but continue with default values
		stats = &models.PaymentStats{
			TotalSpent:          0,
			ProductsPurchased:   0,
			ActiveSubscriptions: 0,
		}
	}

	// Get recent payments (last 5) with sentinel error handling
	recentPayments, err := h.paymentService.GetRecentPayments(c.Request.Context(), userUUID, 5)
	if err != nil {
		// Handle validation errors using Convert() for standardized response
		if errors.Is(err, apierrors.ErrInvalidPaymentLimit) || errors.Is(err, apierrors.ErrPaymentLimitExceeded) {
			c.JSON(http.StatusBadRequest, apierrors.Convert(err))
			return
		}
		// For other errors, continue with empty slice
		recentPayments = []models.PaymentInfo{}
	}

	// Get monthly spending (last 6 months) with sentinel error handling
	monthlySpending, err := h.paymentService.GetMonthlySpending(c.Request.Context(), userUUID, 6)
	if err != nil {
		// Handle validation errors using Convert() for standardized response
		if errors.Is(err, apierrors.ErrInvalidPaymentMonths) || errors.Is(err, apierrors.ErrPaymentMonthsExceeded) {
			c.JSON(http.StatusBadRequest, apierrors.Convert(err))
			return
		}
		// For other errors, continue with empty slice
		monthlySpending = []models.MonthlySpend{}
	}

	// Get next billing summary for active subscriptions
	nextBilling, err := h.paymentService.GetNextBillingSummary(c.Request.Context(), userUUID)
	if err != nil {
		// Log the error but continue with default values
		nextBilling = models.NextBillingSummary{
			Total:  0,
			Dates:  []string{},
		}
	}

	dashboardStats := models.DashboardStats{
		Balance:             user.Balance,
		TotalSpent:          stats.TotalSpent,
		ProductsPurchased:   stats.ProductsPurchased,
		ActiveSubscriptions: stats.ActiveSubscriptions,
		RecentPayments:      recentPayments,
		MonthlySpending:     monthlySpending,
		NextBilling:         nextBilling,
	}

	c.JSON(http.StatusOK, dashboardStats)
}