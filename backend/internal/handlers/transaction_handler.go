package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pixelcraft/api/internal/repository"
)

// TransactionHandler handles transaction-related requests
type TransactionHandler struct {
	repo *repository.TransactionRepository
}

// NewTransactionHandler creates a new TransactionHandler
func NewTransactionHandler(repo *repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{repo: repo}
}

// GetUserTransactions returns the list of transactions for the authenticated user
func (h *TransactionHandler) GetUserTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	transactions, err := h.repo.ListByUserID(userID.(string), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetTransactionStatus returns the status of a specific transaction ID 
// ensuring the requester is the owner of the transaction.
func (h *TransactionHandler) GetTransactionStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	txID := c.Param("id")
	if txID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing transaction id"})
		return
	}

	tx, err := h.repo.GetByID(txID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction"})
		return
	}

	if tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if tx.UserID.String() != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Not your transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": tx.Status})
}
