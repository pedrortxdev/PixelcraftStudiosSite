package service

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// EmailService handles sending emails via SMTP
type EmailService struct {
	config EmailConfig
	db     *sql.DB
}

// NewEmailService creates a new EmailService with optional DB connection
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
	
	return &EmailService{
		config: config,
		db:     db,
	}
}

// GetFromEmail returns the configured FROM email address
func (s *EmailService) GetFromEmail() string {
	config := s.loadConfig()
	return config.From
}

// loadConfig loads configuration from DB if available, falling back to cached/env config
func (s *EmailService) loadConfig() EmailConfig {
	if s.db == nil {
		return s.config
	}

	config := s.config // Start with defaults

	// Query settings
	rows, err := s.db.Query("SELECT key, value FROM system_settings WHERE key LIKE 'smtp_%'")
	if err != nil {
		log.Printf("Warning: Failed to load system settings: %v", err)
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

	// Override with DB values if allowed (only if they are not empty)
	if v, ok := settings["smtp_host"]; ok && v != "" { config.Host = v }
	if v, ok := settings["smtp_port"]; ok && v != "" { config.Port = v }
	if v, ok := settings["smtp_email"]; ok && v != "" { config.Username = v } // smtp_email maps to Username
	if v, ok := settings["smtp_password"]; ok && v != "" { config.Password = v }
	if v, ok := settings["smtp_from"]; ok && v != "" { config.From = v }

	return config
}

// SendEmail sends an email to the specified recipient
func (s *EmailService) SendEmail(to, subject, body string) error {
	return s.SendEmailHTML(to, subject, body, "")
}

// SendEmailHTML sends an HTML email with optional plain text fallback
func (s *EmailService) SendEmailHTML(to, subject, htmlBody, textBody string) error {
	config := s.loadConfig()
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	log.Printf("📧 Attempting to send email to %s via %s", to, addr)
	// SECURITY: Do not log credentials
	log.Printf("📧 SMTP Config: Host=%s, Port=%s, From=%s", config.Host, config.Port, config.From)

	// Build message with proper headers
	msg := s.buildMessage(to, subject, htmlBody, textBody, config)

	// Auth
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// TLS config - SECURE: Proper certificate validation enabled
	tlsConfig := &tls.Config{
		ServerName: config.Host,
		MinVersion: tls.VersionTLS12,
		// InsecureSkipVerify removed - certificates are now properly validated
	}

	// Connect with timeout
	log.Println("📧 Dialing SMTP server...")
	
	// AWS SES uses port 25 with STARTTLS
	if config.Port == "25" || config.Port == "587" {
		return s.sendWithSTARTTLS(addr, auth, to, msg, config)
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		log.Printf("📧 Implicit TLS connection failed, falling back to STARTTLS: %v", err)
		// Fallback to STARTTLS
		return s.sendWithSTARTTLS(addr, auth, to, msg, config)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Auth
	if err := client.Auth(auth); err != nil {
		log.Printf("❌ SMTP authentication failed for host %s", config.Host)
		return fmt.Errorf("SMTP authentication failed - please verify credentials")
	}

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
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close email writer: %w", err)
	}

	log.Printf("✅ Email sent successfully to %s", to)
	return client.Quit()
}

// sendWithSTARTTLS sends email using STARTTLS
func (s *EmailService) sendWithSTARTTLS(addr string, auth smtp.Auth, to string, msg []byte, config EmailConfig) error {
	log.Printf("📧 Starting STARTTLS flow to %s", addr)
	// Connect with timeout
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		log.Printf("❌ Failed to connect to SMTP server %s: %v", addr, err)
		return fmt.Errorf("failed to connect to SMTP server - check network connectivity")
	}
	
	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// STARTTLS - SECURE: Proper certificate validation enabled
	log.Println("📧 Sending STARTTLS command...")
	tlsConfig := &tls.Config{
		ServerName: config.Host,
		MinVersion: tls.VersionTLS12,
		// InsecureSkipVerify removed - certificates are now properly validated
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		log.Printf("❌ STARTTLS failed: %v", err)
		return fmt.Errorf("STARTTLS failed - TLS handshake error")
	}

	// Auth
	log.Println("📧 Authenticating...")
	if err := client.Auth(auth); err != nil {
		log.Printf("❌ SMTP authentication failed")
		return fmt.Errorf("SMTP authentication failed - please verify credentials")
	}

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
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close email writer: %w", err)
	}

	log.Printf("✅ Email sent successfully to %s via STARTTLS", to)
	return client.Quit()
}

