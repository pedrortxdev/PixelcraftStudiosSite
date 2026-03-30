package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// PartnerService handles partner profit distribution
type PartnerService struct {
	roleRepo        *repository.RoleRepository
	transactionRepo *repository.TransactionRepository
	userRepo        *repository.UserRepository
	db              *sql.DB
}

// NewPartnerService creates a new partner service
func NewPartnerService(roleRepo *repository.RoleRepository, transactionRepo *repository.TransactionRepository, userRepo *repository.UserRepository, db *sql.DB) *PartnerService {
	return &PartnerService{
		roleRepo:        roleRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
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
// - Partners are sorted by UUID to prevent deadlocks under concurrent execution
// - Remainder cents from integer division are assigned to the first partner
//
// NOTE: All amounts are in CENTS (int64) to avoid float precision issues
func (s *PartnerService) DistributePartnerProfits(ctx context.Context, saleAmountCents int64, excludeUserID string) error {
	// Calculate 1% of the sale (total commission pool in cents)
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
	eligiblePartners := make([]string, 0, len(partnerIDs))
	for _, pid := range partnerIDs {
		if pid != excludeUserID {
			eligiblePartners = append(eligiblePartners, pid)
		}
	}

	if len(eligiblePartners) == 0 {
		log.Printf("No eligible partners for distribution (buyer was the only partner)")
		return nil
	}

	// DEADLOCK PREVENTION: Always sort partner IDs to guarantee deterministic lock order.
	// If two concurrent transactions update the same partners, they will both acquire
	// row-level locks in the exact same order, making deadlocks impossible.
	sort.Strings(eligiblePartners)

	// Calculate share per partner using INTEGER division (floor division)
	numPartners := int64(len(eligiblePartners))
	sharePerPartnerCents := partnerShareCents / numPartners

	// MINIMUM DISTRIBUTION RULE:
	// If share per partner is less than 1 cent, NO partners receive anything.
	if sharePerPartnerCents < 1 {
		log.Printf("Skipping distribution: share per partner (%d cents) is below minimum threshold", sharePerPartnerCents)
		return nil
	}

	// ROUNDING REMAINDER: Calculate the "dust" that would otherwise vanish.
	// Assign it to the first partner in the sorted list (deterministic).
	remainderCents := partnerShareCents % numPartners

	// Start a database transaction for atomic updates
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Pre-build all transaction records and balance amounts for batch operations
	now := time.Now()
	transactions := make([]*models.Transaction, 0, len(eligiblePartners))
	userIDs := make([]string, 0, len(eligiblePartners))
	amounts := make([]int64, 0, len(eligiblePartners))
	totalDistributedCents := int64(0)

	for i, partnerID := range eligiblePartners {
		partnerUUID, err := uuid.Parse(partnerID)
		if err != nil {
			return fmt.Errorf("invalid partner UUID %s: %w", partnerID, err)
		}

		// First partner gets the remainder cents (rounding correction)
		share := sharePerPartnerCents
		if i == 0 && remainderCents > 0 {
			share += remainderCents
		}

		transactions = append(transactions, &models.Transaction{
			ID:        uuid.New(),
			UserID:    partnerUUID,
			Amount:    share,
			Status:    models.TransactionStatusCompleted,
			Type:      models.TransactionTypePartnerShare,
			CreatedAt: now,
			UpdatedAt: now,
		})

		userIDs = append(userIDs, partnerID)
		amounts = append(amounts, share)
		totalDistributedCents += share
	}

	// BATCH INSERT: One single multi-row INSERT for all transaction records (not N inserts)
	if err := s.transactionRepo.BatchInsertPartnerTransactions(ctx, tx, transactions); err != nil {
		return fmt.Errorf("failed to batch insert partner transactions: %w", err)
	}

	// BATCH UPDATE: One single UPDATE for all partner balances (not N updates)
	if err := s.userRepo.BatchIncrementBalances(ctx, tx, userIDs, amounts); err != nil {
		return fmt.Errorf("failed to batch update partner balances: %w", err)
	}

	// Commit the transaction: all partners receive or none do
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Partner profit distribution completed: %d cents distributed to %d partners from sale of %d cents (remainder %d cents to first partner)",
		totalDistributedCents, len(eligiblePartners), saleAmountCents, remainderCents)

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
