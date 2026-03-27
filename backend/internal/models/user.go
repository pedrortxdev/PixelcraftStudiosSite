package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	Username        string     `json:"username,omitempty"`
	FullName        string     `json:"full_name,omitempty"`
	DiscordHandle   string     `json:"discord_handle,omitempty"`
	WhatsAppPhone   string     `json:"whatsapp_phone,omitempty"`
	AvatarURL       string     `json:"avatar_url,omitempty"` // URL to profile picture
	CPF             *string    `json:"cpf,omitempty"`        // Decrypted CPF (only sent to authorized admins)
	CPFEncrypted    []byte     `json:"-"` // Never sent to client
	Balance         float64    `json:"balance"` // New field for user balance
	ReferralCode    string     `json:"referral_code,omitempty"` // New field for referral code
	ReferredByCode  string     `json:"-"` // Referral code used during registration (not sent to client)
	IsAdmin         bool       `json:"is_admin"` // DEPRECATED - kept for backward compatibility
	Roles           []RoleType `json:"roles,omitempty"` // User's active roles
	HighestRole     *RoleType  `json:"highest_role,omitempty"` // Highest role for display
	TotalSpent      float64    `json:"total_spent"` // Total spent on purchases
	MonthlySpent    float64    `json:"monthly_spent"` // Spent this month
	AssignedEmail   string     `json:"assigned_email,omitempty"` // Email assigned to support staff
	Preferences     map[string]interface{} `json:"preferences,omitempty"` // User UI preferences
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// RegisterRequest represents the request to register a new user
type RegisterRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required,min=8"`
	FullName      string `json:"full_name" binding:"required"`
	Username      string `json:"username,omitempty"`
	DiscordHandle string `json:"discord_handle,omitempty"`
	WhatsAppPhone string `json:"whatsapp_phone,omitempty"`
	CPF           string `json:"cpf" binding:"required,len=11"`
	ReferralCode  string `json:"referral_code,omitempty"` // Referral code used during registration
}

// LoginRequest represents the request to login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UpdateUserRequest represents the request to update user profile
type UpdateUserRequest struct {
	Username      *string `json:"username,omitempty"`
	FullName      *string `json:"full_name,omitempty"`
	DiscordHandle *string `json:"discord_handle,omitempty"`
	WhatsAppPhone *string `json:"whatsapp_phone,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	CPF           *string `json:"cpf,omitempty"`
	Preferences   map[string]interface{} `json:"preferences,omitempty"`
}

// UserStats represents user statistics for the dashboard
type UserStats struct {
	TotalProjects      int `json:"total_projects"`
	ActiveSubscriptions int `json:"active_subscriptions"`
	ProductsPurchased   int `json:"products_purchased"`
}