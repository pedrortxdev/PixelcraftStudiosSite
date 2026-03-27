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

// CheckoutService handles checkout business logic
type CheckoutService struct {
	productRepo      *repository.ProductRepository
	discountRepo     *repository.DiscountRepository
	paymentRepo      *repository.PaymentRepository
	userRepo         *repository.UserRepository
	subscriptionRepo *repository.SubscriptionRepository // Injected
	db               *sql.DB
}

// NewCheckoutService creates a new CheckoutService
func NewCheckoutService(
	db *sql.DB,
	productRepo *repository.ProductRepository,
	discountRepo *repository.DiscountRepository,
	paymentRepo *repository.PaymentRepository,
	userRepo *repository.UserRepository,
	subscriptionRepo *repository.SubscriptionRepository, // Added
) *CheckoutService {
	return &CheckoutService{
		db:               db,
		productRepo:      productRepo,
		discountRepo:     discountRepo,
		paymentRepo:      paymentRepo,
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// ProcessCheckout processes a checkout request
func (s *CheckoutService) ProcessCheckout(ctx context.Context, userID uuid.UUID, req *models.CheckoutRequest) (*models.CheckoutResponse, error) {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Garante rollback em caso de pânico ou erro não tratado
	defer tx.Rollback()

	// Validate products/plans and calculate total
	totalAmount, cartItems, err := s.validateCart(ctx, req.Cart)
	if err != nil {
		return nil, fmt.Errorf("cart validation failed: %w", err)
	}

	// Validate and apply discount if provided
	var discountAmount float64
	var discount *models.Discount
	if req.CouponCode != nil && *req.CouponCode != "" {
		discount, discountAmount, err = s.validateAndApplyDiscount(ctx, *req.CouponCode, totalAmount, cartItems)
		if err != nil {
			return nil, fmt.Errorf("discount validation failed: %w", err)
		}
	}

	// Calculate final amount
	finalAmount := totalAmount - discountAmount

	// Check if this should be marked as a test sale
	isTest := false
	if req.UseBalance {
		// If paying with balance, check if the user has any "Teste" adjustments
		var hasTestBalance bool
		qTest := `SELECT EXISTS(SELECT 1 FROM transactions WHERE user_id = $1 AND type = 'admin_adjustment' AND adjustment_type = 'Teste' AND status = 'completed')`
		err = s.db.QueryRowContext(ctx, qTest, userID).Scan(&hasTestBalance)
		if err == nil && hasTestBalance {
			isTest = true
		}

		userBalance, err := s.userRepo.GetBalance(ctx, userID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get user balance: %w", err)
		}

		// If user doesn't have enough balance, return an error
		if userBalance < finalAmount {
			return nil, fmt.Errorf("insufficient balance: user has %.2f but needs %.2f", userBalance, finalAmount)
		}
	}

	// Check product stock (only for products)
	for _, item := range cartItems {
		if item.IsPlan {
			continue // Plans don't have stock
		}
		hasStock, err := s.productRepo.CheckStock(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to check stock for product %s: %w", item.ProductID, err)
		}
		if !hasStock {
			return nil, fmt.Errorf("insufficient stock for product %s", item.ProductID)
		}
	}

	// Create payment record
	payment := &models.Payment{
		ID:              uuid.New(),
		UserID:          userID,
		Description:     "Purchase", // Will be updated if mixed or single
		Amount:          totalAmount,
		DiscountApplied: discountAmount,
		FinalAmount:     finalAmount,
		Status:          models.PaymentStatusCompleted,
		IsTest:          isTest,
		PaymentMethod:   stringPtr("BALANCE"), // Default to balance for now, or whatever logic
		CreatedAt:       time.Now(),
	}
	
	// If using external gateway, status might be PENDING. But here we assume Balance or immediate.
	// If req.UseBalance is false, we might return a pending payment. 
	// For now, let's assume the existing logic handles balance or it's a placeholder.

	// Insert payment record
	paymentID, err := s.createPayment(ctx, tx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Process items (Decrement stock, Add to Library, Create Subscription)
	for _, item := range cartItems {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for item %s", item.ProductID)
		}

		if item.IsPlan {
			// Check if already subscribed
			activePlan, err := s.subscriptionRepo.GetActiveByUserID(ctx, userID.String())
			if err != nil && err.Error() != "subscription not found" {
				return nil, fmt.Errorf("failed to check existing subscriptions: %w", err)
			}
			if activePlan != nil && activePlan.PlanID != nil && *activePlan.PlanID == item.ProductID {
				return nil, fmt.Errorf("user already has an active subscription for this plan")
			}

			// Create Subscription
			sub := &models.Subscription{
				ID:              uuid.New(),
				UserID:          userID,
				PlanID:          &item.ProductID, // ProductID here is actually PlanID
				PlanName:        item.Name,
				PricePerMonth:   item.Price,
				AgreedPrice:     &item.Price,
				Status:          models.SubscriptionStatusActive,
				ProjectStage:    "Planejamento",
				StartedAt:       time.Now(),
				NextBillingDate: time.Now().AddDate(0, 1, 0), // +1 Month
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			err = s.subscriptionRepo.CreateSubscription(ctx, tx, sub)
			if err != nil {
				return nil, fmt.Errorf("failed to create subscription for plan %s: %w", item.ProductID, err)
			}
		} else {
			// Decrement product stock
			err := s.productRepo.DecrementStock(ctx, tx, item.ProductID, item.Quantity)
			if err != nil {
				return nil, fmt.Errorf("failed to decrement stock for product %s: %w", item.ProductID, err)
			}
			
			// Add to library
			err = s.addProductToLibrary(ctx, tx, userID, item.ProductID, paymentID)
			if err != nil {
				return nil, fmt.Errorf("failed to add product to library: %w", err)
			}
		}
	}

	// Deduct user balance if using balance
	if req.UseBalance {
		// BT-005 Atomic mutation
		err := s.userRepo.IncrementBalance(ctx, tx, userID.String(), -finalAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to subtract user balance: %w", err)
		}
	}

	// Increment discount usage if discount was applied
	if discount != nil {
		err := s.discountRepo.IncrementUsage(ctx, tx, discount.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to increment discount usage: %w", err)
		}
	}

	// Referral System Logic
	if req.ReferralCode != "" {
		referrerID, err := s.userRepo.GetUserByReferralCode(ctx, req.ReferralCode)
		if err != nil {
			log.Printf("Checkout: Error getting referrer: %v", err)
		} else if referrerID != nil && *referrerID != userID.String() {
			commission := finalAmount * 0.05
			// BT-005: Use atomic increment to prevent Race Condition
			err = s.userRepo.IncrementBalance(ctx, tx, *referrerID, commission)
			if err != nil {
				log.Printf("Checkout: Error incrementing referrer balance: %v", err)
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &models.CheckoutResponse{
		Success:         true,
		PaymentID:       paymentID,
		FinalAmount:     finalAmount,
		DiscountApplied: discountAmount,
		Message:         "Checkout completed successfully",
	}, nil
}

// ValidateDiscount validates a discount code and calculates the discount amount
func (s *CheckoutService) ValidateDiscount(ctx context.Context, req *models.ValidateDiscountRequest) (*models.ValidateDiscountResponse, error) {
	// If cart items are provided, use the full validation logic (BT-043)
	if len(req.CartItems) > 0 {
		_, cartItems, err := s.validateCart(ctx, req.CartItems)
		if err != nil {
			return &models.ValidateDiscountResponse{
				IsValid: false,
				Message: err.Error(),
			}, nil
		}

		_, discountAmount, errMsg := s.validateDiscountInternal(ctx, req.Code, req.Amount, cartItems)
		if errMsg != "" {
			return &models.ValidateDiscountResponse{
				IsValid: false,
				Message: errMsg,
			}, nil
		}

		return &models.ValidateDiscountResponse{
			IsValid:        true,
			DiscountAmount: discountAmount,
			FinalAmount:    req.Amount - discountAmount,
			Message:        "Cupom aplicado com sucesso",
		}, nil
	}

	// Fallback to basic validation if no cart items provided (legacy or single item check)
	discount, err := s.discountRepo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to get discount: %w", err)
	}
	if discount == nil {
		return &models.ValidateDiscountResponse{
			IsValid: false,
			Message: "Cupom inválido",
		}, nil
	}

	if !discount.IsActive {
		return &models.ValidateDiscountResponse{
			IsValid: false,
			Message: "Cupom não está ativo",
		}, nil
	}

	if discount.ExpiresAt != nil && time.Now().After(*discount.ExpiresAt) {
		return &models.ValidateDiscountResponse{
			IsValid: false,
			Message: "Cupom expirado",
		}, nil
	}

	if discount.MaxUses != nil && discount.CurrentUses >= *discount.MaxUses {
		return &models.ValidateDiscountResponse{
			IsValid: false,
			Message: "Cupom esgotado",
		}, nil
	}

	if discount.RestrictionType != models.RestrictionAll {
		return &models.ValidateDiscountResponse{
			IsValid: false,
			Message: "Este cupom possui restrições. Adicione itens ao carrinho para validar.",
		}, nil
	}

	var discountAmount float64
	switch discount.Type {
	case models.DiscountTypePercentage:
		discountAmount = req.Amount * (discount.Value / 100)
	case models.DiscountTypeFixedAmount:
		discountAmount = discount.Value
	default:
		return &models.ValidateDiscountResponse{
			IsValid: false,
			Message: "Tipo de cupom inválido",
		}, nil
	}

	return &models.ValidateDiscountResponse{
		IsValid:        true,
		DiscountAmount: discountAmount,
		FinalAmount:    req.Amount - discountAmount,
		Message:        "Cupom aplicado com sucesso",
	}, nil
}

// Internal struct to hold validated item details
type validatedCartItem struct {
	ProductID uuid.UUID
	Quantity  int
	Price     float64
	Name      string
	IsPlan    bool
}

// validateCart validates all products/plans in the cart and calculates the total
func (s *CheckoutService) validateCart(ctx context.Context, cart []models.CartItem) (float64, []validatedCartItem, error) {
	var total float64
	var validatedItems []validatedCartItem

	for _, item := range cart {
		// Try to find as Product first
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to check product %s: %w", item.ProductID, err)
		}

		if product != nil {
			// Validate quantity BT-008
			if item.Quantity <= 0 {
				return 0, nil, fmt.Errorf("invalid quantity for product %s", item.ProductID)
			}
			
			// It's a product
			total += product.Price * float64(item.Quantity)
			validatedItems = append(validatedItems, validatedCartItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.Price,
				Name:      product.Name,
				IsPlan:    false,
			})
			continue
		}

		// If not product, try to find as Plan
		plan, err := s.subscriptionRepo.GetPlanByID(ctx, item.ProductID)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to check plan %s: %w", item.ProductID, err)
		}

		if plan != nil {
			// Check if another plan is already in cart or user tries to buy more than 1 plan BT-009
			if item.Quantity != 1 {
				return 0, nil, fmt.Errorf("plans can only be purchased one at a time")
			}
			
			// It's a plan
			total += plan.Price
			validatedItems = append(validatedItems, validatedCartItem{
				ProductID: item.ProductID,
				Quantity:  1, // strictly single quantity
				Price:     plan.Price,
				Name:      plan.Name,
				IsPlan:    true,
			})
			continue
		}

		// Neither product nor plan found
		return 0, nil, fmt.Errorf("item %s not found (neither product nor plan)", item.ProductID)
	}

	return total, validatedItems, nil
}

// validateAndApplyDiscount validates a discount code and calculates the discount amount
// BT-043: Delegates to validateDiscountInternal to avoid duplicating validation logic from ValidateDiscount
func (s *CheckoutService) validateAndApplyDiscount(ctx context.Context, code string, amount float64, cartItems []validatedCartItem) (*models.Discount, float64, error) {
	discount, discountAmount, errMsg := s.validateDiscountInternal(ctx, code, amount, cartItems)
	if errMsg != "" {
		return nil, 0, fmt.Errorf("%s", errMsg)
	}
	return discount, discountAmount, nil
}

// validateDiscountInternal is the shared validation core used by both validateAndApplyDiscount and ValidateDiscount
func (s *CheckoutService) validateDiscountInternal(ctx context.Context, code string, amount float64, cartItems []validatedCartItem) (*models.Discount, float64, string) {
	discount, err := s.discountRepo.GetByCode(ctx, code)
	if err != nil || discount == nil {
		return nil, 0, "Cupom inválido"
	}
	if !discount.IsActive {
		return nil, 0, "Cupom não está ativo"
	}
	if discount.ExpiresAt != nil && time.Now().After(*discount.ExpiresAt) {
		return nil, 0, "Cupom expirado"
	}
	if discount.MaxUses != nil && discount.CurrentUses >= *discount.MaxUses {
		return nil, 0, "Cupom esgotado"
	}

	// Restriction Validation
	if discount.RestrictionType != models.RestrictionAll {
		applicableAmount := 0.0
		foundApplicable := false

		for _, item := range cartItems {
			isApplicable := false
			switch discount.RestrictionType {
			case models.RestrictionProduct:
				for _, targetID := range discount.TargetIDs {
					if item.ProductID == targetID {
						isApplicable = true
						break
					}
				}
			case models.RestrictionItemCategory:
				// Need to fetch product to check category
				product, _ := s.productRepo.GetByID(ctx, item.ProductID)
				if product != nil && product.CategoryID != nil {
					for _, targetID := range discount.TargetIDs {
						if *product.CategoryID == targetID {
							isApplicable = true
							break
						}
					}
				}
			case models.RestrictionGame:
				// Need to fetch product to check game
				product, _ := s.productRepo.GetByID(ctx, item.ProductID)
				if product != nil && product.GameID != nil {
					for _, targetID := range discount.TargetIDs {
						if *product.GameID == targetID {
							isApplicable = true
							break
						}
					}
				}
			}

			if isApplicable {
				applicableAmount += item.Price * float64(item.Quantity)
				foundApplicable = true
			}
		}

		if !foundApplicable {
			return nil, 0, "Este cupom não tem poder sobre estes itens. Tente em outros produtos do nosso catálogo!"
		}
		
		// Recalculate discount based on applicable amount
		var discountAmount float64
		switch discount.Type {
		case models.DiscountTypePercentage:
			discountAmount = applicableAmount * (discount.Value / 100)
		case models.DiscountTypeFixedAmount:
			discountAmount = discount.Value
			if discountAmount > applicableAmount {
				discountAmount = applicableAmount
			}
		default:
			return nil, 0, "Tipo de cupom inválido"
		}
		return discount, discountAmount, ""
	}

	var discountAmount float64
	switch discount.Type {
	case models.DiscountTypePercentage:
		discountAmount = amount * (discount.Value / 100)
	case models.DiscountTypeFixedAmount:
		discountAmount = discount.Value
		if discountAmount > amount {
			discountAmount = amount
		}
	default:
		return nil, 0, "Tipo de cupom inválido"
	}

	return discount, discountAmount, ""
}

// createPayment creates a payment record in the database
func (s *CheckoutService) createPayment(ctx context.Context, tx *sql.Tx, payment *models.Payment) (uuid.UUID, error) {
	query := `
        INSERT INTO payments (id, user_id, description, amount, discount_applied, final_amount, status, is_test, payment_method, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id
    `

	var paymentID uuid.UUID
	err := tx.QueryRowContext(
		ctx, query,
		payment.ID, payment.UserID, payment.Description,
		payment.Amount, payment.DiscountApplied, payment.FinalAmount,
		payment.Status, payment.IsTest, payment.PaymentMethod, payment.CreatedAt,
	).Scan(&paymentID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return paymentID, nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// addProductToLibrary adds a product to the user's library
func (s *CheckoutService) addProductToLibrary(ctx context.Context, tx *sql.Tx, userID uuid.UUID, productID uuid.UUID, paymentID uuid.UUID) error {
	// Get product price to store in purchase record
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product for purchase record: %w", err)
	}
	if product == nil {
		return fmt.Errorf("product not found: %s", productID.String())
	}

	// ON CONFLICT DO NOTHING evita erros se o usuário já tiver o produto
	query := `
        INSERT INTO user_purchases (user_id, product_id, purchase_price, payment_id, purchased_at)
        VALUES ($1, $2, $3, $4, NOW())
        ON CONFLICT (user_id, product_id) DO NOTHING
    `

	_, err = tx.ExecContext(ctx, query, userID, productID, product.Price, paymentID)
	if err != nil {
		return fmt.Errorf("failed to insert into user_purchases: %w", err)
	}

	return nil
}
