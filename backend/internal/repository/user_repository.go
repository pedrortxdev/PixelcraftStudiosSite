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
			pgp_sym_decrypt(cpf_encrypted, current_setting('app.cpf_encryption_key')) as cpf,
			COALESCE(total_spent, 0),
			COALESCE(monthly_spent, 0)
		FROM users 
		WHERE id = $1
	`

	var user models.User
	var createdAt, updatedAt time.Time
	var preferencesJSON []byte
	var balanceFloat, totalSpentFloat, monthlySpentFloat float64
	
	err = r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.DiscordHandle,
		&user.WhatsAppPhone,
		&balanceFloat,
		&user.ReferralCode,
		&createdAt,
		&updatedAt,
		&user.Username,
		&user.AvatarURL,
		&user.IsAdmin,
		&preferencesJSON,
		&user.CPF,
		&totalSpentFloat,
		&monthlySpentFloat,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	user.Balance = int64(balanceFloat * 100)
	user.TotalSpent = int64(totalSpentFloat * 100)
	user.MonthlySpent = int64(monthlySpentFloat * 100)

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
		log.Printf("ERROR: Failed to update user %s: %v | Query: %s", userID, err, query)
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetBalance retrieves the balance for a user (in cents)
func (r *UserRepository) GetBalance(ctx context.Context, userID string) (int64, error) {
	query := `SELECT balance FROM users WHERE id = $1`

	var balanceFloat float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balanceFloat)
	if err != nil {
		return 0, fmt.Errorf("failed to get user balance: %w", err)
	}

	return int64(balanceFloat * 100), nil
}

// GetBalanceTx retrieves the balance for a user within a transaction with row-level lock (in cents)
func (r *UserRepository) GetBalanceTx(ctx context.Context, tx *sql.Tx, userID string) (int64, error) {
	query := `SELECT balance FROM users WHERE id = $1 FOR UPDATE`

	var balanceFloat float64
	var execTx interface {
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}
	err := execTx.QueryRowContext(ctx, query, userID).Scan(&balanceFloat)
	if err != nil {
		return 0, fmt.Errorf("failed to get user balance (tx): %w", err)
	}

	return int64(balanceFloat * 100), nil
}

// UpdateBalance updates the balance for a user within a transaction (balance in cents)
func (r *UserRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, userID string, newBalance int64) error {
	query := `UPDATE users SET balance = $1 WHERE id = $2`

	// Convert cents to decimal for DB
	balanceDecimal := float64(newBalance) / 100

	// Use the transaction if provided, otherwise use the main database connection
	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}

	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	_, err := execTx.ExecContext(ctx, query, balanceDecimal, userID)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}

// IncrementBalance atomically adds (or subtracts) an amount from the user's balance (amount in cents)
func (r *UserRepository) IncrementBalance(ctx context.Context, tx *sql.Tx, userID string, amount int64) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`

	// Convert cents to decimal for DB
	amountDecimal := float64(amount) / 100

	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}

	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	_, err := execTx.ExecContext(ctx, query, amountDecimal, userID)
	if err != nil {
		return fmt.Errorf("failed to increment user balance: %w", err)
	}
	
	return nil
}

