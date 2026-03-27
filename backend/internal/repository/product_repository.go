package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pixelcraft/api/internal/models"

	"github.com/google/uuid"
)

// ProductRepository handles all database operations for products
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// GetAll retrieves all active products with pagination and filtering
func (r *ProductRepository) GetAll(ctx context.Context, page, pageSize int, productType *models.ProductType, gameID *uuid.UUID, categoryID *uuid.UUID) ([]models.Product, int, error) {
	offset := (page - 1) * pageSize
	
	// Build dynamic query with filters
	baseQuery := `
		SELECT id, name, description, price, type, game_id, category_id, download_url_encrypted, file_id,
		       is_exclusive, stock_quantity, image_url, is_active, created_at, updated_at
		FROM products
		WHERE is_active = true
	`
	countQuery := `SELECT COUNT(*) FROM products WHERE is_active = true`
	
	var args []interface{}
	argIndex := 1
	
	if productType != nil {
		baseQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *productType)
		argIndex++
	}
	
	if gameID != nil {
		baseQuery += fmt.Sprintf(" AND game_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND game_id = $%d", argIndex)
		args = append(args, *gameID)
		argIndex++
	}
	
	if categoryID != nil {
		baseQuery += fmt.Sprintf(" AND category_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND category_id = $%d", argIndex)
		args = append(args, *categoryID)
		argIndex++
	}
	
	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}
	
	// Add pagination
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)
	
	// Get products
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()
	
	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Type,
			&p.GameID, &p.CategoryID, &p.DownloadURLEncrypted, &p.FileID,
			&p.IsExclusive, &p.StockQuantity,
			&p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}
	
	return products, total, nil
}

// GetByID retrieves a single product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT id, name, description, price, type, game_id, category_id, download_url_encrypted, file_id,
		       is_exclusive, stock_quantity, image_url, is_active, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	
	var p models.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Type,
		&p.GameID, &p.CategoryID, &p.DownloadURLEncrypted, &p.FileID,
		&p.IsExclusive, &p.StockQuantity,
		&p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	
	return &p, nil
}

// Create creates a new product (requires encrypted download URL)
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (name, description, price, type, game_id, category_id, download_url_encrypted, file_id,
		                     is_exclusive, stock_quantity, image_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(
		ctx, query,
		product.Name, product.Description, product.Price, product.Type,
		product.GameID, product.CategoryID, product.DownloadURLEncrypted, product.FileID,
		product.IsExclusive, product.StockQuantity,
		product.ImageURL, product.IsActive,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	
	return nil
}

// Update updates an existing product
func (r *ProductRepository) Update(ctx context.Context, id uuid.UUID, product *models.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, type = $4, game_id = $5, category_id = $6,
		    is_exclusive = $7, stock_quantity = $8, image_url = $9, is_active = $10,
			download_url_encrypted = $11, file_id = $12,
		    updated_at = NOW()
		WHERE id = $13
		RETURNING updated_at
	`
	
	err := r.db.QueryRowContext(
		ctx, query,
		product.Name, product.Description, product.Price, product.Type,
		product.GameID, product.CategoryID,
		product.IsExclusive, product.StockQuantity, product.ImageURL,
		product.IsActive, product.DownloadURLEncrypted, product.FileID, id,
	).Scan(&product.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return fmt.Errorf("product not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	
	return nil
}

// Delete soft deletes a product (sets is_active to false)
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET is_active = false, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rows == 0 {
		return fmt.Errorf("product not found")
	}
	
	return nil
}

// CheckStock verifies if a product has sufficient stock
func (r *ProductRepository) CheckStock(ctx context.Context, productID uuid.UUID, quantity int) (bool, error) {
	query := `SELECT stock_quantity FROM products WHERE id = $1 AND is_active = true`
	
	var stock *int
	err := r.db.QueryRowContext(ctx, query, productID).Scan(&stock)
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("product not found or inactive")
	}
	if err != nil {
		return false, fmt.Errorf("failed to check stock: %w", err)
	}
	
	// NULL stock means unlimited
	if stock == nil {
		return true, nil
	}
	
	return *stock >= quantity, nil
}

// DecrementStock decrements the stock of a product (within a transaction)
// Modified to handle NULL stock_quantity as infinite stock
func (r *ProductRepository) DecrementStock(ctx context.Context, tx *sql.Tx, productID uuid.UUID, quantity int) error {
    // A query foi alterada para usar lógica condicional (CASE e OR)
    query := `
        UPDATE products
        SET stock_quantity = CASE
            WHEN stock_quantity IS NULL THEN NULL   -- Mantém infinito (NULL) se já for infinito
            ELSE stock_quantity - $1                -- Subtrai se tiver um número
        END
        WHERE id = $2 
        AND (stock_quantity IS NULL OR stock_quantity >= $1) -- Passa se for infinito OU se tiver saldo suficiente
        RETURNING stock_quantity
    `
    
    var newStock *int
    err := tx.QueryRowContext(ctx, query, quantity, productID).Scan(&newStock)
    
    if err == sql.ErrNoRows {
        // Agora esse erro só acontece se o ID não existir 
        // OU se o estoque for numérico e menor que a quantidade solicitada
        return fmt.Errorf("insufficient stock or product not found")
    }
    if err != nil {
        return fmt.Errorf("failed to decrement stock: %w", err)
    }
    
    return nil
}
