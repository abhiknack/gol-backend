package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/supabase-redis-middleware/internal/repository"
	"go.uber.org/zap"
)

type StoreHandler struct {
	pgRepo *repository.PostgresRepository
	logger *zap.Logger
}

func NewStoreHandler(pgRepo *repository.PostgresRepository, logger *zap.Logger) *StoreHandler {
	return &StoreHandler{
		pgRepo: pgRepo,
		logger: logger,
	}
}

// GetStoreBasicData retrieves basic store information
func (h *StoreHandler) GetStoreBasicData(c *gin.Context) {
	storeID := c.Param("id")

	store, err := h.pgRepo.GetStoreByID(c.Request.Context(), storeID)
	if err != nil {
		h.logger.Error("Failed to get store", zap.String("store_id", storeID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "STORE_NOT_FOUND",
				"message": "Store not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   store,
	})
}

// UpdateStoreStatus updates store active/open status
func (h *StoreHandler) UpdateStoreStatus(c *gin.Context) {
	storeID := c.Param("id")

	var input struct {
		IsActive *bool `json:"is_active"`
		IsOpen   *bool `json:"is_open"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": err.Error(),
			},
		})
		return
	}

	if input.IsActive == nil && input.IsOpen == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "At least one of is_active or is_open must be provided",
			},
		})
		return
	}

	err := h.pgRepo.UpdateStoreStatus(c.Request.Context(), storeID, input.IsActive, input.IsOpen)
	if err != nil {
		h.logger.Error("Failed to update store status",
			zap.String("store_id", storeID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "UPDATE_FAILED",
				"message": "Failed to update store status",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Store status updated successfully",
	})
}

// GetStoreStatus retrieves store status information
func (h *StoreHandler) GetStoreStatus(c *gin.Context) {
	storeID := c.Param("id")

	status, err := h.pgRepo.GetStoreStatus(c.Request.Context(), storeID)
	if err != nil {
		h.logger.Error("Failed to get store status", zap.String("store_id", storeID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "STORE_NOT_FOUND",
				"message": "Store not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   status,
	})
}

// UpdateStoreDetails updates store information
func (h *StoreHandler) UpdateStoreDetails(c *gin.Context) {
	storeID := c.Param("id")

	var input repository.UpdateStoreDetailsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": err.Error(),
			},
		})
		return
	}

	err := h.pgRepo.UpdateStoreDetails(c.Request.Context(), storeID, input)
	if err != nil {
		h.logger.Error("Failed to update store details",
			zap.String("store_id", storeID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "UPDATE_FAILED",
				"message": "Failed to update store details",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Store details updated successfully",
	})
}
