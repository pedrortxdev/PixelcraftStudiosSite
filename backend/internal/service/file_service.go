package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	tempDir      string // Temporary directory for atomic uploads
	maxFileSize  int64  // Maximum file size in bytes
	allowedTypes []string // Allowed file extensions
	apiBaseURL   string   // Base URL for public API endpoints (configurable)
}

// NewFileService creates a new FileService
func NewFileService(db *sql.DB, uploadDir string, maxFileSize int64, allowedTypes []string, apiBaseURL string) *FileService {
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	// Create upload directory if it doesn't exist
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}

	// CRITICAL: Create temp directory INSIDE uploadDir to ensure same partition
	// This prevents "invalid cross-device link" errors on os.Rename()
	// See: https://github.com/golang/go/issues/27945
	tempDir := filepath.Join(uploadDir, ".tmp")
	err = os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp directory: %v", err))
	}

	return &FileService{
		repo:        repository.NewFileRepository(db),
		db:          db,
		uploadDir:   uploadDir,
		tempDir:     tempDir,
		maxFileSize: maxFileSize,
		allowedTypes: allowedTypes,
		apiBaseURL:  apiBaseURL,
	}
}

// GenerateDownloadToken generates a secure one-time download token for a file
// The token expires after 1 hour and can be used once
func (s *FileService) GenerateDownloadToken(ctx context.Context, fileID uuid.UUID) (string, error) {
	// Generate a random token
	token := uuid.New()

	// Store token in database with expiration
	query := `
		INSERT INTO download_tokens (id, file_id, expires_at, max_downloads, current_downloads)
		VALUES ($1, $2, NOW() + INTERVAL '1 hour', 1, 0)
		ON CONFLICT (id) DO UPDATE
		SET file_id = $2, expires_at = NOW() + INTERVAL '1 hour', current_downloads = 0
	`
	_, err := s.db.ExecContext(ctx, query, token, fileID)
	if err != nil {
		return "", fmt.Errorf("failed to create download token: %w", err)
	}

	return token.String(), nil
}

