package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
)

// UserRepository handles all database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByID retrieves a user by their ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	// Validate UUID format
	_, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	query := `
		SELECT 
			id,
			email,
			COALESCE(full_name, ''),
			COALESCE(discord_handle, ''),
			COALESCE(whatsapp_phone, ''),
			COALESCE(balance, 0),
			COALESCE(referral_code, ''),
			created_at,
			updated_at,
			COALESCE(username, ''),
			COALESCE(avatar_url, ''),
			COALESCE(is_admin, false),
			COALESCE(preferences, '{"font": "modern", "density": "comfortable", "backgroundFilter": true}')::jsonb,
			pgp_sym_decrypt(cpf_encrypted, current_setting('app.cpf_encryption_key')) as cpf
		FROM users 
		WHERE id = $1
	`

	var user models.User
	var createdAt, updatedAt time.Time
	var preferencesJSON []byte
	
	err = r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.DiscordHandle,
		&user.WhatsAppPhone,
		&user.Balance,
		&user.ReferralCode,
		&createdAt,
		&updatedAt,
		&user.Username,
		&user.AvatarURL,
		&user.IsAdmin,
		&preferencesJSON,
		&user.CPF,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	if len(preferencesJSON) > 0 {
		if err := json.Unmarshal(preferencesJSON, &user.Preferences); err != nil {
			log.Printf("Warning: failed to unmarshal user preferences: %v", err)
		}
	}

	return &user, nil
}

// UpdateUser updates a user's profile information
func (r *UserRepository) UpdateUser(ctx context.Context, userID string, req *models.UpdateUserRequest) error {
	// Validate UUID format
	_, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// Build dynamic update query based on provided fields
	setClauses := []string{}
	args := []interface{}{userID} // First argument is always userID
	argIndex := 2

	if req.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *req.Username)
		argIndex++
	}

	if req.FullName != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", argIndex))
		args = append(args, *req.FullName)
		argIndex++
	}

	if req.DiscordHandle != nil {
		setClauses = append(setClauses, fmt.Sprintf("discord_handle = $%d", argIndex))
		args = append(args, *req.DiscordHandle)
		argIndex++
	}

	if req.WhatsAppPhone != nil {
		setClauses = append(setClauses, fmt.Sprintf("whatsapp_phone = $%d", argIndex))
		args = append(args, *req.WhatsAppPhone)
		argIndex++
	}

	if req.AvatarURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("avatar_url = $%d", argIndex))
		args = append(args, *req.AvatarURL)
		argIndex++
	}

	if req.CPF != nil {
		setClauses = append(setClauses, fmt.Sprintf("cpf_encrypted = pgp_sym_encrypt($%d, current_setting('app.cpf_encryption_key'))", argIndex))
		args = append(args, *req.CPF)
		argIndex++
	}

	if req.Preferences != nil {
		preferencesJSON, err := json.Marshal(req.Preferences)
		if err == nil {
			setClauses = append(setClauses, fmt.Sprintf("preferences = $%d", argIndex))
			args = append(args, preferencesJSON)
			argIndex++
		}
	}
	
	// Note: models.UpdateUserRequest usually doesn't have Email?
	// I might need to check the model definition.
	// If it doesn't, I should add it or use a separate struct for Admin updates.
	// I'll assume for now I can't easily change the struct definition without checking.
	// But the repository method takes UpdateUserRequest.
	// I will check models/user.go later. If needed I'll make a new method UpdateUserAdmin.

	// If no fields to update, return early
	if len(setClauses) == 0 {
		return nil
	}

	// Add updated_at timestamp
	setClauses = append(setClauses, "updated_at = NOW()")
	
	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $1",
		strings.Join(setClauses, ", "),
	)


	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetBalance retrieves the balance for a user
func (r *UserRepository) GetBalance(ctx context.Context, userID string) (float64, error) {
	query := `SELECT balance FROM users WHERE id = $1`
	
	var balance float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get user balance: %w", err)
	}
	
	return balance, nil
}

// UpdateBalance updates the balance for a user within a transaction
func (r *UserRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, userID string, newBalance float64) error {
	query := `UPDATE users SET balance = $1 WHERE id = $2`
	
	// Use the transaction if provided, otherwise use the main database connection
	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}
	
	_, err := execTx.ExecContext(ctx, query, newBalance, userID)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}
	
	return nil
}

