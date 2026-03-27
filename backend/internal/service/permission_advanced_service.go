package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type PermissionAdvancedService struct {
	db *sql.DB
}

func NewPermissionAdvancedService(db *sql.DB) *PermissionAdvancedService {
	return &PermissionAdvancedService{db: db}
}

// PermissionAuditLog representa um log de auditoria
type PermissionAuditLog struct {
	ID          string                 `json:"id"`
	Role        string                 `json:"role"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Operation   string                 `json:"operation"`
	PerformedBy *string                `json:"performed_by"`
	PerformedAt time.Time              `json:"performed_at"`
	OldValue    map[string]interface{} `json:"old_value,omitempty"`
	NewValue    map[string]interface{} `json:"new_value,omitempty"`
	Reason      *string                `json:"reason,omitempty"`
}

// CustomRole representa um cargo customizado
type CustomRole struct {
	ID             string    `json:"id"`
	RoleName       string    `json:"role_name"`
	DisplayName    string    `json:"display_name"`
	Description    string    `json:"description"`
	Color          string    `json:"color"`
	HierarchyLevel int       `json:"hierarchy_level"`
	IsActive       bool      `json:"is_active"`
	CreatedBy      *string   `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PermissionTemplate representa um template de permissões
type PermissionTemplate struct {
	ID           string                 `json:"id"`
	TemplateName string                 `json:"template_name"`
	Description  string                 `json:"description"`
	TemplateData map[string]interface{} `json:"template_data"`
	CreatedBy    *string                `json:"created_by"`
	CreatedAt    time.Time              `json:"created_at"`
	IsPublic     bool                   `json:"is_public"`
}

// PermissionNotification representa uma notificação
type PermissionNotification struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	Role             string    `json:"role"`
	NotificationType string    `json:"notification_type"`
	Message          string    `json:"message"`
	IsRead           bool      `json:"is_read"`
	CreatedAt        time.Time `json:"created_at"`
}

