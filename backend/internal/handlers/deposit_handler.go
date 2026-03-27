package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"strings"

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
	userIDStr, exists := c.Get("user_id")  // Changed from "userID" to match middleware
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

	log.Printf("Deposit Handler: Payload recebido - Amount: %.2f, Method: %s", req.Amount, req.Method)

	// BT-045: Validate minimum deposit amount
	if req.Amount < 5.00 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valor mínimo de depósito é R$ 5,00"})
		return
	}

	resp, err := h.service.CreateDeposit(userID, req)
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

	// Parse MP webhook payload
	// MP can send data in query params or body depending on configuration.
	// We primarily look for data.id in body for "payment" type.

	// Note: Raw body logging removed for security - may contain sensitive payment data

	var payload struct {
		Type string `json:"type"`
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
		// Sometimes ID is at root or "id" field in query
		ID     interface{} `json:"id"`     // Can be string or int
		Action string      `json:"action"` // e.g. payment.created
	}

	// Try binding JSON first
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Webhook: Erro ao fazer bind do JSON payload - %v", err)
		// If binding fails, it might be query params or just empty body (check query)
	}

	var paymentID string

	// Extract ID logic
	if payload.Data.ID != "" {
		paymentID = payload.Data.ID
		log.Printf("Webhook: Payment ID extraído do payload.data.id: %s", paymentID)
	} else if payload.Type == "payment" || payload.Action == "payment.updated" || payload.Action == "payment.created" {
		// Sometimes ID is int, convert to string
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

		if topic != "payment" && topic != "merchant_order" { // We focus on payment
			log.Printf("Webhook: Topic '%s' não é 'payment' ou 'merchant_order', ignorando", topic)
			// If not payment, just ignore but return 200
		}
	}

	if paymentID == "" {
		log.Printf("Webhook: Nenhum Payment ID encontrado no payload ou query params")
		// No ID found, ignore
		c.JSON(http.StatusOK, gin.H{"status": "ignored", "reason": "no id found"})
		return
	}

	// Validate HMAC Signature (BT-003)
	xSignature := c.GetHeader("x-signature")
	xRequestId := c.GetHeader("x-request-id")
	dataID := c.Query("data.id")
	if dataID == "" {
		dataID = paymentID // fallback to extracted paymentID if not in query
	}

	if h.webhookSecret != "" && xSignature != "" && xRequestId != "" {
		var ts, v1 string
		parts := strings.Split(xSignature, ",")
		for _, part := range parts {
			if strings.HasPrefix(part, "ts=") {
				ts = strings.TrimPrefix(part, "ts=")
			} else if strings.HasPrefix(part, "v1=") {
				v1 = strings.TrimPrefix(part, "v1=")
			}
		}

		if ts != "" && v1 != "" {
			manifest := "id:" + dataID + ";request-id:" + xRequestId + ";ts:" + ts + ";"
			
			mac := hmac.New(sha256.New, []byte(h.webhookSecret))
			mac.Write([]byte(manifest))
			expectedMAC := hex.EncodeToString(mac.Sum(nil))

			if expectedMAC != v1 {
				log.Printf("Webhook Security Error: HMAC signature inválida. Expected: %s, Got: %s", expectedMAC, v1)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
				return
			}
			log.Printf("Webhook Security: HMAC signature validado com sucesso!")
		}
	} else if h.webhookSecret != "" {
		// Se o secret está configurado mas a request não tem headers de assinatura (provável fake)
		log.Println("Webhook Security Error: Headers de assinatura ausentes na request.")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing signature headers"})
		return
	}

	log.Printf("Webhook: Verificando transação ID %s no banco...", paymentID)

	// Process in background or foreground? Foreground is fine for now, MP retries if timeout.
	if err := h.service.ProcessWebhook(paymentID); err != nil {
		log.Printf("Webhook Error: Falha ao processar webhook para Payment ID %s - %v", paymentID, err)
		// Log error but probably still return 200 to MP to stop retries if it's a non-recoverable logic error?
		// Or return 500 to force retry?
		// Prompt says "return 200 OK immediately and exit" if status already completed.
		// Standard practice: if we processed it (even if failed logic), return 200.
		// If internal server error (DB down), return 500.
		// For now, let's return 500 on error to allow retry, 200 on success.
		// But prompt said "Do NOT trust payload... Update DB... On error: ROLLBACK".
		// I'll return 500 if DB update failed so MP retries.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook processing failed"})
		return
	}

	log.Printf("Webhook Success: Payment ID %s processado com sucesso", paymentID)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
