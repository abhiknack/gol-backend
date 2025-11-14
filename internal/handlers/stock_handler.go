package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/supabase-redis-middleware/internal/repository"
	"go.uber.org/zap"
)

type StockHandler struct {
	pgRepo *repository.PostgresRepository
	logger *zap.Logger
}

func NewStockHandler(pgRepo *repository.PostgresRepository, logger *zap.Logger) *StockHandler {
	return &StockHandler{
		pgRepo: pgRepo,
		logger: logger,
	}
}

// UpdateStockRequest represents the stock update payload
type UpdateStockRequest struct {
	StoreID  string               `json:"store_id" binding:"required"`
	Products []StockProductUpdate `json:"products" binding:"required"`
}

// StockProductUpdate represents individual product stock update
type StockProductUpdate struct {
	ID            string               `json:"id" binding:"required"` // External product ID
	StockQuantity float64              `json:"stock_quantity"`
	IsAvailable   bool                 `json:"is_available"`
	Price         float64              `json:"price"`    // Optional: update price
	Variants      []StockVariantUpdate `json:"variants"` // Optional: variation stock updates
}

// StockVariantUpdate represents individual variation stock update
type StockVariantUpdate struct {
	ID            string  `json:"id" binding:"required"` // External variation ID
	StockQuantity float64 `json:"stock_quantity"`
	IsAvailable   bool    `json:"is_available"`
	Price         float64 `json:"price"` // Optional: update price
}

// UpdateStock handles bulk stock updates for a store
// POST /api/v1/products/stock
func (h *StockHandler) UpdateStock(c *gin.Context) {
	var req UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": err.Error(),
			},
		})
		return
	}

	// Convert to repository type
	repoProducts := make([]repository.StockProductUpdate, len(req.Products))
	for i, p := range req.Products {
		// Convert variants
		repoVariants := make([]repository.StockVariantUpdate, len(p.Variants))
		for j, v := range p.Variants {
			repoVariants[j] = repository.StockVariantUpdate{
				ID:            v.ID,
				StockQuantity: v.StockQuantity,
				IsAvailable:   v.IsAvailable,
				Price:         v.Price,
			}
		}

		repoProducts[i] = repository.StockProductUpdate{
			ID:            p.ID,
			StockQuantity: p.StockQuantity,
			IsAvailable:   p.IsAvailable,
			Price:         p.Price,
			Variants:      repoVariants,
		}
	}

	// Update stock
	result, err := h.pgRepo.BulkUpdateStock(c.Request.Context(), req.StoreID, repoProducts)
	if err != nil {
		h.logger.Error("Failed to update stock", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "STOCK_UPDATE_FAILED",
				"message": "Failed to update stock",
			},
		})
		return
	}

	h.logger.Info("Successfully updated stock",
		zap.String("store_id", req.StoreID),
		zap.Int("products_updated", result.Updated),
		zap.Int("products_not_found", result.NotFound),
		zap.Int("variants_updated", result.VariantsUpdated),
		zap.Int("variants_not_found", result.VariantsNotFound))

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"products_updated":   result.Updated,
			"products_not_found": result.NotFound,
			"variants_updated":   result.VariantsUpdated,
			"variants_not_found": result.VariantsNotFound,
		},
		"message": "Stock updated successfully",
	})
}
