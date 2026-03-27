package models

import (
	"time"

	"github.com/google/uuid"
)

// FileType represents the type of file
type FileType string

const (
	FileTypeJar     FileType = "JAR"
	FileTypeZip     FileType = "ZIP"
	FileTypeExe     FileType = "EXE"
	FileTypePng     FileType = "PNG"
	FileTypeJpg     FileType = "JPG"
	FileTypePdf     FileType = "PDF"
	FileTypeOther   FileType = "OTHER"
)

// AccessType represents the access level for a file
type AccessType string

const (
	AccessTypePublic  AccessType = "PUBLIC"  // Anyone with link can download
	AccessTypePrivate AccessType = "PRIVATE" // Only buyers can download
	AccessTypeRole    AccessType = "ROLE"    // Only specific roles can download
)

// File represents a stored file
type File struct {
	ID                     uuid.UUID  `db:"id" json:"id"`
	Name                   string     `db:"name" json:"name"`                 // User-friendly name
	FileName               string     `db:"file_name" json:"file_name"`       // Internal UUID-based name
	FileType               FileType   `db:"file_type" json:"file_type"`
	FilePath               string     `db:"file_path" json:"file_path"`       // Path relative to upload directory
	Size                   int64      `db:"size" json:"size"`                 // File size in bytes
	CreatedBy              uuid.UUID  `db:"created_by" json:"created_by"`     // User ID who uploaded
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
	IsDeleted              bool       `db:"is_deleted" json:"is_deleted"`
	AccessType             AccessType `db:"access_type" json:"access_type"`                           // PUBLIC, PRIVATE, ROLE
	RequiredRole           *string    `db:"required_role" json:"required_role"`                       // Single role required
	AllowedRoles           []byte     `db:"allowed_roles" json:"allowed_roles"`                       // JSONB array of roles
	RequiredProductID      *uuid.UUID `db:"required_product_id" json:"required_product_id"`           // Single product required
	AllowedProductIDs      []byte     `db:"allowed_product_ids" json:"allowed_product_ids"`           // JSONB array of product IDs
	PublicLinkToken        uuid.UUID  `db:"public_link_token" json:"public_link_token"`               // Token for public link
	PublicLinkExpiresAt    *time.Time `db:"public_link_expires_at" json:"public_link_expires_at"`     // Expiration for public link
	DownloadCount          int        `db:"download_count" json:"download_count"`                     // Number of downloads
	MaxDownloads           *int       `db:"max_downloads" json:"max_downloads"`                       // Max downloads limit
}

// CreateFileRequest represents the request to upload a new file
type CreateFileRequest struct {
	Name             string     `json:"name" binding:"required"` // User-friendly name
	AccessType       *AccessType `json:"access_type"`            // Access control type
	RequiredRole     *string    `json:"required_role"`          // Single role required
	AllowedRoles     []string   `json:"allowed_roles"`          // Multiple roles allowed
	RequiredProductID *uuid.UUID `json:"required_product_id"`    // Single product required
	AllowedProductIDs []uuid.UUID `json:"allowed_product_ids"`   // Multiple products allowed
	PublicLinkExpiresAt *time.Time `json:"public_link_expires_at"` // Expiration for public link
	MaxDownloads     *int       `json:"max_downloads"`          // Max downloads limit
}

// UpdateFileRequest represents the request to update a file
type UpdateFileRequest struct {
	Name              *string     `json:"name"`                        // Optional new name
	AccessType        *AccessType `json:"access_type"`                 // Access control type
	RequiredRole      *string     `json:"required_role"`               // Single role required
	AllowedRoles      []string    `json:"allowed_roles"`               // Multiple roles allowed
	RequiredProductID *uuid.UUID  `json:"required_product_id"`         // Single product required
	AllowedProductIDs []uuid.UUID `json:"allowed_product_ids"`         // Multiple products allowed
	PublicLinkExpiresAt *time.Time `json:"public_link_expires_at"`     // Expiration for public link
	MaxDownloads      *int        `json:"max_downloads"`               // Max downloads limit
}

