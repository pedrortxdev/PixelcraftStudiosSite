package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
)

// LibraryRepository handles access to a user's purchased items
type LibraryRepository struct {
	db *sql.DB
}

func NewLibraryRepository(db *sql.DB) *LibraryRepository {
	return &LibraryRepository{db: db}
}

// GetUserLibrary returns the list of purchases with product details for a user
func (r *LibraryRepository) GetUserLibrary(ctx context.Context, userID uuid.UUID) ([]models.UserPurchaseWithProduct, error) {
	query := `
		SELECT
			up.id, up.user_id, up.product_id, up.purchase_price, up.purchased_at,
			p.id, p.name, p.description, p.price, p.type, p.game_id, p.category_id, 
			p.download_url_encrypted, p.file_id, p.is_exclusive, p.stock_quantity, 
			p.image_url, p.is_active, p.created_at, p.updated_at
		FROM user_purchases up
		JOIN products p ON up.product_id = p.id
		WHERE up.user_id = $1
		ORDER BY up.purchased_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user library: %w", err)
	}
	defer rows.Close()

	var items []models.UserPurchaseWithProduct
	for rows.Next() {
		var purchase models.UserPurchase
		var product models.Product

		if err := rows.Scan(
			&purchase.ID, &purchase.UserID, &purchase.ProductID, &purchase.PurchasePrice, &purchase.PurchasedAt,
			&product.ID, &product.Name, &product.Description, &product.Price, &product.Type, 
			&product.GameID, &product.CategoryID, &product.DownloadURLEncrypted, &product.FileID,
			&product.IsExclusive, &product.StockQuantity, &product.ImageURL, &product.IsActive, 
			&product.CreatedAt, &product.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan library item: %w", err)
		}

		items = append(items, models.UserPurchaseWithProduct{Purchase: purchase, Product: product})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}

// GetUserLibraryMinimal returns a minimal list of purchased products for history display (optimized)
func (r *LibraryRepository) GetUserLibraryMinimal(ctx context.Context, userID uuid.UUID) ([]models.ProductMini, error) {
	query := `
		SELECT p.id, p.name, p.price, p.type
		FROM user_purchases up
		JOIN products p ON up.product_id = p.id
		WHERE up.user_id = $1
		ORDER BY up.purchased_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query minimal user library: %w", err)
	}
	defer rows.Close()

	var products []models.ProductMini
	for rows.Next() {
		var p models.ProductMini
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Type); err != nil {
			return nil, fmt.Errorf("failed to scan minimal product: %w", err)
		}
		products = append(products, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

// UserOwnsProduct checks if the user purchased a product
func (r *LibraryRepository) UserOwnsProduct(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (bool, error) {
	query := `SELECT COUNT(1) FROM user_purchases WHERE user_id = $1 AND product_id = $2`
	var count int
	if err := r.db.QueryRowContext(ctx, query, userID, productID).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check ownership: %w", err)
	}
	return count > 0, nil
}

// AddPurchase adds a product to the user's library within a transaction
// Price is in cents (int64) to avoid float precision issues
func (r *LibraryRepository) AddPurchase(ctx context.Context, tx *sql.Tx, userID, productID, paymentID uuid.UUID, price int64) error {
	query := `
		INSERT INTO user_purchases (user_id, product_id, purchase_price, payment_id, purchased_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, product_id) DO NOTHING
	`

	var execTx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	_, err := execTx.ExecContext(ctx, query, userID, productID, price, paymentID)
	return err
}