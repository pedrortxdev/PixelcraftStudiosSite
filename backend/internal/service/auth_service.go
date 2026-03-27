package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/pixelcraft/api/internal/apierrors"
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
	// Validate referral code if provided
	var referredByCode *string
	if req.ReferralCode != "" {
		userID, err := s.userRepo.GetUserByReferralCode(ctx, req.ReferralCode)
		if err != nil {
			return nil, apierrors.ErrInvalidInput
		}
		if userID == nil {
			return nil, apierrors.ErrInvalidInput
		}
		referredByCode = &req.ReferralCode
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate referral code
	referralCode, err := generateReferralCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate referral code: %w", err)
	}

	// Create user object
	userID := uuid.New().String()
	username := req.Username
	if username == "" {
		username = req.Email
	}

	fullName := req.FullName
	if fullName == "" {
		fullName = username
	}

	user := &models.User{
		ID:              userID,
		Email:           req.Email,
		Username:        username,
		FullName:        fullName,
		ReferralCode:    referralCode,
		Balance:         0.0,
		DiscordHandle:   req.DiscordHandle,
		WhatsAppPhone:   req.WhatsAppPhone,
		IsAdmin:         false,
	}

	// Create user in database
	err = s.userRepo.CreateUser(ctx, user, string(hashedPassword), referredByCode)
	if err != nil {
		// Check for unique constraint violation (email already exists)
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique") {
			return nil, apierrors.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user with constant-time comparison to prevent timing attacks
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.User, error) {
	// Get user from database (including password hash)
	userWithPassword, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Perform dummy hash computation even if user doesn't exist to prevent timing attacks
	dummyHash, _ := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.MinCost)

	if userWithPassword == nil {
		// User doesn't exist - perform dummy comparison to maintain constant time
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte("dummy-password"))
		return nil, apierrors.ErrUnauthorized
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword(userWithPassword.PasswordHash, []byte(req.Password))
	if err != nil {
		return nil, apierrors.ErrUnauthorized
	}

	// Create user object without password hash
	user := &models.User{
		ID:            userWithPassword.ID,
		Email:         userWithPassword.Email,
		FullName:      userWithPassword.FullName,
		ReferralCode:  userWithPassword.ReferralCode,
		Balance:       userWithPassword.Balance,
		IsAdmin:       userWithPassword.IsAdmin,
		CreatedAt:     userWithPassword.CreatedAt,
		UpdatedAt:     userWithPassword.UpdatedAt,
		Username:      userWithPassword.Username,
		DiscordHandle: userWithPassword.DiscordHandle,
		WhatsAppPhone: userWithPassword.WhatsAppPhone,
		AvatarURL:     userWithPassword.AvatarURL,
	}

	return user, nil
}

// JWTSecret returns the JWT secret
func (s *AuthService) JWTSecret() string {
	return s.jwtSecret
}

// generateReferralCode generates a unique 8-character referral code with proper error handling
func generateReferralCode() (string, error) {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)

	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random code: %w", err)
		}
		code[i] = chars[n.Int64()]
	}

	return string(code), nil
}

// GenerateResetToken generates a secure temporary JWT for password resets
func (s *AuthService) GenerateResetToken(ctx context.Context, email string) (string, string, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		// Return success even if email doesn't exist to prevent email enumeration
		// Generate a fake token to avoid leaking information
		fakeToken := make([]byte, 32)
		_, _ = rand.Read(fakeToken)
		fakeCode, _ := generateReferralCode()
		return hex.EncodeToString(fakeToken), fakeCode, nil
	}

	// Generate secure token
	bToken := make([]byte, 32)
	_, err = rand.Read(bToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}
	tokenURL := hex.EncodeToString(bToken)

	// Generate verification code
	code, err := generateReferralCode()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate code: %w", err)
	}

	expiresAt := time.Now().Add(15 * time.Minute)

	// Save reset token using repository
	err = s.userRepo.CreatePasswordResetToken(ctx, user.ID, tokenURL, code, expiresAt)
	if err != nil {
		return "", "", fmt.Errorf("failed to save reset data: %w", err)
	}

	return tokenURL, code, nil
}

// ResetPasswordConfirm sets the new password using the validated reset token
func (s *AuthService) ResetPasswordConfirm(ctx context.Context, tokenStr string, code string, newPassword string) error {
	// Get reset token
	token, err := s.userRepo.GetPasswordResetToken(ctx, tokenStr)
	if err != nil {
		return fmt.Errorf("failed to check token: %w", err)
	}
	if token == nil {
		return apierrors.ErrInvalidToken
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return apierrors.ErrTokenExpired
	}

	// Compare verification code using constant-time comparison
	if subtle.ConstantTimeCompare([]byte(code), []byte(token.VerificationCode)) != 1 {
		return apierrors.ErrInvalidVerification
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update password
	err = s.userRepo.UpdatePassword(ctx, token.UserID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	err = s.userRepo.MarkPasswordResetTokenUsed(ctx, token.ID)
	if err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	return tx.Commit()
}
