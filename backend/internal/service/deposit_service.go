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
	repo        *repository.TransactionRepository
	userRepo    *repository.UserRepository
	authService *MercadoPagoAuthService
	webhookURL  string
	client      *http.Client
}

func NewDepositService(repo *repository.TransactionRepository, userRepo *repository.UserRepository, authService *MercadoPagoAuthService, webhookURL string) *DepositService {
	return &DepositService{
		repo:        repo,
		userRepo:    userRepo,
		authService: authService,
		webhookURL:  webhookURL,
		client:      &http.Client{Timeout: 10 * time.Second},
	}
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
func (s *DepositService) CreateDeposit(userID uuid.UUID, req models.DepositRequest) (*models.DepositResponse, error) {
	log.Printf("Deposit Service: Iniciando criação de depósito para user ID: %s, Amount: %.2f, Method: %s", userID, req.Amount, req.Method)

	var providerID string
	var resp models.DepositResponse

	if req.Method == "pix" {
		log.Printf("Deposit Service: Criando pagamento PIX para user ID: %s, Amount: %.2f", userID, req.Amount)
		mpResp, err := s.createPixPayment(userID, req.Amount)
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
		mpResp, err := s.createPreference(userID, req.Amount)
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

func (s *DepositService) createPixPayment(userID uuid.UUID, amount float64) (*MPPaymentResponse, error) {
	log.Printf("Deposit Service: Chamando API do Mercado Pago para criar pagamento PIX - User ID: %s, Amount: %.2f", userID, amount)

	url := "https://api.mercadopago.com/v1/payments"

	// BT-046: Fetch real user email for PIX payload
	payerEmail := fmt.Sprintf("user%s@pixelcraft-studio.store", userID.String()[:8]) // fallback
	user, err := s.userRepo.GetUserByID(context.Background(), userID.String())
	if err == nil && user != nil && user.Email != "" {
		payerEmail = user.Email
	}

	// Prepare the payload with all required fields for PIX
	payload := map[string]interface{}{
		"transaction_amount": amount,
		"payment_method_id":  "pix",
		"payer": map[string]interface{}{
			"email":      payerEmail,
			"first_name": "PIX",
			"last_name":  "Customer",
		},
		"description":           "Add Funds - Pixelcraft Studio",
		"external_reference":    userID.String(),
		"installments":          1,
		"statement_descriptor":  "PIXELCRAFT STUDIO",
	}

	// Add notification URL if configured
	if s.webhookURL != "" {
		payload["notification_url"] = s.webhookURL
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Deposit Service Error: Falha ao serializar payload para pagamento PIX - %v", err)
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Deposit Service Error: Falha ao criar requisição HTTP para pagamento PIX - %v", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Get Token dynamically
	token, err := s.authService.GetToken()
	if err != nil {
		log.Printf("Deposit Service Error: Falha ao obter token do Mercado Pago - %v", err)
		return nil, fmt.Errorf("failed to get MP token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", uuid.New().String())

	log.Printf("Deposit Service: Enviando requisição para API do Mercado Pago - URL: %s", url)

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Deposit Service Error: Falha na chamada da API do Mercado Pago - %v", err)
		return nil, fmt.Errorf("failed to call MP API: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Deposit Service: Resposta da API do Mercado Pago recebida - Status Code: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		// Read response body to get more details about the error
		errorBody, _ := io.ReadAll(resp.Body)
		log.Printf("Deposit Service Error: API do Mercado Pago retornou status inválido: %d - Response: %s", resp.StatusCode, string(errorBody))
		return nil, fmt.Errorf("MP API returned status %d: %s", resp.StatusCode, string(errorBody))
	}

	var mpResp MPPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&mpResp); err != nil {
		log.Printf("Deposit Service Error: Falha ao decodificar resposta da API do Mercado Pago - %v", err)
		return nil, fmt.Errorf("failed to decode MP response: %w", err)
	}

	log.Printf("Deposit Service: Resposta da API do Mercado Pago processada com sucesso - Payment ID: %d, Status: %s", mpResp.ID, mpResp.Status)
	return &mpResp, nil
}

func (s *DepositService) createPreference(userID uuid.UUID, amount float64) (*MPPreferenceResponse, error) {
	log.Printf("Deposit Service: Chamando API do Mercado Pago para criar preferência de pagamento - User ID: %s, Amount: %.2f", userID, amount)

	url := "https://api.mercadopago.com/checkout/preferences"

	// Prepare the payload
	payload := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"title":       "Add Funds",
				"quantity":    1,
				"currency_id": "BRL",
				"unit_price":  amount,
			},
		},
		"external_reference": userID.String(),
		"back_urls": map[string]string{
			"success": "https://pixelcraft-studio.store/dashboard/wallet", // Configure appropriately
			"failure": "https://pixelcraft-studio.store/dashboard/wallet",
			"pending": "https://pixelcraft-studio.store/dashboard/wallet",
		},
	}

	// Add notification URL if configured
	if s.webhookURL != "" {
		log.Printf("Deposit Service: Configurando Webhook URL para: %s", s.webhookURL)
		payload["notification_url"] = s.webhookURL
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Deposit Service Error: Falha ao serializar payload para preferência de pagamento - %v", err)
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Deposit Service Error: Falha ao criar requisição HTTP para preferência de pagamento - %v", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Get Token dynamically
	token, err := s.authService.GetToken()
	if err != nil {
		log.Printf("Deposit Service Error: Falha ao obter token do Mercado Pago - %v", err)
		return nil, fmt.Errorf("failed to get MP token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Deposit Service: Enviando requisição para API do Mercado Pago - URL: %s", url)

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Deposit Service Error: Falha na chamada da API do Mercado Pago - %v", err)
		return nil, fmt.Errorf("failed to call MP API: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Deposit Service: Resposta da API do Mercado Pago recebida - Status Code: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Printf("Deposit Service Error: API do Mercado Pago retornou status inválido: %d", resp.StatusCode)
		return nil, fmt.Errorf("MP API returned status %d", resp.StatusCode)
	}

	var mpResp MPPreferenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&mpResp); err != nil {
		log.Printf("Deposit Service Error: Falha ao decodificar resposta da API do Mercado Pago - %v", err)
		return nil, fmt.Errorf("failed to decode MP response: %w", err)
	}

	log.Printf("Deposit Service: Resposta da API do Mercado Pago processada com sucesso - Preference ID: %s", mpResp.ID)
	return &mpResp, nil
}

// ProcessWebhook handles the incoming webhook from Mercado Pago
func (s *DepositService) ProcessWebhook(paymentID string) error {
	log.Printf("Deposit Service: Processando webhook para Payment ID: %s", paymentID)

	// 1. Verify status with MP
	status, amount, err := s.getPaymentStatus(paymentID)
	if err != nil {
		log.Printf("Webhook Error: Falha ao verificar status do pagamento com MP - Payment ID: %s, Error: %v", paymentID, err)
		return fmt.Errorf("failed to verify payment with MP: %w", err)
	}

	log.Printf("Deposit Service: Status do pagamento recebido do MP - Payment ID: %s, Status: %s, Amount: %.2f", paymentID, status, amount)

	// 2. Check if transaction exists
	log.Printf("Webhook: Verificando transação ID %s no banco...", paymentID)
	tx, err := s.repo.GetByProviderPaymentID(paymentID)
	if err != nil {
		log.Printf("Webhook Error: Falha ao buscar transação no banco - Payment ID: %s, Error: %v", paymentID, err)
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if tx == nil {
		log.Printf("Webhook Error: Transação não encontrada - Payment ID: %s", paymentID)
		// Log warning: received webhook for unknown transaction (maybe from another system or test)
		return nil
	}

	log.Printf("Webhook Success: Transação encontrada - Transaction ID: %s, User ID: %s, Status: %s", tx.ID, tx.UserID, tx.Status)

	// 3. Idempotency check
	if tx.Status == models.TransactionStatusCompleted {
		log.Printf("Webhook: Transação já está completa, ignorando - Transaction ID: %s", tx.ID)
		return nil
	}

	// 4. Update if approved
	if status == "approved" {
		log.Printf("Webhook: Pagamento aprovado, atualizando saldo do usuário - Transaction ID: %s, Amount: %.2f", tx.ID, amount)
		
		// Security check: Verify amount matches expected value
		if amount != tx.Amount {
			log.Printf("Webhook Security Warning: Amount mismatch! Expected: %.2f, Received: %.2f, Transaction ID: %s", tx.Amount, amount, tx.ID)
			return fmt.Errorf("amount mismatch: expected %.2f, got %.2f", tx.Amount, amount)
		}

		if err := s.repo.CompleteDeposit(tx.ID.String(), amount); err != nil {
			log.Printf("Webhook Error: Falha ao completar depósito no banco - Transaction ID: %s, Error: %v", tx.ID, err)
			return fmt.Errorf("failed to complete deposit: %w", err)
		}

		log.Printf("Webhook Success: Saldo atualizado com sucesso - Transaction ID: %s", tx.ID)
	} else if status == "refunded" || status == "charged_back" {
		log.Printf("Webhook: Pagamento estornado/chargeback (%s), deduzindo saldo - Transaction ID: %s, Amount: %.2f", status, tx.ID, amount)

		if tx.Status == models.TransactionStatusCompleted {
			// Refund: Deduct money
			if err := s.repo.RefundDeposit(tx.ID.String(), amount); err != nil {
				log.Printf("Webhook Error: Falha ao estornar depósito no banco - Transaction ID: %s, Error: %v", tx.ID, err)
				return fmt.Errorf("failed to refund deposit: %w", err)
			}
			log.Printf("Webhook Success: Saldo debitado e status atualizado para refunded - Transaction ID: %s", tx.ID)
		} else {
			// If not completed yet, just mark as refunded so it doesn't get completed later
			log.Printf("Webhook: Transação não estava completa (Status: %s), marcando como refunded - Transaction ID: %s", tx.Status, tx.ID)
			if err := s.repo.UpdateStatus(tx.ID.String(), models.TransactionStatusRefunded); err != nil {
				log.Printf("Webhook Error: Falha ao atualizar status para refunded - Transaction ID: %s, Error: %v", tx.ID, err)
				return fmt.Errorf("failed to update status to refunded: %w", err)
			}
		}
	} else if status == "rejected" || status == "cancelled" {
		log.Printf("Webhook: Pagamento rejeitado ou cancelado, atualizando status - Transaction ID: %s, Status: %s", tx.ID, status)
		if err := s.repo.UpdateStatus(tx.ID.String(), models.TransactionStatusRejected); err != nil {
			log.Printf("Webhook Error: Falha ao atualizar status para rejeitado - Transaction ID: %s, Error: %v", tx.ID, err)
			return fmt.Errorf("failed to update status to rejected: %w", err)
		}

		log.Printf("Webhook Success: Status atualizado para rejeitado - Transaction ID: %s", tx.ID)
	} else {
		log.Printf("Webhook: Status não reconhecido, ignorando - Payment ID: %s, Status: %s", paymentID, status)
	}

	return nil
}

func (s *DepositService) getPaymentStatus(id string) (string, float64, error) {
	log.Printf("Deposit Service: Verificando status do pagamento com MP - Payment ID: %s", id)

	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s", id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Deposit Service Error: Falha ao criar requisição HTTP para verificar status - %v", err)
		return "", 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Get Token dynamically
	token, err := s.authService.GetToken()
	if err != nil {
		log.Printf("Deposit Service Error: Falha ao obter token do Mercado Pago - %v", err)
		return "", 0, fmt.Errorf("failed to get MP token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	log.Printf("Deposit Service: Enviando requisição para verificar status do pagamento - URL: %s", url)

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Deposit Service Error: Falha na chamada da API do Mercado Pago para verificar status - %v", err)
		return "", 0, err
	}
	defer resp.Body.Close()

	log.Printf("Deposit Service: Resposta da API do Mercado Pago recebida para verificação de status - Status Code: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Deposit Service Error: API do Mercado Pago retornou status inválido para verificação de status: %d", resp.StatusCode)
		return "", 0, fmt.Errorf("MP API returned status %d", resp.StatusCode)
	}

	var payload struct {
		Status            string  `json:"status"`
		TransactionAmount float64 `json:"transaction_amount"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		log.Printf("Deposit Service Error: Falha ao decodificar resposta da API do Mercado Pago para verificação de status - %v", err)
		return "", 0, err
	}

	log.Printf("Deposit Service: Status do pagamento verificado com sucesso - ID: %s, Status: %s, Amount: %.2f", id, payload.Status, payload.TransactionAmount)
	return payload.Status, payload.TransactionAmount, nil
}

// GetAccountBalance retrieves the Mercado Pago account balance
func (s *DepositService) GetAccountBalance() (*MPBalanceResponse, error) {
	url := "https://api.mercadopago.com/users/me/mercadopago_account/balance"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	token, err := s.authService.GetToken()
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
func (s *DepositService) RefundPayment(paymentID string) error {
	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s/refunds", paymentID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	token, err := s.authService.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	// Idempotency key might be good here, but for now simple call

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
