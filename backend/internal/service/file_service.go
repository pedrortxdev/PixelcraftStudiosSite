package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// FileService handles business logic for file management
type FileService struct {
	repo         *repository.FileRepository
	db           *sql.DB
	uploadDir    string
	maxFileSize  int64  // Maximum file size in bytes
	allowedTypes []string // Allowed file extensions
}

// NewFileService creates a new FileService
func NewFileService(db *sql.DB, uploadDir string, maxFileSize int64, allowedTypes []string) *FileService {
	// Ensure upload directory exists
	if uploadDir == "" {
		uploadDir = "./uploads" // Default directory
	}

	// Create directory if it doesn't exist
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}

	return &FileService{
		repo:         repository.NewFileRepository(db),
		db:           db,
		uploadDir:    uploadDir,
		maxFileSize:  maxFileSize,
		allowedTypes: allowedTypes,
	}
}

// GetDB returns the database connection
func (s *FileService) GetDB() *sql.DB {
	return s.db
}

// SaveFile handles file upload and saves it to the filesystem
func (s *FileService) SaveFile(ctx context.Context, fileHeader *multipart.FileHeader, userID uuid.UUID, friendlyName string) (*models.File, error) {
	// Validate file size
	if fileHeader.Size > s.maxFileSize {
		maxSizeMB := float64(s.maxFileSize) / (1024 * 1024)
		fileSizeMB := float64(fileHeader.Size) / (1024 * 1024)
		return nil, fmt.Errorf("file size too large: %.2f MB. Maximum allowed: %.2f MB", fileSizeMB, maxSizeMB)
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !s.isAllowedType(ext) {
		allowedTypesStr := strings.Join(s.allowedTypes, ", ")
		return nil, fmt.Errorf("file type not allowed: %s. Allowed types: %s", ext, allowedTypesStr)
	}

	// Generate unique internal filename using UUID
	internalFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.uploadDir, internalFileName)

	// Save file to filesystem
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dest, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create file record in database
	fileType := s.getFileType(ext)
	file := &models.File{
		ID:        uuid.New(),
		Name:      friendlyName,
		FileName:  internalFileName,
		FileType:  fileType,
		FilePath:  filePath,
		Size:      fileHeader.Size,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDeleted: false,
	}

	err = s.repo.Create(ctx, file)
	if err != nil {
		// Clean up the saved file if database operation fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	return file, nil
}

// GetFileByID retrieves a file by ID
func (s *FileService) GetFileByID(ctx context.Context, id uuid.UUID) (*models.File, error) {
	return s.repo.GetByID(ctx, id)
}

// GetFilesByUserID retrieves files by user ID with pagination
func (s *FileService) GetFilesByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.File, int, error) {
	return s.repo.GetByUserID(ctx, userID, page, pageSize)
}

// UpdateFileName updates the friendly name of a file
func (s *FileService) UpdateFileName(ctx context.Context, id uuid.UUID, userID uuid.UUID, newName string) error {
	// First verify the file belongs to the user
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found")
	}
	if file.CreatedBy != userID {
		return fmt.Errorf("unauthorized: file does not belong to user")
	}

	// Update the file
	file.Name = newName
	file.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, id, file)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

// DeleteFile soft deletes a file (removes from DB, keeps in filesystem for now)
func (s *FileService) DeleteFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// First verify the file belongs to the user
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found")
	}
	if file.CreatedBy != userID {
		return fmt.Errorf("unauthorized: file does not belong to user")
	}

	// Soft delete the record from database
	err = s.repo.SoftDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Optionally, you could also delete the actual file from the filesystem
	// For now, we'll keep it in case we need to restore it later
	// os.Remove(file.FilePath) // Uncomment if you want to physically delete the file

	return nil
}

// GetFileForDownload returns file details for download
func (s *FileService) GetFileForDownload(ctx context.Context, id uuid.UUID) (*models.File, error) {
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}
	if file.IsDeleted {
		return nil, fmt.Errorf("file not found")
	}

	return file, nil
}

// GetFilePath returns the full path to a file
func (s *FileService) GetFilePath(fileID uuid.UUID, fileName string) string {
	return filepath.Join(s.uploadDir, fileName)
}

// isAllowedType checks if the file extension is allowed
func (s *FileService) isAllowedType(ext string) bool {
	for _, allowedType := range s.allowedTypes {
		if strings.ToLower(allowedType) == ext {
			return true
		}
	}
	return false
}

// getFileType converts file extension to FileType
func (s *FileService) getFileType(ext string) models.FileType {
	switch strings.ToLower(ext) {
	case ".jar":
		return models.FileTypeJar
	case ".zip":
		return models.FileTypeZip
	case ".exe":
		return models.FileTypeExe
	case ".png":
		return models.FileTypePng
	case ".jpg", ".jpeg":
		return models.FileTypeJpg
	case ".pdf":
		return models.FileTypePdf
	default:
		return models.FileTypeOther
	}
}

