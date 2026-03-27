package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	service *service.ProductService
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// ListProducts godoc
// @Summary List all products
// @Description Get a paginated list of all active products
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param type query string false "Filter by product type" Enums(PLUGIN, MOD, MAP, TEXTUREPACK, SERVER_TEMPLATE)
// @Success 200 {object} models.ProductListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	var productType *models.ProductType
	if typeParam := c.Query("type"); typeParam != "" {
		pt := models.ProductType(typeParam)
		productType = &pt
	}
	
	var gameID *uuid.UUID
	if gameParam := c.Query("game_id"); gameParam != "" {
		if id, err := uuid.Parse(gameParam); err == nil {
			gameID = &id
		}
	}
	
	var categoryID *uuid.UUID
	if catParam := c.Query("category_id"); catParam != "" {
		if id, err := uuid.Parse(catParam); err == nil {
			categoryID = &id
		}
	}
	
	// Get products
	response, err := h.service.ListProducts(c.Request.Context(), page, pageSize, productType, gameID, categoryID)
	if err != nil {
		log.Printf("Error listing products: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Get detailed information about a specific product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} models.Product
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	// Parse product ID
	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	
	// Get product
	product, err := h.service.GetProduct(c.Request.Context(), productID)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product"})
		return
	}
	
	c.JSON(http.StatusOK, product)
}

// CreateProduct godoc
// @Summary Create a new product (Admin only)
// @Description Create a new product in the catalog
// @Tags products
// @Accept json
// @Produce json
// @Param product body models.CreateProductRequest true "Product data"
// @Success 201 {object} models.Product
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create product
	product, err := h.service.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}
	
	c.JSON(http.StatusCreated, product)
}

// UpdateProduct godoc
// @Summary Update a product (Admin only)
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param product body models.UpdateProductRequest true "Product data"
// @Success 200 {object} models.Product
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// Parse product ID
	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	
	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Update product
	product, err := h.service.UpdateProduct(c.Request.Context(), productID, &req)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}
	
	c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary Delete a product (Admin only)
// @Description Soft delete a product (sets is_active to false)
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	// Parse product ID
	idParam := c.Param("id")
	productID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	
	// Delete product
	err = h.service.DeleteProduct(c.Request.Context(), productID)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
