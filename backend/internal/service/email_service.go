package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pixelcraft/api/internal/apierrors"
	"github.com/pixelcraft/api/internal/models"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// ConfigCache holds cached SMTP config with TTL
type ConfigCache struct {
	config    EmailConfig
	expiresAt time.Time
	mu        sync.RWMutex
}

// EmailService handles sending emails via SMTP
type EmailService struct {
	config      EmailConfig
	db          *sql.DB
	configCache *ConfigCache
	pool        *SMTPPool
	repo        *EmailRepository
}

// EmailRepository handles email logging operations
type EmailRepository struct {
	db *sql.DB
}

// NewEmailRepository creates a new EmailRepository
func NewEmailRepository(db *sql.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

// NewEmailService creates a new EmailService with connection pooling
func NewEmailService(db *sql.DB) *EmailService {
	// Default config from env - AWS SES configuration
	config := EmailConfig{
		Host:     getEnv("SMTP_HOST", "email-smtp.us-east-1.amazonaws.com"),
		Port:     getEnv("SMTP_PORT", "25"),
		Username: getEnv("SMTP_USERNAME", ""),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     getEnv("SMTP_FROM", "noreply@pixelcraft-studio.store"),
	}

	// Validate required credentials at startup
	if config.Username == "" || config.Password == "" {
		log.Printf("⚠️  WARNING: SMTP credentials not configured. Email sending will fail.")
		log.Printf("⚠️  Please set SMTP_USERNAME and SMTP_PASSWORD environment variables.")
	}

	// Initialize config cache
	cache := &ConfigCache{
		config:    config,
		expiresAt: time.Now().Add(5 * time.Minute), // Cache for 5 minutes
	}

	// Initialize SMTP connection pool
	pool := NewSMTPPool(config, 5) // Pool of 5 connections

	// Initialize email repository
	repo := NewEmailRepository(db)

	return &EmailService{
		config:      config,
		db:          db,
		configCache: cache,
		pool:        pool,
		repo:        repo,
	}
}

// GetFromEmail returns the configured FROM email address
func (s *EmailService) GetFromEmail() string {
	config := s.loadConfig()
	return config.From
}

// loadConfig loads configuration from DB with caching
func (s *EmailService) loadConfig() EmailConfig {
	// Check cache first (fast path)
	s.configCache.mu.RLock()
	if time.Now().Before(s.configCache.expiresAt) {
		config := s.configCache.config
		s.configCache.mu.RUnlock()
		return config
	}
	s.configCache.mu.RUnlock()

	// Cache miss or expired - acquire write lock
	s.configCache.mu.Lock()
	defer s.configCache.mu.Unlock()

	// Double-check after acquiring lock
	if time.Now().Before(s.configCache.expiresAt) {
		return s.configCache.config
	}

	// Load from DB if available
	if s.db == nil {
		s.configCache.expiresAt = time.Now().Add(5 * time.Minute)
		return s.config
	}

	config := s.config // Start with defaults

	// Query settings with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, "SELECT key, value FROM system_settings WHERE key LIKE 'smtp_%'")
	if err != nil {
		log.Printf("Warning: Failed to load system settings: %v", err)
		s.configCache.expiresAt = time.Now().Add(1 * time.Minute) // Retry sooner on error
		return config
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err == nil {
			settings[key] = value
		}
	}

	// Override with DB values if they are not empty
	if v, ok := settings["smtp_host"]; ok && v != "" {
		config.Host = v
	}
	if v, ok := settings["smtp_port"]; ok && v != "" {
		config.Port = v
	}
	if v, ok := settings["smtp_email"]; ok && v != "" {
		config.Username = v
	}
	if v, ok := settings["smtp_password"]; ok && v != "" {
		config.Password = v
	}
	if v, ok := settings["smtp_from"]; ok && v != "" {
		config.From = v
	}

	// Update cache
	s.configCache.config = config
	s.configCache.expiresAt = time.Now().Add(5 * time.Minute)

	return config
}

// SendEmail sends an email to the specified recipient
func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	return s.SendEmailHTML(ctx, to, subject, body, "")
}

