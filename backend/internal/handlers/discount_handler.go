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

type DiscountHandler struct {
	service *service.DiscountService
}

func NewDiscountHandler(service *service.DiscountService) *DiscountHandler {
	return &DiscountHandler{service: service}
}

func (h *DiscountHandler) ListDiscounts(c *gin.Context) {
	discounts, err := h.service.ListDiscounts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list discounts"})
		return
	}
	c.JSON(http.StatusOK, discounts)
}

func (h *DiscountHandler) GetDiscount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discount ID"})
		return
	}

	discount, err := h.service.GetDiscount(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, apierrors.ErrDiscountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Discount not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get discount"})
		return
	}
	if discount == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Discount not found"})
		return
	}

	c.JSON(http.StatusOK, discount)
}

func (h *DiscountHandler) CreateDiscount(c *gin.Context) {
	var req models.CreateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminIDStr, _ := c.Get("user_id")
	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid admin ID"})
		return
	}

	discount, err := h.service.CreateDiscount(c.Request.Context(), &req, adminID)
	if err != nil {
		switch {
		case errors.Is(err, apierrors.ErrDiscountCodeAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Discount code already exists"})
		case errors.Is(err, apierrors.ErrDiscountInvalidValue):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discount value"})
		case errors.Is(err, apierrors.ErrDiscountInvalidPercentage):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Percentage must be between 0 and 100"})
		case errors.Is(err, apierrors.ErrDiscountNegativeValue):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Discount value cannot be negative"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create discount"})
		}
		return
	}

	c.JSON(http.StatusCreated, discount)
}

func (h *DiscountHandler) UpdateDiscount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discount ID"})
		return
	}

	var req models.UpdateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateDiscount(c.Request.Context(), id, &req)
	if err != nil {
		switch {
		case errors.Is(err, apierrors.ErrDiscountNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Discount not found"})
		case errors.Is(err, apierrors.ErrDiscountCodeAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Discount code already exists"})
		case errors.Is(err, apierrors.ErrDiscountInvalidValue):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discount value"})
		case errors.Is(err, apierrors.ErrDiscountInvalidPercentage):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Percentage must be between 0 and 100"})
		case errors.Is(err, apierrors.ErrDiscountNegativeValue):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Discount value cannot be negative"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update discount"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Discount updated successfully"})
}

func (h *DiscountHandler) DeleteDiscount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discount ID"})
		return
	}

	err = h.service.DeleteDiscount(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, apierrors.ErrDiscountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Discount not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete discount"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Discount deactivated successfully"})
}