// IncrementBalance atomically adds (or subtracts) an amount from the user's balance
func (r *UserRepository) IncrementBalance(ctx context.Context, tx *sql.Tx, userID string, amount float64) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`
	
	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}
	
	_, err := execTx.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to increment user balance: %w", err)
	}
	
	return nil
}

// GetUserByReferralCode retrieves a user by their referral code
func (r *UserRepository) GetUserByReferralCode(ctx context.Context, referralCode string) (*string, error) {
	query := `SELECT id FROM users WHERE referral_code = $1`
	
	var userID string
	err := r.db.QueryRowContext(ctx, query, referralCode).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by referral code: %w", err)
	}
	
	return &userID, nil
}

// ListAll retrieves a list of users with pagination and search
func (r *UserRepository) ListAll(ctx context.Context, page, limit int, search string) ([]models.User, int, error) {
	offset := (page - 1) * limit
	
	// Query with subquery to get highest role
	baseQuery := `SELECT u.id, u.email, COALESCE(u.full_name, ''), COALESCE(u.discord_handle, ''), COALESCE(u.whatsapp_phone, ''), COALESCE(u.balance, 0), COALESCE(u.referral_code, ''), u.created_at, u.updated_at, COALESCE(u.username, ''), COALESCE(u.avatar_url, ''), COALESCE(u.is_admin, false), COALESCE(u.preferences, '{"font": "modern", "density": "comfortable", "backgroundFilter": true}')::jsonb, pgp_sym_decrypt(u.cpf_encrypted, current_setting('app.cpf_encryption_key')),
		(SELECT ur.role FROM user_roles ur WHERE ur.user_id = u.id 
		 ORDER BY CASE ur.role 
		   WHEN 'DIRECTION' THEN 8
		   WHEN 'ENGINEERING' THEN 7
		   WHEN 'DEVELOPMENT' THEN 6
		   WHEN 'ADMIN' THEN 5
		   WHEN 'SUPPORT' THEN 4
		   WHEN 'CLIENT_VIP' THEN 3
		   WHEN 'CLIENT' THEN 2
		   WHEN 'PARTNER' THEN 1
		   ELSE 0 END DESC LIMIT 1) as highest_role
	FROM users u`
	countQuery := `SELECT COUNT(*) FROM users u`
	
	var args []interface{}
	var whereClauses []string
	argIndex := 1

	if search != "" {
		// Encapsulate search pattern for LIKE
		searchPattern := "%" + search + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(u.email ILIKE $%d OR u.full_name ILIKE $%d OR u.username ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	if len(whereClauses) > 0 {
		whereSQL := " WHERE " + strings.Join(whereClauses, " AND ")
		baseQuery += whereSQL
		countQuery += whereSQL
	}

	// Add Order and Limit
	baseQuery += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Get Total Count
	var total int
	// For count query, we only need search args, not limit/offset
	countArgs := args[:len(args)-2] 
	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var createdAt, updatedAt time.Time
		var highestRole *string
		var preferencesJSON []byte
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FullName,
			&user.DiscordHandle,
			&user.WhatsAppPhone,
			&user.Balance,
			&user.ReferralCode,
			&createdAt,
			&updatedAt,
			&user.Username,
			&user.AvatarURL,
			&user.IsAdmin,
			&preferencesJSON,
			&user.CPF,
			&highestRole,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		user.CreatedAt = createdAt
		user.UpdatedAt = updatedAt
		if len(preferencesJSON) > 0 {
			if err := json.Unmarshal(preferencesJSON, &user.Preferences); err != nil {
				log.Printf("Warning: failed to unmarshal user preferences in list: %v", err)
			}
		}
		if highestRole != nil {
			role := models.RoleType(*highestRole)
			user.HighestRole = &role
		}
		users = append(users, user)
	}

	return users, total, nil
}

// UpdatePassword updates the user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	return err
}

// UpdateUserAdmin updates a user's profile information including email (Admin only)
func (r *UserRepository) UpdateUserAdmin(ctx context.Context, userID string, updates map[string]interface{}) error {
	// Validate UUID format
	_, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{userID} // First argument is always userID
	argIndex := 2

	for key, value := range updates {
		if key == "cpf" {
			setClauses = append(setClauses, fmt.Sprintf("cpf_encrypted = pgp_sym_encrypt($%d, current_setting('app.cpf_encryption_key'))", argIndex))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, argIndex))
		}
		args = append(args, value)
		argIndex++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	
	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $1",
		strings.Join(setClauses, ", "),
	)

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user (admin): %w", err)
	}

	return nil
}