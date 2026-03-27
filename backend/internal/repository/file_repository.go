package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
)

// FileRepository handles database operations for files
type FileRepository struct {
	db *sql.DB
}

// NewFileRepository creates a new FileRepository
func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create creates a new file record
func (r *FileRepository) Create(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (id, name, file_name, file_type, file_path, size, created_by, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		file.ID,
		file.Name,
		file.FileName,
		file.FileType,
		file.FilePath,
		file.Size,
		file.CreatedBy,
		file.CreatedAt,
		file.UpdatedAt,
		file.IsDeleted,
	)
	return err
}

// GetByID retrieves a file by ID
func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.File, error) {
	query := `
		SELECT id, name, file_name, file_type, file_path, size, created_by, created_at, updated_at, is_deleted
		FROM files
		WHERE id = $1 AND is_deleted = false
	`

	var file models.File
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&file.ID,
		&file.Name,
		&file.FileName,
		&file.FileType,
		&file.FilePath,
		&file.Size,
		&file.CreatedBy,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.IsDeleted,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &file, nil
}

// GetByUserID retrieves files by user ID with pagination
func (r *FileRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.File, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// Count total
	countQuery := `
		SELECT COUNT(*)
		FROM files
		WHERE created_by = $1 AND is_deleted = false
	`

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get files
	query := `
		SELECT id, name, file_name, file_type, file_path, size, created_by, created_at, updated_at, is_deleted
		FROM files
		WHERE created_by = $1 AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.FileName,
			&file.FileType,
			&file.FilePath,
			&file.Size,
			&file.CreatedBy,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.IsDeleted,
		)
		if err != nil {
			return nil, 0, err
		}
		files = append(files, file)
	}

	return files, total, nil
}

// Update updates a file record
func (r *FileRepository) Update(ctx context.Context, id uuid.UUID, file *models.File) error {
	query := `
		UPDATE files
		SET name = $1, updated_at = $2
		WHERE id = $3 AND is_deleted = false
	`

	_, err := r.db.ExecContext(ctx, query,
		file.Name,
		file.UpdatedAt,
		id,
	)
	return err
}

// SoftDelete marks a file as deleted
func (r *FileRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE files
		SET is_deleted = true, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetByFileName retrieves a file by its file name
func (r *FileRepository) GetByFileName(ctx context.Context, fileName string) (*models.File, error) {
	query := `
		SELECT id, name, file_name, file_type, file_path, size, created_by, created_at, updated_at, is_deleted
		FROM files
		WHERE file_name = $1 AND is_deleted = false
	`

	var file models.File
	err := r.db.QueryRowContext(ctx, query, fileName).Scan(
		&file.ID,
		&file.Name,
		&file.FileName,
		&file.FileType,
		&file.FilePath,
		&file.Size,
		&file.CreatedBy,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.IsDeleted,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &file, nil
}

// GetAllWithUsers retrieves all files with user information (admin only)
func (r *FileRepository) GetAllWithUsers(ctx context.Context, page, pageSize int, search string) ([]models.FileWithUser, int, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// Build query with search
	query := `
		SELECT f.id, f.name, f.file_name, f.file_type, f.file_path, f.size,
		       f.created_by, f.created_at, f.updated_at, f.is_deleted,
		       u.email as user_email, u.username as user_name
		FROM files f
		LEFT JOIN users u ON f.created_by = u.id
		WHERE f.is_deleted = false
	`

	countQuery := `SELECT COUNT(*) FROM files WHERE is_deleted = false`

	args := []interface{}{}
	argCount := 1

	if search != "" {
		query += ` AND f.name ILIKE $` + string(rune(argCount+'0'))
		countQuery += ` AND name ILIKE $1`
		args = append(args, "%"+search+"%")
		argCount++
	}

	query += ` ORDER BY f.created_at DESC LIMIT $` + string(rune(argCount+'0')) + ` OFFSET $` + string(rune(argCount+'1'))
	args = append(args, pageSize, offset)

	// Get total count
	var total int
	if search != "" {
		err := r.db.QueryRowContext(ctx, countQuery, "%"+search+"%").Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var files []models.FileWithUser
	for rows.Next() {
		var file models.FileWithUser
		err := rows.Scan(
			&file.ID, &file.Name, &file.FileName, &file.FileType, &file.FilePath,
			&file.Size, &file.CreatedBy, &file.CreatedAt, &file.UpdatedAt, &file.IsDeleted,
			&file.UserEmail, &file.UserName,
		)
		if err != nil {
			return nil, 0, err
		}
		files = append(files, file)
	}

	return files, total, nil
}

// UpdatePermissions updates file access permissions
func (r *FileRepository) UpdatePermissions(ctx context.Context, id uuid.UUID, file *models.File) error {
	query := `
		UPDATE files
		SET access_type = $1,
		    required_role = $2,
		    allowed_roles = $3,
		    required_product_id = $4,
		    allowed_product_ids = $5,
		    public_link_expires_at = $6,
		    max_downloads = $7,
		    updated_at = NOW()
		WHERE id = $8 AND is_deleted = false
	`

	_, err := r.db.ExecContext(ctx, query,
		file.AccessType,
		file.RequiredRole,
		file.AllowedRoles,
		file.RequiredProductID,
		file.AllowedProductIDs,
		file.PublicLinkExpiresAt,
		file.MaxDownloads,
		id,
	)
	return err
}

// GetFileWithPermissions retrieves a file with full permission details
func (r *FileRepository) GetFileWithPermissions(ctx context.Context, id uuid.UUID) (*models.File, error) {
	query := `
		SELECT id, name, file_name, file_type, file_path, size,
		       created_by, created_at, updated_at, is_deleted,
		       access_type, required_role, allowed_roles, required_product_id, allowed_product_ids,
		       public_link_token, public_link_expires_at, download_count, max_downloads
		FROM files
		WHERE id = $1 AND is_deleted = false
	`

	var file models.File
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&file.ID,
		&file.Name,
		&file.FileName,
		&file.FileType,
		&file.FilePath,
		&file.Size,
		&file.CreatedBy,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.IsDeleted,
		&file.AccessType,
		&file.RequiredRole,
		&file.AllowedRoles,
		&file.RequiredProductID,
		&file.AllowedProductIDs,
		&file.PublicLinkToken,
		&file.PublicLinkExpiresAt,
		&file.DownloadCount,
		&file.MaxDownloads,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &file, nil
}

// CheckAccess checks if a user has access to a file using the database function
func (r *FileRepository) CheckAccess(ctx context.Context, fileID, userID uuid.UUID) (bool, error) {
	query := `SELECT check_file_access($1, $2)`

	var hasAccess bool
	err := r.db.QueryRowContext(ctx, query, fileID, userID).Scan(&hasAccess)
	if err != nil {
		return false, err
	}

	return hasAccess, nil
}

// LogAccess logs a file access attempt
func (r *FileRepository) LogAccess(ctx context.Context, log *models.FileAccessLog) error {
	query := `
		INSERT INTO file_access_logs (file_id, user_id, action, access_granted, reason, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.FileID,
		log.UserID,
		log.Action,
		log.AccessGranted,
		log.Reason,
		log.IPAddress,
		log.UserAgent,
	)
	return err
}

