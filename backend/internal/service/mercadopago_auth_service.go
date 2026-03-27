package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pixelcraft/api/internal/config"
)

type MercadoPagoAuthService struct {
	clientID     string
	clientSecret string
	token        string
	expiresAt    time.Time
	mu           sync.RWMutex
	client       *http.Client
}

func NewMercadoPagoAuthService(cfg config.MercadoPagoConfig) *MercadoPagoAuthService {
	return &MercadoPagoAuthService{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		token:        cfg.AccessToken, // Initialize with static token as fallback/initial
		client:       &http.Client{Timeout: 10 * time.Second},
	}
}

// GetToken returns a valid access token, refreshing it if necessary.
// It uses double-check locking to avoid holding the mutex during HTTP calls.
func (s *MercadoPagoAuthService) GetToken(ctx context.Context) (string, error) {
	// Fast path: check if current token is valid without acquiring write lock
	s.mu.RLock()
	if s.clientID == "" || s.clientSecret == "" {
		s.mu.RUnlock()
		// No credentials configured, use static token if available
		if s.token != "" {
			return s.token, nil
		}
		return "", fmt.Errorf("mercado pago credentials (client_id/secret) not configured and no static token provided")
	}

	if s.token != "" && (s.expiresAt.IsZero() || time.Now().Add(5*time.Minute).Before(s.expiresAt)) {
		token := s.token
		s.mu.RUnlock()
		return token, nil
	}
	s.mu.RUnlock()

	// Slow path: need to refresh token, acquire write lock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check: another goroutine may have refreshed the token while we were waiting for the lock
	if s.token != "" && (s.expiresAt.IsZero() || time.Now().Add(5*time.Minute).Before(s.expiresAt)) {
		return s.token, nil
	}

	// Refresh the token with context propagation
	newToken, err := s.refreshToken(ctx)
	if err != nil {
		// FAIL FAST: Do NOT return expired token. Let the caller handle the error.
		log.Printf("MercadoPagoAuthService: Token refresh failed - %v", err)
		return "", fmt.Errorf("failed to refresh access token: %w", err)
	}

	return newToken, nil
}

func (s *MercadoPagoAuthService) refreshToken(ctx context.Context) (string, error) {
	log.Println("MercadoPagoAuthService: Refreshing access token...")
	url := "https://api.mercadopago.com/oauth/token"

	// Use json.Marshal for safe JSON encoding (handles special characters properly)
	payload := map[string]string{
		"client_secret": s.clientSecret,
		"client_id":     s.clientID,
		"grant_type":    "client_credentials",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OAuth payload: %w", err)
	}

	// Use http.NewRequestWithContext to propagate context cancellation
	req, err := http.NewRequestWithContext(ctx, "POST", url, &byteReader{data: body})
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call MP OAuth API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errorResp); decodeErr != nil {
			log.Printf("MercadoPagoAuthService Error: Status %d, Failed to parse error response", resp.StatusCode)
		} else {
			log.Printf("MercadoPagoAuthService Error: Status %d, Response: %v", resp.StatusCode, errorResp)
		}
		return "", fmt.Errorf("MP OAuth API returned status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		UserID       int    `json:"user_id"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode MP OAuth response: %w", err)
	}

	s.token = result.AccessToken
	// expires_in is in seconds
	s.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	log.Printf("MercadoPagoAuthService: Token refreshed successfully. Expires in %d seconds.", result.ExpiresIn)

	return s.token, nil
}

// byteReader implements io.Reader for byte slices
type byteReader struct {
	data []byte
	pos  int
}

func (b *byteReader) Read(p []byte) (int, error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}