// SaveFile handles file upload with atomic operations and MIME type validation
func (s *FileService) SaveFile(ctx context.Context, fileHeader *multipart.FileHeader, userID uuid.UUID, friendlyName string) (*models.File, error) {
	// Validate file size
	if fileHeader.Size > s.maxFileSize {
		maxSizeMB := float64(s.maxFileSize) / (1024 * 1024)
		fileSizeMB := float64(fileHeader.Size) / (1024 * 1024)
		return nil, fmt.Errorf("file size too large: %.2f MB. Maximum allowed: %.2f MB", fileSizeMB, maxSizeMB)
	}

	// Open uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read first 512 bytes for MIME type detection (Magic Number)
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	// Detect actual MIME type from magic bytes
	detectedMIME := http.DetectContentType(buffer[:n])
	
	// Validate MIME type against allowed types
	if !s.isAllowedMIMEType(detectedMIME) {
		return nil, fmt.Errorf("file type not allowed: %s (detected from magic bytes)", detectedMIME)
	}

	// Get file extension and validate
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !s.isAllowedType(ext) {
		allowedTypesStr := strings.Join(s.allowedTypes, ", ")
		return nil, fmt.Errorf("file extension not allowed: %s. Allowed types: %s", ext, allowedTypesStr)
	}

	// Generate unique internal filename using UUID
	internalFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// ATOMIC UPLOAD: Save to temp directory first
	// Use timestamp prefix for efficient cleanup (avoids Lstat calls)
	timestampPrefix := time.Now().Format("20060102150405")
	tempFileName := fmt.Sprintf("%s_%s", timestampPrefix, internalFileName)
	tempFilePath := filepath.Join(s.tempDir, tempFileName)
	dest, err := os.Create(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write the buffer we already read + rest of file
	_, err = dest.Write(buffer[:n])
	if err != nil {
		dest.Close()
		os.Remove(tempFilePath)
		return nil, fmt.Errorf("failed to write file buffer: %w", err)
	}

	_, err = io.Copy(dest, src)
	if err != nil {
		dest.Close()
		os.Remove(tempFilePath)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	err = dest.Close()
	if err != nil {
		os.Remove(tempFilePath)
		return nil, fmt.Errorf("failed to close file: %w", err)
	}

	// ATOMIC UPLOAD: Move from temp to final location FIRST (before DB)
	// This prevents DB/file desync: file must exist on disk before we record it
	finalPath := filepath.Join(s.uploadDir, internalFileName)
	err = os.Rename(tempFilePath, finalPath)
	if err != nil {
		// Move failed - clean up temp file
		os.Remove(tempFilePath)
		return nil, fmt.Errorf("failed to move file to final location: %w", err)
	}

	// File is now safely on disk - create database record
	// If DB fails, we have an orphan file (cleaned up by periodic cleanup job)
	fileType := s.getFileType(ext)
	file := &models.File{
		ID:        uuid.New(),
		Name:      friendlyName,
		FileName:  internalFileName,
		FileType:  fileType,
		FilePath:  finalPath,
		Size:      fileHeader.Size,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDeleted: false,
	}

	err = s.repo.Create(ctx, file)
	if err != nil {
		// DB failed - file exists on disk but record not created
		// This is safer than the reverse: orphan files can be cleaned up
		// but phantom DB records pointing to non-existent files cause errors
		os.Remove(finalPath) // Clean up orphaned file
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	return file, nil
}

// GetFilePath returns the sanitized full path to a file (prevents directory traversal)
func (s *FileService) GetFilePath(fileID uuid.UUID, fileName string) string {
	// SECURITY: Use filepath.Base to strip any directory traversal attempts
	// ../../etc/passwd becomes just etc/passwd, then filepath.Join makes it safe
	safeFileName := filepath.Base(fileName)
	return filepath.Join(s.uploadDir, safeFileName)
}

// isAllowedMIMEType checks if the detected MIME type is allowed
// Handles both standard files and binary game distributions
func (s *FileService) isAllowedMIMEType(mime string) bool {
	// Map MIME types to allowed categories
	allowedMIMEPrefixes := []string{
		// Images
		"image/",
		
		// Documents
		"application/pdf",
		"text/plain",
		
		// Archives (zip, jar, etc.)
		"application/zip",
		"application/x-zip-compressed",
		"application/x-tar",
		"application/x-gtar",
		"application/x-rar-compressed",
		"application/java-archive", // .jar files
		
		// Executables (game launchers, mods, etc.)
		"application/x-executable",
		"application/x-dosexec",     // Windows .exe
		"application/x-mach-binary", // macOS binaries
		"application/x-elf",         // Linux binaries
		
		// Game-specific MIME types
		"application/octet-stream",  // Generic binary (ACCEPTED for game files)
	}

	for _, prefix := range allowedMIMEPrefixes {
		if strings.HasPrefix(mime, prefix) {
			return true
		}
	}
	return false
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

// UpdateFilePermissions updates the access permissions for a file
// Uses relational tables for AllowedRoles and AllowedProductIDs (best practice)
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
		// Use relational tables instead of JSON (best practice)
		if err := s.repo.SyncFilePermissions(ctx, id, req.AllowedRoles, nil); err != nil {
			return nil, fmt.Errorf("failed to sync role permissions: %w", err)
		}
		// Also update JSON for backward compatibility (transitional)
		rolesJSON, err := json.Marshal(req.AllowedRoles)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal allowed roles: %w", err)
		}
		file.AllowedRoles = rolesJSON
	}
	if req.RequiredProductID != nil {
		file.RequiredProductID = req.RequiredProductID
	}
	if req.AllowedProductIDs != nil {
		// Use relational tables instead of JSON (best practice)
		if err := s.repo.SyncFilePermissions(ctx, id, nil, req.AllowedProductIDs); err != nil {
			return nil, fmt.Errorf("failed to sync product permissions: %w", err)
		}
		// Also update JSON for backward compatibility (transitional)
		productIDsJSON, err := json.Marshal(req.AllowedProductIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal allowed product IDs: %w", err)
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
// Uses relational tables as primary source, JSON as fallback (transitional)
func (s *FileService) GetFilePermissions(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.FilePermissionResponse, error) {
	file, err := s.repo.GetFileWithPermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}

	// Check if user has access to view permissions (owner or admin)
	if file.CreatedBy != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// PRIMARY: Read from relational tables (best practice)
	// FALLBACK: Parse JSON if relational tables are empty (transitional)
	var allowedRoles []string
	allowedRoles, err = s.repo.GetRolePermissionsForFile(ctx, id)
	if err != nil {
		// Fallback to JSON parsing
		if file.AllowedRoles != nil && len(file.AllowedRoles) > 0 {
			if jsonErr := json.Unmarshal(file.AllowedRoles, &allowedRoles); jsonErr != nil {
				return nil, fmt.Errorf("failed to unmarshal allowed roles: %w", jsonErr)
			}
		}
	}

	var allowedProductIDs []uuid.UUID
	allowedProductIDs, err = s.repo.GetProductPermissionsForFile(ctx, id)
	if err != nil {
		// Fallback to JSON parsing
		if file.AllowedProductIDs != nil && len(file.AllowedProductIDs) > 0 {
			if jsonErr := json.Unmarshal(file.AllowedProductIDs, &allowedProductIDs); jsonErr != nil {
				return nil, fmt.Errorf("failed to unmarshal allowed product IDs: %w", jsonErr)
			}
		}
	}

	// Generate public link URL using configurable base URL
	publicLinkURL := ""
	if file.AccessType == models.AccessTypePublic && file.PublicLinkToken != uuid.Nil {
		publicLinkURL = fmt.Sprintf("%s/api/v1/files/public/%s/download", s.apiBaseURL, file.PublicLinkToken.String())
	}

	// Return properly typed response DTO
	return &models.FilePermissionResponse{
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

// GetFileByPublicToken retrieves a file by its public link token
// Used for public file downloads (no authentication required)
func (s *FileService) GetFileByPublicToken(ctx context.Context, token uuid.UUID) (*models.File, error) {
	query := `
		SELECT id, name, file_name, file_type, file_path, size,
		       created_by, created_at, updated_at, is_deleted,
		       access_type, required_role, allowed_roles, required_product_id, allowed_product_ids,
		       public_link_token, public_link_expires_at, download_count, max_downloads
		FROM files
		WHERE public_link_token = $1 AND is_deleted = false
	`

	file := &models.File{}
	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&file.ID, &file.Name, &file.FileName, &file.FileType, &file.FilePath,
		&file.Size, &file.CreatedBy, &file.CreatedAt, &file.UpdatedAt, &file.IsDeleted,
		&file.AccessType, &file.RequiredRole, &file.AllowedRoles, &file.RequiredProductID,
		&file.AllowedProductIDs, &file.PublicLinkToken, &file.PublicLinkExpiresAt,
		&file.DownloadCount, &file.MaxDownloads,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file by public token: %w", err)
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

	file.Name = newName
	file.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, id, file)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

// DeleteFile soft deletes a file
func (s *FileService) DeleteFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
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

	err = s.repo.SoftDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

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

// GetAllFiles retrieves all files with pagination and search (admin only)
func (s *FileService) GetAllFiles(ctx context.Context, page, pageSize int, search string) ([]models.FileWithUser, int, error) {
	return s.repo.GetAllWithUsers(ctx, page, pageSize, search)
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

	publicLinkURL := fmt.Sprintf("%s/api/v1/files/public/%s/download", s.apiBaseURL, newToken.String())
	return publicLinkURL, nil
}

// GenerateOneTimeDownloadToken generates a one-time download token for a file
func (s *FileService) GenerateOneTimeDownloadToken(ctx context.Context, fileID, userID uuid.UUID, expiresInMinutes, maxDownloads int) (*models.OneTimeDownloadToken, error) {
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

	if expiresInMinutes <= 0 {
		expiresInMinutes = 15
	}
	if maxDownloads <= 0 {
		maxDownloads = 1
	}

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

// CleanupTempFiles removes orphaned temp files (run periodically)
// Uses timestamp-based naming pattern to avoid expensive Lstat calls
func (s *FileService) CleanupTempFiles(ctx context.Context) error {
	entries, err := os.ReadDir(s.tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	now := time.Now()
	cleanupThreshold := 1 * time.Hour // Files older than 1 hour

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Naming pattern: {timestamp}_{filename}
		// This allows us to extract the timestamp from the filename itself
		// without needing expensive Lstat calls
		name := entry.Name()
		var fileTime time.Time

		// Try to parse timestamp from filename (format: YYYYMMDDHHMMSS_filename)
		if len(name) >= 15 && name[14] == '_' {
			if parsedTime, err := time.ParseInLocation("20060102150405", name[:14], now.Location()); err == nil {
				fileTime = parsedTime
			} else {
				// Fallback to ModTime if timestamp parsing fails
				info, err := entry.Info()
				if err != nil {
					continue
				}
				fileTime = info.ModTime()
			}
		} else {
			// Fallback to ModTime for files without timestamp prefix
			info, err := entry.Info()
			if err != nil {
				continue
			}
			fileTime = info.ModTime()
		}

		if now.Sub(fileTime) > cleanupThreshold {
			fullPath := filepath.Join(s.tempDir, entry.Name())
			os.Remove(fullPath) // Best effort cleanup
		}
	}

	return nil
}

// AddRoleToFilePermission adds a role permission to a file
// Uses relational tables (best practice) with JSON fallback (transitional)
func (s *FileService) AddRoleToFilePermission(ctx context.Context, fileID uuid.UUID, role string) error {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found")
	}

	// PRIMARY: Add to relational table (best practice)
	if err := s.repo.AddRolePermissionToFile(ctx, fileID, role); err != nil {
		return fmt.Errorf("failed to add role permission: %w", err)
	}

	// ALSO update JSON for backward compatibility (transitional)
	var allowedRoles []string
	if file.AllowedRoles != nil && len(file.AllowedRoles) > 0 {
		if err := json.Unmarshal(file.AllowedRoles, &allowedRoles); err != nil {
			allowedRoles = []string{}
		}
	}

	// Add new role if not already present
	for _, r := range allowedRoles {
		if r == role {
			return nil // Already exists in JSON
		}
	}
	allowedRoles = append(allowedRoles, role)

	rolesJSON, err := json.Marshal(allowedRoles)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed roles: %w", err)
	}

	file.AllowedRoles = rolesJSON
	return s.repo.UpdatePermissions(ctx, fileID, file)
}

// RemoveRoleFromFilePermission removes a role permission from a file
// Uses relational tables (best practice) with JSON fallback (transitional)
func (s *FileService) RemoveRoleFromFilePermission(ctx context.Context, fileID uuid.UUID, role string) error {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found")
	}

	// PRIMARY: Remove from relational table (best practice)
	if err := s.repo.RemoveRolePermissionFromFile(ctx, fileID, role); err != nil {
		return fmt.Errorf("failed to remove role permission: %w", err)
	}

	// ALSO update JSON for backward compatibility (transitional)
	var allowedRoles []string
	if file.AllowedRoles != nil && len(file.AllowedRoles) > 0 {
		if err := json.Unmarshal(file.AllowedRoles, &allowedRoles); err != nil {
			allowedRoles = []string{}
		}
	}

	// Remove role
	newRoles := []string{}
	for _, r := range allowedRoles {
		if r != role {
			newRoles = append(newRoles, r)
		}
	}

	rolesJSON, err := json.Marshal(newRoles)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed roles: %w", err)
	}

	file.AllowedRoles = rolesJSON
	return s.repo.UpdatePermissions(ctx, fileID, file)
}

// AddProductToFilePermission adds a product permission to a file
// Uses relational tables (best practice) with JSON fallback (transitional)
func (s *FileService) AddProductToFilePermission(ctx context.Context, fileID, productID uuid.UUID) error {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found")
	}

	// PRIMARY: Add to relational table (best practice)
	if err := s.repo.AddProductPermissionToFile(ctx, fileID, productID); err != nil {
		return fmt.Errorf("failed to add product permission: %w", err)
	}

	// ALSO update JSON for backward compatibility (transitional)
	var allowedProductIDs []uuid.UUID
	if file.AllowedProductIDs != nil && len(file.AllowedProductIDs) > 0 {
		if err := json.Unmarshal(file.AllowedProductIDs, &allowedProductIDs); err != nil {
			allowedProductIDs = []uuid.UUID{}
		}
	}

	// Add new product if not already present
	for _, pid := range allowedProductIDs {
		if pid == productID {
			return nil // Already exists in JSON
		}
	}
	allowedProductIDs = append(allowedProductIDs, productID)

	productIDsJSON, err := json.Marshal(allowedProductIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed product IDs: %w", err)
	}

	file.AllowedProductIDs = productIDsJSON
	return s.repo.UpdatePermissions(ctx, fileID, file)
}

// RemoveProductFromFilePermission removes a product permission from a file
// Uses relational tables (best practice) with JSON fallback (transitional)
func (s *FileService) RemoveProductFromFilePermission(ctx context.Context, fileID, productID uuid.UUID) error {
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found")
	}

	// PRIMARY: Remove from relational table (best practice)
	if err := s.repo.RemoveProductPermissionFromFile(ctx, fileID, productID); err != nil {
		return fmt.Errorf("failed to remove product permission: %w", err)
	}

	// ALSO update JSON for backward compatibility (transitional)
	var allowedProductIDs []uuid.UUID
	if file.AllowedProductIDs != nil && len(file.AllowedProductIDs) > 0 {
		if err := json.Unmarshal(file.AllowedProductIDs, &allowedProductIDs); err != nil {
			allowedProductIDs = []uuid.UUID{}
		}
	}

	// Remove product
	newProductIDs := []uuid.UUID{}
	for _, pid := range allowedProductIDs {
		if pid != productID {
			newProductIDs = append(newProductIDs, pid)
		}
	}

	productIDsJSON, err := json.Marshal(newProductIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed product IDs: %w", err)
	}

	file.AllowedProductIDs = productIDsJSON
	return s.repo.UpdatePermissions(ctx, fileID, file)
}
