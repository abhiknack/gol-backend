package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpsertProductsWithMatching creates or updates products using the product matching engine
func (r *PostgresRepository) UpsertProductsWithMatching(
	ctx context.Context,
	storeExternalID string,
	products []ProductInput,
	variations []VariationInput,
	storeProducts []StoreProductInput,
) (*UpsertResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result := &UpsertResult{}

	// Get store UUID from external_id
	var storeUUID string
	err = tx.QueryRow(ctx, `SELECT id FROM stores WHERE external_id = $1`, storeExternalID).Scan(&storeUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to find store: %w", err)
	}

	// Map external product IDs to internal UUIDs
	productIDMap := make(map[string]string)      // external_product_id -> product_uuid
	storeProductIDMap := make(map[string]string) // external_product_id -> store_product_uuid

	// Process each product
	for _, p := range products {
		var productUUID string
		var matchType string
		var confidence float64

		// Try to find matching product using the matching engine
		err := tx.QueryRow(ctx, `
			SELECT product_id, match_type, confidence
			FROM find_matching_product($1, $2, $3, $4, $5, $6)
		`, p.Name, p.Barcode, p.SKU, p.EAN, storeUUID, p.ExternalProductID).Scan(&productUUID, &matchType, &confidence)

		if err != nil {
			// No match found - create new product
			r.logger.Info("No matching product found, creating new",
				zap.String("external_product_id", p.ExternalProductID),
				zap.String("name", p.Name))

			// Find or create brand
			var brandUUID *string
			if p.Brand != "" {
				var brandID string
				err := tx.QueryRow(ctx, `SELECT find_or_create_brand($1)`, p.Brand).Scan(&brandID)
				if err == nil && brandID != "" {
					brandUUID = &brandID
				}
			}

			// Find category UUID from external_id
			var categoryUUID *string
			if p.CategoryID != "" {
				var catID string
				err := tx.QueryRow(ctx, `SELECT id FROM categories WHERE external_id = $1`, p.CategoryID).Scan(&catID)
				if err == nil {
					categoryUUID = &catID
				}
			}

			// Create new product
			productUUID = uuid.New().String()
			_, err = tx.Exec(ctx, `
				INSERT INTO products (
					id, sku, name, slug, description, category_id, brand_id,
					base_price, currency, unit, unit_quantity, primary_image_url,
					manufacturer, barcode, ean, is_active, is_featured,
					is_customizable, is_addon
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
				)
			`, productUUID, p.SKU, p.Name, p.Slug, p.Description, categoryUUID, brandUUID,
				p.BasePrice, p.Currency, p.Unit, p.UnitQuantity, p.PrimaryImageURL,
				p.Manufacturer, p.Barcode, p.EAN, p.IsActive, p.IsFeatured,
				p.IsCustomizable, p.IsAddon)

			if err != nil {
				return nil, fmt.Errorf("failed to create product: %w", err)
			}

			result.Created++
		} else {
			// Match found
			r.logger.Info("Found matching product",
				zap.String("external_product_id", p.ExternalProductID),
				zap.String("product_uuid", productUUID),
				zap.String("match_type", matchType),
				zap.Float64("confidence", confidence))

			// Update existing product
			_, err = tx.Exec(ctx, `
				UPDATE products SET
					name = $2,
					description = $3,
					base_price = $4,
					primary_image_url = $5,
					manufacturer = $6,
					is_active = $7,
					is_featured = $8,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = $1
			`, productUUID, p.Name, p.Description, p.BasePrice, p.PrimaryImageURL,
				p.Manufacturer, p.IsActive, p.IsFeatured)

			if err != nil {
				return nil, fmt.Errorf("failed to update product: %w", err)
			}

			result.Updated++
		}

		// Store mapping
		productIDMap[p.ExternalProductID] = productUUID

		// Create/update store_product_mapping
		// Store the external_product_id mapping for later use
		// (will be stored in store_products.external_id)

		// Insert product images
		if len(p.Images) > 0 {
			for idx, imgURL := range p.Images {
				_, err := tx.Exec(ctx, `
					INSERT INTO product_images (product_id, image_url, display_order, is_primary)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (product_id, image_url) DO UPDATE SET
						display_order = EXCLUDED.display_order
				`, productUUID, imgURL, idx, idx == 0)
				if err != nil {
					r.logger.Warn("Failed to insert product image", zap.Error(err))
				}
			}
		}
	}

	// Upsert store products FIRST (before variations, so we have store_product_id)
	if len(storeProducts) > 0 {
		for _, sp := range storeProducts {
			productUUID, ok := productIDMap[sp.ExternalProductID]
			if !ok {
				r.logger.Warn("Product not found for store product", zap.String("external_product_id", sp.ExternalProductID))
				continue
			}

			// Upsert store_product
			var storeProductUUID string
			err := tx.QueryRow(ctx, `
				INSERT INTO store_products (
					external_id, store_id, product_id, price, stock_quantity, is_in_stock, is_available
				) VALUES ($1, $2, $3, $4, $5, $6, true)
				ON CONFLICT (store_id, product_id) DO UPDATE SET
					external_id = EXCLUDED.external_id,
					price = EXCLUDED.price,
					stock_quantity = EXCLUDED.stock_quantity,
					is_in_stock = EXCLUDED.is_in_stock,
					updated_at = CURRENT_TIMESTAMP
				RETURNING id
			`, sp.ExternalProductID, storeUUID, productUUID, sp.Price, sp.StockQuantity, sp.IsInStock).Scan(&storeProductUUID)

			if err != nil {
				r.logger.Error("Failed to upsert store product", zap.String("external_product_id", sp.ExternalProductID), zap.Error(err))
				return nil, fmt.Errorf("failed to upsert store product: %w", err)
			}

			// Store the mapping for variations
			storeProductIDMap[sp.ExternalProductID] = storeProductUUID

			// Upsert store product taxes
			if len(sp.Taxes) > 0 {
				for _, taxExternalID := range sp.Taxes {
					// Find tax UUID by external_id (ERP's tax ID)
					var taxUUID string
					err := tx.QueryRow(ctx, `
						SELECT id FROM taxes 
						WHERE store_id = $1 AND external_id = $2
					`, storeUUID, taxExternalID).Scan(&taxUUID)

					if err != nil {
						r.logger.Warn("Tax not found by external_id",
							zap.String("external_id", taxExternalID),
							zap.String("store_id", storeUUID))
						continue
					}

					// Insert store_product_tax using internal UUID
					_, err = tx.Exec(ctx, `
						INSERT INTO store_product_taxes (store_id, store_product_id, tax_id, is_active)
						VALUES ($1, $2, $3, true)
						ON CONFLICT (store_id, store_product_id, tax_id) DO UPDATE SET
							is_active = true,
							updated_at = CURRENT_TIMESTAMP
					`, storeUUID, storeProductUUID, taxUUID)

					if err != nil {
						r.logger.Warn("Failed to insert store product tax", zap.Error(err))
					} else {
						result.TaxesProcessed++
					}
				}
			}

			result.StoreProductsProcessed++
		}
	}

	// Upsert variations AFTER store_products (so we have store_product_id mapping)
	if len(variations) > 0 {
		for _, v := range variations {
			storeProductUUID, ok := storeProductIDMap[v.ExternalProductID]
			if !ok {
				r.logger.Warn("Store product not found for variation",
					zap.String("external_product_id", v.ExternalProductID),
					zap.String("variation_id", v.ExternalID))
				continue
			}

			_, err := tx.Exec(ctx, `
				INSERT INTO product_variations (
					external_id, store_product_id, name, display_name, price, is_default, is_active
				) VALUES ($1, $2, $3, $4, $5, $6, true)
				ON CONFLICT (store_product_id, name) DO UPDATE SET
					external_id = EXCLUDED.external_id,
					display_name = EXCLUDED.display_name,
					price = EXCLUDED.price,
					is_default = EXCLUDED.is_default,
					updated_at = CURRENT_TIMESTAMP
			`, v.ExternalID, storeProductUUID, v.Name, v.DisplayName, v.Price, v.IsDefault)

			if err != nil {
				r.logger.Error("Failed to upsert variation",
					zap.String("external_product_id", v.ExternalProductID),
					zap.String("variation_id", v.ExternalID),
					zap.Error(err))
				return nil, fmt.Errorf("failed to upsert variation: %w", err)
			}
			result.VariationsProcessed++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Successfully upserted products with matching",
		zap.Int("created", result.Created),
		zap.Int("updated", result.Updated),
		zap.Int("variations", result.VariationsProcessed),
		zap.Int("store_products", result.StoreProductsProcessed),
		zap.Int("taxes", result.TaxesProcessed))

	return result, nil
}
