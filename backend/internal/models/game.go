package models

import (
	"time"

	"github.com/google/uuid"
)

// Game represents a game in the system (e.g., Minecraft, Roblox)
type Game struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Slug         string    `db:"slug" json:"slug"`
	IconURL      *string   `db:"icon_url" json:"icon_url,omitempty"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	DisplayOrder int       `db:"display_order" json:"display_order"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// Category represents a product category within a game
type Category struct {
	ID           uuid.UUID `db:"id" json:"id"`
	GameID       uuid.UUID `db:"game_id" json:"game_id"`
	Name         string    `db:"name" json:"name"`
	Slug         string    `db:"slug" json:"slug"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	DisplayOrder int       `db:"display_order" json:"display_order"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// GameWithCategories includes a game with its categories
type GameWithCategories struct {
	Game       Game       `json:"game"`
	Categories []Category `json:"categories"`
}

// CreateGameRequest represents the request to create a new game
type CreateGameRequest struct {
	Name         string  `json:"name" binding:"required"`
	Slug         string  `json:"slug" binding:"required"`
	IconURL      *string `json:"icon_url"`
	DisplayOrder int     `json:"display_order"`
}

// UpdateGameRequest represents the request to update a game
type UpdateGameRequest struct {
	Name         *string `json:"name"`
	Slug         *string `json:"slug"`
	IconURL      *string `json:"icon_url"`
	IsActive     *bool   `json:"is_active"`
	DisplayOrder *int    `json:"display_order"`
}

// CreateCategoryRequest represents the request to create a new category
type CreateCategoryRequest struct {
	GameID       uuid.UUID `json:"game_id" binding:"required"`
	Name         string    `json:"name" binding:"required"`
	Slug         string    `json:"slug" binding:"required"`
	DisplayOrder int       `json:"display_order"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name         *string `json:"name"`
	Slug         *string `json:"slug"`
	IsActive     *bool   `json:"is_active"`
	DisplayOrder *int    `json:"display_order"`
}