// GetPermissionAuditLog retorna o histórico de mudanças
func (s *PermissionAdvancedService) GetPermissionAuditLog(page, limit int, role string) ([]PermissionAuditLog, int, error) {
	offset := (page - 1) * limit

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if role != "" {
		whereClause += fmt.Sprintf(" AND role = $%d", argCount)
		args = append(args, role)
		argCount++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM permission_audit_log %s", whereClause)
	var total int
	if err := s.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get logs
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id, role, resource, action, operation, performed_by, performed_at, old_value, new_value, reason
		FROM permission_audit_log
		%s
		ORDER BY performed_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []PermissionAuditLog
	for rows.Next() {
		var log PermissionAuditLog
		var oldValueJSON, newValueJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.Role,
			&log.Resource,
			&log.Action,
			&log.Operation,
			&log.PerformedBy,
			&log.PerformedAt,
			&oldValueJSON,
			&newValueJSON,
			&log.Reason,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(oldValueJSON) > 0 {
			json.Unmarshal(oldValueJSON, &log.OldValue)
		}
		if len(newValueJSON) > 0 {
			json.Unmarshal(newValueJSON, &log.NewValue)
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// InheritPermissions herda permissões de um cargo
func (s *PermissionAdvancedService) InheritPermissions(targetRole, sourceRole, performedBy string) (int, error) {
	query := "SELECT inherit_permissions_from_role($1, $2)"
	var count int
	err := s.db.QueryRow(query, targetRole, sourceRole).Scan(&count)
	if err != nil {
		return 0, err
	}

	// Log the operation
	logQuery := `
		INSERT INTO permission_audit_log (role, resource, action, operation, performed_by, reason)
		VALUES ($1, 'ALL', 'INHERIT', 'INHERIT', $2, $3)
	`
	reason := fmt.Sprintf("Inherited %d permissions from %s", count, sourceRole)
	s.db.Exec(logQuery, targetRole, performedBy, reason)

	return count, nil
}

// RemoveInheritedPermissions remove permissões herdadas
func (s *PermissionAdvancedService) RemoveInheritedPermissions(role, performedBy string) (int, error) {
	query := "SELECT remove_inherited_permissions($1)"
	var count int
	err := s.db.QueryRow(query, role).Scan(&count)
	if err != nil {
		return 0, err
	}

	// Log the operation
	logQuery := `
		INSERT INTO permission_audit_log (role, resource, action, operation, performed_by, reason)
		VALUES ($1, 'ALL', 'REMOVE_INHERITED', 'REMOVE', $2, $3)
	`
	reason := fmt.Sprintf("Removed %d inherited permissions", count)
	s.db.Exec(logQuery, role, performedBy, reason)

	return count, nil
}

// CreateCustomRole cria um cargo customizado
func (s *PermissionAdvancedService) CreateCustomRole(roleName, displayName, description, color string, hierarchyLevel int, createdBy string) (*CustomRole, error) {
	if color == "" {
		color = "#999999"
	}

	// Convert createdBy to UUID or use NULL
	var createdByUUID interface{}
	if createdBy != "" {
		createdByUUID = createdBy
	} else {
		createdByUUID = nil
	}

	query := `
		INSERT INTO custom_roles (role_name, display_name, description, color, hierarchy_level, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, role_name, display_name, description, color, hierarchy_level, is_active, created_by, created_at, updated_at
	`

	var role CustomRole
	err := s.db.QueryRow(query, roleName, displayName, description, color, hierarchyLevel, createdByUUID).Scan(
		&role.ID,
		&role.RoleName,
		&role.DisplayName,
		&role.Description,
		&role.Color,
		&role.HierarchyLevel,
		&role.IsActive,
		&role.CreatedBy,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create custom role: %w", err)
	}

	return &role, nil
}

// GetCustomRoles retorna todos os cargos customizados
func (s *PermissionAdvancedService) GetCustomRoles() ([]CustomRole, error) {
	query := `
		SELECT id, role_name, display_name, description, color, hierarchy_level, is_active, created_by, created_at, updated_at
		FROM custom_roles
		WHERE is_active = TRUE
		ORDER BY hierarchy_level DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []CustomRole
	for rows.Next() {
		var role CustomRole
		err := rows.Scan(
			&role.ID,
			&role.RoleName,
			&role.DisplayName,
			&role.Description,
			&role.Color,
			&role.HierarchyLevel,
			&role.IsActive,
			&role.CreatedBy,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// DeleteCustomRole deleta um cargo customizado
func (s *PermissionAdvancedService) DeleteCustomRole(roleID string) error {
	query := "UPDATE custom_roles SET is_active = FALSE WHERE id = $1"
	_, err := s.db.Exec(query, roleID)
	return err
}

// ExportPermissions exporta configurações de permissões
func (s *PermissionAdvancedService) ExportPermissions(roles []string) (map[string]interface{}, error) {
	whereClause := ""
	args := []interface{}{}

	if len(roles) > 0 {
		whereClause = "WHERE role = ANY($1)"
		args = append(args, roles)
	}

	query := fmt.Sprintf(`
		SELECT role, resource, action, is_inherited, inherited_from
		FROM role_permissions
		%s
		ORDER BY role, resource, action
	`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]interface{})
	permissions := make(map[string][]map[string]interface{})

	for rows.Next() {
		var role, resource, action string
		var isInherited bool
		var inheritedFrom *string

		err := rows.Scan(&role, &resource, &action, &isInherited, &inheritedFrom)
		if err != nil {
			return nil, err
		}

		perm := map[string]interface{}{
			"resource":       resource,
			"action":         action,
			"is_inherited":   isInherited,
			"inherited_from": inheritedFrom,
		}

		permissions[role] = append(permissions[role], perm)
	}

	result["permissions"] = permissions
	result["exported_at"] = time.Now()
	result["version"] = "1.0"

	return result, nil
}

// ImportPermissions importa configurações de permissões
func (s *PermissionAdvancedService) ImportPermissions(templateData map[string]interface{}, overwrite bool, performedBy string) (map[string]interface{}, error) {
	permissions, ok := templateData["permissions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid template data format")
	}

	imported := 0
	skipped := 0

	for role, perms := range permissions {
		permList, ok := perms.([]interface{})
		if !ok {
			continue
		}

		for _, p := range permList {
			perm, ok := p.(map[string]interface{})
			if !ok {
				continue
			}

			resource := perm["resource"].(string)
			action := perm["action"].(string)

			if overwrite {
				query := `
					INSERT INTO role_permissions (role, resource, action)
					VALUES ($1, $2, $3)
					ON CONFLICT (role, resource, action) DO UPDATE
					SET is_inherited = FALSE
				`
				_, err := s.db.Exec(query, role, resource, action)
				if err == nil {
					imported++
				}
			} else {
				query := `
					INSERT INTO role_permissions (role, resource, action)
					VALUES ($1, $2, $3)
					ON CONFLICT (role, resource, action) DO NOTHING
				`
				result, err := s.db.Exec(query, role, resource, action)
				if err == nil {
					rows, _ := result.RowsAffected()
					if rows > 0 {
						imported++
					} else {
						skipped++
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"imported": imported,
		"skipped":  skipped,
		"message":  fmt.Sprintf("Imported %d permissions, skipped %d", imported, skipped),
	}, nil
}

// SavePermissionTemplate salva um template
func (s *PermissionAdvancedService) SavePermissionTemplate(name, description string, templateData map[string]interface{}, isPublic bool, createdBy string) (*PermissionTemplate, error) {
	dataJSON, err := json.Marshal(templateData)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO permission_templates (template_name, description, template_data, created_by, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, template_name, description, template_data, created_by, created_at, is_public
	`

	var template PermissionTemplate
	var dataBytes []byte

	err = s.db.QueryRow(query, name, description, dataJSON, createdBy, isPublic).Scan(
		&template.ID,
		&template.TemplateName,
		&template.Description,
		&dataBytes,
		&template.CreatedBy,
		&template.CreatedAt,
		&template.IsPublic,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(dataBytes, &template.TemplateData)
	return &template, nil
}

// GetPermissionTemplates retorna todos os templates
func (s *PermissionAdvancedService) GetPermissionTemplates() ([]PermissionTemplate, error) {
	query := `
		SELECT id, template_name, description, template_data, created_by, created_at, is_public
		FROM permission_templates
		WHERE is_public = TRUE
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []PermissionTemplate
	for rows.Next() {
		var template PermissionTemplate
		var dataBytes []byte

		err := rows.Scan(
			&template.ID,
			&template.TemplateName,
			&template.Description,
			&dataBytes,
			&template.CreatedBy,
			&template.CreatedAt,
			&template.IsPublic,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(dataBytes, &template.TemplateData)
		templates = append(templates, template)
	}

	return templates, nil
}

// GetUserNotifications retorna notificações do usuário
func (s *PermissionAdvancedService) GetUserNotifications(userID string) ([]PermissionNotification, error) {
	query := `
		SELECT id, user_id, role, notification_type, message, is_read, created_at
		FROM permission_notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []PermissionNotification
	for rows.Next() {
		var notif PermissionNotification
		err := rows.Scan(
			&notif.ID,
			&notif.UserID,
			&notif.Role,
			&notif.NotificationType,
			&notif.Message,
			&notif.IsRead,
			&notif.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}

// MarkNotificationAsRead marca notificação como lida
func (s *PermissionAdvancedService) MarkNotificationAsRead(notificationID string) error {
	query := "UPDATE permission_notifications SET is_read = TRUE WHERE id = $1"
	_, err := s.db.Exec(query, notificationID)
	return err
}

// GetPermissionDashboard retorna estatísticas
func (s *PermissionAdvancedService) GetPermissionDashboard() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total permissions
	var totalPerms int
	s.db.QueryRow("SELECT COUNT(*) FROM role_permissions").Scan(&totalPerms)
	stats["total_permissions"] = totalPerms

	// Inherited permissions
	var inheritedPerms int
	s.db.QueryRow("SELECT COUNT(*) FROM role_permissions WHERE is_inherited = TRUE").Scan(&inheritedPerms)
	stats["inherited_permissions"] = inheritedPerms

	// Custom roles
	var customRoles int
	s.db.QueryRow("SELECT COUNT(*) FROM custom_roles WHERE is_active = TRUE").Scan(&customRoles)
	stats["custom_roles"] = customRoles

	// Recent changes (last 7 days)
	var recentChanges int
	s.db.QueryRow("SELECT COUNT(*) FROM permission_audit_log WHERE performed_at > NOW() - INTERVAL '7 days'").Scan(&recentChanges)
	stats["recent_changes"] = recentChanges

	// Permissions by role
	rows, err := s.db.Query(`
		SELECT role, COUNT(*) as count
		FROM role_permissions
		GROUP BY role
		ORDER BY count DESC
	`)
	if err == nil {
		defer rows.Close()
		permsByRole := make(map[string]int)
		for rows.Next() {
			var role string
			var count int
			rows.Scan(&role, &count)
			permsByRole[role] = count
		}
		stats["permissions_by_role"] = permsByRole
	}

	return stats, nil
}
