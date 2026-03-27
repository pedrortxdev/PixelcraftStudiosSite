package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// LibraryService contains business logic for user library and downloads
type LibraryService struct {
	libraryRepo    *repository.LibraryRepository
	productService *ProductService
}

func NewLibraryService(libraryRepo *repository.LibraryRepository, productService *ProductService) *LibraryService {
	return &LibraryService{libraryRepo: libraryRepo, productService: productService}
}

// GetUserLibrary returns purchased items for a user
func (s *LibraryService) GetUserLibrary(ctx context.Context, userID string) (interface{}, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	items, err := s.libraryRepo.GetUserLibrary(ctx, uid)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetDownloadInfo verifies ownership and returns download information
func (s *LibraryService) GetDownloadInfo(ctx context.Context, userID string, productIDStr string) (string, bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", false, fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate the product ID format before processing
	pid, err := uuid.Parse(productIDStr)
	if err != nil {
		return "", false, fmt.Errorf("invalid product ID format: %w", err)
	}

	owns, err := s.libraryRepo.UserOwnsProduct(ctx, uid, pid)
	if err != nil {
		return "", false, err
	}
	if !owns {
		return "", false, fmt.Errorf("user does not own this product")
	}

	url, isFile, err := s.productService.GetDownloadInfo(ctx, pid)
	if err != nil {
		return "", false, err
	}
	return url, isFile, nil
}

func (s *LibraryService) GetProduct(ctx context.Context, productID uuid.UUID) (*models.Product, error) {
	return s.productService.GetProduct(ctx, productID)
}	