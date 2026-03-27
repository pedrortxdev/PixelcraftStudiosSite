package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

type DepositService struct {
	repo            *repository.TransactionRepository
	userRepo        *repository.UserRepository
	paymentRepo     *repository.PaymentRepository
	checkoutService *CheckoutService
	authService     *MercadoPagoAuthService
	webhookURL      string
	depositURLs     DepositURLs
	client          *http.Client
}

// DepositURLs holds configurable deposit callback URLs
type DepositURLs struct {
	Success string
	Failure string
	Pending string
}

func NewDepositService(
	repo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	paymentRepo *repository.PaymentRepository,
	authService *MercadoPagoAuthService,
	webhookURL string,
	depositURLs DepositURLs,
) *DepositService {
	return &DepositService{
		repo:        repo,
		userRepo:    userRepo,
		paymentRepo: paymentRepo,
		authService: authService,
		webhookURL:  webhookURL,
		depositURLs: depositURLs,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

// SetCheckoutService breaks circular dependency
func (s *DepositService) SetCheckoutService(cs *CheckoutService) {
	s.checkoutService = cs
}

// MPPaymentResponse partial struct for Pix response
type MPPaymentResponse struct {
	ID                 int64  `json:"id"`
	Status             string `json:"status"`
	StatusDetail       string `json:"status_detail"`
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
	TotalAmount     float64 `json:"total_amount"`
	AvailableAmount float64 `json:"available_amount"`
	UnavailableAmount float64 `json:"unavailable_amount"`
}

// CreateDeposit initiates a deposit
func (s *DepositService) CreateDeposit(ctx context.Context, userID uuid.UUID, req models.DepositRequest) (*models.DepositResponse, error) {
	log.Printf("Deposit Service: Iniciando criação de depósito para user ID: %s, Amount: %.2f, Method: %s", userID, req.Amount, req.Method)

	var providerID string
	var resp models.DepositResponse

	if req.Method == "pix" {
		log.Printf("Deposit Service: Criando pagamento PIX para user ID: %s, Amount: %.2f", userID, req.Amount)
		mpResp, err := s.createPixPayment(ctx, userID, req.Amount)
		if err != nil {
			log.Printf("Deposit Service Error: Falha ao criar pagamento PIX - %v", err)
			return nil, err
		}
		providerID = fmt.Sprintf("%d", mpResp.ID)
		resp.QRCode = mpResp.PointOfInteraction.TransactionData.QRCode
		resp.QRCodeBase64 = mpResp.PointOfInteraction.TransactionData.QRCodeBase64
		log.Printf("Deposit Service: Pagamento PIX criado com sucesso - Payment ID: %d", mpResp.ID)
	} else if req.Method == "link" {
		log.Printf("Deposit Service: Criando preferência de pagamento para user ID: %s, Amount: %.2f", userID, req.Amount)
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

func (s *DepositService) createPixPayment(ctx context.Context, userID uuid.UUID, amount float64) (*MPPaymentResponse, error) {
	log.Printf("Deposit Service: Chamando API do Mercado Pago para criar pagamento PIX - User ID: %s, Amount: %.2f", userID, amount)

	url := "https://api.mercadopago.com/v1/payments"

	// Try to get user's real email first
	payerEmail := ""
	user, err := s.userRepo.GetUserByID(ctx, userID.String())
	if err == nil && user != nil && user.Email != "" {
		payerEmail = user.Email
	} else {
		// Fallback: Use a generic email that MP accepts for PIX
		// PIX doesn't require email validation, so this is safe
		payerEmail = fmt.Sprintf("deposit-%s@noreply.mercadopago.com", userID.String()[:8])
	}

	payload := map[string]interface{}{
		"transaction_amount": amount,
		"payment_method_id":  "pix",
		"payer": map[string]interface{}{
			"email":      payerEmail,
			"first_name": "Cliente",
			"last_name":  "PIX",
		},
		"description":           "Deposito em Carteira - Pixelcraft Studio",
		"external_reference":    fmt.Sprintf("DEPOSIT_%s", userID.String()),
		"installments":          1,
		"statement_descriptor":  "PIXELCRAFT STUDIO",
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

func (s *DepositService) CreatePreference(ctx context.Context, userID uuid.UUID, amount float64, externalRef string) (*MPPreferenceResponse, error) {
	log.Printf("Deposit Service: Chamando API do Mercado Pago para criar preferência de pagamento - User ID: %s, Amount: %.2f, Ref: %s", userID, amount, externalRef)

	url := "https://api.mercadopago.com/checkout/preferences"

	payload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"title":       "Pixelcraft Studio - Deposito",
				"quantity":    1,
				"currency_id": "BRL",
				"unit_price":  amount,
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
func (s *DepositService) ProcessWebhook(ctx context.Context, paymentID string) error {
	log.Printf("Deposit Service: Processando webhook para Payment ID: %s", paymentID)

	status, amount, externalRef, err := s.getPaymentStatus(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to verify payment with MP: %w", err)
	}

	log.Printf("Deposit Service: Webhook - Status: %s, Amount: %.2f, Ref: %s", status, amount, externalRef)

	// Route by External Reference
	// Case A: Direct Purchase (UUID without prefix)
	if _, err := uuid.Parse(externalRef); err == nil {
		log.Printf("Webhook: Identificada COMPRA DIRETA - Payment ID: %s", externalRef)
		if status == "approved" {
			if s.checkoutService == nil {
				return fmt.Errorf("checkout service not initialized in deposit service")
			}
			return s.checkoutService.FinalizeDirectPurchase(ctx, externalRef, paymentID)
		}
		if status == "rejected" || status == "cancelled" {
			return s.paymentRepo.UpdateStatus(ctx, nil, externalRef, models.PaymentStatusFailed, &paymentID)
		}
		return nil
	}

	// Case B: Wallet Deposit (has DEPOSIT_ prefix)
	if len(externalRef) > 8 && externalRef[:8] == "DEPOSIT_" {
		userIDStr := externalRef[8:]
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fmt.Errorf("invalid user ID in deposit reference: %w", err)
		}

		log.Printf("Webhook: Identificado DEPÓSITO EM CARTEIRA - User ID: %s, Provider ID: %s", userID, paymentID)
		
		// Get transaction by provider payment ID
		tx, err := s.repo.GetByProviderPaymentID(paymentID)
		if err != nil {
			return fmt.Errorf("failed to get transaction: %w", err)
		}
		if tx == nil {
			log.Printf("Webhook Warning: Transação não encontrada para ID %s", paymentID)
			return nil
		}

		// CompleteDeposit already handles race condition with FOR UPDATE lock
		if status == "approved" {
			if err := s.repo.CompleteDeposit(tx.ID.String(), amount); err != nil {
				return fmt.Errorf("failed to complete deposit: %w", err)
			}
		} else if status == "rejected" || status == "cancelled" {
			if err := s.repo.UpdateStatus(tx.ID.String(), models.TransactionStatusRejected); err != nil {
				return fmt.Errorf("failed to update transaction status: %w", err)
			}
		}
		return nil
	}

	// Unknown reference format
	log.Printf("Webhook Warning: External reference format unknown: %s", externalRef)
	return nil
}

func (s *DepositService) getPaymentStatus(ctx context.Context, id string) (status string, amount float64, externalRef string, err error) {
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

	var payload struct {
		Status            string  `json:"status"`
		TransactionAmount float64 `json:"transaction_amount"`
		ExternalReference string  `json:"external_reference"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", 0, "", err
	}

	return payload.Status, payload.TransactionAmount, payload.ExternalReference, nil
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
