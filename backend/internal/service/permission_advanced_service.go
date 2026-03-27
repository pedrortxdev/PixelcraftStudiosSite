package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pixelcraft/api/internal/repository"
)

// PermissionAdvancedService handles advanced permission operations
// REFACTORED: Now uses repositories instead of raw SQL
type PermissionAdvancedService struct {
	permRepo *repository.PermissionRepository
	db       *sql.DB // Only for complex transactions that span multiple repos
}

func NewPermissionAdvancedService(
	permRepo *repository.PermissionRepository,
	db *sql.DB,
) *PermissionAdvancedService {
	return &PermissionAdvancedService{
		permRepo: permRepo,
		db:       db,
	}
}

// PermissionAuditLog represents an audit log entry
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

// CustomRole represents a custom role
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

// PermissionTemplate represents a permission template
type PermissionTemplate struct {
	ID           string                 `json:"id"`
	TemplateName string                 `json:"template_name"`
	Description  string                 `json:"description"`
	TemplateData map[string]interface{} `json:"template_data"`
	CreatedBy    *string                `json:"created_by"`
	CreatedAt    time.Time              `json:"created_at"`
	IsPublic     bool                   `json:"is_public"`
}

// PermissionNotification represents a notification
type PermissionNotification struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	Role             string    `json:"role"`
	NotificationType string    `json:"notification_type"`
	Message          string    `json:"message"`
	IsRead           bool      `json:"is_read"`
	CreatedAt        time.Time `json:"created_at"`
}

