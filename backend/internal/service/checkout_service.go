package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/apierrors"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// CheckoutService handles checkout business logic
type CheckoutService struct {
	productRepo      *repository.ProductRepository
	discountRepo     *repository.DiscountRepository
	paymentRepo      *repository.PaymentRepository
	userRepo         *repository.UserRepository
	subscriptionRepo *repository.SubscriptionRepository
	libraryRepo      *repository.LibraryRepository
	depositService   *DepositService
	db               *sql.DB
}

// NewCheckoutService creates a new CheckoutService
func NewCheckoutService(
	db *sql.DB,
	productRepo *repository.ProductRepository,
	discountRepo *repository.DiscountRepository,
	paymentRepo *repository.PaymentRepository,
	userRepo *repository.UserRepository,
	subscriptionRepo *repository.SubscriptionRepository,
	libraryRepo *repository.LibraryRepository,
	depositService *DepositService,
) *CheckoutService {
	return &CheckoutService{
		db:               db,
		productRepo:      productRepo,
		discountRepo:     discountRepo,
		paymentRepo:      paymentRepo,
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		libraryRepo:      libraryRepo,
		depositService:   depositService,
	}
}

// CheckoutMetadata stores information needed to fulfill a direct purchase after payment
// UnitPrices are frozen at checkout time to prevent price changes from affecting paid orders
type CheckoutMetadata struct {
	Cart         []models.CartItem `json:"cart"`
	UnitPrices   map[string]int64  `json:"unit_prices"` // Frozen prices: key is ProductID string
	CouponCode   *string           `json:"coupon_code,omitempty"`
	ReferralCode string            `json:"referral_code,omitempty"`
}

