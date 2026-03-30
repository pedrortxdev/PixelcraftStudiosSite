package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

// AIConfig holds configuration for the AI service
type AIConfig struct {
	GeminiAPIKey     string
	GeminiAPIURL     string
	GeminiModel      string
	RequestTimeout   time.Duration
	MaxRetries       int
}

// DefaultAIConfig returns a production-ready default configuration
func DefaultAIConfig() AIConfig {
	return AIConfig{
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		GeminiAPIURL:   os.Getenv("GEMINI_API_URL"), // Optional, falls back to default
		GeminiModel:    os.Getenv("GEMINI_MODEL"),   // Optional, falls back to default
		RequestTimeout: 30 * time.Second,
		MaxRetries:     2,
	}
}

// AIService handles AI operations (text formatting, image generation)
// Uses connection pooling, timeouts, and proper context propagation
type AIService struct {
	apiKey     string
	apiURL     string
	model      string
	httpClient *http.Client
}

// NewAIService creates a new AIService with the given configuration
// Returns an error if required configuration is missing (fail-fast)
func NewAIService(cfg AIConfig) (*AIService, error) {
	// Fail-fast: API key is required
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required but not set")
	}

	// Use defaults if not configured
	apiURL := cfg.GeminiAPIURL
	if apiURL == "" {
		// Default to gemini-1.5-flash via v1beta API
		apiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"
	}

	model := cfg.GeminiModel
	if model == "" {
		model = "gemini-1.5-flash"
	}

	// Create a shared HTTP client with proper timeout and connection pooling
	// This client will be reused for all requests (connection pooling)
	httpClient := &http.Client{
		Timeout: cfg.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false, // Keep connections alive for pooling
		},
	}

	return &AIService{
		apiKey:     cfg.GeminiAPIKey,
		apiURL:     apiURL,
		model:      model,
		httpClient: httpClient,
	}, nil
}

// GeminiRequest represents the request payload for Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in Gemini API
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a text part in Gemini API
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

// GeminiCandidate represents a candidate response from Gemini API
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// FormatText reformats text using AI to improve UX and readability
// Uses context for timeout/cancellation and shared HTTP client for connection pooling
func (s *AIService) FormatText(ctx context.Context, input string) (string, error) {
	prompt := fmt.Sprintf(`Você é um especialista em UX Writing e formatação de texto.
Sua tarefa é reformatar a seguinte descrição de produto para torná-la visualmente atraente, organizada e profissional, usando Markdown.

Diretrizes:
1. Use listas com marcadores (bullet points) para itens, características ou especificações.
2. Use negrito (**texto**) para destacar chaves ou títulos importantes (ex: "Versões:", "Incluído no pacote:").
3. Quebre os parágrafos de forma lógica.
4. Mantenha os emojis existentes se ajudarem na leitura, ou reorganize-os para ficar mais limpo.
5. Corrija erros gramaticais e de pontuação óbvios.
6. NÃO remova nenhuma informação técnica.
7. NÃO alucine ou invente informações. Apenas reformate o que foi dado.

Texto Original:
%s`, input)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("%s?key=%s", s.apiURL, s.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute with context (respects timeout and cancellation)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// GenerateAvatar generates an avatar image using Pollinations AI
// Uses context for timeout/cancellation and global rand for efficiency
func (s *AIService) GenerateAvatar(ctx context.Context, userPrompt string) (string, error) {
	// Construct the Pollinations URL
	// We add "pixel art minecraft style" to the prompt for better consistency
	enhancedPrompt := fmt.Sprintf("pixel art minecraft style avatar, high quality, %s", userPrompt)
	encodedPrompt := url.QueryEscape(enhancedPrompt)

	// Use global rand (already seeded in Go 1.20+) - no need to create new instance
	seed := rand.Intn(1000000)

	// Use Flux model via Pollinations
	apiURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=512&height=512&nologo=true&seed=%d&model=flux", encodedPrompt, seed)

	// Create request with context for timeout/cancellation
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Use shared HTTP client (connection pooling)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Pollinations API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Pollinations API returned error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read image bytes
	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data from Pollinations: %w", err)
}

	// Convert to Base64 (to maintain compatibility with existing handler)
	return base64.StdEncoding.EncodeToString(imgBytes), nil
}
