package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// LibraryService contains business logic for user library and downloads
type LibraryService struct {
	libraryRepo *repository.LibraryRepository
	productRepo *repository.ProductRepository
	fileService *FileService
	apiBaseURL  string // Base URL for public API endpoints (configurable)
}

func NewLibraryService(
	libraryRepo *repository.LibraryRepository,
	productRepo *repository.ProductRepository,
	fileService *FileService,
	apiBaseURL string,
) *LibraryService {
	return &LibraryService{
		libraryRepo: libraryRepo,
		productRepo: productRepo,
		fileService: fileService,
		apiBaseURL:  apiBaseURL,
	}
}

// LibraryItem represents a typed library item (no more interface{})
type LibraryItem struct {
	models.UserPurchaseWithProduct
	DownloadToken     *string    `json:"download_token,omitempty"`      // Ephemeral token for OneTimeDownload
	DownloadExpiresAt *time.Time `json:"download_expires_at,omitempty"` // Token expiration
}

// GetUserLibrary returns purchased items for a user (PROPERLY TYPED)
func (s *LibraryService) GetUserLibrary(ctx context.Context, userID uuid.UUID) ([]LibraryItem, error) {
	items, err := s.libraryRepo.GetUserLibrary(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert to LibraryItem with proper type
	result := make([]LibraryItem, len(items))
	for i, item := range items {
		result[i] = LibraryItem{
			UserPurchaseWithProduct: item,
			// DownloadToken is generated on-demand in GetDownloadInfo
		}
	}

	return result, nil
}

// GetDownloadInfo verifies ownership and generates ONE-TIME download token
// This prevents link sharing/piracy - tokens expire and can only be used once
func (s *LibraryService) GetDownloadInfo(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*DownloadInfo, error) {
	// 1. Verify user owns the product
	owns, err := s.libraryRepo.UserOwnsProduct(ctx, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !owns {
		return nil, fmt.Errorf("user does not own this product")
	}

	// 2. Get product details
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// 3. Generate ONE-TIME download token (prevents link sharing)
	token, err := s.fileService.GenerateOneTimeDownloadToken(ctx, productID, userID, 15, 1) // 15 min, 1 download
	if err != nil {
		return nil, fmt.Errorf("failed to generate download token: %w", err)
	}

	// 4. Return ephemeral URL (not static!) using configurable base URL
	return &DownloadInfo{
		ProductID:       productID,
		ProductName:     product.Name,
		DownloadURL:     fmt.Sprintf("%s/api/v1/files/onedownload/%s", s.apiBaseURL, token.Token.String()),
		Token:           token.Token,
		ExpiresAt:       token.ExpiresAt,
		MaxDownloads:    token.MaxDownloads,
		RemainingUses:   token.MaxDownloads,
		IsOneTime:       true,
	}, nil
}

// DownloadInfo contains ephemeral download information (not static URLs)
type DownloadInfo struct {
	ProductID     uuid.UUID  `json:"product_id"`
	ProductName   string     `json:"product_name"`
	DownloadURL   string     `json:"download_url"` // One-time token URL
	Token         uuid.UUID  `json:"token"`
	ExpiresAt     time.Time  `json:"expires_at"`
	MaxDownloads  int        `json:"max_downloads"`
	RemainingUses int        `json:"remaining_uses"`
	IsOneTime     bool       `json:"is_one_time"`
}

// ValidateDownloadToken validates a one-time download token and returns file path
func (s *LibraryService) ValidateDownloadToken(ctx context.Context, token uuid.UUID, ipAddress, userAgent string) (*models.File, error) {
	// This would call fileService to validate and use the token
	// The token is consumed (decremented) on each use
	fileID, valid, reason, err := s.fileService.ValidateOneTimeDownloadToken(ctx, token, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid or expired token: %s", reason)
	}

	// Get file details
	file, err := s.fileService.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file, nil
}

// UserOwnsProduct checks if user purchased a specific product
func (s *LibraryService) UserOwnsProduct(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (bool, error) {
	return s.libraryRepo.UserOwnsProduct(ctx, userID, productID)
}
