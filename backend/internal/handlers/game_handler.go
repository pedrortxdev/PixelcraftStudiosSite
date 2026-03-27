package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pixelcraft/api/internal/models"
	"github.com/pixelcraft/api/internal/repository"
)

// GameHandler handles HTTP requests for games and categories
type GameHandler struct {
	repo *repository.GameRepository
}

// NewGameHandler creates a new GameHandler
func NewGameHandler(repo *repository.GameRepository) *GameHandler {
	return &GameHandler{repo: repo}
}

// ListGames handles GET /games - returns all active games
func (h *GameHandler) ListGames(c *gin.Context) {
	games, err := h.repo.GetAllGames(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch games"})
		return
	}
	c.JSON(http.StatusOK, games)
}

// ListGamesWithCategories handles GET /games/with-categories
func (h *GameHandler) ListGamesWithCategories(c *gin.Context) {
	gamesWithCategories, err := h.repo.GetGamesWithCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch games with categories"})
		return
	}
	c.JSON(http.StatusOK, gamesWithCategories)
}

// GetGameCategories handles GET /games/:id/categories
func (h *GameHandler) GetGameCategories(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	categories, err := h.repo.GetCategoriesByGameID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// CreateGame handles POST /admin/games
func (h *GameHandler) CreateGame(c *gin.Context) {
	var req models.CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game := &models.Game{
		Name:         req.Name,
		Slug:         req.Slug,
		IconURL:      req.IconURL,
		IsActive:     true,
		DisplayOrder: req.DisplayOrder,
	}

	if err := h.repo.CreateGame(c.Request.Context(), game); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
		return
	}

	c.JSON(http.StatusCreated, game)
}

// UpdateGame handles PUT /admin/games/:id
func (h *GameHandler) UpdateGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	var req models.UpdateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing game
	game, err := h.repo.GetGameByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get game"})
		return
	}
	if game == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	// Update fields
	if req.Name != nil {
		game.Name = *req.Name
	}
	if req.Slug != nil {
		game.Slug = *req.Slug
	}
	if req.IconURL != nil {
		game.IconURL = req.IconURL
	}
	if req.IsActive != nil {
		game.IsActive = *req.IsActive
	}
	if req.DisplayOrder != nil {
		game.DisplayOrder = *req.DisplayOrder
	}

	if err := h.repo.UpdateGame(c.Request.Context(), id, game); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game"})
		return
	}

	c.JSON(http.StatusOK, game)
}

// DeleteGame handles DELETE /admin/games/:id
func (h *GameHandler) DeleteGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	if err := h.repo.DeleteGame(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game deleted successfully"})
}

// CreateCategory handles POST /admin/categories
func (h *GameHandler) CreateCategory(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := &models.Category{
		GameID:       req.GameID,
		Name:         req.Name,
		Slug:         req.Slug,
		IsActive:     true,
		DisplayOrder: req.DisplayOrder,
	}

	if err := h.repo.CreateCategory(c.Request.Context(), category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory handles PUT /admin/categories/:id
func (h *GameHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing category
	category, err := h.repo.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get category"})
		return
	}
	if category == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Update fields
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Slug != nil {
		category.Slug = *req.Slug
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.DisplayOrder != nil {
		category.DisplayOrder = *req.DisplayOrder
	}

	if err := h.repo.UpdateCategory(c.Request.Context(), id, category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory handles DELETE /admin/categories/:id
func (h *GameHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	if err := h.repo.DeleteCategory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// ListAllCategories handles GET /categories
func (h *GameHandler) ListAllCategories(c *gin.Context) {
	categories, err := h.repo.GetAllCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}
