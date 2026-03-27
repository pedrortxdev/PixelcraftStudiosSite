package repository

import (
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

// GetUserPermissions retorna todas as permissões de um usuário baseado em seus cargos
func (r *PermissionRepository) GetUserPermissions(userID string) (*models.UserPermissions, error) {
	// Buscar cargos do usuário
	rolesQuery := `
		SELECT DISTINCT role 
		FROM user_roles 
		WHERE user_id = $1 
		AND (expires_at IS NULL OR expires_at > NOW())
	`

	rows, err := r.db.Query(rolesQuery, userID)
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
		// Usuário sem cargos - retorna permissões vazias
		return &models.UserPermissions{
			UserID:      userID,
			Roles:       []string{},
			Permissions: make(map[models.ResourceType][]models.ActionType),
		}, nil
	}

	// Buscar permissões dos cargos
	permQuery := `
		SELECT DISTINCT resource, action 
		FROM role_permissions 
		WHERE role = ANY($1)
	`

	permRows, err := r.db.Query(permQuery, pq.Array(roles))
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

// GetRolePermissions retorna todas as permissões de um cargo específico
func (r *PermissionRepository) GetRolePermissions(role string) ([]models.RolePermission, error) {
	query := `
		SELECT id, role, resource, action, created_at
		FROM role_permissions
		WHERE role = $1
		ORDER BY resource, action
	`

	rows, err := r.db.Query(query, role)
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

// AddRolePermission adiciona uma permissão a um cargo
func (r *PermissionRepository) AddRolePermission(role string, resource models.ResourceType, action models.ActionType) error {
	query := `
		INSERT INTO role_permissions (role, resource, action)
		VALUES ($1, $2, $3)
		ON CONFLICT (role, resource, action) DO NOTHING
	`

	_, err := r.db.Exec(query, role, resource, action)
	return err
}

// RemoveRolePermission remove uma permissão de um cargo
func (r *PermissionRepository) RemoveRolePermission(role string, resource models.ResourceType, action models.ActionType) error {
	query := `
		DELETE FROM role_permissions
		WHERE role = $1 AND resource = $2 AND action = $3
	`

	_, err := r.db.Exec(query, role, resource, action)
	return err
}

// GetAllRolePermissions retorna todas as permissões de todos os cargos
func (r *PermissionRepository) GetAllRolePermissions() (map[string][]models.RolePermission, error) {
	query := `
		SELECT id, role, resource, action, created_at
		FROM role_permissions
		ORDER BY role, resource, action
	`

	rows, err := r.db.Query(query)
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

// LogEmail registra um email enviado
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

	err = r.db.QueryRow(
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

// GetEmailLogs retorna o histórico de emails com paginação e filtros
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
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
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

	rows, err := r.db.Query(query, args...)
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

// GetEmailLogByID retorna um email específico por ID
func (r *PermissionRepository) GetEmailLogByID(id string) (*models.EmailLog, error) {
	query := `
		SELECT id, from_email, to_email, subject, body, status, error_message, sent_by, sent_at, message_id, metadata
		FROM email_logs
		WHERE id = $1
	`

	var log models.EmailLog
	var metadataJSON []byte

	err := r.db.QueryRow(query, id).Scan(
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