// BatchIncrementBalances atomically increments balance for multiple users in a single query.
// userIDs and amountsCents must have the same length. Amounts are in cents.
func (r *UserRepository) BatchIncrementBalances(ctx context.Context, tx *sql.Tx, userIDs []string, amountsCents []int64) error {
	if len(userIDs) != len(amountsCents) {
		return fmt.Errorf("userIDs and amountsCents must have the same length")
	}
	if len(userIDs) == 0 {
		return nil
	}

	// Build a multi-row VALUES clause for a single UPDATE
	// UPDATE users SET balance = balance + v.amount FROM (VALUES ($1,$2), ($3,$4), ...) AS v(id, amount) WHERE users.id = v.id::uuid
	valuesClauses := make([]string, 0, len(userIDs))
	args := make([]interface{}, 0, len(userIDs)*2)
	for i, uid := range userIDs {
		amountDecimal := float64(amountsCents[i]) / 100
		argIdx := i * 2
		valuesClauses = append(valuesClauses, fmt.Sprintf("($%d, $%d)", argIdx+1, argIdx+2))
		args = append(args, uid, amountDecimal)
	}

	query := fmt.Sprintf(`
		UPDATE users 
		SET balance = balance + v.amount 
		FROM (VALUES %s) AS v(id, amount) 
		WHERE users.id = v.id::uuid
	`, strings.Join(valuesClauses, ", "))

	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	_, err := execTx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to batch increment balances: %w", err)
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

// EmailExists checks if an email already exists in the database
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User, passwordHash string, referredByCode *string) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, full_name, referral_code, referred_by_code, balance, discord_handle, whatsapp_phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, email, full_name, referral_code, balance, is_admin, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	var balanceFloat float64
	
	// Convert cents to decimal for DB
	balanceDecimal := float64(user.Balance) / 100

	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Email, user.Username, passwordHash, user.FullName,
		user.ReferralCode, referredByCode, balanceDecimal, user.DiscordHandle, user.WhatsAppPhone,
	).Scan(
		&user.ID, &user.Email, &user.FullName, &user.ReferralCode, &balanceFloat,
		&user.IsAdmin, &createdAt, &updatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.Balance = int64(balanceFloat * 100)
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	return nil
}

// GetUserByEmail retrieves a user by their email address including password hash
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.UserWithPassword, error) {
	query := `
		SELECT id, email, password_hash, full_name, referral_code, balance, is_admin, created_at, updated_at, username, discord_handle, whatsapp_phone, avatar_url, 
		       COALESCE(total_spent, 0), COALESCE(monthly_spent, 0)
		FROM users
		WHERE email = $1
	`

	var user models.UserWithPassword
	var fullName, referralCode, username, discordHandle, whatsappPhone, avatarUrl sql.NullString
	var balanceFloat, totalSpentFloat, monthlySpentFloat float64
	var isAdmin bool
	var createdAt, updatedAt time.Time
	var hashedPassword []byte

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &hashedPassword, &fullName, &referralCode, &balanceFloat, &isAdmin, &createdAt, &updatedAt,
		&username, &discordHandle, &whatsappPhone, &avatarUrl, &totalSpentFloat, &monthlySpentFloat,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.PasswordHash = hashedPassword
	user.FullName = fullName.String
	user.ReferralCode = referralCode.String
	user.Balance = int64(balanceFloat * 100)
	user.TotalSpent = int64(totalSpentFloat * 100)
	user.MonthlySpent = int64(monthlySpentFloat * 100)
	user.IsAdmin = isAdmin
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	user.Username = username.String
	user.DiscordHandle = discordHandle.String
	user.WhatsAppPhone = whatsappPhone.String
	user.AvatarURL = avatarUrl.String

	return &user, nil
}

// UpdatePassword updates the user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return r.UpdatePasswordTx(ctx, nil, userID, passwordHash)
}

// UpdatePasswordTx updates the user's password hash within a transaction
func (r *UserRepository) UpdatePasswordTx(ctx context.Context, tx *sql.Tx, userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}
	_, err := execTx.ExecContext(ctx, query, passwordHash, userID)
	return err
}