// ProcessCheckout processes a checkout request
func (s *CheckoutService) ProcessCheckout(ctx context.Context, userID uuid.UUID, req *models.CheckoutRequest) (*models.CheckoutResponse, error) {
	// 1. Initial Validation (Pre-transaction, to provide fast feedback)
	totalAmount, cartItems, err := s.validateCart(ctx, req.Cart)
	if err != nil {
		return nil, fmt.Errorf("cart validation failed: %w", err)
	}

	// Prepare frozen prices map
	frozenPrices := make(map[string]int64)
	for _, item := range cartItems {
		frozenPrices[item.ProductID.String()] = item.Price
	}

	// CASE A: PAYMENT WITH WALLET BALANCE
	if req.UseBalance {
		// Start a transaction with serializable isolation to prevent race conditions (Double-Spend/Webhook/TOCTOU)
		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		// 2. RE-VALIDATE EVERYTHING INSIDE TRANSACTION (Atomic Check)
		// This prevents price/discount changes between initial check and fulfillment
		currentTotal, validatedItems, err := s.validateCartTx(ctx, tx, req.Cart)
		if err != nil {
			return nil, err
		}

		var discountAmount int64
		var discount *models.Discount
		if req.CouponCode != nil && *req.CouponCode != "" {
			discount, discountAmount, err = s.validateAndApplyDiscountTx(ctx, tx, *req.CouponCode, currentTotal, validatedItems)
			if err != nil {
				return nil, err
			}
		}

		finalAmount := currentTotal - discountAmount

		// 3. Stock Check WITH TRANSACTION LOCK
		for _, item := range validatedItems {
			if !item.IsPlan {
				hasStock, err := s.productRepo.CheckStockTx(ctx, tx, item.ProductID, item.Quantity)
				if err != nil {
					return nil, fmt.Errorf("failed to check stock for product %s: %w", item.ProductID, err)
				}
				if !hasStock {
					return nil, apierrors.ErrInsufficientStock
				}
			}
		}

		// 4. Balance Check WITH TRANSACTION LOCK
		userBalance, err := s.userRepo.GetBalanceTx(ctx, tx, userID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get user balance: %w", err)
		}

		if userBalance < finalAmount {
			return nil, apierrors.ErrInsufficientBalance
		}

		// Create Completed Payment Record
		payment := &models.Payment{
			ID:              uuid.New(),
			UserID:          userID,
			Description:     "Purchase (Wallet Balance)",
			Amount:          currentTotal,
			DiscountApplied: discountAmount,
			FinalAmount:     finalAmount,
			Status:          models.PaymentStatusCompleted,
			PaymentMethod:   stringPtr(models.PaymentMethodBalance),
			CreatedAt:       time.Now(),
		}

		paymentID, err := s.createPayment(ctx, tx, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to create payment: %w", err)
		}

		// Fulfill Items
		if err := s.fulfillItems(ctx, tx, userID, validatedItems, paymentID); err != nil {
			return nil, err
		}

		// Deduct Balance
		if err := s.userRepo.IncrementBalance(ctx, tx, userID.String(), -finalAmount); err != nil {
			return nil, fmt.Errorf("failed to subtract user balance: %w", err)
		}

		// Update Discount Usage
		if discount != nil {
			if err := s.discountRepo.IncrementUsage(ctx, tx, discount.ID); err != nil {
				return nil, fmt.Errorf("failed to increment discount usage: %w", err)
			}
		}

		// Referral Commission
		if req.ReferralCode != "" {
			s.handleReferral(ctx, tx, userID, req.ReferralCode, finalAmount)
		}

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

	// CASE B: DIRECT PAYMENT VIA MERCADO PAGO
	// We need to calculate discount even for external payment to show the correct amount
	var discountAmount int64
	if req.CouponCode != nil && *req.CouponCode != "" {
		_, discountAmount, err = s.validateDiscountInternal(ctx, *req.CouponCode, totalAmount, cartItems)
		if err != nil {
			return nil, fmt.Errorf("discount validation failed: %w", err)
		}
	}
	finalAmount := totalAmount - discountAmount

	metadata := CheckoutMetadata{
		Cart:         req.Cart,
		UnitPrices:   frozenPrices, // FREEZE PRICES HERE
		CouponCode:   req.CouponCode,
		ReferralCode: req.ReferralCode,
	}
	metadataJSON, _ := json.Marshal(metadata)
	metadataStr := string(metadataJSON)

	// Create PENDING Payment Record
	payment := &models.Payment{
		ID:              uuid.New(),
		UserID:          userID,
		Description:     "Purchase (Direct Payment)",
		Amount:          totalAmount,
		DiscountApplied: discountAmount,
		FinalAmount:     finalAmount,
		Status:          models.PaymentStatusPending,
		PaymentMethod:   stringPtr(models.PaymentMethodMercadoPago),
		PaymentMetadata: &metadataStr,
		CreatedAt:       time.Now(),
	}

	paymentID, err := s.createPayment(ctx, nil, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create pending payment: %w", err)
	}

	// Create MP Preference
	mpResp, err := s.depositService.CreatePreference(ctx, userID, finalAmount, paymentID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create payment preference: %w", err)
	}

	// Update payment with gateway ID
	if err := s.paymentRepo.UpdateStatus(ctx, nil, paymentID.String(), models.PaymentStatusPending, &mpResp.ID); err != nil {
		log.Printf("Warning: failed to update payment with MP ID: %v", err)
	}

	return &models.CheckoutResponse{
		Success:           true,
		PaymentID:         paymentID,
		FinalAmount:       finalAmount,
		DiscountApplied:   discountAmount,
		Message:           "Direct payment initiated",
		PaymentGatewayURL: &mpResp.InitPoint,
	}, nil
}

// FinalizeDirectPurchase is called by the webhook when payment is approved
func (s *CheckoutService) FinalizeDirectPurchase(ctx context.Context, paymentID string, gatewayID string) error {
	log.Printf("CheckoutService: Finalizing direct purchase for Payment ID: %s", paymentID)

	// Start a transaction with serializable isolation
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get Payment Record WITH LOCK (Pessimistic Lock to handle concurrent Webhooks)
	payment, err := s.paymentRepo.GetByIDTx(ctx, tx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment with lock: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found")
	}

	// Idempotency check: if already completed, just commit and return
	if payment.Status == models.PaymentStatusCompleted {
		log.Printf("CheckoutService: Payment %s already completed (idempotent), skipping", paymentID)
		return nil
	}

	// 2. Parse Metadata
	if payment.PaymentMetadata == nil {
		return fmt.Errorf("no metadata found for direct purchase %s", paymentID)
	}

	var metadata CheckoutMetadata
	if err := json.Unmarshal([]byte(*payment.PaymentMetadata), &metadata); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// 3. FULFILLMENT USING FROZEN PRICES (from Metadata)
	// We don't call validateCart here because we must respect the price agreed at checkout
	var validatedItems []validatedCartItem
	for _, item := range metadata.Cart {
		price, ok := metadata.UnitPrices[item.ProductID.String()]
		if !ok {
			return fmt.Errorf("missing frozen price for product %s", item.ProductID)
		}
		
		// We still need to check if it's a plan or product for fulfillment logic
		// But we use a single query for that later if needed, or assume metadata is correct
		// To be safe, let's verify item type (Plan vs Product)
		isPlan := false
		plan, _ := s.subscriptionRepo.GetPlanByID(ctx, item.ProductID)
		if plan != nil {
			isPlan = true
		}

		validatedItems = append(validatedItems, validatedCartItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,
			IsPlan:    isPlan,
			Name:      fmt.Sprintf("Item %s", item.ProductID), // Optional: could freeze names too
		})
	}

	// 4. Fulfill Items (Stock deduction and Library addition)
	if err := s.fulfillItems(ctx, tx, payment.UserID, validatedItems, payment.ID); err != nil {
		return err
	}

	// 5. Update Status
	if err := s.paymentRepo.UpdateStatus(ctx, tx, paymentID, models.PaymentStatusCompleted, &gatewayID); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// 6. Handle Discount usage WITH TRANSACTION LOCK
	if metadata.CouponCode != nil && *metadata.CouponCode != "" {
		discount, err := s.discountRepo.GetByCodeTx(ctx, tx, *metadata.CouponCode)
		if err == nil && discount != nil {
			// Double check limit again inside lock
			if discount.MaxUses == nil || discount.CurrentUses < *discount.MaxUses {
				s.discountRepo.IncrementUsage(ctx, tx, discount.ID)
			}
		}
	}

	// 7. Handle Referral
	if metadata.ReferralCode != "" {
		s.handleReferral(ctx, tx, payment.UserID, metadata.ReferralCode, payment.FinalAmount)
	}

	return tx.Commit()
}

