package service

import (
	"bytes"
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

type AIService struct {
	apiKey string
	apiURL string
}

func NewAIService() *AIService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	return &AIService{
		apiKey: apiKey,
		apiURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent",
	}
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

func (s *AIService) FormatText(input string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

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
		return "", err
	}

	url := fmt.Sprintf("%s?key=%s", s.apiURL, s.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func (s *AIService) GenerateAvatar(userPrompt string) (string, error) {
	// Construct the Pollinations URL
	// We add "pixel art minecraft style" to the prompt for better consistency
	enhancedPrompt := fmt.Sprintf("pixel art minecraft style avatar, high quality, %s", userPrompt)
	encodedPrompt := url.QueryEscape(enhancedPrompt)
	
	// Add a random seed to ensure unique results each time
	seed := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000)
	
	// Use Flux model via Pollinations (it's generally superior)
	apiURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=512&height=512&nologo=true&seed=%d&model=flux", encodedPrompt, seed)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to call Pollinations API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("pollinations API returned error status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read image bytes
	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data from Pollinations: %v", err)
	}

	// Convert to Base64 (to maintain compatibility with existing handler)
	return base64.StdEncoding.EncodeToString(imgBytes), nil
}
