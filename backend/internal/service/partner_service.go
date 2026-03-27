package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// PartnerService handles partner profit distribution
type PartnerService struct {
	roleRepo        *repository.RoleRepository
	transactionRepo *repository.TransactionRepository
	db              *sql.DB
}

// NewPartnerService creates a new partner service
func NewPartnerService(roleRepo *repository.RoleRepository, transactionRepo *repository.TransactionRepository, db *sql.DB) *PartnerService {
	return &PartnerService{
		roleRepo:        roleRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

// DistributePartnerProfits distributes 1% of a sale to all eligible partners
// This is called after any successful purchase
//
// Business Rules:
// - 1% of sale amount is distributed among partners
// - Buyer is excluded from distribution (cannot earn commission on own purchase)
// - Minimum distribution: 1 cent (if share < 1 cent, NO partners receive anything)
// - All updates are atomic: either all partners receive, or none do (rollback on failure)
// - Each credit creates a transaction record for audit trail
//
// NOTE: All amounts are in CENTS (int64) to avoid float precision issues
func (s *PartnerService) DistributePartnerProfits(ctx context.Context, saleAmountCents int64, excludeUserID string) error {
	// Calculate 1% of the sale (total commission pool in cents)
	// Using integer division: 1% = divide by 100
	partnerShareCents := saleAmountCents / 100

	if partnerShareCents <= 0 {
		return nil
	}

	// Get all partners
	partnerIDs, err := s.roleRepo.GetPartnerUserIDs(ctx)
	if err != nil {
		log.Printf("Error getting partner IDs: %v", err)
		return fmt.Errorf("failed to get partner IDs: %w", err)
	}

	if len(partnerIDs) == 0 {
		return nil
	}

	// Filter out the buyer from the partner list BEFORE calculating shares
	// This ensures correct math: we divide by actual recipients, not total partners
	eligiblePartners := make([]string, 0, len(partnerIDs))
	for _, pid := range partnerIDs {
		if pid != excludeUserID {
			eligiblePartners = append(eligiblePartners, pid)
		}
	}

	// If buyer was the only partner, nothing to distribute
	if len(eligiblePartners) == 0 {
		log.Printf("No eligible partners for distribution (buyer was the only partner)")
		return nil
	}

	// Calculate share per partner using INTEGER division (floor division)
	// This automatically rounds DOWN to avoid precision issues
	sharePerPartnerCents := partnerShareCents / int64(len(eligiblePartners))

	// MINIMUM DISTRIBUTION RULE:
	// If share per partner is less than 1 cent, NO partners receive anything.
	// This prevents micro-dust amounts and database pollution.
	if sharePerPartnerCents < 1 {
		log.Printf("Skipping distribution: share per partner (%d cents) is below minimum threshold", sharePerPartnerCents)
		return nil
	}

	// Start a database transaction for atomic updates
	// Either ALL partners receive their share, or NONE do (rollback on any failure)
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if commit fails or we return early

	// Track total distributed for logging (in cents)
	totalDistributedCents := int64(0)

	// Distribute to each eligible partner WITH transaction record
	for _, partnerID := range eligiblePartners {
		// Parse partner UUID for transaction record
		partnerUUID, err := uuid.Parse(partnerID)
		if err != nil {
			return fmt.Errorf("invalid partner UUID %s: %w", partnerID, err)
		}

		// 1. Create transaction record FIRST (audit trail)
		now := time.Now()
		transaction := &models.Transaction{
			ID:             uuid.New(),
			UserID:         partnerUUID,
			Amount:         sharePerPartnerCents,
			Status:         models.TransactionStatusCompleted,
			Type:           models.TransactionTypePartnerShare,
			AdjustmentType: nil, // Not an adjustment, it's a partner share
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		// Insert transaction record
		insertQuery := `
			INSERT INTO transactions (id, user_id, amount, status, type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = tx.ExecContext(ctx, insertQuery,
			transaction.ID,
			transaction.UserID,
			transaction.Amount,
			transaction.Status,
			transaction.Type,
			transaction.CreatedAt,
			transaction.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create transaction record for partner %s: %w", partnerID, err)
		}

		// 2. Update user balance (atomic within same transaction)
		balanceQuery := `UPDATE users SET balance = balance + $1 WHERE id = $2`
		_, err = tx.ExecContext(ctx, balanceQuery, sharePerPartnerCents, partnerID)
		if err != nil {
			return fmt.Errorf("failed to update balance for partner %s: %w", partnerID, err)
		}

		totalDistributedCents += sharePerPartnerCents
		log.Printf("Partner %s received %d cents from sale of %d cents", partnerID, sharePerPartnerCents, saleAmountCents)
	}

	// Commit the transaction: all partners receive or none do
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Partner profit distribution completed: %d cents distributed to %d partners from sale of %d cents",
		totalDistributedCents, len(eligiblePartners), saleAmountCents)

	return nil
}

// GetPartners returns all users with the PARTNER role
func (s *PartnerService) GetPartners(ctx context.Context) ([]string, error) {
	return s.roleRepo.GetUsersWithRole(ctx, models.RolePartner)
}

// GetPartnerEarnings returns total earnings for a partner from transaction records (in cents)
func (s *PartnerService) GetPartnerEarnings(ctx context.Context, partnerID string) (int64, error) {
	partnerUUID, err := uuid.Parse(partnerID)
	if err != nil {
		return 0, fmt.Errorf("invalid partner UUID: %w", err)
	}

	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1 AND type = $2 AND status = $3
	`

	var total int64
	err = s.db.QueryRowContext(ctx, query, partnerUUID, models.TransactionTypePartnerShare, models.TransactionStatusCompleted).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get partner earnings: %w", err)
	}

	return total, nil
}
