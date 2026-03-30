package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// CheckoutGateway defines the interface for checkout operations (decouples circular dependency)
type CheckoutGateway interface {
	FinalizeDirectPurchase(ctx context.Context, paymentID string, gatewayID string) error
}

// DepositService handles deposit operations with Mercado Pago
// Handles currency conversion (cents ↔ BRL), idempotency, and webhook security
type DepositService struct {
	repo            *repository.TransactionRepository
	userRepo        *repository.UserRepository
	paymentRepo     *repository.PaymentRepository
	checkoutGateway CheckoutGateway
	authService     *MercadoPagoAuthService
	webhookURL      string
	webhookSecret   string // For signature verification
	depositURLs     DepositURLs
	client          *http.Client
}

// DepositURLs holds configurable deposit callback URLs
type DepositURLs struct {
	Success string
	Failure string
	Pending string
}

// NewDepositService creates a new DepositService
func NewDepositService(
	repo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	paymentRepo *repository.PaymentRepository,
	authService *MercadoPagoAuthService,
	webhookURL string,
	depositURLs DepositURLs,
) *DepositService {
	// Load webhook secret for signature verification (optional but recommended)
	webhookSecret := os.Getenv("MERCADO_PAGO_WEBHOOK_SECRET")

	return &DepositService{
		repo:          repo,
		userRepo:      userRepo,
		paymentRepo:   paymentRepo,
		authService:   authService,
		webhookURL:    webhookURL,
		webhookSecret: webhookSecret,
		depositURLs:   depositURLs,
		client:        &http.Client{Timeout: 10 * time.Second},
	}
}

// SetCheckoutGateway sets the checkout gateway (breaks circular dependency via interface)
func (s *DepositService) SetCheckoutGateway(gw CheckoutGateway) {
	s.checkoutGateway = gw
}

// MPPaymentResponse partial struct for Pix response
type MPPaymentResponse struct {
	ID                 int64   `json:"id"`
	Status             string  `json:"status"`
	StatusDetail       string  `json:"status_detail"`
	TransactionAmount  float64 `json:"transaction_amount"` // BRL as float
	PointOfInteraction struct {
		TransactionData struct {
			QRCode       string `json:"qr_code"`
			QRCodeBase64 string `json:"qr_code_base64"`
		} `json:"transaction_data"`
	} `json:"point_of_interaction"`
}

// MPPreferenceResponse partial struct for Preference response
type MPPreferenceResponse struct {
	ID        string `json:"id"`
	InitPoint string `json:"init_point"`
}

// MPBalanceResponse represents the balance response from Mercado Pago
type MPBalanceResponse struct {
	TotalAmount       float64 `json:"total_amount"`       // BRL as float
	AvailableAmount   float64 `json:"available_amount"`   // BRL as float
	UnavailableAmount float64 `json:"unavailable_amount"` // BRL as float
}

// centsToBRL converts cents (int64) to BRL (float64) for MP API
// Example: 5000 cents → 50.00 BRL
func centsToBRL(cents int64) float64 {
	return float64(cents) / 100.0
}

// bRLToCents converts BRL (float64) from MP webhook to cents (int64)
// Example: 50.00 BRL → 5000 cents
// Uses rounding to handle floating point precision issues
func bRLToCents(brl float64) int64 {
	return int64(math.Round(brl * 100.0))
}