// SendEmailHTML sends an HTML email with optional plain text fallback
func (s *EmailService) SendEmailHTML(ctx context.Context, to, subject, htmlBody, textBody string) error {
	config := s.loadConfig()
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	log.Printf("📧 Attempting to send email to %s via %s", to, addr)

	// Build message with proper headers and unique boundary
	msg := s.buildMessage(to, subject, htmlBody, textBody, config)

	// Get connection from pool
	client, _, err := s.pool.Get(ctx)
	if err != nil {
		log.Printf("📧 Failed to get connection from pool: %v", err)
		return fmt.Errorf("failed to get SMTP connection: %w", err)
	}
	defer s.pool.Put(client) // Return to pool

	// Set sender
	if err := client.Mail(config.From); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}

	// Set recipient
	if err := client.Rcpt(to); err != nil {
		log.Printf("❌ Invalid recipient address: %s", to)
		return fmt.Errorf("invalid recipient email address")
	}

	// Send body
	w, err := client.Data()
	if err != nil {
		// Check for rate limit errors
		if strings.Contains(err.Error(), "454") || strings.Contains(err.Error(), "throttl") {
			log.Printf("⚠️  AWS SES rate limit exceeded")
			return fmt.Errorf("email sending rate limit exceeded - please try again later")
		}
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		w.Close()
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close email writer: %w", err)
	}

	log.Printf("✅ Email sent successfully to %s", to)
	return client.Quit()
}

