package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/supabase-redis-middleware/internal/repository"
	"go.uber.org/zap"
)

type ProductHandler struct {
	pgRepo *repository.PostgresRepository
	logger *zap.Logger
}

func NewProductHandler(pgRepo *repository.PostgresRepository, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		pgRepo: pgRepo,
		logger: logger,
	}
}

// PushProductsRequest represents the incoming payload structure
type PushProductsRequest struct {
	Categories    []Category     `json:"categories"`
	Taxes         []Tax          `json:"taxes"`
	Products      []Product      `json:"products" binding:"required"`
	Variations    []Variation    `json:"variations"`
	StoreProducts []StoreProduct `json:"store_products"`
	StoreDetails  StoreDetails   `json:"store_details" binding:"required"`
}

type Category struct {
	ID           string  `json:"id" binding:"required"`
	ParentID     *string `json:"parent_id"`
	Name         string  `json:"name" binding:"required"`
	Slug         string  `json:"slug" binding:"required"`
	Description  string  `json:"description"`
	DisplayOrder int     `json:"display_order"`
	IsActive     bool    `json:"is_active"`
}

type Tax struct {
	ID          string  `json:"id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	TaxID       string  `json:"tax_id" binding:"required"`
	Description string  `json:"description"`
	Rate        float64 `json:"rate" binding:"required"`
	TaxType     string  `json:"tax_type" binding:"required"`
	IsInclusive bool    `json:"is_inclusive"`
	IsActive    bool    `json:"is_active"`
}

type Product struct {
	ID              string   `json:"id" binding:"required"` // External product ID from ERP
	SKU             string   `json:"sku" binding:"required"`
	Name            string   `json:"name" binding:"required"`
	Slug            string   `json:"slug"`
	Description     string   `json:"description"`
	CategoryID      string   `json:"category_id"`
	Price           float64  `json:"price" binding:"required"`
	Currency        string   `json:"currency"`
	Unit            string   `json:"unit"`
	UnitQuantity    float64  `json:"unit_quantity"`
	PrimaryImageURL string   `json:"primary_image_url"`
	Images          []string `json:"images"`
	Brand           string   `json:"brand"`
	Manufacturer    string   `json:"manufacturer"`
	Barcode         string   `json:"barcode"`
	EAN             string   `json:"ean"`
	Taxes           []string `json:"taxes"` // Tax external IDs for this product
	IsActive        bool     `json:"is_active"`
	IsFeatured      bool     `json:"is_featured"`
	IsCustomizable  bool     `json:"is_customizable"`
	IsAddon         bool     `json:"is_addon"`
}

type Variation struct {
	ID          string  `json:"id"` // External variation ID
	ProductID   string  `json:"product_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	DisplayName string  `json:"display_name" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	IsDefault   bool    `json:"is_default"`
}

type StoreProduct struct {
	ProductID     string   `json:"product_id" binding:"required"` // Links to Product.id
	Price         float64  `json:"price" binding:"required"`
	StockQuantity float64  `json:"stock_quantity"`
	IsInStock     bool     `json:"is_in_stock"`
	Taxes         []string `json:"taxes"` // Tax IDs for this store-product
}

