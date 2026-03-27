package models

import (
	"time"
)

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID              string
	UserID          string
	VerificationCode string
	ExpiresAt       time.Time
}

// PasswordResetRequest represents a request to reset password
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmRequest represents a request to confirm password reset
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