// GetAllFiles retrieves all files with pagination and search (admin only)
func (s *FileService) GetAllFiles(ctx context.Context, page, pageSize int, search string) ([]models.FileWithUser, int, error) {
	return s.repo.GetAllWithUsers(ctx, page, pageSize, search)
}

// UpdateFilePermissions updates the access permissions for a file
func (s *FileService) UpdateFilePermissions(ctx context.Context, id uuid.UUID, userID uuid.UUID, req models.UpdateFilePermissionsRequest) (*models.File, error) {
	// First verify the file exists and belongs to the user or user is admin
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}

	// For now, only allow file owner or admin to update permissions
	// Admin check should be done in handler
	if file.CreatedBy != userID {
		return nil, fmt.Errorf("unauthorized: file does not belong to user")
	}

	// Update file permissions
	if req.AccessType != nil {
		file.AccessType = *req.AccessType
	}
	if req.RequiredRole != nil {
		file.RequiredRole = req.RequiredRole
	}
	if req.AllowedRoles != nil {
		// Convert to JSONB
		rolesJSON := []byte("[]")
		if len(req.AllowedRoles) > 0 {
			rolesJSON = []byte("[" + strings.Join(func() []string {
				result := make([]string, len(req.AllowedRoles))
				for i, role := range req.AllowedRoles {
					result[i] = fmt.Sprintf("\"%s\"", role)
				}
				return result
			}(), ",") + "]")
		}
		file.AllowedRoles = rolesJSON
	}
	if req.RequiredProductID != nil {
		file.RequiredProductID = req.RequiredProductID
	}
	if req.AllowedProductIDs != nil {
		// Convert to JSONB
		productIDsJSON := []byte("[]")
		if len(req.AllowedProductIDs) > 0 {
			productIDsJSON = []byte("[" + strings.Join(func() []string {
				result := make([]string, len(req.AllowedProductIDs))
				for i, pid := range req.AllowedProductIDs {
					result[i] = fmt.Sprintf("\"%s\"", pid.String())
				}
				return result
			}(), ",") + "]")
		}
		file.AllowedProductIDs = productIDsJSON
	}
	if req.PublicLinkExpiresAt != nil {
		file.PublicLinkExpiresAt = req.PublicLinkExpiresAt
	}
	if req.MaxDownloads != nil {
		file.MaxDownloads = req.MaxDownloads
	}

	file.UpdatedAt = time.Now()

	err = s.repo.UpdatePermissions(ctx, id, file)
	if err != nil {
		return nil, fmt.Errorf("failed to update permissions: %w", err)
	}

	// Get updated file
	updatedFile, err := s.repo.GetFileWithPermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated file: %w", err)
	}

	return updatedFile, nil
}

// GetFilePermissions retrieves the permission configuration for a file
func (s *FileService) GetFilePermissions(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.FilePermission, error) {
	file, err := s.repo.GetFileWithPermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}

	// Check if user has access to view permissions (owner or admin)
	if file.CreatedBy != userID {
		// Check if user is admin (this should be done in handler)
		// For now, allow only owner
		return nil, fmt.Errorf("unauthorized")
	}

	// Parse allowed roles
	var allowedRoles []string
	if file.AllowedRoles != nil && len(file.AllowedRoles) > 0 {
		// Simple JSON parsing - in production use proper JSON parser
		rolesStr := string(file.AllowedRoles)
		rolesStr = strings.Trim(rolesStr, "[]")
		if rolesStr != "" {
			parts := strings.Split(rolesStr, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				part = strings.Trim(part, "\"")
				if part != "" {
					allowedRoles = append(allowedRoles, part)
				}
			}
		}
	}

	// Parse allowed product IDs
	var allowedProductIDs []string
	if file.AllowedProductIDs != nil && len(file.AllowedProductIDs) > 0 {
		productIDsStr := string(file.AllowedProductIDs)
		productIDsStr = strings.Trim(productIDsStr, "[]")
		if productIDsStr != "" {
			parts := strings.Split(productIDsStr, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				part = strings.Trim(part, "\"")
				if part != "" {
					allowedProductIDs = append(allowedProductIDs, part)
				}
			}
		}
	}

	// Generate public link URL
	publicLinkURL := ""
	if file.AccessType == models.AccessTypePublic {
		publicLinkURL = fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/public/%s/download", file.PublicLinkToken.String())
	}

	return &models.FilePermission{
		FileID:              file.ID,
		AccessType:          file.AccessType,
		RequiredRole:        file.RequiredRole,
		AllowedRoles:        allowedRoles,
		RequiredProductID:   file.RequiredProductID,
		AllowedProductIDs:   allowedProductIDs,
		PublicLinkToken:     file.PublicLinkToken,
		PublicLinkExpiresAt: file.PublicLinkExpiresAt,
		DownloadCount:       file.DownloadCount,
		MaxDownloads:        file.MaxDownloads,
		PublicLinkURL:       publicLinkURL,
	}, nil
}