// CreateDeposit initiates a deposit
func (s *DepositService) CreateDeposit(ctx context.Context, userID uuid.UUID, req models.DepositRequest) (*models.DepositResponse, error) {
	log.Printf("Deposit Service: Iniciando criação de depósito para user ID: %s, Amount: %d cents (%.2f BRL), Method: %s", userID, req.Amount, centsToBRL(req.Amount), req.Method)

	var providerID string
	var resp models.DepositResponse

	if req.Method == "pix" {
		log.Printf("Deposit Service: Criando pagamento PIX para user ID: %s, Amount: %d cents", userID, req.Amount)
		mpResp, err := s.createPixPayment(ctx, userID, req.Amount)
		if err != nil {
			log.Printf("Deposit Service Error: Falha ao criar pagamento PIX - %v", err)
			return nil, err
		}
		providerID = fmt.Sprintf("%d", mpResp.ID)
		resp.QRCode = mpResp.PointOfInteraction.TransactionData.QRCode
		resp.QRCodeBase64 = mpResp.PointOfInteraction.TransactionData.QRCodeBase64
		log.Printf("Deposit Service: Pagamento PIX criado com sucesso - Payment ID: %d, Amount: %.2f BRL", mpResp.ID, mpResp.TransactionAmount)
	} else if req.Method == "link" {
		log.Printf("Deposit Service: Criando preferência de pagamento para user ID: %s, Amount: %d cents", userID, req.Amount)
		// Use prefix "DEPOSIT_" to distinguish from direct purchases
		externalRef := fmt.Sprintf("DEPOSIT_%s", userID.String())
		mpResp, err := s.CreatePreference(ctx, userID, req.Amount, externalRef)
		if err != nil {
			log.Printf("Deposit Service Error: Falha ao criar preferência de pagamento - %v", err)
			return nil, err
		}
		providerID = mpResp.ID
		resp.PaymentLink = mpResp.InitPoint
		log.Printf("Deposit Service: Preferência de pagamento criada com sucesso - Preference ID: %s", mpResp.ID)
	} else {
		log.Printf("Deposit Service Error: Método de pagamento inválido: %s", req.Method)
		return nil, fmt.Errorf("invalid payment method: %s", req.Method)
	}

	// 2. Create Transaction in DB
	txID := uuid.New()
	log.Printf("Deposit Service: Criando registro de transação no banco - Transaction ID: %s, Provider Payment ID: %s", txID, providerID)

	tx := &models.Transaction{
		ID:                txID,
		UserID:            userID,
		ProviderPaymentID: &providerID,
		Amount:            req.Amount,
		Status:            models.TransactionStatusPending,
		Type:              models.TransactionTypeDeposit,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.Create(tx); err != nil {
		log.Printf("Deposit Service Error: Falha ao criar registro de transação no banco - %v", err)
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	log.Printf("Deposit Service: Transação criada com sucesso - Transaction ID: %s, Provider Payment ID: %s", txID, providerID)
	resp.TransactionID = txID
	return &resp, nil
}

func (s *DepositService) createPixPayment(ctx context.Context, userID uuid.UUID, amountCents int64) (*MPPaymentResponse, error) {
	log.Printf("Deposit Service: Chamando API do Mercado Pago para criar pagamento PIX - User ID: %s, Amount: %d cents (%.2f BRL)", userID, amountCents, centsToBRL(amountCents))

	url := "https://api.mercadopago.com/v1/payments"

	// Convert cents to BRL float for MP API (CRITICAL FIX #1)
	amountBRL := centsToBRL(amountCents)

	// Try to get user's real email first
	payerEmail := ""
	user, err := s.userRepo.GetUserByID(ctx, userID.String())
	if err == nil && user != nil && user.Email != "" {
		payerEmail = user.Email
	} else {
		// Fallback: Use the production domain for compliance
		// PIX doesn't require email validation, but MP API may reject invalid TLDs.
		// Using pixelcraft-studio.store ensures it passes regex validation.
		payerEmail = fmt.Sprintf("pix-%s@pixelcraft-studio.store", userID.String()[:8])
	}

	payload := map[string]interface{}{
		"transaction_amount": amountBRL, // Send as BRL float, NOT cents!
		"payment_method_id":  "pix",
		"payer": map[string]interface{}{
			"email":      payerEmail,
			"first_name": "Cliente",
			"last_name":  "PIX",
		},
		"description":          "Deposito em Carteira - Pixelcraft Studio",
		"external_reference":   fmt.Sprintf("DEPOSIT_%s", userID.String()),
		"installments":         1,
		"statement_descriptor": "PIXELCRAFT STUDIO",
	}

	if s.webhookURL != "" {
		payload["notification_url"] = s.webhookURL
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	token, err := s.authService.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get MP token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", uuid.New().String())

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call MP API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MP API returned status %d: %s", resp.StatusCode, string(errorBody))
	}

	var mpResp MPPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&mpResp); err != nil {
		return nil, fmt.Errorf("failed to decode MP response: %w", err)
	}

	return &mpResp, nil
}

func (s *DepositService) CreatePreference(ctx context.Context, userID uuid.UUID, amountCents int64, externalRef string) (*MPPreferenceResponse, error) {
	log.Printf("Deposit Service: Chamando API do Mercado Pago para criar preferência de pagamento - User ID: %s, Amount: %d cents (%.2f BRL), Ref: %s", userID, amountCents, centsToBRL(amountCents), externalRef)

	url := "https://api.mercadopago.com/checkout/preferences"

	// Convert cents to BRL float for MP API (CRITICAL FIX #1)
	amountBRL := centsToBRL(amountCents)

	payload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"title":       "Pixelcraft Studio - Deposito",
				"quantity":    1,
				"currency_id": "BRL",
				"unit_price":  amountBRL, // Send as BRL float, NOT cents!
			},
		},
		"external_reference": externalRef,
		"back_urls": map[string]string{
			"success":   s.depositURLs.Success,
			"failure":   s.depositURLs.Failure,
			"pending":   s.depositURLs.Pending,
		},
		"auto_return": "approved",
	}

	if s.webhookURL != "" {
		payload["notification_url"] = s.webhookURL
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	token, err := s.authService.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get MP token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call MP API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MP API returned status %d", resp.StatusCode)
	}

	var mpResp MPPreferenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&mpResp); err != nil {
		return nil, fmt.Errorf("failed to decode MP response: %w", err)
	}

	return &mpResp, nil
}