// buildMessage builds the email message with unique boundary
func (s *EmailService) buildMessage(to, subject, htmlBody, textBody string, config EmailConfig) []byte {
	var msg bytes.Buffer

	// Headers
	fmt.Fprintf(&msg, "From: %s\r\n", config.From)
	fmt.Fprintf(&msg, "To: %s\r\n", to)
	fmt.Fprintf(&msg, "Subject: %s\r\n", subject)
	msg.WriteString("MIME-Version: 1.0\r\n")

	if htmlBody != "" && textBody != "" {
		// Multipart message with UNIQUE boundary (crypto-random)
		boundary := generateUniqueBoundary()
		fmt.Fprintf(&msg, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
		msg.WriteString("\r\n")

		// Plain text part
		fmt.Fprintf(&msg, "--%s\r\n", boundary)
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(textBody)
		msg.WriteString("\r\n")

		// HTML part
		fmt.Fprintf(&msg, "--%s\r\n", boundary)
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(htmlBody)
		msg.WriteString("\r\n")

		fmt.Fprintf(&msg, "--%s--\r\n", boundary)
	} else if htmlBody != "" {
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(htmlBody)
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(textBody)
	}

	return msg.Bytes()
}

// generateUniqueBoundary generates a cryptographically secure unique boundary
func generateUniqueBoundary() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based (should never happen)
		return fmt.Sprintf("Boundary_%d", time.Now().UnixNano())
	}
	return "Pixelcraft_" + hex.EncodeToString(b)
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *EmailService) SendWelcomeEmail(ctx context.Context, to, username string) error {
	subject := "Bem-vindo à Pixelcraft Studio! 🎮"
	
	// Use html/template to escape any dangerous content
	tmpl := template.Must(template.New("welcome").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #0a0a0f; color: #ffffff; }
        .container { max-width: 600px; margin: 0 auto; padding: 40px; }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 28px; font-weight: bold; color: #8b5cf6; }
        .content { background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); padding: 30px; border-radius: 12px; }
        .button { display: inline-block; background: linear-gradient(135deg, #8b5cf6, #06b6d4); color: white; padding: 14px 28px; text-decoration: none; border-radius: 8px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🎮 Pixelcraft Studio</div>
        </div>
        <div class="content">
            <h2>Olá, {{.Username}}! 👋</h2>
            <p>Seja bem-vindo à Pixelcraft Studio! Estamos muito felizes em ter você conosco.</p>
            <p>Aqui você encontra os melhores jogos e experiências gaming.</p>
            <a href="https://pixelcraft-studio.store" class="button">Explorar Loja</a>
        </div>
    </div>
</body>
</html>
`))

	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, map[string]string{"Username": username}); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	text := fmt.Sprintf("Olá, %s! Bem-vindo à Pixelcraft Studio! Acesse: https://pixelcraft-studio.store", username)

	return s.SendEmailHTML(ctx, to, subject, htmlBuf.String(), text)
}

// SendOrderConfirmation sends an order confirmation email
func (s *EmailService) SendOrderConfirmation(ctx context.Context, to, orderID string, total float64, items []string) error {
	subject := fmt.Sprintf("Pedido Confirmado #%s - Pixelcraft Studio", orderID[:8])

	// Use html/template to escape XSS in items
	tmpl := template.Must(template.New("order").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #0a0a0f; color: #ffffff; }
        .container { max-width: 600px; margin: 0 auto; padding: 40px; }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 28px; font-weight: bold; color: #8b5cf6; }
        .content { background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%); padding: 30px; border-radius: 12px; }
        .total { font-size: 24px; color: #10b981; font-weight: bold; }
        .items { background: rgba(0,0,0,0.3); padding: 15px; border-radius: 8px; margin: 15px 0; }
        .items ul { margin: 10px 0; padding-left: 20px; }
        .items li { margin: 5px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🎮 Pixelcraft Studio</div>
        </div>
        <div class="content">
            <h2>Pedido Confirmado! ✅</h2>
            <p><strong>Pedido:</strong> #{{.OrderID}}</p>
            <div class="items">
                <strong>Itens:</strong>
                <ul>
                    {{range .Items}}<li>{{.}}</li>{{end}}
                </ul>
            </div>
            <p class="total">Total: R$ {{.Total}}</p>
            <p>Seus jogos já estão disponíveis na sua biblioteca!</p>
        </div>
    </div>
</body>
</html>
`))

	var htmlBuf bytes.Buffer
	data := map[string]interface{}{
		"OrderID": orderID[:8],
		"Total":   fmt.Sprintf("%.2f", total),
		"Items":   items,
	}
	
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	return s.SendEmailHTML(ctx, to, subject, htmlBuf.String(), "")
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SMTPPool manages a pool of reusable SMTP connections
type SMTPPool struct {
	config    EmailConfig
	connections chan *SMTPConnection
	mu          sync.Mutex
	maxSize     int
}

// SMTPConnection holds an SMTP client and its connection
type SMTPConnection struct {
	client    *smtp.Client
	lastUsed  time.Time
}

// NewSMTPPool creates a new SMTP connection pool
func NewSMTPPool(config EmailConfig, maxSize int) *SMTPPool {
	return &SMTPPool{
		config:      config,
		connections: make(chan *SMTPConnection, maxSize),
		maxSize:     maxSize,
	}
}

// Get retrieves a connection from the pool or creates a new one
func (p *SMTPPool) Get(ctx context.Context) (*smtp.Client, net.Conn, error) {
	select {
	case conn := <-p.connections:
		// Check if connection is still alive
		if err := conn.client.Noop(); err == nil {
			return conn.client, nil, nil
		}
		// Connection dead, create new one
		conn.client.Close()
	default:
		// No connections available, create new one
	}

	// Create new connection with context support
	addr := fmt.Sprintf("%s:%s", p.config.Host, p.config.Port)
	auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)

	// Dial with context - NATIVE cancellation, no goroutine leaks
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial SMTP: %w", err)
	}

	// Create client
	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to create SMTP client: %w", err)
	}

	// STARTTLS for ports 25 and 587
	if p.config.Port == "25" || p.config.Port == "587" {
		tlsConfig := &tls.Config{
			ServerName: p.config.Host,
			MinVersion: tls.VersionTLS12,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			client.Close()
			conn.Close()
			return nil, nil, fmt.Errorf("STARTTLS failed: %w", err)
		}
	}

	// Authenticate
	if err := client.Auth(auth); err != nil {
		client.Close()
		conn.Close()
		return nil, nil, fmt.Errorf("SMTP auth failed: %w", err)
	}

	return client, conn, nil
}

// Put returns a connection to the pool
func (p *SMTPPool) Put(client *smtp.Client) {
	if client == nil {
		return
	}

	// Check if client is still usable
	if err := client.Noop(); err != nil {
		client.Close()
		return
	}

	// Try to return to pool (non-blocking)
	select {
	case p.connections <- &SMTPConnection{client: client, lastUsed: time.Now()}:
		// Successfully returned to pool
	default:
		// Pool full, close connection
		client.Close()
	}
}

// Close closes all connections in the pool
func (p *SMTPPool) Close() {
	close(p.connections)
	for conn := range p.connections {
		if conn.client != nil {
			conn.client.Quit()
			conn.client.Close()
		}
	}
}

// Email logging methods (SRP - moved from PermissionService)

