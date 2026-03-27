package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
)

// SubscriptionRepository handles database operations for subscriptions
type SubscriptionRepository struct {
	db *sql.DB
}

// NewSubscriptionRepository creates a new SubscriptionRepository
func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// ListActivePlans returns all active plans
func (r *SubscriptionRepository) ListActivePlans(ctx context.Context) ([]models.Plan, error) {
	query := `
		SELECT id, name, description, price, image_url, is_active, features, created_at, updated_at
		FROM plans
		WHERE is_active = true
		ORDER BY price ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %w", err)
	}
	defer rows.Close()

	var plans []models.Plan
	for rows.Next() {
		var p models.Plan
		var featuresJSON sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.IsActive, &featuresJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if featuresJSON.Valid {
			if err := json.Unmarshal([]byte(featuresJSON.String), &p.Features); err != nil {
				// Log error but don't fail the whole request? Or return empty features?
				// For now, let's just initialize it as empty if fail
				p.Features = []string{}
			}
		} else {
			p.Features = []string{}
		}
		plans = append(plans, p)
	}
	return plans, nil
}

// GetPlanByID returns a plan by ID
func (r *SubscriptionRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*models.Plan, error) {
	query := `
		SELECT id, name, description, price, image_url, is_active, features, created_at, updated_at
		FROM plans
		WHERE id = $1
	`
	var p models.Plan
	var featuresJSON sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.IsActive, &featuresJSON, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if featuresJSON.Valid {
		if err := json.Unmarshal([]byte(featuresJSON.String), &p.Features); err != nil {
			log.Printf("Failed to decode features JSON: %v", err)
		}
	} else {
		p.Features = []string{}
	}
	return &p, nil
}

// CreatePlan creates a new plan
func (r *SubscriptionRepository) CreatePlan(ctx context.Context, plan *models.Plan) error {
	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return fmt.Errorf("failed to marshal features: %w", err)
	}

	query := `
		INSERT INTO plans (id, name, description, price, image_url, features, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())
	`
	_, err = r.db.ExecContext(ctx, query, plan.ID, plan.Name, plan.Description, plan.Price, plan.ImageURL, string(featuresJSON))
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}
	return nil
}

// UpdatePlan updates an existing plan
func (r *SubscriptionRepository) UpdatePlan(ctx context.Context, plan *models.Plan) error {
	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return fmt.Errorf("failed to marshal features: %w", err)
	}

	query := `
		UPDATE plans 
		SET name=$1, description=$2, price=$3, features=$4, updated_at=NOW() 
		WHERE id=$5
	`
	_, err = r.db.ExecContext(ctx, query, plan.Name, plan.Description, plan.Price, string(featuresJSON), plan.ID)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}
	return nil
}

// DeletePlan soft deletes a plan
func (r *SubscriptionRepository) DeletePlan(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE plans SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}
	return nil
}

// CreateSubscription creates a new subscription
func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, tx *sql.Tx, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, user_id, plan_id, plan_name, price_per_month, agreed_price, status, 
			project_stage, started_at, next_billing_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	
	// Use transaction if provided, otherwise use db
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query,
			sub.ID, sub.UserID, sub.PlanID, sub.PlanName, sub.PricePerMonth, sub.AgreedPrice,
			sub.Status, sub.ProjectStage, sub.StartedAt, sub.NextBillingDate,
			sub.CreatedAt, sub.UpdatedAt,
		)
	} else {
		_, err = r.db.ExecContext(ctx, query,
			sub.ID, sub.UserID, sub.PlanID, sub.PlanName, sub.PricePerMonth, sub.AgreedPrice,
			sub.Status, sub.ProjectStage, sub.StartedAt, sub.NextBillingDate,
			sub.CreatedAt, sub.UpdatedAt,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

// GetByUserID returns all subscriptions for a user
func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID string) ([]models.Subscription, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.plan_id, s.plan_name, s.price_per_month, s.agreed_price,
			s.status, s.project_stage, s.started_at, s.next_billing_date, s.canceled_at,
			s.created_at, s.updated_at,
			p.id, p.name, p.description, p.features
		FROM subscriptions s
		LEFT JOIN plans p ON s.plan_id = p.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	for rows.Next() {
		var s models.Subscription
		var pID uuid.NullUUID
		var pName sql.NullString
		var pDesc sql.NullString
		var pFeatures sql.NullString

		err := rows.Scan(
			&s.ID, &s.UserID, &s.PlanID, &s.PlanName, &s.PricePerMonth, &s.AgreedPrice,
			&s.Status, &s.ProjectStage, &s.StartedAt, &s.NextBillingDate, &s.CanceledAt,
			&s.CreatedAt, &s.UpdatedAt,
			&pID, &pName, &pDesc, &pFeatures,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		if pID.Valid {
			s.PlanName = pName.String
			s.Plan = &models.Plan{
				ID:          pID.UUID,
				Name:        pName.String,
				Description: pDesc.String,
			}
			if pFeatures.Valid {
				if err := json.Unmarshal([]byte(pFeatures.String), &s.Plan.Features); err != nil {
					log.Printf("Failed to decode plan features JSON: %v", err)
				}
			}
		}
		
		subscriptions = append(subscriptions, s)
	}

	// Populate logs for each subscription
	for i := range subscriptions {
		logs, err := r.GetLogs(ctx, subscriptions[i].ID)
		if err != nil {
			logs = []models.ProjectLog{}
		}
		subscriptions[i].Logs = logs
	}

	return subscriptions, nil
}

// GetActiveByUserID returns the active subscription for a user if one exists
func (r *SubscriptionRepository) GetActiveByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.plan_id, s.plan_name, s.price_per_month, s.agreed_price,
			s.status, s.project_stage, s.started_at, s.next_billing_date, s.canceled_at,
			s.created_at, s.updated_at,
			p.id, p.name, p.description, p.features
		FROM subscriptions s
		LEFT JOIN plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status = 'ACTIVE'
		ORDER BY s.created_at DESC
		LIMIT 1
	`

	var s models.Subscription
	var pID uuid.NullUUID
	var pName sql.NullString
	var pDesc sql.NullString
	var pFeatures sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.PlanID, &s.PlanName, &s.PricePerMonth, &s.AgreedPrice,
		&s.Status, &s.ProjectStage, &s.StartedAt, &s.NextBillingDate, &s.CanceledAt,
		&s.CreatedAt, &s.UpdatedAt,
		&pID, &pName, &pDesc, &pFeatures,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No active subscription found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscription: %w", err)
	}

	if pID.Valid {
		s.PlanName = pName.String
		s.Plan = &models.Plan{
			ID:          pID.UUID,
			Name:        pName.String,
			Description: pDesc.String,
		}
		if pFeatures.Valid {
			if err := json.Unmarshal([]byte(pFeatures.String), &s.Plan.Features); err != nil {
				log.Printf("Failed to decode plan features JSON: %v", err)
			}
		}
	}

	return &s, nil
}