// ProcessWebhook handles the incoming webhook from Mercado Pago
// Implements proper security: signature verification, early idempotency check, replay attack protection
func (s *DepositService) ProcessWebhook(ctx context.Context, paymentID string, headers http.Header, rawBody []byte) error {
	log.Printf("Deposit Service: Processando webhook para Payment ID: %s", paymentID)

	// SECURITY: Verify webhook signature FIRST (before any processing)
	if s.webhookSecret != "" {
		if err := s.verifyWebhookSignature(headers, rawBody, paymentID); err != nil {
			return fmt.Errorf("webhook signature verification failed: %w", err)
		}
		log.Printf("Deposit Service: Webhook signature verified successfully")
	}

	// REPLAY ATTACK PROTECTION: Check local DB FIRST for idempotency
	// This prevents DDoS via replay attacks - if already processed, return immediately
	existingTx, err := s.repo.GetByProviderPaymentID(paymentID)
	if err != nil {
		return fmt.Errorf("failed to check existing transaction: %w", err)
	}

	// EARLY EXIT: If already completed, return immediately (no MP API call needed)
	if existingTx != nil && existingTx.Status == models.TransactionStatusCompleted {
		log.Printf("Webhook: Transaction %s already completed, skipping (idempotency - no API call)", existingTx.ID)
		return nil
	}

	// Only call MP API if transaction is pending or not found locally
	// This prevents DDoS via replay attacks
	status, amountCents, externalRef, err := s.getPaymentStatus(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to verify payment with MP: %w", err)
	}

	log.Printf("Deposit Service: Webhook - Status: %s, Amount: %d cents (%.2f BRL), Ref: %s", status, amountCents, centsToBRL(amountCents), externalRef)

	// ROUTING: Check DEPOSIT_ prefix FIRST (before UUID parse)
	// This prevents UUID collision since DEPOSIT_<uuid> is not a valid UUID format
	if strings.HasPrefix(externalRef, "DEPOSIT_") {
		// Case B: Wallet Deposit
		userIDStr := strings.TrimPrefix(externalRef, "DEPOSIT_")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fmt.Errorf("invalid user ID in deposit reference: %w", err)
		}

		log.Printf("Webhook: Identificado DEPÓSITO EM CARTEIRA - User ID: %s, Provider ID: %s", userID, paymentID)

		// ATOMIC STATE TRANSITION: Use CompleteDeposit which has FOR UPDATE lock
		// and checks status atomically (CRITICAL FIX #2)
		if status == "approved" {
			// CompleteDeposit internally:
			// 1. Locks the transaction row with FOR UPDATE
			// 2. Checks if status is already 'completed' (idempotency)
			// 3. Only updates if status is still 'pending'
			// This prevents double-funding from duplicate webhooks
			if err := s.repo.CompleteDepositWithAmount(paymentID, amountCents); err != nil {
				return fmt.Errorf("failed to complete deposit: %w", err)
			}
			log.Printf("Webhook: Depósito completado - Payment ID: %s, Amount: %d cents", paymentID, amountCents)
			return nil
		} else if status == "rejected" || status == "cancelled" {
			// Atomic status update (checks current status)
			if err := s.repo.UpdateStatusByProviderID(paymentID, models.TransactionStatusRejected); err != nil {
				return fmt.Errorf("failed to update transaction status: %w", err)
			}
			log.Printf("Webhook: Transação rejeitada/cancelada - Payment ID: %s", paymentID)
			return nil
		}

		// For 'pending' or other statuses, just acknowledge
		return nil
	}

	// Case A: Direct Purchase (UUID without prefix)
	// Only reach here if NOT a DEPOSIT_ prefixed reference
	if _, err := uuid.Parse(externalRef); err == nil {
		log.Printf("Webhook: Identificada COMPRA DIRETA - Payment ID: %s", externalRef)
		if s.checkoutGateway == nil {
			return fmt.Errorf("checkout gateway not initialized in deposit service")
		}
		if status == "approved" {
			return s.checkoutGateway.FinalizeDirectPurchase(ctx, externalRef, paymentID)
		}
		if status == "rejected" || status == "cancelled" {
			return s.paymentRepo.UpdateStatus(ctx, nil, externalRef, models.PaymentStatusFailed, &paymentID)
		}
		return nil
	}

	// Unknown reference format
	log.Printf("Webhook Warning: External reference format unknown: %s", externalRef)
	return nil
}

