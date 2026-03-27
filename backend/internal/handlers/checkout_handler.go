package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/apierrors"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

// CheckoutHandler handles HTTP requests for checkout
type CheckoutHandler struct {
	service *service.CheckoutService
}

// NewCheckoutHandler creates a new CheckoutHandler
func NewCheckoutHandler(service *service.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{service: service}
}

// ProcessCheckout godoc
// @Summary Process checkout
// @Description Process a checkout with cart items, optional coupon, and balance usage
// @Tags checkout
// @Accept json
// @Produce json
// @Param checkout body models.CheckoutRequest true "Checkout data"
// @Success 200 {object} models.CheckoutResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /checkout [post]
func (h *CheckoutHandler) ProcessCheckout(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.New("Dados de entrada inválidos", "ERR_INVALID_BODY"))
		return
	}

	// BT-044: Apply context timeout to critical checkout operation
	timeoutCtx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	response, err := h.service.ProcessCheckout(timeoutCtx, userID, &req)
	if err != nil {
		log.Printf("[Checkout Error] user=%v: %v", userID, err)
		c.JSON(http.StatusBadRequest, apierrors.New("Falha ao processar checkout", "ERR_CHECKOUT"))
		return
	}

	c.JSON(http.StatusOK, response)
}

// ValidateDiscount godoc
// @Summary Validate discount code
// @Description Validate a discount code and calculate the discount amount
// @Tags checkout
// @Accept json
// @Produce json
// @Param discount body models.ValidateDiscountRequest true "Discount validation data"
// @Success 200 {object} models.ValidateDiscountResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /discounts/validate [post]
func (h *CheckoutHandler) ValidateDiscount(c *gin.Context) {
	var req models.ValidateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.New("Dados de entrada inválidos", "ERR_INVALID_BODY"))
		return
	}

	// Call the service to validate the discount
	response, err := h.service.ValidateDiscount(c.Request.Context(), &req)
	if err != nil {
		log.Printf("[Discount Validation Error] %v", err)
		c.JSON(http.StatusBadRequest, apierrors.New("Falha ao validar cupom", "ERR_DISCOUNT"))
		return
	}

	c.JSON(http.StatusOK, response)
}