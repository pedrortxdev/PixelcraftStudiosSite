package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pixelcraft/api/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// DiscountRepository handles all database operations for discounts
type DiscountRepository struct {
	db *sql.DB
}

// NewDiscountRepository creates a new DiscountRepository
func NewDiscountRepository(db *sql.DB) *DiscountRepository {
	return &DiscountRepository{db: db}
}

// GetByCode retrieves a discount by its code
func (r *DiscountRepository) GetByCode(ctx context.Context, code string) (*models.Discount, error) {
	query := `
		SELECT id, code, type, value, is_referral, created_by_user_id, 
		       expires_at, max_uses, current_uses, is_active, created_at,
		       restriction_type, target_ids
		FROM discounts
		WHERE code = $1
	`

	var d models.Discount
	var createdByUserID *uuid.UUID
	
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&d.ID, &d.Code, &d.Type, &d.Value, &d.IsReferral, &createdByUserID,
		&d.ExpiresAt, &d.MaxUses, &d.CurrentUses, &d.IsActive, &d.CreatedAt,
		&d.RestrictionType, pq.Array(&d.TargetIDs),
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get discount: %w", err)
	}
	
	d.CreatedByUserID = createdByUserID
	
	return &d, nil
}

// GetByID retrieves a discount by its ID
func (r *DiscountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Discount, error) {
	query := `
		SELECT id, code, type, value, is_referral, created_by_user_id, 
		       expires_at, max_uses, current_uses, is_active, created_at,
		       restriction_type, target_ids
		FROM discounts
		WHERE id = $1
	`

	var d models.Discount
	var createdByUserID *uuid.UUID
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.Code, &d.Type, &d.Value, &d.IsReferral, &createdByUserID,
		&d.ExpiresAt, &d.MaxUses, &d.CurrentUses, &d.IsActive, &d.CreatedAt,
		&d.RestrictionType, pq.Array(&d.TargetIDs),
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get discount: %w", err)
	}
	
	d.CreatedByUserID = createdByUserID
	
	return &d, nil
}

// List retrieves all discounts
func (r *DiscountRepository) List(ctx context.Context) ([]models.Discount, error) {
	query := `
		SELECT id, code, type, value, is_referral, created_by_user_id, 
		       expires_at, max_uses, current_uses, is_active, created_at,
		       restriction_type, target_ids
		FROM discounts
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list discounts: %w", err)
	}
	defer rows.Close()

	var discounts []models.Discount
	for rows.Next() {
		var d models.Discount
		var createdByUserID *uuid.UUID
		
		err := rows.Scan(
			&d.ID, &d.Code, &d.Type, &d.Value, &d.IsReferral, &createdByUserID,
			&d.ExpiresAt, &d.MaxUses, &d.CurrentUses, &d.IsActive, &d.CreatedAt,
			&d.RestrictionType, pq.Array(&d.TargetIDs),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan discount: %w", err)
		}
		
		d.CreatedByUserID = createdByUserID
		discounts = append(discounts, d)
	}
	
	return discounts, nil
}

// Create creates a new discount
func (r *DiscountRepository) Create(ctx context.Context, d *models.Discount) error {
	query := `
		INSERT INTO discounts (id, code, type, value, is_referral, created_by_user_id, 
		                      expires_at, max_uses, restriction_type, target_ids, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.Code, d.Type, d.Value, d.IsReferral, d.CreatedByUserID,
		d.ExpiresAt, d.MaxUses, d.RestrictionType, pq.Array(d.TargetIDs), d.IsActive, d.CreatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create discount: %w", err)
	}
	
	return nil
}

// Update updates an existing discount
func (r *DiscountRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE discounts SET "
	args := []interface{}{}
	i := 1

	for key, value := range updates {
		if key == "target_ids" {
			query += fmt.Sprintf("%s = $%d, ", key, i)
			args = append(args, pq.Array(value.([]uuid.UUID)))
		} else {
			query += fmt.Sprintf("%s = $%d, ", key, i)
			args = append(args, value)
		}
		i++
	}

	// Remove trailing comma and space
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d", i)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update discount: %w", err)
	}

	return nil
}

// Delete deletes a discount
func (r *DiscountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM discounts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// IncrementUsage increments the current_uses counter for a discount
func (r *DiscountRepository) IncrementUsage(ctx context.Context, tx *sql.Tx, discountID uuid.UUID) error {
	query := `
		UPDATE discounts
		SET current_uses = current_uses + 1
		WHERE id = $1
	`
	
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, discountID)
	} else {
		_, err = r.db.ExecContext(ctx, query, discountID)
	}
	
	if err != nil {
		return fmt.Errorf("failed to increment discount usage: %w", err)
	}
	
	return nil
}
