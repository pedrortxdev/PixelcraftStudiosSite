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
	"math"

	"github.com/pixelcraft/api/internal/config"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"

	"github.com/google/uuid"
)

// ProductService handles business logic for products
type ProductService struct {
	repo              *repository.ProductRepository
	fileService       *FileService  // Need to add file service to handle file downloads
	encryptionKey     string
}

// NewProductService creates a new ProductService
func NewProductService(db *sql.DB, cfg *config.Config, fileService *FileService) *ProductService {
	return &ProductService{
		repo:              repository.NewProductRepository(db),
		fileService:       fileService,
		encryptionKey:     cfg.FileEncryptionKey,
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
	
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	
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
	var price float64
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

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, req *models.UpdateProductRequest) (*models.Product, error) {
	// Get existing product
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Update fields if provided
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Type != nil {
		product.Type = *req.Type
	}
	if req.GameID != nil {
		product.GameID = req.GameID
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.DownloadURL != nil {
		// Re-encrypt the download URL
		encryptedURL, err := encryptData(*req.DownloadURL, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt download URL: %w", err)
		}
		product.DownloadURLEncrypted = encryptedURL
		// Clear file ID if download URL is provided
		product.FileID = nil
	}
	if req.FileID != nil {
		// Set file ID reference (and clear download URL)
		product.FileID = req.FileID
		product.DownloadURLEncrypted = nil
	}
	if req.IsExclusive != nil {
		product.IsExclusive = *req.IsExclusive
	}
	if req.StockQuantity != nil {
		product.StockQuantity = req.StockQuantity
	}
	if req.ImageURL != nil {
		product.ImageURL = req.ImageURL
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, id, product)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// DeleteProduct soft deletes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// GetDownloadInfo returns download information for a product (URL or file path)
// This should only be called after verifying the user owns the product
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
		// Get the file details
		file, err := s.fileService.GetFileForDownload(ctx, *product.FileID)
		if err != nil {
			return "", false, fmt.Errorf("failed to get file: %w", err)
		}
		if file == nil {
			return "", false, fmt.Errorf("file not found")
		}

		// Return the file path and indicate it's a file download (not URL)
		return s.fileService.GetFilePath(file.ID, file.FileName), true, nil
	}

	// It's a URL-based download
	downloadURL, err := decryptData(product.DownloadURLEncrypted, s.encryptionKey)
	if err != nil {
		return "", false, fmt.Errorf("failed to decrypt download URL: %w", err)
	}

	return downloadURL, false, nil
}

// encryptData encrypts data using AES-GCM
func encryptData(data, key string) ([]byte, error) {
	// Decode the hex key
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		// If key is not hex, use it directly (padded/truncated to 32 bytes for AES-256)
		keyBytes = []byte(key)
		if len(keyBytes) < 32 {
			// Pad to 32 bytes
			padded := make([]byte, 32)
			copy(padded, keyBytes)
			keyBytes = padded
		} else if len(keyBytes) > 32 {
			keyBytes = keyBytes[:32]
		}
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

// decryptData decrypts data using AES-GCM
func decryptData(encrypted []byte, key string) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}

	// Decode the hex key
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		// If key is not hex, use it directly
		keyBytes = []byte(key)
		if len(keyBytes) < 32 {
			padded := make([]byte, 32)
			copy(padded, keyBytes)
			keyBytes = padded
		} else if len(keyBytes) > 32 {
			keyBytes = keyBytes[:32]
		}
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