// CreatePasswordResetToken creates a password reset token for a user
func (r *UserRepository) CreatePasswordResetToken(ctx context.Context, userID, tokenURL, code string, expiresAt time.Time) error {
	// First invalidate any existing tokens
	_, err := r.db.ExecContext(ctx, "UPDATE password_resets SET used = true WHERE user_id = $1 AND used = false", userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate existing tokens: %w", err)
	}

	query := `
		INSERT INTO password_resets (user_id, token_url, verification_code, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = r.db.ExecContext(ctx, query, userID, tokenURL, code, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	return nil
}

// GetPasswordResetToken retrieves a password reset token by token URL
func (r *UserRepository) GetPasswordResetToken(ctx context.Context, tokenURL string) (*models.PasswordResetToken, error) {
	return r.GetPasswordResetTokenTx(ctx, nil, tokenURL)
}

// GetPasswordResetTokenTx retrieves a password reset token by token URL with a row-level lock
func (r *UserRepository) GetPasswordResetTokenTx(ctx context.Context, tx *sql.Tx, tokenURL string) (*models.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, verification_code, expires_at
		FROM password_resets
		WHERE token_url = $1 AND used = false
		FOR UPDATE
	`

	var execTx interface {
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	var token models.PasswordResetToken
	err := execTx.QueryRowContext(ctx, query, tokenURL).Scan(&token.ID, &token.UserID, &token.VerificationCode, &token.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}

	return &token, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used
func (r *UserRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenID string) error {
	return r.MarkPasswordResetTokenUsedTx(ctx, nil, tokenID)
}

// MarkPasswordResetTokenUsedTx marks a password reset token as used within a transaction
func (r *UserRepository) MarkPasswordResetTokenUsedTx(ctx context.Context, tx *sql.Tx, tokenID string) error {
	query := `UPDATE password_resets SET used = true WHERE id = $1`
	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}
	_, err := execTx.ExecContext(ctx, query, tokenID)
	return err
}

// ListAll retrieves a list of users with pagination and search
func (r *UserRepository) ListAll(ctx context.Context, page, limit int, search string) ([]models.User, int, error) {
	offset := (page - 1) * limit
	
	// Query with subquery to get highest role
	baseQuery := `SELECT u.id, u.email, COALESCE(u.full_name, ''), COALESCE(u.discord_handle, ''), COALESCE(u.whatsapp_phone, ''), COALESCE(u.balance, 0), COALESCE(u.referral_code, ''), u.created_at, u.updated_at, COALESCE(u.username, ''), COALESCE(u.avatar_url, ''), COALESCE(u.is_admin, false), COALESCE(u.preferences, '{"font": "modern", "density": "comfortable", "backgroundFilter": true}')::jsonb, pgp_sym_decrypt(u.cpf_encrypted, current_setting('app.cpf_encryption_key')),
		COALESCE(u.total_spent, 0), COALESCE(u.monthly_spent, 0),
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
		var balanceFloat, totalSpentFloat, monthlySpentFloat float64
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FullName,
			&user.DiscordHandle,
			&user.WhatsAppPhone,
			&balanceFloat,
			&user.ReferralCode,
			&createdAt,
			&updatedAt,
			&user.Username,
			&user.AvatarURL,
			&user.IsAdmin,
			&preferencesJSON,
			&user.CPF,
			&totalSpentFloat,
			&monthlySpentFloat,
			&highestRole,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		user.CreatedAt = createdAt
		user.UpdatedAt = updatedAt
		user.Balance = int64(balanceFloat * 100)
		user.TotalSpent = int64(totalSpentFloat * 100)
		user.MonthlySpent = int64(monthlySpentFloat * 100)
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

// UpdateUserAdmin updates a user's profile information including email (Admin only)
func (r *UserRepository) UpdateUserAdmin(ctx context.Context, userID string, updates map[string]interface{}) error {
	return r.UpdateUserAdminTx(ctx, nil, userID, updates)
}

// UpdateUserAdminTx updates a user's profile information within an existing transaction (Admin only)
func (r *UserRepository) UpdateUserAdminTx(ctx context.Context, tx *sql.Tx, userID string, updates map[string]interface{}) error {
	// Validate UUID format
	_, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// BT-042: Lista de campos permitidos na tabela users para evitar erros SQL
	// com campos extras vindos do frontend (como highest_role, balance, etc)
	allowedFields := map[string]bool{
		"username": true, "email": true, "full_name": true, 
		"discord_handle": true, "whatsapp_phone": true, 
		"avatar_url": true, "is_admin": true, "cpf": true,
	}

	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{userID} // First argument is always userID
	argIndex := 2

	for key, value := range updates {
		if !allowedFields[key] {
			log.Printf("UpdateUserAdminTx: Saltando campo inválido '%s'", key)
			continue
		}

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

	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}

	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	_, err = execTx.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("ERROR Admin UpdateUserTx: %v | UserID: %s | Query: %s", err, userID, query)
		return fmt.Errorf("failed to update user (admin tx): %w", err)
	}

	return nil
}