// UpdateFilePermissionsRequest represents the request to update file permissions
type UpdateFilePermissionsRequest struct {
	AccessType        *AccessType `json:"access_type"`                 // Access control type
	RequiredRole      *string     `json:"required_role"`               // Single role required
	AllowedRoles      []string    `json:"allowed_roles"`               // Multiple roles allowed
	RequiredProductID *uuid.UUID  `json:"required_product_id"`         // Single product required
	AllowedProductIDs []uuid.UUID `json:"allowed_product_ids"`         // Multiple products allowed
	PublicLinkExpiresAt *time.Time `json:"public_link_expires_at"`     // Expiration for public link
	MaxDownloads      *int        `json:"max_downloads"`               // Max downloads limit
}

// FilePermission represents the permission configuration for a file
type FilePermission struct {
	FileID              uuid.UUID  `json:"file_id"`
	AccessType          AccessType `json:"access_type"`
	RequiredRole        *string    `json:"required_role,omitempty"`
	AllowedRoles        []string   `json:"allowed_roles,omitempty"`
	RequiredProductID   *uuid.UUID `json:"required_product_id,omitempty"`
	AllowedProductIDs   []string   `json:"allowed_product_ids,omitempty"`
	PublicLinkToken     uuid.UUID  `json:"public_link_token"`
	PublicLinkExpiresAt *time.Time `json:"public_link_expires_at,omitempty"`
	DownloadCount       int        `json:"download_count"`
	MaxDownloads        *int       `json:"max_downloads,omitempty"`
	PublicLinkURL       string     `json:"public_link_url,omitempty"`
}

// FilePermissionResponse is the DTO for file permission responses (properly typed)
type FilePermissionResponse struct {
	FileID              uuid.UUID  `json:"file_id"`
	AccessType          AccessType `json:"access_type"`
	RequiredRole        *string    `json:"required_role,omitempty"`
	AllowedRoles        []string   `json:"allowed_roles,omitempty"`
	RequiredProductID   *uuid.UUID `json:"required_product_id,omitempty"`
	AllowedProductIDs   []uuid.UUID `json:"allowed_product_ids,omitempty"` // UUIDs, not strings
	PublicLinkToken     uuid.UUID  `json:"public_link_token"`
	PublicLinkExpiresAt *time.Time `json:"public_link_expires_at,omitempty"`
	DownloadCount       int        `json:"download_count"`
	MaxDownloads        *int       `json:"max_downloads,omitempty"`
	PublicLinkURL       string     `json:"public_link_url,omitempty"`
}