type StoreDetails struct {
	StoreID  string   `json:"store_id" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Address  Address  `json:"address" binding:"required"`
	Location Location `json:"location" binding:"required"`
}

type Address struct {
	Line1      string `json:"line1" binding:"required"`
	City       string `json:"city" binding:"required"`
	State      string `json:"state" binding:"required"`
	PostalCode string `json:"postal_code" binding:"required"`
}

type Location struct {
	Lat float64 `json:"lat" binding:"required"`
	Lng float64 `json:"lng" binding:"required"`
}

// PushProducts handles bulk product upsert
func (h *ProductHandler) PushProducts(c *gin.Context) {
	var req PushProductsRequest
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

	// Validate store exists or create/update it
	storeInput := repository.StoreDetailsInput{
		StoreID: req.StoreDetails.StoreID,
		Name:    req.StoreDetails.Name,
		Address: repository.AddressInput{
			Line1:      req.StoreDetails.Address.Line1,
			City:       req.StoreDetails.Address.City,
			State:      req.StoreDetails.Address.State,
			PostalCode: req.StoreDetails.Address.PostalCode,
		},
		Location: repository.LocationInput{
			Lat: req.StoreDetails.Location.Lat,
			Lng: req.StoreDetails.Location.Lng,
		},
	}
	if err := h.pgRepo.UpsertStore(c.Request.Context(), storeInput); err != nil {
		h.logger.Error("Failed to upsert store", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "STORE_UPSERT_FAILED",
				"message": "Failed to create or update store",
			},
		})
		return
	}

	// Upsert categories
	if len(req.Categories) > 0 {
		categoryInputs := make([]repository.CategoryInput, len(req.Categories))
		for i, cat := range req.Categories {
			categoryInputs[i] = repository.CategoryInput{
				ID:           cat.ID,
				ParentID:     cat.ParentID,
				Name:         cat.Name,
				Slug:         cat.Slug,
				Description:  cat.Description,
				DisplayOrder: cat.DisplayOrder,
				IsActive:     cat.IsActive,
			}
		}
		if err := h.pgRepo.UpsertCategories(c.Request.Context(), categoryInputs); err != nil {
			h.logger.Error("Failed to upsert categories", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error": gin.H{
					"code":    "CATEGORY_UPSERT_FAILED",
					"message": "Failed to create or update categories",
				},
			})
			return
		}
	}

	// Upsert taxes
	if len(req.Taxes) > 0 {
		taxInputs := make([]repository.TaxInput, len(req.Taxes))
		for i, tax := range req.Taxes {
			taxInputs[i] = repository.TaxInput{
				ID:          tax.ID,
				Name:        tax.Name,
				TaxID:       tax.TaxID,
				Description: tax.Description,
				Rate:        tax.Rate,
				TaxType:     tax.TaxType,
				IsInclusive: tax.IsInclusive,
				IsActive:    tax.IsActive,
			}
		}
		if err := h.pgRepo.UpsertTaxes(c.Request.Context(), taxInputs, req.StoreDetails.StoreID); err != nil {
			h.logger.Error("Failed to upsert taxes", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error": gin.H{
					"code":    "TAX_UPSERT_FAILED",
					"message": "Failed to create or update taxes",
				},
			})
			return
		}
	}

	// Convert products - map payload fields to internal structure
	productInputs := make([]repository.ProductInput, len(req.Products))
	for i, prod := range req.Products {
		// Generate slug if not provided
		slug := prod.Slug
		if slug == "" {
			slug = prod.SKU // Fallback to SKU
		}

		productInputs[i] = repository.ProductInput{
			ExternalProductID: prod.ID, // Map id -> ExternalProductID
			SKU:               prod.SKU,
			Name:              prod.Name,
			Slug:              slug,
			Description:       prod.Description,
			CategoryID:        prod.CategoryID,
			BasePrice:         prod.Price,
			Currency:          prod.Currency,
			Unit:              prod.Unit,
			UnitQuantity:      prod.UnitQuantity,
			PrimaryImageURL:   prod.PrimaryImageURL,
			Images:            prod.Images,
			Brand:             prod.Brand,
			Manufacturer:      prod.Manufacturer,
			Barcode:           prod.Barcode,
			EAN:               prod.EAN,
			IsActive:          prod.IsActive,
			IsFeatured:        prod.IsFeatured,
			IsCustomizable:    prod.IsCustomizable,
			IsAddon:           prod.IsAddon,
		}
	}

	// Convert variations
	variationInputs := make([]repository.VariationInput, len(req.Variations))
	for i, v := range req.Variations {
		variationInputs[i] = repository.VariationInput{
			ExternalID:        v.ID, // Map variation ID to external_id
			ExternalProductID: v.ProductID,
			Name:              v.Name,
			DisplayName:       v.DisplayName,
			Price:             v.Price,
			IsDefault:         v.IsDefault,
		}
	}

	// Convert store products
	// If store_products array is provided, use it; otherwise auto-generate from products
	var storeProductInputs []repository.StoreProductInput

	if len(req.StoreProducts) > 0 {
		// Use provided store_products
		storeProductInputs = make([]repository.StoreProductInput, len(req.StoreProducts))
		for i, sp := range req.StoreProducts {
			storeProductInputs[i] = repository.StoreProductInput{
				ExternalProductID:    sp.ProductID,
				ExternalStoreProduct: "",
				StoreID:              req.StoreDetails.StoreID,
				Price:                sp.Price,
				StockQuantity:        sp.StockQuantity,
				IsInStock:            sp.IsInStock,
				Taxes:                sp.Taxes,
			}
		}
	} else {
		// Auto-generate store_products from products
		storeProductInputs = make([]repository.StoreProductInput, len(req.Products))
		for i, prod := range req.Products {
			storeProductInputs[i] = repository.StoreProductInput{
				ExternalProductID:    prod.ID,
				ExternalStoreProduct: "",
				StoreID:              req.StoreDetails.StoreID,
				Price:                prod.Price,
				StockQuantity:        0,          // Default stock
				IsInStock:            true,       // Default in stock
				Taxes:                prod.Taxes, // Use taxes from product
			}
		}
	}

	// Upsert products (main operation)
	result, err := h.pgRepo.UpsertProductsWithMatching(
		c.Request.Context(),
		req.StoreDetails.StoreID,
		productInputs,
		variationInputs,
		storeProductInputs,
	)
	if err != nil {
		h.logger.Error("Failed to upsert products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "PRODUCT_UPSERT_FAILED",
				"message": "Failed to create or update products",
			},
		})
		return
	}

	h.logger.Info("Successfully pushed products",
		zap.Int("products_created", result.Created),
		zap.Int("products_updated", result.Updated),
		zap.Int("variations_processed", result.VariationsProcessed),
		zap.Int("store_products_processed", result.StoreProductsProcessed),
		zap.Int("taxes_processed", result.TaxesProcessed))

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"products_created":         result.Created,
			"products_updated":         result.Updated,
			"variations_processed":     result.VariationsProcessed,
			"store_products_processed": result.StoreProductsProcessed,
			"taxes_processed":          result.TaxesProcessed,
		},
		"message": "Products pushed successfully",
	})
}