// GetAccessLogs retrieves access logs for a file with pagination
func (r *FileRepository) GetAccessLogs(ctx context.Context, fileID uuid.UUID, page, pageSize int) ([]models.FileAccessLog, int, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM file_access_logs WHERE file_id = $1`

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, fileID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get logs
	query := `
		SELECT id, file_id, user_id, action, access_granted, reason, ip_address, user_agent, created_at
		FROM file_access_logs
		WHERE file_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, fileID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.FileAccessLog
	for rows.Next() {
		var log models.FileAccessLog
		err := rows.Scan(
			&log.ID, &log.FileID, &log.UserID, &log.Action, &log.AccessGranted,
			&log.Reason, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

// AddRolePermission adds a role permission to a file
func (r *FileRepository) AddRolePermission(ctx context.Context, fileID uuid.UUID, role string) error {
	query := `
		INSERT INTO file_role_permissions (file_id, role)
		VALUES ($1, $2)
		ON CONFLICT (file_id, role) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, fileID, role)
	return err
}

// RemoveRolePermission removes a role permission from a file
func (r *FileRepository) RemoveRolePermission(ctx context.Context, fileID uuid.UUID, role string) error {
	query := `
		DELETE FROM file_role_permissions
		WHERE file_id = $1 AND role = $2
	`

	_, err := r.db.ExecContext(ctx, query, fileID, role)
	return err
}

// AddProductPermission adds a product permission to a file
func (r *FileRepository) AddProductPermission(ctx context.Context, fileID, productID uuid.UUID) error {
	query := `
		INSERT INTO file_product_permissions (file_id, product_id)
		VALUES ($1, $2)
		ON CONFLICT (file_id, product_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, fileID, productID)
	return err
}

// RemoveProductPermission removes a product permission from a file
func (r *FileRepository) RemoveProductPermission(ctx context.Context, fileID, productID uuid.UUID) error {
	query := `
		DELETE FROM file_product_permissions
		WHERE file_id = $1 AND product_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, fileID, productID)
	return err
}

// RegeneratePublicLinkToken generates a new public link token for a file
func (r *FileRepository) RegeneratePublicLinkToken(ctx context.Context, fileID uuid.UUID) (uuid.UUID, error) {
	newToken := uuid.New()
	query := `
		UPDATE files
		SET public_link_token = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, newToken, fileID)
	if err != nil {
		return uuid.Nil, err
	}

	return newToken, nil
}

// CreateOneTimeDownloadToken creates a new one-time download token
func (r *FileRepository) CreateOneTimeDownloadToken(ctx context.Context, token *models.OneTimeDownloadToken) error {
	query := `
		INSERT INTO one_time_download_tokens (id, file_id, user_id, token, created_at, expires_at, max_downloads)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.FileID,
		token.UserID,
		token.Token,
		token.CreatedAt,
		token.ExpiresAt,
		token.MaxDownloads,
	)
	return err
}

// GetOneTimeTokenByUUID retrieves a one-time download token by its UUID
func (r *FileRepository) GetOneTimeTokenByUUID(ctx context.Context, tokenID uuid.UUID) (*models.OneTimeDownloadToken, error) {
	query := `
		SELECT id, file_id, user_id, token, created_at, expires_at, used_at, is_used, download_count, max_downloads,
		       COALESCE(ip_address, ''), COALESCE(user_agent, '')
		FROM one_time_download_tokens
		WHERE id = $1
	`

	var t models.OneTimeDownloadToken
	err := r.db.QueryRowContext(ctx, query, tokenID).Scan(
		&t.ID, &t.FileID, &t.UserID, &t.Token, &t.CreatedAt, &t.ExpiresAt,
		&t.UsedAt, &t.IsUsed, &t.DownloadCount, &t.MaxDownloads,
		&t.IPAddress, &t.UserAgent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &t, nil
}

// GetOneTimeTokenByTokenString retrieves a one-time download token by its token string
func (r *FileRepository) GetOneTimeTokenByTokenString(ctx context.Context, tokenStr uuid.UUID) (*models.OneTimeDownloadToken, error) {
	query := `
		SELECT id, file_id, user_id, token, created_at, expires_at, used_at, is_used, download_count, max_downloads,
		       COALESCE(ip_address, ''), COALESCE(user_agent, '')
		FROM one_time_download_tokens
		WHERE token = $1
	`

	var t models.OneTimeDownloadToken
	err := r.db.QueryRowContext(ctx, query, tokenStr).Scan(
		&t.ID, &t.FileID, &t.UserID, &t.Token, &t.CreatedAt, &t.ExpiresAt,
		&t.UsedAt, &t.IsUsed, &t.DownloadCount, &t.MaxDownloads,
		&t.IPAddress, &t.UserAgent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &t, nil
}

// ValidateAndUseOneTimeToken validates a token and increments its usage counter using database function
func (r *FileRepository) ValidateAndUseOneTimeToken(ctx context.Context, token uuid.UUID, ipAddress string, userAgent string) (uuid.UUID, bool, string, error) {
	query := `SELECT * FROM validate_and_use_download_token($1, $2, $3)`

	var fileID uuid.UUID
	var isValid bool
	var errorMessage string

	err := r.db.QueryRowContext(ctx, query, token, ipAddress, userAgent).Scan(&fileID, &isValid, &errorMessage)
	if err != nil {
		return uuid.Nil, false, "Failed to validate token", err
	}

	return fileID, isValid, errorMessage, nil
}

// CleanupExpiredDownloadTokens removes expired and used tokens older than 24 hours
func (r *FileRepository) CleanupExpiredDownloadTokens(ctx context.Context) (int, error) {
	query := `SELECT cleanup_expired_download_tokens()`

	var deletedCount int
	err := r.db.QueryRowContext(ctx, query).Scan(&deletedCount)
	if err != nil {
		return 0, err
	}

	return deletedCount, nil
}
