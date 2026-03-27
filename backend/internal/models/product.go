package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductType represents the type of digital product
type ProductType string

const (
	ProductTypePlugin         ProductType = "PLUGIN"
	ProductTypeMod            ProductType = "MOD"
	ProductTypeMap            ProductType = "MAP"
	ProductTypeTexturePack    ProductType = "TEXTUREPACK"
	ProductTypeServerTemplate ProductType = "SERVER_TEMPLATE"
)

// Product represents a digital product in the store
type Product struct {
	ID                    uuid.UUID   `db:"id" json:"id"`
	Name                  string      `db:"name" json:"name"`
	Description           *string     `db:"description" json:"description,omitempty"`
	Price                 float64     `db:"price" json:"price"`
	Type                  ProductType `db:"type" json:"type"`
	GameID                *uuid.UUID  `db:"game_id" json:"game_id,omitempty"`
	CategoryID            *uuid.UUID  `db:"category_id" json:"category_id,omitempty"`
	DownloadURLEncrypted  []byte      `db:"download_url_encrypted" json:"-"` // Never expose in API (for external URLs)
	FileID                *uuid.UUID  `db:"file_id" json:"file_id,omitempty"` // Reference to uploaded file
	IsExclusive           bool        `db:"is_exclusive" json:"is_exclusive"`
	StockQuantity         *int        `db:"stock_quantity" json:"stock_quantity,omitempty"` // NULL = unlimited
	ImageURL              *string     `db:"image_url" json:"image_url,omitempty"`
	IsActive              bool        `db:"is_active" json:"is_active"`
	CreatedAt             time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time   `db:"updated_at" json:"updated_at"`
}

// UserPurchase represents a product purchased by a user
type UserPurchase struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UserID        uuid.UUID `db:"user_id" json:"user_id"`
	ProductID     uuid.UUID `db:"product_id" json:"product_id"`
	PurchasePrice float64   `db:"purchase_price" json:"purchase_price"`
	PurchasedAt   time.Time `db:"purchased_at" json:"purchased_at"`
}

// UserPurchaseWithProduct combines purchase with product details
type UserPurchaseWithProduct struct {
	Purchase UserPurchase `json:"purchase"`
	Product  Product      `json:"product"`
}

// CreateProductRequest represents the request to create a new product
type CreateProductRequest struct {
	Name             string      `json:"name" binding:"required"`
	Description      *string     `json:"description"`
	Price            *float64    `json:"price" binding:"required,min=0"`
	Type             ProductType `json:"type" binding:"required"`
	GameID           *uuid.UUID  `json:"game_id"`
	CategoryID       *uuid.UUID  `json:"category_id"`
	DownloadURL      *string     `json:"download_url"`
	FileID           *uuid.UUID  `json:"file_id"`
	IsExclusive      bool        `json:"is_exclusive"`
	StockQuantity    *int        `json:"stock_quantity"`
	ImageURL         *string     `json:"image_url"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Name          *string      `json:"name"`
	Description   *string      `json:"description"`
	Price         *float64     `json:"price" binding:"omitempty,min=0"`
	Type          *ProductType `json:"type"`
	GameID        *uuid.UUID   `json:"game_id"`
	CategoryID    *uuid.UUID   `json:"category_id"`
	DownloadURL   *string      `json:"download_url"`
	FileID        *uuid.UUID   `json:"file_id"`
	IsExclusive   *bool        `json:"is_exclusive"`
	StockQuantity *int         `json:"stock_quantity"`
	ImageURL      *string      `json:"image_url"`
	IsActive      *bool        `json:"is_active"`
}

// ProductListResponse represents a paginated list of products
type ProductListResponse struct {
	Products   []Product `json:"products"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}