// CheckFileAccess checks if a user has access to download a file
func (s *FileService) CheckFileAccess(ctx context.Context, fileID, userID uuid.UUID) (bool, error) {
	return s.repo.CheckAccess(ctx, fileID, userID)
}

// LogFileAccess logs a file access attempt
func (s *FileService) LogFileAccess(ctx context.Context, fileID, userID uuid.UUID, action string, accessGranted bool, reason, ipAddress, userAgent string) error {
	log := &models.FileAccessLog{
		FileID:        fileID,
		UserID:        userID,
		Action:        action,
		AccessGranted: accessGranted,
		Reason:        reason,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}
	return s.repo.LogAccess(ctx, log)
}

// GetFileAccessLogs retrieves access logs for a file
func (s *FileService) GetFileAccessLogs(ctx context.Context, fileID uuid.UUID, page, pageSize int) ([]models.FileAccessLog, int, error) {
	return s.repo.GetAccessLogs(ctx, fileID, page, pageSize)
}

// AddRoleToFilePermission adds a role permission to a file
func (s *FileService) AddRoleToFilePermission(ctx context.Context, fileID uuid.UUID, role string) error {
	return s.repo.AddRolePermission(ctx, fileID, role)
}

// RemoveRoleFromFilePermission removes a role permission from a file
func (s *FileService) RemoveRoleFromFilePermission(ctx context.Context, fileID uuid.UUID, role string) error {
	return s.repo.RemoveRolePermission(ctx, fileID, role)
}

// AddProductToFilePermission adds a product permission to a file
func (s *FileService) AddProductToFilePermission(ctx context.Context, fileID, productID uuid.UUID) error {
	return s.repo.AddProductPermission(ctx, fileID, productID)
}

// RemoveProductFromFilePermission removes a product permission from a file
func (s *FileService) RemoveProductFromFilePermission(ctx context.Context, fileID, productID uuid.UUID) error {
	return s.repo.RemoveProductPermission(ctx, fileID, productID)
}

// RegeneratePublicLink regenerates the public link token for a file
func (s *FileService) RegeneratePublicLink(ctx context.Context, fileID, userID uuid.UUID) (string, error) {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return "", fmt.Errorf("file not found")
	}
	if file.CreatedBy != userID {
		return "", fmt.Errorf("unauthorized")
	}

	newToken, err := s.repo.RegeneratePublicLinkToken(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("failed to regenerate token: %w", err)
	}

	publicLinkURL := fmt.Sprintf("https://api.pixelcraft-studio.store/api/v1/files/public/%s/download", newToken.String())
	return publicLinkURL, nil
}

// GetFileByPublicToken retrieves a file by its public link token
func (s *FileService) GetFileByPublicToken(ctx context.Context, token uuid.UUID) (*models.File, error) {
	// This would need a new repository method
	// For now, we'll handle it in the handler
	return nil, fmt.Errorf("not implemented")
}

// GenerateOneTimeDownloadToken generates a one-time download token for a file
func (s *FileService) GenerateOneTimeDownloadToken(ctx context.Context, fileID, userID uuid.UUID, expiresInMinutes, maxDownloads int) (*models.OneTimeDownloadToken, error) {
	// Verify file exists and belongs to user
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}
	if file.CreatedBy != userID {
		return nil, fmt.Errorf("unauthorized: file does not belong to user")
	}

	// Default values
	if expiresInMinutes <= 0 {
		expiresInMinutes = 15 // Default: 15 minutes
	}
	if maxDownloads <= 0 {
		maxDownloads = 1 // Default: 1 download
	}

	// Create token
	token := &models.OneTimeDownloadToken{
		ID:           uuid.New(),
		FileID:       fileID,
		UserID:       userID,
		Token:        uuid.New(),
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(expiresInMinutes) * time.Minute),
		MaxDownloads: maxDownloads,
	}

	err = s.repo.CreateOneTimeDownloadToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return token, nil
}

// ValidateOneTimeDownloadToken validates a one-time download token
func (s *FileService) ValidateOneTimeDownloadToken(ctx context.Context, token uuid.UUID, ipAddress, userAgent string) (uuid.UUID, bool, string, error) {
	return s.repo.ValidateAndUseOneTimeToken(ctx, token, ipAddress, userAgent)
}

// GetOneTimeDownloadToken retrieves a one-time download token by its token string
func (s *FileService) GetOneTimeDownloadToken(ctx context.Context, token uuid.UUID) (*models.OneTimeDownloadToken, error) {
	return s.repo.GetOneTimeTokenByTokenString(ctx, token)
}

// CleanupExpiredDownloadTokens removes expired tokens
func (s *FileService) CleanupExpiredDownloadTokens(ctx context.Context) (int, error) {
	return s.repo.CleanupExpiredDownloadTokens(ctx)
}
