package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo  *repository.UserRepository
	db        *sql.DB
	jwtSecret string
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo *repository.UserRepository, db *sql.DB, jwtConfig struct{ Secret string }) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		db:        db,
		jwtSecret: jwtConfig.Secret,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	log.Printf("Registering user with email: %s", req.Email)

	// Check if email already exists
	var existingID string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err != nil {
			log.Printf("Error checking existing user: %v", err)
			return nil, fmt.Errorf("failed to check existing user: %w", err)
		}
		log.Printf("Email already exists: %s", req.Email)
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	log.Printf("Hashing password for user: %s", req.Email)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate referral code
	log.Printf("Generating referral code for user: %s", req.Email)
	referralCode := s.generateReferralCode()
	log.Printf("Generated referral code: %s", referralCode)

	// Validate referral code if provided
	var referredByCode *string
	if req.ReferralCode != "" {
		log.Printf("Checking referral code: %s", req.ReferralCode)
		userID, err := s.userRepo.GetUserByReferralCode(ctx, req.ReferralCode)
		if err != nil {
			log.Printf("Error checking referral code: %v", err)
			return nil, fmt.Errorf("invalid referral code: %w", err)
		}
		if userID == nil {
			log.Printf("Referral code not found: %s", req.ReferralCode)
			return nil, fmt.Errorf("invalid referral code")
		}
		referredByCode = &req.ReferralCode
		log.Printf("Referral code validated for user ID: %s", *userID)
	}

	// Create user in database
	userID := uuid.New().String()
	log.Printf("Creating user with ID: %s", userID)

	// Use email as username if not provided
	username := req.Username
	if username == "" {
		username = req.Email
	}

	// Use full name from request or username if not provided
	fullName := req.FullName
	if fullName == "" {
		fullName = username
	}

	// ATUALIZADO: Adicionei is_admin no RETURNING para já devolver o estado correto
	// Nota: Não inserimos is_admin no VALUES porque deixamos o banco usar o DEFAULT false
	query := `
		INSERT INTO users (id, email, username, password_hash, full_name, referral_code, referred_by_code, balance, discord_handle, whatsapp_phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, email, full_name, referral_code, balance, is_admin, created_at, updated_at
	`

	var user models.User
	var createdAt, updatedAt time.Time
	
	// ATUALIZADO: Adicionei &user.IsAdmin no Scan
	err = s.db.QueryRowContext(ctx, query, userID, req.Email, username, hashedPassword, fullName, referralCode, referredByCode, 0.0, req.DiscordHandle, req.WhatsAppPhone).Scan(
		&user.ID, &user.Email, &user.FullName, &user.ReferralCode, &user.Balance, &user.IsAdmin, &createdAt, &updatedAt)
	
	if err != nil {
		log.Printf("Error creating user in database: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	log.Printf("User created successfully: %s (Admin: %v)", user.ID, user.IsAdmin)
	return &user, nil
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.User, error) {
	// Get user from database
	var userID, email string
	var fullName, referralCode, username, discordHandle, whatsappPhone, avatarUrl sql.NullString
	var balance float64
	var hashedPassword []byte
	var isAdmin bool
	var createdAt, updatedAt time.Time

	// Updated to include profile fields
	query := `
		SELECT id, email, password_hash, full_name, referral_code, balance, is_admin, created_at, updated_at, username, discord_handle, whatsapp_phone, avatar_url
		FROM users
		WHERE email = $1
	`

	err := s.db.QueryRowContext(ctx, query, req.Email).Scan(
		&userID, &email, &hashedPassword, &fullName, &referralCode, &balance, &isAdmin, &createdAt, &updatedAt, &username, &discordHandle, &whatsappPhone, &avatarUrl)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Create user object
	user := &models.User{
		ID:            userID,
		Email:         email,
		FullName:      fullName.String,
		ReferralCode:  referralCode.String,
		Balance:       balance,
		IsAdmin:       isAdmin, 
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		Username:      username.String,
		DiscordHandle: discordHandle.String,
		WhatsAppPhone: whatsappPhone.String,
		AvatarURL:     avatarUrl.String,
	}

	return user, nil
}

// JWTSecret returns the JWT secret
func (s *AuthService) JWTSecret() string {
	return s.jwtSecret
}

// generateReferralCode generates a unique 8-character referral code
func (s *AuthService) generateReferralCode() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code[i] = chars[n.Int64()]
	}
	
	return string(code)
}



// GenerateResetToken generates a secure temporary JWT for password resets
func (s *AuthService) GenerateResetToken(ctx context.Context, email string) (string, string, error) {
	var userID string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("email not found")
		}
		return "", "", fmt.Errorf("failed to find user: %w", err)
	}

	_, _ = s.db.ExecContext(ctx, "UPDATE password_resets SET used = true WHERE user_id = $1 AND used = false", userID)

	bToken := make([]byte, 32)
	_, err = rand.Read(bToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}
	tokenURL := fmt.Sprintf("%x", bToken)

	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeBytes := make([]byte, 8)
	for i := range codeBytes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", "", fmt.Errorf("failed to generate code: %w", err)
		}
		codeBytes[i] = chars[n.Int64()]
	}
	code := string(codeBytes)

	expiresAt := time.Now().Add(15 * time.Minute)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO password_resets (user_id, token_url, verification_code, expires_at)
		VALUES ($1, $2, $3, $4)
	`, userID, tokenURL, code, expiresAt)
	if err != nil {
		return "", "", fmt.Errorf("failed to save reset data: %w", err)
	}

	return tokenURL, code, nil
}

// ResetPasswordConfirm sets the new password using the validated reset token
func (s *AuthService) ResetPasswordConfirm(ctx context.Context, tokenStr string, code string, newPassword string) error {
	var resetID, userID, dbCode string
	var expiresAt time.Time

	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, verification_code, expires_at 
		FROM password_resets 
		WHERE token_url = $1 AND used = false
	`, tokenStr).Scan(&resetID, &userID, &dbCode, &expiresAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("token inválido ou já utilizado")
		}
		return fmt.Errorf("failed to check token: %w", err)
	}

	if time.Now().After(expiresAt) {
		return fmt.Errorf("token expirado")
	}

	if subtle.ConstantTimeCompare([]byte(code), []byte(dbCode)) != 1 {
		return fmt.Errorf("código de verificação incorreto")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE password_resets SET used = true WHERE id = $1", resetID)
	if err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	return tx.Commit()
}