// FileAccessLog represents a file access log entry
type FileAccessLog struct {
	ID            uuid.UUID `db:"id" json:"id"`
	FileID        uuid.UUID `db:"file_id" json:"file_id"`
	UserID        uuid.UUID `db:"user_id" json:"user_id"`
	Action        string    `db:"action" json:"action"`
	AccessGranted bool      `db:"access_granted" json:"access_granted"`
	Reason        string    `db:"reason" json:"reason"`
	IPAddress     string    `db:"ip_address" json:"ip_address"`
	UserAgent     string    `db:"user_agent" json:"user_agent"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// ListFileAccessLogsResponse represents paginated access logs
type ListFileAccessLogsResponse struct {
	Logs       []FileAccessLog `json:"logs"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// FileResponse represents the response for file operations
type FileResponse struct {
	ID                    uuid.UUID  `json:"id"`
	Name                  string     `json:"name"`
	FileName              string     `json:"file_name"`
	FileType              FileType   `json:"file_type"`
	Size                  int64      `json:"size"`
	Url                   string     `json:"url"` // Download URL
	CreatedAt             time.Time  `json:"created_at"`
	AccessType            AccessType `json:"access_type"`
	RequiredRole          *string    `json:"required_role,omitempty"`
	AllowedRoles          []string   `json:"allowed_roles,omitempty"`
	RequiredProductID     *uuid.UUID `json:"required_product_id,omitempty"`
	AllowedProductIDs     []string   `json:"allowed_product_ids,omitempty"`
	PublicLinkToken       uuid.UUID  `json:"public_link_token"`
	PublicLinkExpiresAt   *time.Time `json:"public_link_expires_at,omitempty"`
	DownloadCount         int        `json:"download_count"`
	MaxDownloads          *int       `json:"max_downloads,omitempty"`
	PublicLinkURL         string     `json:"public_link_url,omitempty"`
}

// ListFilesResponse represents a paginated list of files

type ListFilesResponse struct {

	Files    []FileResponse `json:"files"`

	Total    int            `json:"total"`

	Page     int            `json:"page"`

	PageSize int            `json:"page_size"`

	TotalPages int          `json:"total_pages"`

}

// FileWithUser represents a file with user information
type FileWithUser struct {
	File
	UserEmail string `db:"user_email" json:"user_email"`
	UserName  string `db:"user_name" json:"user_name"`
}

// FileResponseWithUser represents file response with user info
type FileResponseWithUser struct {
	ID                    uuid.UUID  `json:"id"`
	Name                  string     `json:"name"`
	FileName              string     `json:"file_name"`
	FileType              FileType   `json:"file_type"`
	Size                  int64      `json:"size"`
	Url                   string     `json:"url"`
	CreatedAt             time.Time  `json:"created_at"`
	CreatedBy             uuid.UUID  `json:"created_by"`
	UserEmail             string     `json:"user_email"`
	UserName              string     `json:"user_name"`
	AccessType            AccessType `json:"access_type"`
	RequiredRole          *string    `json:"required_role,omitempty"`
	AllowedRoles          []string   `json:"allowed_roles,omitempty"`
	RequiredProductID     *uuid.UUID `json:"required_product_id,omitempty"`
	AllowedProductIDs     []string   `json:"allowed_product_ids,omitempty"`
	PublicLinkToken       uuid.UUID  `json:"public_link_token"`
	PublicLinkExpiresAt   *time.Time `json:"public_link_expires_at,omitempty"`
	DownloadCount         int        `json:"download_count"`
	MaxDownloads          *int       `json:"max_downloads,omitempty"`
	PublicLinkURL         string     `json:"public_link_url,omitempty"`
}

// ListFilesAdminResponse represents admin file list response
type ListFilesAdminResponse struct {
	Files      []FileResponseWithUser `json:"files"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// OneTimeDownloadToken represents a one-time use download token
type OneTimeDownloadToken struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	FileID        uuid.UUID  `db:"file_id" json:"file_id"`
	UserID        uuid.UUID  `db:"user_id" json:"user_id"`
	Token         uuid.UUID  `db:"token" json:"token"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	ExpiresAt     time.Time  `db:"expires_at" json:"expires_at"`
	UsedAt        *time.Time `db:"used_at" json:"used_at"`
	IsUsed        bool       `db:"is_used" json:"is_used"`
	DownloadCount int        `db:"download_count" json:"download_count"`
	MaxDownloads  int        `db:"max_downloads" json:"max_downloads"`
	IPAddress     string     `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent     string     `db:"user_agent" json:"user_agent,omitempty"`
}

// GenerateOneTimeTokenRequest represents the request to generate a one-time download token
type GenerateOneTimeTokenRequest struct {
	FileID       uuid.UUID  `json:"file_id" binding:"required"`
	ExpiresIn    int        `json:"expires_in_minutes"`   // Token expiration in minutes (default: 15)
	MaxDownloads int        `json:"max_downloads"`        // Max downloads allowed (default: 1)
}

// OneTimeDownloadLinkResponse represents the response with a one-time download link
type OneTimeDownloadLinkResponse struct {
	TokenID       uuid.UUID `json:"token_id"`
	FileID        uuid.UUID `json:"file_id"`
	FileName      string    `json:"file_name"`
	DownloadURL   string    `json:"download_url"`
	ExpiresAt     time.Time `json:"expires_at"`
	MaxDownloads  int       `json:"max_downloads"`
	CreatedAt     time.Time `json:"created_at"`
}

// ValidateTokenResponse represents the response for token validation
type ValidateTokenResponse struct {
	FileID      uuid.UUID `json:"file_id"`
	IsValid     bool      `json:"is_valid"`
	ErrorMessage string   `json:"error_message,omitempty"`
}
