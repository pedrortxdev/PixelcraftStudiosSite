package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Database         DatabaseConfig
	JWT              JWTConfig
	Server           ServerConfig
	CORS             CORSConfig
	CPFEncryptionKey string
	FileEncryptionKey string // For encrypting download URLs
	EmailEncryptionKey string // For encrypting SMTP credentials
	MercadoPago      MercadoPagoConfig
	APIBaseURL       string // Base URL for public API endpoints
}

// MercadoPagoConfig holds Mercado Pago settings
type MercadoPagoConfig struct {
	AccessToken   string
	WebhookURL    string
	WebhookSecret string
	ClientID      string
	ClientSecret  string
	// Deposit callback URLs
	DepositSuccessURL string
	DepositFailureURL string
	DepositPendingURL string
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT authentication settings
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// ServerConfig holds server settings
type ServerConfig struct {
	Host string
	Port string
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	AllowedOrigins []string
}

// Load reads configuration from environment variables
// IMPORTANT: All secrets MUST be configured in .env file
func Load() (*Config, error) {
	var missing []string

	getRequired := func(key string) string {
		value := os.Getenv(key)
		if value == "" {
			missing = append(missing, key)
		}
		return value
	}

	getOptional := func(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}

	jwtExpHoursStr := getOptional("JWT_EXPIRATION_HOURS", "72")
	jwtExpHours, parseErr := strconv.Atoi(jwtExpHoursStr)
	if parseErr != nil {
		jwtExpHours = 72
	}

	corsOrigins := strings.Split(getOptional("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), ",")
	for i, origin := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(origin)
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getOptional("DB_HOST", "localhost"),
			Port:     getOptional("DB_PORT", "5432"),
			User:     getRequired("DB_USER"),
			Password: getRequired("DB_PASSWORD"),
			DBName:   getRequired("DB_NAME"),
			SSLMode:  getOptional("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     getRequired("JWT_SECRET"),
			Expiration: time.Hour * time.Duration(jwtExpHours),
		},
		Server: ServerConfig{
			Host: getOptional("SERVER_HOST", "0.0.0.0"),
			Port: getOptional("SERVER_PORT", "8080"),
		},
		CORS: CORSConfig{
			AllowedOrigins: corsOrigins,
		},
		CPFEncryptionKey:   getRequired("CPF_ENCRYPTION_KEY"),
		FileEncryptionKey:  getRequired("FILE_ENCRYPTION_KEY"),
		EmailEncryptionKey: getRequired("EMAIL_ENCRYPTION_KEY"),
		APIBaseURL:         getOptional("API_BASE_URL", "https://api.pixelcraft-studio.store"),
		MercadoPago: MercadoPagoConfig{
			AccessToken:       getOptional("MP_ACCESS_TOKEN", ""),
			WebhookURL:        getOptional("MP_WEBHOOK_URL", ""),
			WebhookSecret:     getOptional("MP_WEBHOOK_SECRET", ""),
			ClientID:          getRequired("MP_CLIENT_ID"),
			ClientSecret:      getRequired("MP_CLIENT_SECRET"),
			DepositSuccessURL: getOptional("MP_DEPOSIT_SUCCESS_URL", "https://pixelcraft-studio.store/dashboard?deposit=success"),
			DepositFailureURL: getOptional("MP_DEPOSIT_FAILURE_URL", "https://pixelcraft-studio.store/dashboard?deposit=failure"),
			DepositPendingURL: getOptional("MP_DEPOSIT_PENDING_URL", "https://pixelcraft-studio.store/dashboard?deposit=pending"),
		},
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("required environment variables not set: %s. Please configure .env file", strings.Join(missing, ", "))
	}

	return cfg, nil
}
