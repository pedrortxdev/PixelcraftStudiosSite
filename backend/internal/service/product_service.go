package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/pixelcraft/api/internal/config"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"

	"github.com/google/uuid"
)

// ProductService handles business logic for products
type ProductService struct {
	repo          *repository.ProductRepository
	fileService   *FileService
	encryptionKey []byte // Pre-validated 32-byte key for AES-256
}

// NewProductService creates a new ProductService
// IMPORTANT: This function will PANIC if FileEncryptionKey is not exactly 32 bytes (64 hex chars)
// This is intentional - cryptography with weak keys is worse than no cryptography
func NewProductService(db *sql.DB, cfg *config.Config, fileService *FileService) *ProductService {
	// Validate and decode the encryption key
	keyBytes, err := hex.DecodeString(cfg.FileEncryptionKey)
	if err != nil {
		panic(fmt.Sprintf("FileEncryptionKey is not valid hex: %v", err))
	}

	if len(keyBytes) != 32 {
		panic(fmt.Sprintf(
			"FileEncryptionKey must be exactly 32 bytes (64 hex characters), got %d bytes. "+
				"Generate a secure key with: openssl rand -hex 32",
			len(keyBytes),
		))
	}

	return &ProductService{
		repo:          repository.NewProductRepository(db),
		fileService:   fileService,
		encryptionKey: keyBytes,
	}
}

// ListProducts retrieves all active products with pagination
func (s *ProductService) ListProducts(ctx context.Context, page, pageSize int, productType *models.ProductType, gameID *uuid.UUID, categoryID *uuid.UUID) (*models.ProductListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, total, err := s.repo.GetAll(ctx, page, pageSize, productType, gameID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Integer-only ceiling division (no float conversion overhead)
	// Formula: (total + pageSize - 1) / pageSize
	totalPages := (total + pageSize - 1) / pageSize

	return &models.ProductListResponse{
		Products:   products,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetProduct retrieves a single product by ID
func (s *ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	return product, nil
}

// CreateProduct creates a new product with encrypted download URL or file reference
func (s *ProductService) CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	// Since Price is required in CreateProductRequest, it should not be nil
	// but we'll safely dereference it anyway
	var price int64
	if req.Price != nil {
		price = *req.Price
	} else {
		return nil, fmt.Errorf("price is required")
	}

	product := &models.Product{
		Name:          req.Name,
		Description:   req.Description,
		Price:         price,
		Type:          req.Type,
		GameID:        req.GameID,
		CategoryID:    req.CategoryID,
		IsExclusive:   req.IsExclusive,
		StockQuantity: req.StockQuantity,
		ImageURL:      req.ImageURL,
		IsActive:      true,
	}

	// Set download URL or file ID based on the request
	if req.DownloadURL != nil {
		// Encrypt the download URL
		encryptedURL, err := encryptData(*req.DownloadURL, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt download URL: %w", err)
		}
		product.DownloadURLEncrypted = encryptedURL
	} else if req.FileID != nil {
		// Set file ID reference
		product.FileID = req.FileID
	} else {
		return nil, fmt.Errorf("either download_url or file_id must be provided")
	}

	err := s.repo.Create(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// UpdateProduct updates an existing product with partial update semantics
// Only fields that are explicitly set in the request will be updated
// This prevents race conditions where concurrent updates overwrite each other
func (s *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req *models.UpdateProductRequest) (*models.Product, error) {
	// Verify product exists
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Build partial update request
	partialReq := &repository.UpdateProductRequestPartial{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Type:        req.Type,
		GameID:      req.GameID,
		CategoryID:  req.CategoryID,
		IsExclusive: req.IsExclusive,
		StockQuantity: req.StockQuantity,
		ImageURL:    req.ImageURL,
		IsActive:    req.IsActive,
	}

	// Handle DownloadURL - encrypt if provided
	if req.DownloadURL != nil {
		encryptedURL, err := encryptData(*req.DownloadURL, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt download URL: %w", err)
		}
		partialReq.DownloadURLEncrypted = &encryptedURL
		// Explicitly clear FileID when DownloadURL is provided
		partialReq.FileID = &uuid.Nil
	} else if req.FileID != nil {
		// Set file ID and clear download URL
		partialReq.FileID = req.FileID
		// Explicitly clear DownloadURL when FileID is provided
		partialReq.DownloadURLEncrypted = &[]byte{}
	}

	// Use partial update to avoid race conditions
	err = s.repo.UpdatePartial(ctx, id, partialReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Fetch updated product
	updatedProduct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated product: %w", err)
	}

	return updatedProduct, nil
}

// DeleteProduct soft deletes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// GetDownloadInfo returns download information for a product (URL or secure download link)
// This should only be called after verifying the user owns the product
// Returns: downloadURL, isFileDownload, error
func (s *ProductService) GetDownloadInfo(ctx context.Context, productID uuid.UUID) (string, bool, error) {
	product, err := s.repo.GetByID(ctx, productID)
	if err != nil {
		return "", false, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return "", false, fmt.Errorf("product not found")
	}

	// Check if it's a file-based download
	if product.FileID != nil && s.fileService != nil {
		// Generate a secure one-time download token instead of exposing filesystem path
		// This token will be validated by the file download handler
		token, err := s.fileService.GenerateDownloadToken(ctx, *product.FileID)
		if err != nil {
			return "", false, fmt.Errorf("failed to generate download token: %w", err)
		}

		// Return secure download URL (not filesystem path)
		// The handler will validate the token and serve the file
		return fmt.Sprintf("/api/v1/files/download?token=%s", token), true, nil
	}

	// It's a URL-based download
	downloadURL, err := decryptData(product.DownloadURLEncrypted, s.encryptionKey)
	if err != nil {
		return "", false, fmt.Errorf("failed to decrypt download URL: %w", err)
	}

	return downloadURL, false, nil
}

// encryptData encrypts data using AES-GCM with a pre-validated 32-byte key
func encryptData(data string, keyBytes []byte) ([]byte, error) {
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes, got %d", len(keyBytes))
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return ciphertext, nil
}

// decryptData decrypts data using AES-GCM with a pre-validated 32-byte key
func decryptData(encrypted []byte, keyBytes []byte) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}

	if len(keyBytes) != 32 {
		return "", fmt.Errorf("encryption key must be exactly 32 bytes, got %d", len(keyBytes))
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// GenerateSecureKey generates a new 32-byte (256-bit) encryption key
// This should be used to generate the FileEncryptionKey for production
func GenerateSecureKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate secure key: %w", err)
	}
	return hex.EncodeToString(key), nil
}

// ValidateEncryptionKey validates that a key string is a valid 64-character hex string
func ValidateEncryptionKey(key string) error {
	key = strings.TrimSpace(key)
	if len(key) != 64 {
		return fmt.Errorf("key must be exactly 64 hex characters (32 bytes), got %d", len(key))
	}

	_, err := hex.DecodeString(key)
	if err != nil {
		return fmt.Errorf("key must be valid hexadecimal: %w", err)
	}

	return nil
}