// verifyWebhookSignature validates the webhook signature from Mercado Pago
// Uses HMAC-SHA256 to verify the request originated from MP (CRITICAL FIX #4)
// MP sends: x-signature: ts={timestamp},v1={hmac_signature}
// The HMAC is computed over: "id:{payment_id};request-id:{request_id};ts:{timestamp};"
func (s *DepositService) verifyWebhookSignature(headers http.Header, rawBody []byte, paymentID string) error {
	xSignature := headers.Get("x-signature")
	xRequestID := headers.Get("x-request-id")

	if xSignature == "" {
		return fmt.Errorf("missing x-signature header")
	}

	if xRequestID == "" {
		return fmt.Errorf("missing x-request-id header")
	}

	// Parse signature format: "ts={timestamp},v1={hmac}"
	var timestamp, v1Signature string
	parts := strings.Split(xSignature, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "ts=") {
			timestamp = strings.TrimPrefix(part, "ts=")
		} else if strings.HasPrefix(part, "v1=") {
			v1Signature = strings.TrimPrefix(part, "v1=")
		}
	}

	if timestamp == "" || v1Signature == "" {
		return fmt.Errorf("invalid x-signature format: expected ts=...,v1=...")
	}

	// Validate timestamp is recent (prevent replay attacks older than 5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	now := time.Now().Unix()
	maxAge := int64(5 * 60) // 5 minutes
	if now-ts > maxAge {
		return fmt.Errorf("webhook timestamp too old: %d seconds ago", now-ts)
	}

	// Reconstruct the manifest that MP signed
	// Format: "id:{payment_id};request-id:{request_id};ts:{timestamp};"
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", paymentID, xRequestID, timestamp)

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write([]byte(manifest))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(expectedSignature), []byte(v1Signature)) {
		return fmt.Errorf("HMAC signature mismatch")
	}

	return nil
}

// getPaymentStatus fetches payment status from Mercado Pago
// Converts BRL float to cents (CRITICAL FIX #2)
func (s *DepositService) getPaymentStatus(ctx context.Context, id string) (status string, amountCents int64, externalRef string, err error) {
	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s", id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", 0, "", err
	}

	token, err := s.authService.GetToken(ctx)
	if err != nil {
		return "", 0, "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, "", fmt.Errorf("MP API returned status %d", resp.StatusCode)
	}

	// CRITICAL FIX #2: MP returns BRL as float, convert to cents
	var payload struct {
		Status            string  `json:"status"`
		TransactionAmount float64 `json:"transaction_amount"` // BRL as float from MP
		ExternalReference string  `json:"external_reference"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", 0, "", err
	}

	// Convert BRL float to cents int64 with proper rounding
	// Example: 50.00 BRL → 5000 cents, 50.50 BRL → 5050 cents
	amountCents = bRLToCents(payload.TransactionAmount)

	return payload.Status, amountCents, payload.ExternalReference, nil
}

// GetAccountBalance retrieves the Mercado Pago account balance
func (s *DepositService) GetAccountBalance(ctx context.Context) (*MPBalanceResponse, error) {
	url := "https://api.mercadopago.com/users/me/mercadopago_account/balance"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	token, err := s.authService.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call MP API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MP API returned status %d", resp.StatusCode)
	}

	var balanceResp MPBalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &balanceResp, nil
}

// RefundPayment refunds a payment in Mercado Pago
func (s *DepositService) RefundPayment(ctx context.Context, paymentID string) error {
	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s/refunds", paymentID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	token, err := s.authService.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call MP API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("MP API returned status %d", resp.StatusCode)
	}

	return nil
}

// Helper function to convert string to int64 (for header parsing)
func parseIntHeader(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.ParseInt(value, 10, 64)
}
