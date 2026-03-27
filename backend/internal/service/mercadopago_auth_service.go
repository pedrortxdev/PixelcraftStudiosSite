package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pixelcraft/api/internal/config"
)

type MercadoPagoAuthService struct {
	clientID     string
	clientSecret string
	token        string
	expiresAt    time.Time
	mu           sync.Mutex
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

// GetToken returns a valid access token, refreshing it if necessary
func (s *MercadoPagoAuthService) GetToken() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. If we have no credentials, rely on static token
	if s.clientID == "" || s.clientSecret == "" {
		if s.token != "" {
			return s.token, nil
		}
		return "", fmt.Errorf("mercado pago credentials (client_id/secret) not configured and no static token provided")
	}

	// 2. If token is set and not expired (buffer of 5 mins), return it
	if s.token != "" && (s.expiresAt.IsZero() || time.Now().Add(5*time.Minute).Before(s.expiresAt)) {
		return s.token, nil
	}

	// 3. Otherwise, fetch a new token
	newToken, err := s.refreshToken()
	if err != nil && s.token != "" {
		log.Printf("MercadoPagoAuthService: Refresh failed, falling back to static token: %v", err)
		return s.token, nil
	}
	return newToken, err
}

func (s *MercadoPagoAuthService) refreshToken() (string, error) {
	log.Println("MercadoPagoAuthService: Refreshing access token...")
	url := "https://api.mercadopago.com/oauth/token"

	payload := strings.NewReader(fmt.Sprintf(
		`{"client_secret":"%s","client_id":"%s","grant_type":"client_credentials"}`,
		s.clientSecret, s.clientID,
	))

	req, err := http.NewRequest("POST", url, payload)
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
		json.NewDecoder(resp.Body).Decode(&errorResp)
		log.Printf("MercadoPagoAuthService Error: Status %d, Response: %v", resp.StatusCode, errorResp)
		return "", fmt.Errorf("MP OAuth API returned status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
		UserID      int    `json:"user_id"`
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
