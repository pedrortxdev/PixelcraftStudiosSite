package service

import (
	"context"
	"database/sql"
	"log"

	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// PartnerService handles partner profit distribution
type PartnerService struct {
	roleRepo *repository.RoleRepository
	db       *sql.DB
}

// NewPartnerService creates a new partner service
func NewPartnerService(roleRepo *repository.RoleRepository, db *sql.DB) *PartnerService {
	return &PartnerService{
		roleRepo: roleRepo,
		db:       db,
	}
}

// DistributePartnerProfits distributes 1% of a sale to all partners
// This is called after any successful purchase
func (s *PartnerService) DistributePartnerProfits(ctx context.Context, saleAmount float64, excludeUserID string) error {
	// Calculate 1% of the sale
	partnerShare := saleAmount * 0.01
	
	if partnerShare <= 0 {
		return nil
	}
	
	// Get all partners
	partnerIDs, err := s.roleRepo.GetPartnerUserIDs(ctx)
	if err != nil {
		log.Printf("Error getting partner IDs: %v", err)
		return err
	}
	
	if len(partnerIDs) == 0 {
		return nil
	}
	
	// Calculate share per partner
	sharePerPartner := partnerShare / float64(len(partnerIDs))
	
	if sharePerPartner < 0.01 {
		// Don't distribute if share is less than 1 cent
		return nil
	}
	
	// Distribute to each partner
	for _, partnerID := range partnerIDs {
		// Skip the buyer themselves if they are a partner
		if partnerID == excludeUserID {
			continue
		}
		
		// Add to partner's balance
		if err := s.addToBalance(ctx, partnerID, sharePerPartner); err != nil {
			log.Printf("Error adding partner share to user %s: %v", partnerID, err)
			// Continue with other partners even if one fails
			continue
		}
		
		// Log the distribution
		log.Printf("Partner %s received R$%.2f from sale of R$%.2f", partnerID, sharePerPartner, saleAmount)
		
		// TODO: Create a transaction record for the partner share
		// This could be done by creating a transaction entry in the transactions table
	}
	
	return nil
}

// addToBalance adds an amount to a user's balance
func (s *PartnerService) addToBalance(ctx context.Context, userID string, amount float64) error {
	query := `UPDATE users SET balance = balance + $2 WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, userID, amount)
	return err
}

// GetPartners returns all users with the PARTNER role
func (s *PartnerService) GetPartners(ctx context.Context) ([]string, error) {
	return s.roleRepo.GetUsersWithRole(ctx, models.RolePartner)
}

// GetPartnerEarnings returns total earnings for a partner (would need transaction tracking)
// For now, this is a placeholder that would query the transactions table
func (s *PartnerService) GetPartnerEarnings(ctx context.Context, partnerID string) (float64, error) {
	// TODO: Implement when transaction tracking for partner shares is added
	// This would query something like:
	// SELECT SUM(amount) FROM transactions WHERE user_id = $1 AND type = 'PARTNER_SHARE'
	return 0, nil
}