// LogEmail logs an sent email
func (s *EmailService) LogEmail(ctx context.Context, log *models.EmailLog) error {
	return s.repo.LogEmail(ctx, log)
}

// GetEmailLogs returns email logs with pagination and proper validation
func (s *EmailService) GetEmailLogs(ctx context.Context, page, limit int, filters map[string]string) ([]models.EmailLog, int, error) {
	// PROPER VALIDATION: Return sentinel errors instead of hardcoded strings
	if limit < 1 {
		return nil, 0, apierrors.ErrInvalidPaginationLimit
	}
	if limit > 100 {
		return nil, 0, fmt.Errorf("%w: limit cannot exceed 100 (requested: %d)", apierrors.ErrInvalidPaginationLimit, limit)
	}
	if page < 1 {
		return nil, 0, apierrors.ErrInvalidPaginationPage
	}

	return s.repo.GetEmailLogs(ctx, page, limit, filters)
}

// GetEmailLogByID returns a specific email log by ID
func (s *EmailService) GetEmailLogByID(ctx context.Context, id string) (*models.EmailLog, error) {
	return s.repo.GetEmailLogByID(ctx, id)
}

// EmailRepository interface for email logging (to be created)
type EmailRepositoryInterface interface {
	LogEmail(ctx context.Context, log *models.EmailLog) error
	GetEmailLogs(ctx context.Context, page, limit int, filters map[string]string) ([]models.EmailLog, int, error)
	GetEmailLogByID(ctx context.Context, id string) (*models.EmailLog, error)
}

// LogEmail logs a sent email
func (r *EmailRepository) LogEmail(ctx context.Context, log *models.EmailLog) error {
	metadataJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO email_logs (from_email, to_email, subject, body, status, error_message, sent_by, message_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, sent_at
	`

	err = r.db.QueryRowContext(ctx,
		query,
		log.FromEmail,
		log.ToEmail,
		log.Subject,
		log.Body,
		log.Status,
		log.ErrorMessage,
		log.SentBy,
		log.MessageID,
		metadataJSON,
	).Scan(&log.ID, &log.SentAt)

	return err
}

// GetEmailLogs returns email logs with pagination
func (r *EmailRepository) GetEmailLogs(ctx context.Context, page, limit int, filters map[string]string) ([]models.EmailLog, int, error) {
	offset := (page - 1) * limit

	// Build query with filters
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if from, ok := filters["from"]; ok && from != "" {
		whereClause += fmt.Sprintf(" AND from_email ILIKE $%d", argCount)
		args = append(args, "%"+from+"%")
		argCount++
	}

	if to, ok := filters["to"]; ok && to != "" {
		whereClause += fmt.Sprintf(" AND to_email ILIKE $%d", argCount)
		args = append(args, "%"+to+"%")
		argCount++
	}

	if status, ok := filters["status"]; ok && status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	if sentBy, ok := filters["sent_by"]; ok && sentBy != "" {
		whereClause += fmt.Sprintf(" AND sent_by = $%d", argCount)
		args = append(args, sentBy)
		argCount++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM email_logs %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get logs
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id, from_email, to_email, subject, body, status, error_message, sent_by, sent_at, message_id, metadata
		FROM email_logs
		%s
		ORDER BY sent_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.EmailLog
	for rows.Next() {
		var log models.EmailLog
		var metadataJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.FromEmail,
			&log.ToEmail,
			&log.Subject,
			&log.Body,
			&log.Status,
			&log.ErrorMessage,
			&log.SentBy,
			&log.SentAt,
			&log.MessageID,
			&metadataJSON,
		)
		if err != nil {
			return nil, 0, err
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
				log.Metadata = nil
			}
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// GetEmailLogByID returns a specific email log by ID
func (r *EmailRepository) GetEmailLogByID(ctx context.Context, id string) (*models.EmailLog, error) {
	query := `
		SELECT id, from_email, to_email, subject, body, status, error_message, sent_by, sent_at, message_id, metadata
		FROM email_logs
		WHERE id = $1
	`

	var log models.EmailLog
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.FromEmail,
		&log.ToEmail,
		&log.Subject,
		&log.Body,
		&log.Status,
		&log.ErrorMessage,
		&log.SentBy,
		&log.SentAt,
		&log.MessageID,
		&metadataJSON,
	)

	if err != nil {
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
			log.Metadata = nil
		}
	}

	return &log, nil
}
