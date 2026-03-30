package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
)

// UpdateProductRequestPartial represents a partial update request
// Only non-nil fields will be updated to prevent race conditions
type UpdateProductRequestPartial struct {
	Name                 *string
	Description          *string
	Price                *int64
	Type                 *models.ProductType
	GameID               *uuid.UUID
	CategoryID           *uuid.UUID
	DownloadURLEncrypted *[]byte // Encrypted URL
	FileID               *uuid.UUID
	IsExclusive          *bool
	StockQuantity        *int
	ImageURL             *string
	IsActive             *bool
}

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
	return r.GetByIDTx(ctx, nil, id)
}

// GetByIDTx retrieves a single product by ID within a transaction with optional lock
func (r *ProductRepository) GetByIDTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT id, name, description, price, type, game_id, category_id, download_url_encrypted, file_id,
		       is_exclusive, stock_quantity, image_url, is_active, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	if tx != nil {
		query += " FOR UPDATE"
	}

	var p models.Product
	var execTx interface {
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}

	err := execTx.QueryRowContext(ctx, query, id).Scan(
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

// GetByIDs retrieves multiple products by their IDs using a single query
func (r *ProductRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Product, error) {
	if len(ids) == 0 {
		return []models.Product{}, nil
	}

	query := `
		SELECT id, name, description, price, type, game_id, category_id, download_url_encrypted, file_id,
		       is_exclusive, stock_quantity, image_url, is_active, created_at, updated_at
		FROM products
		WHERE id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to query products by IDs: %w", err)
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
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	return products, nil
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

// UpdatePartial updates only the fields that are non-nil in the request
// This prevents race conditions where concurrent updates overwrite each other
func (r *ProductRepository) UpdatePartial(ctx context.Context, id uuid.UUID, req *UpdateProductRequestPartial) error {
	// Build dynamic SET clause with only provided fields
	setClauses := []string{"updated_at = NOW()"}
	var args []interface{}
	argIndex := 2 // Start at $2 since $1 is the ID

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.Price != nil {
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", argIndex))
		args = append(args, *req.Price)
		argIndex++
	}
	if req.Type != nil {
		setClauses = append(setClauses, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *req.Type)
		argIndex++
	}
	if req.GameID != nil {
		setClauses = append(setClauses, fmt.Sprintf("game_id = $%d", argIndex))
		args = append(args, *req.GameID)
		argIndex++
	}
	if req.CategoryID != nil {
		setClauses = append(setClauses, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *req.CategoryID)
		argIndex++
	}
	if req.DownloadURLEncrypted != nil {
		setClauses = append(setClauses, fmt.Sprintf("download_url_encrypted = $%d", argIndex))
		args = append(args, *req.DownloadURLEncrypted)
		argIndex++
	}
	if req.FileID != nil {
		setClauses = append(setClauses, fmt.Sprintf("file_id = $%d", argIndex))
		args = append(args, *req.FileID)
		argIndex++
	}
	if req.IsExclusive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_exclusive = $%d", argIndex))
		args = append(args, *req.IsExclusive)
		argIndex++
	}
	if req.StockQuantity != nil {
		setClauses = append(setClauses, fmt.Sprintf("stock_quantity = $%d", argIndex))
		args = append(args, *req.StockQuantity)
		argIndex++
	}
	if req.ImageURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("image_url = $%d", argIndex))
		args = append(args, *req.ImageURL)
		argIndex++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	query := fmt.Sprintf(`
		UPDATE products
		SET %s
		WHERE id = $1
	`, strings.Join(setClauses, ", "))

	args = append([]interface{}{id}, args...)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
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

// CheckStockTx verifies if a product has sufficient stock within a transaction with row-level lock
func (r *ProductRepository) CheckStockTx(ctx context.Context, tx *sql.Tx, productID uuid.UUID, quantity int) (bool, error) {
	query := `SELECT stock_quantity FROM products WHERE id = $1 AND is_active = true FOR UPDATE`

	var stock *int
	var execTx interface {
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}
	if tx != nil {
		execTx = tx
	} else {
		execTx = r.db
	}
	err := execTx.QueryRowContext(ctx, query, productID).Scan(&stock)
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("product not found or inactive")
	}
	if err != nil {
		return false, fmt.Errorf("failed to check stock (tx): %w", err)
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
