package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
)

// GameRepository handles database operations for games and categories
type GameRepository struct {
	db *sql.DB
}

// NewGameRepository creates a new GameRepository
func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

// GetAllGames retrieves all active games ordered by display_order
func (r *GameRepository) GetAllGames(ctx context.Context) ([]models.Game, error) {
	query := `
		SELECT id, name, slug, icon_url, is_active, display_order, created_at
		FROM games
		WHERE is_active = true
		ORDER BY display_order ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var g models.Game
		if err := rows.Scan(&g.ID, &g.Name, &g.Slug, &g.IconURL, &g.IsActive, &g.DisplayOrder, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, g)
	}

	return games, nil
}

// GetGameByID retrieves a game by its ID
func (r *GameRepository) GetGameByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	query := `
		SELECT id, name, slug, icon_url, is_active, display_order, created_at
		FROM games
		WHERE id = $1
	`

	var g models.Game
	err := r.db.QueryRowContext(ctx, query, id).Scan(&g.ID, &g.Name, &g.Slug, &g.IconURL, &g.IsActive, &g.DisplayOrder, &g.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return &g, nil
}

// CreateGame creates a new game
func (r *GameRepository) CreateGame(ctx context.Context, game *models.Game) error {
	query := `
		INSERT INTO games (id, name, slug, icon_url, is_active, display_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at
	`

	game.ID = uuid.New()
	err := r.db.QueryRowContext(ctx, query, game.ID, game.Name, game.Slug, game.IconURL, game.IsActive, game.DisplayOrder).
		Scan(&game.ID, &game.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	return nil
}

// UpdateGame updates an existing game
func (r *GameRepository) UpdateGame(ctx context.Context, id uuid.UUID, game *models.Game) error {
	query := `
		UPDATE games
		SET name = $2, slug = $3, icon_url = $4, is_active = $5, display_order = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, game.Name, game.Slug, game.IconURL, game.IsActive, game.DisplayOrder)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

// DeleteGame deletes a game (soft delete by setting is_active = false)
func (r *GameRepository) DeleteGame(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE games SET is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}
	return nil
}

// GetCategoriesByGameID retrieves all categories for a game
func (r *GameRepository) GetCategoriesByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Category, error) {
	query := `
		SELECT id, game_id, name, slug, is_active, display_order, created_at
		FROM categories
		WHERE game_id = $1 AND is_active = true
		ORDER BY display_order ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.GameID, &c.Name, &c.Slug, &c.IsActive, &c.DisplayOrder, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}

	return categories, nil
}

// GetAllCategories retrieves all active categories
func (r *GameRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, game_id, name, slug, is_active, display_order, created_at
		FROM categories
		WHERE is_active = true
		ORDER BY display_order ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.GameID, &c.Name, &c.Slug, &c.IsActive, &c.DisplayOrder, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}

	return categories, nil
}

// GetCategoryByID retrieves a category by its ID
func (r *GameRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	query := `
		SELECT id, game_id, name, slug, is_active, display_order, created_at
		FROM categories
		WHERE id = $1
	`

	var c models.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.GameID, &c.Name, &c.Slug, &c.IsActive, &c.DisplayOrder, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &c, nil
}

// CreateCategory creates a new category
func (r *GameRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (id, game_id, name, slug, is_active, display_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at
	`

	category.ID = uuid.New()
	err := r.db.QueryRowContext(ctx, query, category.ID, category.GameID, category.Name, category.Slug, category.IsActive, category.DisplayOrder).
		Scan(&category.ID, &category.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// UpdateCategory updates an existing category
func (r *GameRepository) UpdateCategory(ctx context.Context, id uuid.UUID, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $2, slug = $3, is_active = $4, display_order = $5
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, category.Name, category.Slug, category.IsActive, category.DisplayOrder)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// DeleteCategory deletes a category (soft delete)
func (r *GameRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE categories SET is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}

// GetGamesWithCategories retrieves all games with their categories
func (r *GameRepository) GetGamesWithCategories(ctx context.Context) ([]models.GameWithCategories, error) {
	games, err := r.GetAllGames(ctx)
	if err != nil {
		return nil, err
	}

	var result []models.GameWithCategories
	for _, game := range games {
		categories, err := r.GetCategoriesByGameID(ctx, game.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, models.GameWithCategories{
			Game:       game,
			Categories: categories,
		})
	}

	return result, nil
}