// buildMessage builds the email message with headers
func (s *EmailService) buildMessage(to, subject, htmlBody, textBody string, config EmailConfig) []byte {
	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if htmlBody != "" && textBody != "" {
		// Multipart message
		boundary := "PixelcraftBoundary"
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		msg.WriteString("\r\n")

		// Plain text part
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(textBody)
		msg.WriteString("\r\n")

		// HTML part
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(htmlBody)
		msg.WriteString("\r\n")

		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if htmlBody != "" {
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(htmlBody)
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(textBody)
	}

	return []byte(msg.String())
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *EmailService) SendWelcomeEmail(to, username string) error {
	subject := "Bem-vindo à Pixelcraft Studio! 🎮"
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #0a0a0f; color: #ffffff; }
        .container { max-width: 600px; margin: 0 auto; padding: 40px; }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 28px; font-weight: bold; color: #8b5cf6; }
        .content { background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%); padding: 30px; border-radius: 12px; }
        .button { display: inline-block; background: linear-gradient(135deg, #8b5cf6, #06b6d4); color: white; padding: 14px 28px; text-decoration: none; border-radius: 8px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🎮 Pixelcraft Studio</div>
        </div>
        <div class="content">
            <h2>Olá, %s! 👋</h2>
            <p>Seja bem-vindo à Pixelcraft Studio! Estamos muito felizes em ter você conosco.</p>
            <p>Aqui você encontra os melhores jogos e experiências gaming.</p>
            <a href="https://pixelcraft-studio.store" class="button">Explorar Loja</a>
        </div>
    </div>
</body>
</html>
`, username)

	text := fmt.Sprintf("Olá, %s! Bem-vindo à Pixelcraft Studio! Acesse: https://pixelcraft-studio.store", username)

	return s.SendEmailHTML(to, subject, html, text)
}

// SendOrderConfirmation sends an order confirmation email
func (s *EmailService) SendOrderConfirmation(to, orderID string, total float64, items []string) error {
	subject := fmt.Sprintf("Pedido Confirmado #%s - Pixelcraft Studio", orderID[:8])
	
	itemsHTML := ""
	for _, item := range items {
		itemsHTML += fmt.Sprintf("<li>%s</li>", item)
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #0a0a0f; color: #ffffff; }
        .container { max-width: 600px; margin: 0 auto; padding: 40px; }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 28px; font-weight: bold; color: #8b5cf6; }
        .content { background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%); padding: 30px; border-radius: 12px; }
        .total { font-size: 24px; color: #10b981; font-weight: bold; }
        .items { background: rgba(0,0,0,0.3); padding: 15px; border-radius: 8px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🎮 Pixelcraft Studio</div>
        </div>
        <div class="content">
            <h2>Pedido Confirmado! ✅</h2>
            <p><strong>Pedido:</strong> #%s</p>
            <div class="items">
                <strong>Itens:</strong>
                <ul>%s</ul>
            </div>
            <p class="total">Total: R$ %.2f</p>
            <p>Seus jogos já estão disponíveis na sua biblioteca!</p>
        </div>
    </div>
</body>
</html>
`, orderID[:8], itemsHTML, total)

	return s.SendEmailHTML(to, subject, html, "")
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