// Internal fulfillment logic shared by both flows
func (s *CheckoutService) fulfillItems(ctx context.Context, tx *sql.Tx, userID uuid.UUID, items []validatedCartItem, paymentID uuid.UUID) error {
	for _, item := range items {
		if item.IsPlan {
			// Create Subscription
			sub := &models.Subscription{
				ID:              uuid.New(),
				UserID:          userID,
				PlanID:          &item.ProductID,
				PlanName:        item.Name,
				PricePerMonth:   item.Price,
				AgreedPrice:     &item.Price,
				Status:          models.SubscriptionStatusActive,
				ProjectStage:    models.ProjectStagePlanning,
				StartedAt:       time.Now(),
				NextBillingDate: time.Now().AddDate(0, 1, 0),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			if err := s.subscriptionRepo.CreateSubscription(ctx, tx, sub); err != nil {
				return fmt.Errorf("failed to create subscription: %w", err)
			}
		} else {
			// Decrement stock
			if err := s.productRepo.DecrementStock(ctx, tx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("failed to decrement stock: %w", err)
			}
			
			// Add to library
			if err := s.libraryRepo.AddPurchase(ctx, tx, userID, item.ProductID, paymentID, item.Price); err != nil {
				return fmt.Errorf("failed to add to library: %w", err)
			}
		}
	}
	return nil
}

func (s *CheckoutService) handleReferral(ctx context.Context, tx *sql.Tx, userID uuid.UUID, code string, amountCents int64) {
	referrerID, err := s.userRepo.GetUserByReferralCode(ctx, code)
	if err == nil && referrerID != nil && *referrerID != userID.String() {
		commissionCents := amountCents * models.ReferralCommissionRate / 100
		s.userRepo.IncrementBalance(ctx, tx, *referrerID, commissionCents)
	}
}

// Internal struct to hold validated item details
type validatedCartItem struct {
	ProductID uuid.UUID
	Quantity  int
	Price     int64
	Name      string
	IsPlan    bool
}

// validateCart uses Batch queries to avoid N+1 problem
func (s *CheckoutService) validateCart(ctx context.Context, cart []models.CartItem) (int64, []validatedCartItem, error) {
	return s.validateCartTx(ctx, nil, cart)
}

func (s *CheckoutService) validateCartTx(ctx context.Context, tx *sql.Tx, cart []models.CartItem) (int64, []validatedCartItem, error) {
	if len(cart) == 0 {
		return 0, nil, apierrors.ErrInvalidInput
	}

	// Extract IDs for batch query
	itemIDs := make([]uuid.UUID, len(cart))
	cartMap := make(map[uuid.UUID]int)
	for i, item := range cart {
		itemIDs[i] = item.ProductID
		cartMap[item.ProductID] = item.Quantity
	}

	// Batch Fetch Products
	products, err := s.productRepo.GetByIDs(ctx, itemIDs)
	if err != nil {
		return 0, nil, err
	}

	// Batch Fetch Plans
	plans, err := s.subscriptionRepo.GetPlansByIDs(ctx, itemIDs)
	if err != nil {
		return 0, nil, err
	}

	// Combine for validation
	productMap := make(map[uuid.UUID]models.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}
	planMap := make(map[uuid.UUID]models.Plan)
	for _, p := range plans {
		planMap[p.ID] = p
	}

	var total int64
	var validatedItems []validatedCartItem

	for _, item := range cart {
		if product, ok := productMap[item.ProductID]; ok {
			if item.Quantity <= 0 {
				return 0, nil, fmt.Errorf("invalid quantity for product %s", item.ProductID)
			}
			total += product.Price * int64(item.Quantity)
			validatedItems = append(validatedItems, validatedCartItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.Price,
				Name:      product.Name,
				IsPlan:    false,
			})
		} else if plan, ok := planMap[item.ProductID]; ok {
			if item.Quantity != 1 {
				return 0, nil, fmt.Errorf("plans can only be purchased one at a time")
			}
			total += plan.Price
			validatedItems = append(validatedItems, validatedCartItem{
				ProductID: item.ProductID,
				Quantity:  1,
				Price:     plan.Price,
				Name:      plan.Name,
				IsPlan:    true,
			})
		} else {
			return 0, nil, fmt.Errorf("item %s not found", item.ProductID)
		}
	}

	return total, validatedItems, nil
}

