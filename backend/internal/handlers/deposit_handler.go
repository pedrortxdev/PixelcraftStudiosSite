package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/service"
)

type DepositHandler struct {
	service       *service.DepositService
	webhookSecret string
}

func NewDepositHandler(service *service.DepositService, webhookSecret string) *DepositHandler {
	return &DepositHandler{service: service, webhookSecret: webhookSecret}
}

// Deposit handles requests to add funds
func (h *DepositHandler) Deposit(c *gin.Context) {
	log.Printf("Deposit Handler: Iniciando transação para User ID: [buscando do contexto gin]")

	// Get user ID from context (set by middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		log.Printf("Critical: User ID missing from context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Deposit Error: Invalid user ID format - Value: %s, Error: %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("Deposit Handler: Procesando depósito para usuário ID: %s", userID)

	var req models.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Deposit Error: JSON Binding falhou - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Deposit Handler: Payload recebido - Amount: %d cents, Method: %s", req.Amount, req.Method)

	// BT-045: Validate minimum deposit amount (500 cents = R$ 5.00)
	if req.Amount < 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valor mínimo de depósito é R$ 5,00"})
		return
	}

	resp, err := h.service.CreateDeposit(c.Request.Context(), userID, req)
	if err != nil {
		log.Printf("Deposit Error: Falha ao criar depósito - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deposit"})
		return
	}

	log.Printf("Deposit Handler: Depósito criado com sucesso - Transaction ID: %s", resp.TransactionID)
	c.JSON(http.StatusOK, resp)
}

// Webhook handles Mercado Pago notifications
func (h *DepositHandler) Webhook(c *gin.Context) {
	log.Printf("Webhook: Recebido payload [iniciando processamento]")

	// Read raw body for signature verification (must be done before ShouldBindJSON)
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Webhook Error: Failed to read body - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	// Restore body for later JSON binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	// Parse MP webhook payload
	var payload struct {
		Type   string `json:"type"`
		Data   struct {
			ID string `json:"id"`
		} `json:"data"`
		ID     interface{} `json:"id"`
		Action string      `json:"action"`
	}

	// Try binding JSON
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Webhook: Erro ao fazer bind do JSON payload - %v", err)
	}

	var paymentID string

	// Extract ID logic
	if payload.Data.ID != "" {
		paymentID = payload.Data.ID
		log.Printf("Webhook: Payment ID extraído do payload.data.id: %s", paymentID)
	} else if payload.Type == "payment" || payload.Action == "payment.updated" || payload.Action == "payment.created" {
		if payload.Data.ID != "" {
			paymentID = payload.Data.ID
		} else if idStr, ok := payload.ID.(string); ok {
			paymentID = idStr
			log.Printf("Webhook: Payment ID extraído do payload.id (string): %s", paymentID)
		} else if idFloat, ok := payload.ID.(float64); ok {
			paymentID = strconv.FormatFloat(idFloat, 'f', 0, 64)
			log.Printf("Webhook: Payment ID extraído do payload.id (float64): %s", paymentID)
		}
	}

	// Fallback to query param
	if paymentID == "" {
		paymentID = c.Query("id")
		topic := c.Query("topic")
		log.Printf("Webhook: Fallback para query param - ID: %s, Topic: %s", paymentID, topic)

		if topic != "payment" && topic != "merchant_order" {
			log.Printf("Webhook: Topic '%s' não é 'payment' ou 'merchant_order', ignorando", topic)
		}
	}

	if paymentID == "" {
		log.Printf("Webhook: Nenhum Payment ID encontrado no payload ou query params")
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "no id found"})
		return
	}

	// Process webhook via service (pass headers and raw body for signature verification)
	if err := h.service.ProcessWebhook(c.Request.Context(), paymentID, c.Request.Header, rawBody); err != nil {
		log.Printf("Webhook Error: Falha ao processar webhook - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process webhook"})
		return
	}

	log.Printf("Webhook: Processado com sucesso para Payment ID: %s", paymentID)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