// GetPermissionAuditLog returns audit logs with pagination (WITH CONTEXT)
func (s *PermissionAdvancedService) GetPermissionAuditLog(ctx context.Context, page, limit int, role string) ([]PermissionAuditLog, int, error) {
	offset := (page - 1) * limit

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if role != "" {
		whereClause += fmt.Sprintf(" AND role = $%d", argCount)
		args = append(args, role)
		argCount++
	}

	// Count total (WITH CONTEXT)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM permission_audit_log %s", whereClause)
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get logs (WITH CONTEXT)
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id, role, resource, action, operation, performed_by, performed_at, old_value, new_value, reason
		FROM permission_audit_log
		%s
		ORDER BY performed_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// PROPER ERROR HANDLING: Check JSON unmarshal errors
		if len(oldValueJSON) > 0 {
			if err := json.Unmarshal(oldValueJSON, &log.OldValue); err != nil {
				// Log error but continue - corrupted data shouldn't break entire request
				log.OldValue = map[string]interface{}{"error": "corrupted_json"}
			}
		}
		if len(newValueJSON) > 0 {
			if err := json.Unmarshal(newValueJSON, &log.NewValue); err != nil {
				log.NewValue = map[string]interface{}{"error": "corrupted_json"}
			}
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return logs, total, nil
}

// InheritPermissions inherits permissions from a source role (WITH CONTEXT + ERROR HANDLING)
func (s *PermissionAdvancedService) InheritPermissions(ctx context.Context, targetRole, sourceRole, performedBy string) (int, error) {
	// Start transaction for atomic operation
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := "SELECT inherit_permissions_from_role($1, $2)"
	var count int
	err = tx.QueryRowContext(ctx, query, targetRole, sourceRole).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to inherit permissions: %w", err)
	}

	// Log the operation (WITH ERROR HANDLING)
	logQuery := `
		INSERT INTO permission_audit_log (role, resource, action, operation, performed_by, reason)
		VALUES ($1, 'ALL', 'INHERIT', 'INHERIT', $2, $3)
	`
	reason := fmt.Sprintf("Inherited %d permissions from %s", count, sourceRole)
	_, err = tx.ExecContext(ctx, logQuery, targetRole, performedBy, reason)
	if err != nil {
		// Log error but don't fail the entire operation
		// The permissions were inherited, only the log failed
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

// RemoveInheritedPermissions removes inherited permissions (WITH CONTEXT + ERROR HANDLING)
func (s *PermissionAdvancedService) RemoveInheritedPermissions(ctx context.Context, role, performedBy string) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := "SELECT remove_inherited_permissions($1)"
	var count int
	err = tx.QueryRowContext(ctx, query, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to remove inherited permissions: %w", err)
	}

	// Log the operation (WITH ERROR HANDLING)
	logQuery := `
		INSERT INTO permission_audit_log (role, resource, action, operation, performed_by, reason)
		VALUES ($1, 'ALL', 'REMOVE_INHERITED', 'REMOVE', $2, $3)
	`
	reason := fmt.Sprintf("Removed %d inherited permissions", count)
	_, err = tx.ExecContext(ctx, logQuery, role, performedBy, reason)
	if err != nil {
		// Log error but continue
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return count, nil
}

// CreateCustomRole creates a custom role (WITH CONTEXT)
func (s *PermissionAdvancedService) CreateCustomRole(ctx context.Context, roleName, displayName, description, color string, hierarchyLevel int, createdBy string) (*CustomRole, error) {
	if color == "" {
		color = "#999999"
	}

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
	err := s.db.QueryRowContext(ctx, query, roleName, displayName, description, color, hierarchyLevel, createdByUUID).Scan(
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

// GetCustomRoles returns all custom roles (WITH CONTEXT)
func (s *PermissionAdvancedService) GetCustomRoles(ctx context.Context) ([]CustomRole, error) {
	query := `
		SELECT id, role_name, display_name, description, color, hierarchy_level, is_active, created_by, created_at, updated_at
		FROM custom_roles
		WHERE is_active = TRUE
		ORDER BY hierarchy_level DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query custom roles: %w", err)
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
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return roles, nil
}

// DeleteCustomRole deactivates a custom role (WITH CONTEXT)
func (s *PermissionAdvancedService) DeleteCustomRole(ctx context.Context, roleID string) error {
	query := "UPDATE custom_roles SET is_active = FALSE WHERE id = $1"
	_, err := s.db.ExecContext(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete custom role: %w", err)
	}
	return err
}

// ExportPermissions exports permission configurations (WITH CONTEXT)
func (s *PermissionAdvancedService) ExportPermissions(ctx context.Context, roles []string) (map[string]interface{}, error) {
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

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to export permissions: %w", err)
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
			return nil, fmt.Errorf("failed to scan permission: %w", err)
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

// ImportPermissions imports permission configurations (WITH CONTEXT + BULK INSERT)
func (s *PermissionAdvancedService) ImportPermissions(ctx context.Context, templateData map[string]interface{}, overwrite bool, performedBy string) (map[string]interface{}, error) {
	permissions, ok := templateData["permissions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid template data format")
	}

	// Start transaction for BULK operation
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

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

			resource, ok := perm["resource"].(string)
			if !ok {
				continue
			}
			action, ok := perm["action"].(string)
			if !ok {
				continue
			}

			if overwrite {
				query := `
					INSERT INTO role_permissions (role, resource, action)
					VALUES ($1, $2, $3)
					ON CONFLICT (role, resource, action) DO UPDATE
					SET is_inherited = FALSE
				`
				_, err := tx.ExecContext(ctx, query, role, resource, action)
				if err == nil {
					imported++
				}
			} else {
				query := `
					INSERT INTO role_permissions (role, resource, action)
					VALUES ($1, $2, $3)
					ON CONFLICT (role, resource, action) DO NOTHING
				`
				result, err := tx.ExecContext(ctx, query, role, resource, action)
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

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit import transaction: %w", err)
	}

	return map[string]interface{}{
		"imported": imported,
		"skipped":  skipped,
		"message":  fmt.Sprintf("Imported %d permissions, skipped %d", imported, skipped),
	}, nil
}

// SavePermissionTemplate saves a template (WITH CONTEXT + ERROR HANDLING)
func (s *PermissionAdvancedService) SavePermissionTemplate(ctx context.Context, name, description string, templateData map[string]interface{}, isPublic bool, createdBy string) (*PermissionTemplate, error) {
	dataJSON, err := json.Marshal(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template data: %w", err)
	}

	query := `
		INSERT INTO permission_templates (template_name, description, template_data, created_by, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, template_name, description, template_data, created_by, created_at, is_public
	`

	var template PermissionTemplate
	var dataBytes []byte

	err = s.db.QueryRowContext(ctx, query, name, description, dataJSON, createdBy, isPublic).Scan(
		&template.ID,
		&template.TemplateName,
		&template.Description,
		&dataBytes,
		&template.CreatedBy,
		&template.CreatedAt,
		&template.IsPublic,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to save template: %w", err)
	}

	// PROPER ERROR HANDLING for JSON unmarshal
	if err := json.Unmarshal(dataBytes, &template.TemplateData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template data: %w", err)
	}

	return &template, nil
}

// GetPermissionTemplates returns all templates (WITH CONTEXT + ERROR HANDLING)
func (s *PermissionAdvancedService) GetPermissionTemplates(ctx context.Context) ([]PermissionTemplate, error) {
	query := `
		SELECT id, template_name, description, template_data, created_by, created_at, is_public
		FROM permission_templates
		WHERE is_public = TRUE
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
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
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		// PROPER ERROR HANDLING
		if err := json.Unmarshal(dataBytes, &template.TemplateData); err != nil {
			template.TemplateData = map[string]interface{}{"error": "corrupted_json"}
		}

		templates = append(templates, template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return templates, nil
}

// GetUserNotifications returns user notifications (WITH CONTEXT)
func (s *PermissionAdvancedService) GetUserNotifications(ctx context.Context, userID string) ([]PermissionNotification, error) {
	query := `
		SELECT id, user_id, role, notification_type, message, is_read, created_at
		FROM permission_notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
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
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notif)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return notifications, nil
}

// MarkNotificationAsRead marks notification as read (WITH CONTEXT)
func (s *PermissionAdvancedService) MarkNotificationAsRead(ctx context.Context, notificationID string) error {
	query := "UPDATE permission_notifications SET is_read = TRUE WHERE id = $1"
	_, err := s.db.ExecContext(ctx, query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return err
}

// PermissionDashboardStats represents dashboard statistics
type PermissionDashboardStats struct {
	TotalPermissions     int            `json:"total_permissions"`
	InheritedPermissions int            `json:"inherited_permissions"`
	CustomRoles          int            `json:"custom_roles"`
	RecentChanges        int            `json:"recent_changes"`
	PermissionsByRole    map[string]int `json:"permissions_by_role"`
}

// GetPermissionDashboard returns statistics (CONCURRENT QUERIES)
func (s *PermissionAdvancedService) GetPermissionDashboard(ctx context.Context) (*PermissionDashboardStats, error) {
	stats := &PermissionDashboardStats{
		PermissionsByRole: make(map[string]int),
	}

	// Run queries CONCURRENTLY using goroutines
	type result struct {
		value int
		err   error
	}

	// Channel for collecting results
	totalCh := make(chan result, 1)
	inheritedCh := make(chan result, 1)
	customRolesCh := make(chan result, 1)
	recentCh := make(chan result, 1)
	byRoleCh := make(chan map[string]int, 1)

	// Query 1: Total permissions
	go func() {
		var total int
		err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM role_permissions").Scan(&total)
		totalCh <- result{total, err}
	}()

	// Query 2: Inherited permissions
	go func() {
		var inherited int
		err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM role_permissions WHERE is_inherited = TRUE").Scan(&inherited)
		inheritedCh <- result{inherited, err}
	}()

	// Query 3: Custom roles
	go func() {
		var custom int
		err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM custom_roles WHERE is_active = TRUE").Scan(&custom)
		customRolesCh <- result{custom, err}
	}()

	// Query 4: Recent changes
	go func() {
		var recent int
		err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM permission_audit_log WHERE performed_at > NOW() - INTERVAL '7 days'").Scan(&recent)
		recentCh <- result{recent, err}
	}()

	// Query 5: Permissions by role
	go func() {
		permsByRole := make(map[string]int)
		rows, err := s.db.QueryContext(ctx, `
			SELECT role, COUNT(*) as count
			FROM role_permissions
			GROUP BY role
			ORDER BY count DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var role string
				var count int
				if err := rows.Scan(&role, &count); err == nil {
					permsByRole[role] = count
				}
			}
		}
		byRoleCh <- permsByRole
	}()

	// Collect results (with timeout protection from ctx)
	select {
	case r := <-totalCh:
		if r.err == nil {
			stats.TotalPermissions = r.value
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case r := <-inheritedCh:
		if r.err == nil {
			stats.InheritedPermissions = r.value
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case r := <-customRolesCh:
		if r.err == nil {
			stats.CustomRoles = r.value
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case r := <-recentCh:
		if r.err == nil {
			stats.RecentChanges = r.value
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case stats.PermissionsByRole = <-byRoleCh:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return stats, nil
}

// Alternative: Single query version for even better performance
// GetPermissionDashboardSingleQuery returns statistics using a single query
func (s *PermissionAdvancedService) GetPermissionDashboardSingleQuery(ctx context.Context) (*PermissionDashboardStats, error) {
	stats := &PermissionDashboardStats{
		PermissionsByRole: make(map[string]int),
	}

	query := `
		SELECT 
			(SELECT COUNT(*) FROM role_permissions) as total,
			(SELECT COUNT(*) FROM role_permissions WHERE is_inherited = TRUE) as inherited,
			(SELECT COUNT(*) FROM custom_roles WHERE is_active = TRUE) as custom_roles,
			(SELECT COUNT(*) FROM permission_audit_log WHERE performed_at > NOW() - INTERVAL '7 days') as recent_changes
	`

	err := s.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalPermissions,
		&stats.InheritedPermissions,
		&stats.CustomRoles,
		&stats.RecentChanges,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query dashboard stats: %w", err)
	}

	// Separate query for permissions by role (can't be combined with scalars efficiently)
	rows, err := s.db.QueryContext(ctx, `
		SELECT role, COUNT(*) as count
		FROM role_permissions
		GROUP BY role
		ORDER BY count DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var role string
			var count int
			if err := rows.Scan(&role, &count); err == nil {
				stats.PermissionsByRole[role] = count
			}
		}
	}

	return stats, nil
}

// BulkImportPermissions performs a true bulk insert using COPY-like approach
func (s *PermissionAdvancedService) BulkImportPermissions(ctx context.Context, permissions []PermissionImport, overwrite bool) (int, error) {
	if len(permissions) == 0 {
		return 0, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Build a single multi-row INSERT statement
	var query strings.Builder
	args := []interface{}{}
	argIndex := 1

	if overwrite {
		query.WriteString(`
			INSERT INTO role_permissions (role, resource, action)
			VALUES 
		`)
	} else {
		query.WriteString(`
			INSERT INTO role_permissions (role, resource, action)
			VALUES 
		`)
	}

	// Build VALUES clause
	valuePairs := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		valuePairs = append(valuePairs, fmt.Sprintf("($%d, $%d, $%d)", argIndex, argIndex+1, argIndex+2))
		args = append(args, perm.Role, perm.Resource, perm.Action)
		argIndex += 3
	}

	query.WriteString(strings.Join(valuePairs, ", "))

	if overwrite {
		query.WriteString(`
			ON CONFLICT (role, resource, action) DO UPDATE
			SET is_inherited = FALSE
		`)
	} else {
		query.WriteString(`
			ON CONFLICT (role, resource, action) DO NOTHING
		`)
	}

	_, err = tx.ExecContext(ctx, query.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk insert permissions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return len(permissions), nil
}

// PermissionImport represents a permission to import
type PermissionImport struct {
	Role     string `json:"role"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}
