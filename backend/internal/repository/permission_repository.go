package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"github.com/pixelcraft/api/internal/models"
)

type PermissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository(db *sql.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// GetUserPermissions returns all permissions for a user based on their roles (WITH CONTEXT)
func (r *PermissionRepository) GetUserPermissions(ctx context.Context, userID string) (*models.UserPermissions, error) {
	// Get user roles
	rolesQuery := `
		SELECT DISTINCT role
		FROM user_roles
		WHERE user_id = $1
		AND (expires_at IS NULL OR expires_at > NOW())
	`

	rows, err := r.db.QueryContext(ctx, rolesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if len(roles) == 0 {
		// User without roles - return empty permissions
		return &models.UserPermissions{
			UserID:      userID,
			Roles:       []string{},
			Permissions: make(map[models.ResourceType][]models.ActionType),
		}, nil
	}

	// Get role permissions
	permQuery := `
		SELECT DISTINCT resource, action
		FROM role_permissions
		WHERE role = ANY($1)
	`

	permRows, err := r.db.QueryContext(ctx, permQuery, pq.Array(roles))
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer permRows.Close()

	permissions := make(map[models.ResourceType][]models.ActionType)
	for permRows.Next() {
		var resource models.ResourceType
		var action models.ActionType
		if err := permRows.Scan(&resource, &action); err != nil {
			return nil, err
		}

		permissions[resource] = append(permissions[resource], action)
	}

	return &models.UserPermissions{
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

// HasPermission checks if a user has a SPECIFIC permission (FAST - SELECT EXISTS)
// This is the CORRECT way - returns single boolean, no memory waste
func (r *PermissionRepository) HasPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM user_roles ur
			JOIN role_permissions rp ON ur.role = rp.role
			WHERE ur.user_id = $1
			AND rp.resource = $2
			AND rp.action = $3
			AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
			LIMIT 1
		)
	`

	var hasPerm bool
	err := r.db.QueryRowContext(ctx, query, userID, resource, action).Scan(&hasPerm)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return hasPerm, nil
}

// GetRolePermissions returns all permissions for a specific role (WITH CONTEXT)
func (r *PermissionRepository) GetRolePermissions(ctx context.Context, role string) ([]models.RolePermission, error) {
	query := `
		SELECT id, role, resource, action, created_at
		FROM role_permissions
		WHERE role = $1
		ORDER BY resource, action
	`

	rows, err := r.db.QueryContext(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.RolePermission
	for rows.Next() {
		var perm models.RolePermission
		if err := rows.Scan(&perm.ID, &perm.Role, &perm.Resource, &perm.Action, &perm.CreatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// AddRolePermission adds a permission to a role (WITH CONTEXT)
func (r *PermissionRepository) AddRolePermission(ctx context.Context, role string, resource models.ResourceType, action models.ActionType) error {
	query := `
		INSERT INTO role_permissions (role, resource, action)
		VALUES ($1, $2, $3)
		ON CONFLICT (role, resource, action) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, role, resource, action)
	return err
}

// RemoveRolePermission removes a permission from a role (WITH CONTEXT)
func (r *PermissionRepository) RemoveRolePermission(ctx context.Context, role string, resource models.ResourceType, action models.ActionType) error {
	query := `
		DELETE FROM role_permissions
		WHERE role = $1 AND resource = $2 AND action = $3
	`

	_, err := r.db.ExecContext(ctx, query, role, resource, action)
	return err
}

// GetAllRolePermissions returns all permissions for all roles (WITH CONTEXT)
func (r *PermissionRepository) GetAllRolePermissions(ctx context.Context) (map[string][]models.RolePermission, error) {
	query := `
		SELECT id, role, resource, action, created_at
		FROM role_permissions
		ORDER BY role, resource, action
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]models.RolePermission)
	for rows.Next() {
		var perm models.RolePermission
		if err := rows.Scan(&perm.ID, &perm.Role, &perm.Resource, &perm.Action, &perm.CreatedAt); err != nil {
			return nil, err
		}
		result[perm.Role] = append(result[perm.Role], perm)
	}

	return result, nil
}

// AssignRoleToUser assigns a role to a user (WITH CONTEXT)
func (r *PermissionRepository) AssignRoleToUser(ctx context.Context, userID, role string) error {
	query := `
		INSERT INTO user_roles (user_id, role, assigned_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, role) DO UPDATE SET expires_at = NULL
	`

	_, err := r.db.ExecContext(ctx, query, userID, role)
	return err
}

// RemoveRoleFromUser removes a role from a user (WITH CONTEXT)
func (r *PermissionRepository) RemoveRoleFromUser(ctx context.Context, userID, role string) error {
	query := `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role = $2
	`

	_, err := r.db.ExecContext(ctx, query, userID, role)
	return err
}

// GetUserRoles returns all roles for a user (WITH CONTEXT)
func (r *PermissionRepository) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT role
		FROM user_roles
		WHERE user_id = $1
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY role
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// Email logging methods moved to EmailRepository (SRP violation fix)
// These methods are deprecated and will be removed in future versions
// Use EmailRepository.LogEmail, GetEmailLogs, GetEmailLogByID instead

// LogEmail registra um email enviado (DEPRECATED - use EmailRepository)
func (r *PermissionRepository) LogEmail(log *models.EmailLog) error {
	metadataJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO email_logs (from_email, to_email, subject, body, status, error_message, sent_by, message_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, sent_at
	`

	err = r.db.QueryRowContext(context.Background(),
		query,
		log.FromEmail,
		log.ToEmail,
		log.Subject,
		log.Body,
		log.Status,
		log.ErrorMessage,
		log.SentBy,
		log.MessageID,
		metadataJSON,
	).Scan(&log.ID, &log.SentAt)

	return err
}

// GetEmailLogs retorna o histórico de emails (DEPRECATED - use EmailRepository)
func (r *PermissionRepository) GetEmailLogs(page, limit int, filters map[string]string) ([]models.EmailLog, int, error) {
	offset := (page - 1) * limit

	// Construir query com filtros
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if from, ok := filters["from"]; ok && from != "" {
		whereClause += fmt.Sprintf(" AND from_email ILIKE $%d", argCount)
		args = append(args, "%"+from+"%")
		argCount++
	}

	if to, ok := filters["to"]; ok && to != "" {
		whereClause += fmt.Sprintf(" AND to_email ILIKE $%d", argCount)
		args = append(args, "%"+to+"%")
		argCount++
	}

	if status, ok := filters["status"]; ok && status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	if sentBy, ok := filters["sent_by"]; ok && sentBy != "" {
		whereClause += fmt.Sprintf(" AND sent_by = $%d", argCount)
		args = append(args, sentBy)
		argCount++
	}

	// Contar total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM email_logs %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(context.Background(), countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Buscar logs
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id, from_email, to_email, subject, body, status, error_message, sent_by, sent_at, message_id, metadata
		FROM email_logs
		%s
		ORDER BY sent_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	rows, err := r.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.EmailLog
	for rows.Next() {
		var log models.EmailLog
		var metadataJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.FromEmail,
			&log.ToEmail,
			&log.Subject,
			&log.Body,
			&log.Status,
			&log.ErrorMessage,
			&log.SentBy,
			&log.SentAt,
			&log.MessageID,
			&metadataJSON,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
				log.Metadata = nil
			}
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetEmailLogByID retorna um email específico por ID (DEPRECATED - use EmailRepository)
func (r *PermissionRepository) GetEmailLogByID(id string) (*models.EmailLog, error) {
	query := `
		SELECT id, from_email, to_email, subject, body, status, error_message, sent_by, sent_at, message_id, metadata
		FROM email_logs
		WHERE id = $1
	`

	var log models.EmailLog
	var metadataJSON []byte

	err := r.db.QueryRowContext(context.Background(), query, id).Scan(
		&log.ID,
		&log.FromEmail,
		&log.ToEmail,
		&log.Subject,
		&log.Body,
		&log.Status,
		&log.ErrorMessage,
		&log.SentBy,
		&log.SentAt,
		&log.MessageID,
		&metadataJSON,
	)

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
			log.Metadata = nil
		}
	}

	return &log, nil
}
