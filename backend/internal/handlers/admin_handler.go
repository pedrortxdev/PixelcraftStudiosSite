package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/apierrors"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
	"github.com/pixelcraft/api/internal/service"
)

type AdminHandler struct {
	service *service.AdminService
	worker  *service.AnalyticsWorker
}

func NewAdminHandler(service *service.AdminService, worker *service.AnalyticsWorker) *AdminHandler {
	return &AdminHandler{service: service, worker: worker}
}

// RefreshStats godoc
// @Summary Manually trigger analytics snapshot refresh
// @Tags admin
// @Produce json
// @Success 200 {object} gin.H
// @Router /admin/stats/refresh [post]
func (h *AdminHandler) RefreshStats(c *gin.Context) {
	// Execute immediately in background or wait? Since it's quick enough, we can wait.
	// But let's use a goroutine if we want to be super safe with timeouts.
	// For now, synchronous call is fine for admin utility.
	h.worker.RefreshNow() 

	c.JSON(http.StatusOK, gin.H{
		"message": "Dashboard statistics refreshed successfully",
		"lastUpdated": time.Now(),
	})
}

// GetStats godoc
// @Summary Get admin dashboard statistics
// @Tags admin
// @Produce json
// @Success 200 {object} repository.AnalyticsSnapshot
// @Router /admin/stats [get]
func (h *AdminHandler) GetStats(c *gin.Context) {

	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		log.Printf("Erro Admin Stats: Falha ao buscar estatísticas - %v", err)
		defaultStats := &repository.AnalyticsSnapshot{
			TotalRevenue:   0,
			TotalUsers:     0,
			ActiveProducts: 0,
			TotalSales:     0,
			RevenueGrowth:  0,
			UsersGrowth:    0,
			ProductsStatus: "",
			SalesGrowth:    0,
			LastUpdated:    time.Now(),
		}
		c.JSON(http.StatusOK, defaultStats)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetRecentOrders godoc
// @Summary Get recent orders
// @Tags admin
// @Produce json
// @Success 200 {array} repository.RecentOrder
// @Router /admin/orders/recent [get]
func (h *AdminHandler) GetRecentOrders(c *gin.Context) {

	orders, err := h.service.GetRecentOrders(c.Request.Context())
	if err != nil {
		log.Printf("Erro Admin Stats: Falha ao buscar ordens recentes - %v", err)
		c.JSON(http.StatusInternalServerError, apierrors.Convert(err))
		return
	}
	c.JSON(http.StatusOK, orders)
}

// GetTopProducts godoc
// @Summary Get top selling products
// @Tags admin
// @Produce json
// @Success 200 {array} repository.TopProduct
// @Router /admin/products/top [get]
func (h *AdminHandler) GetTopProducts(c *gin.Context) {

	products, err := h.service.GetTopProducts(c.Request.Context())
	if err != nil {
		log.Printf("Erro Admin Stats: Falha ao buscar produtos top - %v", err)
		c.JSON(http.StatusInternalServerError, apierrors.Convert(err))
		return
	}
	c.JSON(http.StatusOK, products)
}

// ListTransactions lists transactions with pagination
func (h *AdminHandler) ListTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	} else if limit > 100 {
		limit = 100 // BT-031
	}

	txs, total, err := h.service.ListTransactions(c.Request.Context(), page, limit, status)
	if err != nil {
		log.Printf("Erro Admin Transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  txs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetMPBalance gets the Mercado Pago account balance
func (h *AdminHandler) GetMPBalance(c *gin.Context) {
	balance, err := h.service.GetMercadoPagoBalance(c.Request.Context())
	if err != nil {
		// Check for permission error
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Forbidden") {
			log.Printf("Aviso: Token sem permissão de leitura de saldo. Retornando 0. Erro original: %v", err)
			// Return zeroed balance instead of error
			c.JSON(http.StatusOK, gin.H{
				"total_amount":      0.0,
				"available_amount":  0.0,
				"unavailable_amount": 0.0,
			})
			return
		}
		
		log.Printf("Erro Admin MP Balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Mercado Pago balance", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, balance)
}

// RefundTransaction handles transaction refunds
func (h *AdminHandler) RefundTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID is required"})
		return
	}

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido. Use o UUID da transação interna, não o ID do Mercado Pago"})
		return
	}

	if err := h.service.RefundTransaction(c.Request.Context(), id); err != nil {
		log.Printf("Erro Admin Refund: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refund transaction", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction refunded successfully"})
}

// ListUsers lists users with pagination
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	} else if limit > 100 {
		limit = 100 // BT-031
	}

	users, total, err := h.service.ListUsers(c.Request.Context(), page, limit, search)
	if err != nil {
		log.Printf("Erro Admin ListUsers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	canViewCPF := false
	if permsX, exists := c.Get("user_permissions"); exists {
		if up, ok := permsX.(*models.UserPermissions); ok {
			canViewCPF = up.CanViewCPF()
		}
	}

	if !canViewCPF {
		for i := range users {
			users[i].CPF = nil
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetUserDetail returns full details of a user
func (h *AdminHandler) GetUserDetail(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	detail, err := h.service.GetUserDetail(c.Request.Context(), id)
	if err != nil {
		log.Printf("Erro Admin GetUserDetail: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user details"})
		return
	}

	canViewCPF := false
	if permsX, exists := c.Get("user_permissions"); exists {
		if up, ok := permsX.(*models.UserPermissions); ok {
			canViewCPF = up.CanViewCPF()
		}
	}

	if !canViewCPF && detail.User != nil {
		detail.User.CPF = nil
	}

	c.JSON(http.StatusOK, detail)
}

// AdminUpdateUserRequest defines allowed fields for admin user updates
// This prevents mass assignment vulnerabilities by whitelisting allowed fields
type AdminUpdateUserRequest struct {
	Username       *string  `json:"username"`
	FullName       *string  `json:"full_name"`
	DisplayName    *string  `json:"display_name"`
	Email          *string  `json:"email"`
	Phone          *string  `json:"phone"`
	IsActive       *bool    `json:"is_active"`
	IsAdmin        *bool    `json:"is_admin"`
	Balance        *float64 `json:"balance"`
	AdjustmentType *string  `json:"adjustment_type"` // "Teste" or "Pix Direto"
}

// UpdateUser updates user details
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Build updates map from allowed fields only
	updates := make(map[string]interface{})
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}
	if req.Balance != nil {
		updates["balance"] = *req.Balance
	}
	if req.AdjustmentType != nil {
		updates["adjustment_type"] = *req.AdjustmentType
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	// Extract admin ID
	adminIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	adminID := adminIDStr.(string)

	if err := h.service.UpdateUser(c.Request.Context(), id, adminID, updates); err != nil {
		log.Printf("Erro Admin UpdateUser: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// UpdateUserPassword updates user password
func (h *AdminHandler) UpdateUserPassword(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password required (min 6 chars)"})
		return
	}

	if err := h.service.UpdateUserPassword(c.Request.Context(), id, req.Password); err != nil {
		log.Printf("Erro Admin UpdateUserPassword: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