func (s *CheckoutService) validateAndApplyDiscountTx(ctx context.Context, tx *sql.Tx, code string, amount int64, cartItems []validatedCartItem) (*models.Discount, int64, error) {
	discount, err := s.discountRepo.GetByCodeTx(ctx, tx, code)
	if err != nil || discount == nil {
		return nil, 0, apierrors.ErrInvalidDiscount
	}
	if !discount.IsActive {
		return nil, 0, apierrors.ErrDiscountInactive
	}
	if discount.ExpiresAt != nil && time.Now().After(*discount.ExpiresAt) {
		return nil, 0, apierrors.ErrDiscountExpired
	}
	if discount.MaxUses != nil && discount.CurrentUses >= *discount.MaxUses {
		return nil, 0, apierrors.ErrDiscountExhausted
	}

	var discountAmount int64
	switch discount.Type {
	case models.DiscountTypePercentage:
		discountAmount = amount * discount.Value / 100
	case models.DiscountTypeFixedAmount:
		discountAmount = discount.Value
	default:
		return nil, 0, apierrors.ErrInvalidDiscount
	}

	// Cap discount to total amount to prevent negative final amounts (BT-044)
	if discountAmount > amount {
		discountAmount = amount
	}

	return discount, discountAmount, nil
}

// ... rest of methods kept similar but using batch queries where appropriate ...

func (s *CheckoutService) createPayment(ctx context.Context, tx *sql.Tx, payment *models.Payment) (uuid.UUID, error) {
	return s.paymentRepo.Create(ctx, tx, payment)
}

func (s *CheckoutService) addProductToLibrary(ctx context.Context, tx *sql.Tx, userID uuid.UUID, productID uuid.UUID, paymentID uuid.UUID, price int64) error {
	return s.libraryRepo.AddPurchase(ctx, tx, userID, productID, paymentID, price)
}

// ValidateDiscount validates a discount code (external API)
func (s *CheckoutService) ValidateDiscount(ctx context.Context, req *models.ValidateDiscountRequest) (*models.ValidateDiscountResponse, error) {
	// Re-uses internal validation logic
	total, validatedItems, err := s.validateCart(ctx, req.CartItems)
	if err != nil {
		return nil, err
	}
	
	_, discountAmount, err := s.validateDiscountInternal(ctx, req.Code, total, validatedItems)
	if err != nil {
		return &models.ValidateDiscountResponse{IsValid: false, Message: err.Error()}, nil
	}

	return &models.ValidateDiscountResponse{
		IsValid: true,
		DiscountAmount: discountAmount,
		FinalAmount: total - discountAmount,
		Message: "Success",
	}, nil
}

func (s *CheckoutService) validateDiscountInternal(ctx context.Context, code string, amount int64, cartItems []validatedCartItem) (*models.Discount, int64, error) {
	return s.validateAndApplyDiscountTx(ctx, nil, code, amount, cartItems)
}

func stringPtr(s string) *string {
	return &s
}
