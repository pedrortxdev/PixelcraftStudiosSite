package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pixelcraft/api/internal/config"
	"golang.org/x/sync/singleflight"
)

type MercadoPagoAuthService struct {
	clientID     string
	clientSecret string

	// token and expiresAt are protected by mu for reads.
	// Writes happen only inside the singleflight callback, which is serialized.
	mu        sync.RWMutex
	token     string
	expiresAt time.Time

	// lastErr caches the last refresh error along with its expiration
	// to implement a "negative cache" backoff against thundering herd on failures.
	lastErr       error
	lastErrExpiry time.Time

	client *http.Client
	sfg    singleflight.Group
}

// errBackoffDuration is the time during which a failed refresh result is cached.
// Prevents hammering the MP API when it's down.
const errBackoffDuration = 10 * time.Second

func NewMercadoPagoAuthService(cfg config.MercadoPagoConfig) *MercadoPagoAuthService {
	return &MercadoPagoAuthService{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		token:        cfg.AccessToken, // Initialize with static token as fallback/initial
		client:       &http.Client{Timeout: 10 * time.Second},
	}
}

// GetToken returns a valid access token, refreshing it if necessary.
// Uses singleflight to coalesce concurrent refresh attempts WITHOUT blocking
// readers behind a write mutex during I/O.
func (s *MercadoPagoAuthService) GetToken(ctx context.Context) (string, error) {
	// Fast path: read-lock to check if the current token is still valid
	s.mu.RLock()
	if s.clientID == "" || s.clientSecret == "" {
		token := s.token
		s.mu.RUnlock()
		// No credentials configured, use static token if available
		if token != "" {
			return token, nil
		}
		return "", fmt.Errorf("mercado pago credentials (client_id/secret) not configured and no static token provided")
	}

	if s.token != "" && (s.expiresAt.IsZero() || time.Now().Add(5*time.Minute).Before(s.expiresAt)) {
		token := s.token
		s.mu.RUnlock()
		return token, nil
	}

	// Check negative cache (backoff on recent failures)
	if s.lastErr != nil && time.Now().Before(s.lastErrExpiry) {
		cachedErr := s.lastErr
		s.mu.RUnlock()
		return "", fmt.Errorf("token refresh in backoff (retry after %s): %w", time.Until(s.lastErrExpiry).Truncate(time.Second), cachedErr)
	}
	s.mu.RUnlock()

	// Slow path: token needs refresh.
	// singleflight ensures only ONE goroutine performs the HTTP call.
	// All other goroutines wait for that result WITHOUT holding any mutex.
	result, err, _ := s.sfg.Do("refresh_token", func() (interface{}, error) {
		// Double-check inside singleflight: another call may have just completed
		s.mu.RLock()
		if s.token != "" && (s.expiresAt.IsZero() || time.Now().Add(5*time.Minute).Before(s.expiresAt)) {
			token := s.token
			s.mu.RUnlock()
			return token, nil
		}
		s.mu.RUnlock()

		// Perform the HTTP call OUTSIDE of any lock
		newToken, refreshErr := s.refreshToken(ctx)
		if refreshErr != nil {
			// Cache the error for backoff
			s.mu.Lock()
			s.lastErr = refreshErr
			s.lastErrExpiry = time.Now().Add(errBackoffDuration)
			s.mu.Unlock()

			log.Printf("MercadoPagoAuthService: Token refresh failed - %v", refreshErr)
			return "", fmt.Errorf("failed to refresh access token: %w", refreshErr)
		}

		// Store the new token under write lock (fast, no I/O)
		s.mu.Lock()
		s.token = newToken.accessToken
		s.expiresAt = time.Now().Add(time.Duration(newToken.expiresIn) * time.Second)
		s.lastErr = nil // Clear any cached error on success
		s.lastErrExpiry = time.Time{}
		s.mu.Unlock()

		log.Printf("MercadoPagoAuthService: Token refreshed successfully. Expires in %d seconds.", newToken.expiresIn)
		return newToken.accessToken, nil
	})

	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// tokenResult holds the raw response from the MP OAuth API.
type tokenResult struct {
	accessToken string
	expiresIn   int64
}

// refreshToken performs the HTTP call to the Mercado Pago OAuth API.
// It does NOT modify any shared state — the caller is responsible for that.
func (s *MercadoPagoAuthService) refreshToken(ctx context.Context) (*tokenResult, error) {
	log.Println("MercadoPagoAuthService: Refreshing access token...")
	url := "https://api.mercadopago.com/oauth/token"

	payload := map[string]string{
		"client_secret": s.clientSecret,
		"client_id":     s.clientID,
		"grant_type":    "client_credentials",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OAuth payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call MP OAuth API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errorResp); decodeErr != nil {
			log.Printf("MercadoPagoAuthService Error: Status %d, Failed to parse error response", resp.StatusCode)
		} else {
			log.Printf("MercadoPagoAuthService Error: Status %d, Response: %v", resp.StatusCode, errorResp)
		}
		return nil, fmt.Errorf("MP OAuth API returned status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
		UserID       int64  `json:"user_id"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode MP OAuth response: %w", err)
	}

	return &tokenResult{
		accessToken: result.AccessToken,
		expiresIn:   result.ExpiresIn,
	}, nil
}