// UpdateSubscription updates subscription details
func (r *SubscriptionRepository) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET status = $1, project_stage = $2, next_billing_date = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, sub.Status, sub.ProjectStage, sub.NextBillingDate, sub.ID)
	return err
}

// GetSubscriptionByID returns a subscription with details
func (r *SubscriptionRepository) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT 
			s.id, s.user_id, s.plan_id, s.plan_name, s.price_per_month, s.agreed_price,
			s.status, s.project_stage, s.started_at, s.next_billing_date, s.canceled_at,
			s.created_at, s.updated_at,
			p.id, p.name, p.description, p.features,
			u.full_name, u.email
		FROM subscriptions s
		LEFT JOIN plans p ON s.plan_id = p.id
		LEFT JOIN users u ON s.user_id = u.id 
		WHERE s.id = $1
	`
	var s models.Subscription
	var pID uuid.NullUUID
	var pName sql.NullString
	var pDesc sql.NullString
	var pFeatures sql.NullString
	// Variáveis para capturar dados do usuário
	var uFullName sql.NullString
	var uEmail sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.PlanID, &s.PlanName, &s.PricePerMonth, &s.AgreedPrice,
		&s.Status, &s.ProjectStage, &s.StartedAt, &s.NextBillingDate, &s.CanceledAt,
		&s.CreatedAt, &s.UpdatedAt,
		&pID, &pName, &pDesc, &pFeatures,
		&uFullName, &uEmail, // Scan dos novos campos
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if pID.Valid {
		s.PlanName = pName.String 
		s.Plan = &models.Plan{
			ID:          pID.UUID,
			Name:        pName.String,
			Description: pDesc.String,
		}
		if pFeatures.Valid {
			if err := json.Unmarshal([]byte(pFeatures.String), &s.Plan.Features); err != nil {
				log.Printf("Failed to decode plan features JSON: %v", err)
			}
		}
	}

	// Popula o objeto User se encontrou dados
	if uFullName.Valid || uEmail.Valid {
		s.User = &models.User{
			ID:       s.UserID.String(), // Assume conversão de UUID pra string aqui se necessário
			FullName: uFullName.String,
			Email:    uEmail.String,
		}
	}
	
	return &s, nil
}

// AddLog adds a project log
func (r *SubscriptionRepository) AddLog(ctx context.Context, log *models.ProjectLog) error {
	query := `
		INSERT INTO project_logs (id, subscription_id, message, created_by_user_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, log.ID, log.SubscriptionID, log.Message, log.CreatedBy, log.CreatedAt)
	return err
}

