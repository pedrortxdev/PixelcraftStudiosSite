package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/pixelcraft/api/internal/models"
)

// RoleRepository handles role-related database operations
type RoleRepository struct {
	db *sql.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// GetUserRoles returns all active roles for a user
func (r *RoleRepository) GetUserRoles(ctx context.Context, userID string) ([]models.RoleType, error) {
	query := `
		SELECT role FROM user_roles 
		WHERE user_id = $1 
		AND (expires_at IS NULL OR expires_at > NOW())
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.RoleType
	for rows.Next() {
		var role models.RoleType
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// GetUserRolesWithDetails returns all active roles with full details
func (r *RoleRepository) GetUserRolesWithDetails(ctx context.Context, userID string) ([]models.UserRole, error) {
	query := `
		SELECT id, user_id, role, granted_at, granted_by, expires_at
		FROM user_roles 
		WHERE user_id = $1 
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY granted_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.UserRole
	for rows.Next() {
		var role models.UserRole
		if err := rows.Scan(&role.ID, &role.UserID, &role.Role, &role.GrantedAt, &role.GrantedBy, &role.ExpiresAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// AddRole adds a role to a user
func (r *RoleRepository) AddRole(ctx context.Context, userID string, role models.RoleType, grantedBy *string, expiresAt *time.Time) error {
	query := `
		INSERT INTO user_roles (user_id, role, granted_by, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role) DO UPDATE SET
			expires_at = EXCLUDED.expires_at,
			granted_at = NOW(),
			granted_by = EXCLUDED.granted_by
	`
	_, err := r.db.ExecContext(ctx, query, userID, role, grantedBy, expiresAt)
	return err
}

// RemoveRole removes a role from a user
func (r *RoleRepository) RemoveRole(ctx context.Context, userID string, role models.RoleType) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role = $2`
	_, err := r.db.ExecContext(ctx, query, userID, role)
	return err
}

// HasRole checks if a user has a specific role
func (r *RoleRepository) HasRole(ctx context.Context, userID string, role models.RoleType) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_roles 
			WHERE user_id = $1 AND role = $2
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, role).Scan(&exists)
	return exists, err
}

// HasAnyRole checks if a user has any of the specified roles
func (r *RoleRepository) HasAnyRole(ctx context.Context, userID string, roles ...models.RoleType) (bool, error) {
	if len(roles) == 0 {
		return false, nil
	}

	// Build query with dynamic parameters
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_roles 
			WHERE user_id = $1 
			AND role = ANY($2)
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, pq.Array(roles)).Scan(&exists)
	return exists, err
}

// GetUsersWithRole returns all users with a specific role
func (r *RoleRepository) GetUsersWithRole(ctx context.Context, role models.RoleType) ([]string, error) {
	query := `
		SELECT user_id FROM user_roles 
		WHERE role = $1 
		AND (expires_at IS NULL OR expires_at > NOW())
	`
	rows, err := r.db.QueryContext(ctx, query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

// CleanExpiredRoles removes all expired roles
func (r *RoleRepository) CleanExpiredRoles(ctx context.Context) (int64, error) {
	query := `DELETE FROM user_roles WHERE expires_at IS NOT NULL AND expires_at <= NOW()`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// UpdateUserSpending updates user's spending tracking
func (r *RoleRepository) UpdateUserSpending(ctx context.Context, userID string, amount float64) error {
	query := `
		UPDATE users SET 
			total_spent = total_spent + $2,
			monthly_spent = CASE 
				WHEN monthly_spent_reset_at < date_trunc('month', NOW()) THEN $2
				ELSE monthly_spent + $2
			END,
			monthly_spent_reset_at = CASE 
				WHEN monthly_spent_reset_at < date_trunc('month', NOW()) THEN NOW()
				ELSE monthly_spent_reset_at
			END
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, userID, amount)
	return err
}

// GetUserSpending returns user's spending information
func (r *RoleRepository) GetUserSpending(ctx context.Context, userID string) (totalSpent, monthlySpent float64, err error) {
	query := `
		SELECT 
			COALESCE(total_spent, 0),
			CASE 
				WHEN monthly_spent_reset_at < date_trunc('month', NOW()) THEN 0
				ELSE COALESCE(monthly_spent, 0)
			END
		FROM users WHERE id = $1
	`
	err = r.db.QueryRowContext(ctx, query, userID).Scan(&totalSpent, &monthlySpent)
	return
}

// GetPartnerUserIDs returns all user IDs with the PARTNER role
func (r *RoleRepository) GetPartnerUserIDs(ctx context.Context) ([]string, error) {
	return r.GetUsersWithRole(ctx, models.RolePartner)
}
