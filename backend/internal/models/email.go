package models

import "time"

// EmailAccount represents an email account in the mail server
type EmailAccount struct {
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name,omitempty"`
	Quota       string    `json:"quota,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateEmailRequest is the request body for creating an email account
type CreateEmailRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name,omitempty"`
}

// UpdateEmailPasswordRequest is the request body for updating an email password
type UpdateEmailPasswordRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}