// GetLogs returns logs for a subscription
func (r *SubscriptionRepository) GetLogs(ctx context.Context, subscriptionID uuid.UUID) ([]models.ProjectLog, error) {
	query := `
		SELECT id, subscription_id, message, created_by_user_id, created_at
		FROM project_logs
		WHERE subscription_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ProjectLog
	for rows.Next() {
		var l models.ProjectLog
		err := rows.Scan(&l.ID, &l.SubscriptionID, &l.Message, &l.CreatedBy, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	// Retornar array vazio em vez de nil pro JSON ficar bonito []
	if logs == nil {
		logs = []models.ProjectLog{}
	}
	return logs, nil
}

// ListActiveSubscriptions returns all active subscriptions with user and plan details
func (r *SubscriptionRepository) ListActiveSubscriptions(ctx context.Context) ([]models.ActiveSubscriptionDTO, error) {
	query := `
		SELECT 
			s.id, s.user_id, u.full_name, u.email,
			COALESCE(p.name, s.plan_name), s.price_per_month,
			s.status, s.project_stage, s.next_billing_date
		FROM subscriptions s
		INNER JOIN users u ON s.user_id = u.id
		LEFT JOIN plans p ON s.plan_id = p.id
		ORDER BY s.next_billing_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []models.ActiveSubscriptionDTO
	for rows.Next() {
		var s models.ActiveSubscriptionDTO
		var nextBilling time.Time
		var id, userID uuid.UUID
		var userName, userEmail, planName sql.NullString
		
		err := rows.Scan(
			&id, &userID, &userName, &userEmail,
			&planName, &s.Price,
			&s.Status, &s.ProjectStage, &nextBilling,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active subscription: %w", err)
		}
		
		s.ID = id.String()
		s.UserID = userID.String()
		s.UserName = userName.String
		s.UserEmail = userEmail.String
		s.PlanName = planName.String
		s.NextBillingDate = nextBilling.Format(time.RFC3339)
		subscriptions = append(subscriptions, s)
	}

	if subscriptions == nil {
		subscriptions = []models.ActiveSubscriptionDTO{}
	}

	return subscriptions, nil
}