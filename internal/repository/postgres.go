package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PostgresRepository handles PostgreSQL database operations
type PostgresRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(databaseURL string, logger *zap.Logger) (*PostgresRepository, error) {
	// Parse and validate the connection string
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL",
		zap.String("database", config.ConnConfig.Database),
		zap.String("host", config.ConnConfig.Host),
		zap.Uint16("port", config.ConnConfig.Port),
	)

	return &PostgresRepository{
		pool:   pool,
		logger: logger,
	}, nil
}

// Close closes the database connection pool
func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
		r.logger.Info("PostgreSQL connection pool closed")
	}
}

// Ping checks if the database connection is alive
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// GetPool returns the underlying connection pool for direct access
func (r *PostgresRepository) GetPool() *pgxpool.Pool {
	return r.pool
}

// QuerySupermarketProducts retrieves supermarket products with optional filters
func (r *PostgresRepository) QuerySupermarketProducts(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, category, price, stock, description, created_at, updated_at
		FROM supermarket_products
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	// Add category filter if provided
	if category, ok := filters["category"].(string); ok && category != "" {
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
		argCount++
	}

	// Add search filter if provided
	if search, ok := filters["search"].(string); ok && search != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to query supermarket products", zap.Error(err))
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var name, category, description string
		var price float64
		var stock int
		var createdAt, updatedAt interface{}

		if err := rows.Scan(&id, &name, &category, &price, &stock, &description, &createdAt, &updatedAt); err != nil {
			r.logger.Error("Failed to scan product row", zap.Error(err))
			continue
		}

		results = append(results, map[string]interface{}{
			"id":          id,
			"name":        name,
			"category":    category,
			"price":       price,
			"stock":       stock,
			"description": description,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GetSupermarketProductByID retrieves a single supermarket product by ID
func (r *PostgresRepository) GetSupermarketProductByID(ctx context.Context, id int) (map[string]interface{}, error) {
	query := `
		SELECT id, name, category, price, stock, description, created_at, updated_at
		FROM supermarket_products
		WHERE id = $1
	`

	var productID int
	var name, category, description string
	var price float64
	var stock int
	var createdAt, updatedAt interface{}

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&productID, &name, &category, &price, &stock, &description, &createdAt, &updatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to get product by ID", zap.Int("id", id), zap.Error(err))
		return nil, fmt.Errorf("product not found: %w", err)
	}

	return map[string]interface{}{
		"id":          productID,
		"name":        name,
		"category":    category,
		"price":       price,
		"stock":       stock,
		"description": description,
		"created_at":  createdAt,
		"updated_at":  updatedAt,
	}, nil
}

// QueryMovies retrieves movies with optional filters
func (r *PostgresRepository) QueryMovies(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT id, title, genre, duration, rating, release_date, description, created_at, updated_at
		FROM movies
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	// Add genre filter if provided
	if genre, ok := filters["genre"].(string); ok && genre != "" {
		query += fmt.Sprintf(" AND genre = $%d", argCount)
		args = append(args, genre)
		argCount++
	}

	// Add ordering and pagination
	query += " ORDER BY release_date DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to query movies", zap.Error(err))
		return nil, fmt.Errorf("failed to query movies: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, duration int
		var title, genre, description string
		var rating float64
		var releaseDate, createdAt, updatedAt interface{}

		if err := rows.Scan(&id, &title, &genre, &duration, &rating, &releaseDate, &description, &createdAt, &updatedAt); err != nil {
			r.logger.Error("Failed to scan movie row", zap.Error(err))
			continue
		}

		results = append(results, map[string]interface{}{
			"id":           id,
			"title":        title,
			"genre":        genre,
			"duration":     duration,
			"rating":       rating,
			"release_date": releaseDate,
			"description":  description,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryMedicines retrieves medicines with optional filters
func (r *PostgresRepository) QueryMedicines(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, category, price, prescription_required, stock, description, created_at, updated_at
		FROM medicines
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	// Add category filter if provided
	if category, ok := filters["category"].(string); ok && category != "" {
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
		argCount++
	}

	// Add search filter if provided
	if search, ok := filters["search"].(string); ok && search != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to query medicines", zap.Error(err))
		return nil, fmt.Errorf("failed to query medicines: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, stock int
		var name, category, description string
		var price float64
		var prescriptionRequired bool
		var createdAt, updatedAt interface{}

		if err := rows.Scan(&id, &name, &category, &price, &prescriptionRequired, &stock, &description, &createdAt, &updatedAt); err != nil {
			r.logger.Error("Failed to scan medicine row", zap.Error(err))
			continue
		}

		results = append(results, map[string]interface{}{
			"id":                    id,
			"name":                  name,
			"category":              category,
			"price":                 price,
			"prescription_required": prescriptionRequired,
			"stock":                 stock,
			"description":           description,
			"created_at":            createdAt,
			"updated_at":            updatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// ExecuteQuery executes a raw SQL query (for advanced use cases)
func (r *PostgresRepository) ExecuteQuery(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to execute query", zap.String("query", query), zap.Error(err))
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column descriptions
	fieldDescriptions := rows.FieldDescriptions()
	var results []map[string]interface{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			r.logger.Error("Failed to get row values", zap.Error(err))
			continue
		}

		row := make(map[string]interface{})
		for i, col := range fieldDescriptions {
			row[string(col.Name)] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// ProductCreate represents data for creating a new product
type ProductCreate struct {
	SKU                  string   `json:"sku" binding:"required"`
	Name                 string   `json:"name" binding:"required"`
	Description          string   `json:"description"`
	CategoryID           *string  `json:"category_id"`
	BasePrice            float64  `json:"base_price" binding:"required,min=0"`
	SalePrice            *float64 `json:"sale_price"`
	Unit                 string   `json:"unit"`
	UnitQuantity         float64  `json:"unit_quantity" binding:"min=0"`
	Brand                string   `json:"brand"`
	IsActive             bool     `json:"is_active"`
	RequiresPrescription bool     `json:"requires_prescription"`
}

// BulkCreateProducts creates multiple products in a single transaction
func (r *PostgresRepository) BulkCreateProducts(ctx context.Context, products []ProductCreate) ([]map[string]interface{}, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO products (sku, name, description, category_id, base_price, sale_price, 
			unit, unit_quantity, brand, is_active, requires_prescription, slug)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, sku, name, base_price, is_active, created_at
	`

	var createdProducts []map[string]interface{}

	for _, product := range products {
		slug := generateSlug(product.Name)

		var id, sku, name string
		var basePrice float64
		var isActive bool
		var createdAt interface{}

		err := tx.QueryRow(ctx, query,
			product.SKU,
			product.Name,
			product.Description,
			product.CategoryID,
			product.BasePrice,
			product.SalePrice,
			product.Unit,
			product.UnitQuantity,
			product.Brand,
			product.IsActive,
			product.RequiresPrescription,
			slug,
		).Scan(&id, &sku, &name, &basePrice, &isActive, &createdAt)

		if err != nil {
			r.logger.Error("Failed to insert product",
				zap.String("sku", product.SKU),
				zap.Error(err))
			return nil, fmt.Errorf("failed to insert product %s: %w", product.SKU, err)
		}

		createdProducts = append(createdProducts, map[string]interface{}{
			"id":         id,
			"sku":        sku,
			"name":       name,
			"base_price": basePrice,
			"is_active":  isActive,
			"created_at": createdAt,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Bulk created products", zap.Int("count", len(createdProducts)))
	return createdProducts, nil
}

// UpdateProductStock updates the stock quantity for a product
func (r *PostgresRepository) UpdateProductStock(ctx context.Context, productID string, stockQuantity float64) error {
	query := `
		UPDATE store_products
		SET stock_quantity = $1,
		    is_in_stock = CASE WHEN $1 > 0 THEN true ELSE false END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE product_id = $2
	`

	result, err := r.pool.Exec(ctx, query, stockQuantity, productID)
	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found in any store")
	}

	r.logger.Info("Updated product stock",
		zap.String("product_id", productID),
		zap.Float64("stock", stockQuantity),
		zap.Int64("rows_affected", result.RowsAffected()))

	return nil
}

// UpdateProductStatus updates the active status of a product
func (r *PostgresRepository) UpdateProductStatus(ctx context.Context, productID string, isActive bool) error {
	query := `
		UPDATE products
		SET is_active = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, isActive, productID)
	if err != nil {
		return fmt.Errorf("failed to update product status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	r.logger.Info("Updated product status",
		zap.String("product_id", productID),
		zap.Bool("is_active", isActive))

	return nil
}

// BulkUpdateProductStock updates stock for multiple products
func (r *PostgresRepository) BulkUpdateProductStock(ctx context.Context, updates []struct {
	ProductID     string  `json:"product_id"`
	StockQuantity float64 `json:"stock_quantity"`
}) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE store_products
		SET stock_quantity = $1,
		    is_in_stock = CASE WHEN $1 > 0 THEN true ELSE false END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE product_id = $2
	`

	for _, update := range updates {
		_, err := tx.Exec(ctx, query, update.StockQuantity, update.ProductID)
		if err != nil {
			r.logger.Error("Failed to update product stock in bulk",
				zap.String("product_id", update.ProductID),
				zap.Error(err))
			return fmt.Errorf("failed to update product %s: %w", update.ProductID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Bulk updated product stock", zap.Int("count", len(updates)))
	return nil
}

// GetStoreByID retrieves basic store information
func (r *PostgresRepository) GetStoreByID(ctx context.Context, storeID string) (map[string]interface{}, error) {
	query := `
		SELECT id, name, slug, description, store_type, phone, email,
		       address_line1, city, state, postal_code, country,
		       latitude, longitude, rating, total_ratings,
		       min_order_amount, delivery_fee, estimated_delivery_time,
		       is_active, is_open, created_at, updated_at
		FROM stores
		WHERE id = $1
	`

	var id, name, slug, storeType, addressLine1, city, country string
	var description, phone, email, state, postalCode *string
	var latitude, longitude, rating, minOrderAmount, deliveryFee float64
	var totalRatings, estimatedDeliveryTime *int
	var isActive, isOpen bool
	var createdAt, updatedAt interface{}

	err := r.pool.QueryRow(ctx, query, storeID).Scan(
		&id, &name, &slug, &description, &storeType, &phone, &email,
		&addressLine1, &city, &state, &postalCode, &country,
		&latitude, &longitude, &rating, &totalRatings,
		&minOrderAmount, &deliveryFee, &estimatedDeliveryTime,
		&isActive, &isOpen, &createdAt, &updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("store not found: %w", err)
	}

	return map[string]interface{}{
		"id":                      id,
		"name":                    name,
		"slug":                    slug,
		"description":             description,
		"store_type":              storeType,
		"phone":                   phone,
		"email":                   email,
		"address_line1":           addressLine1,
		"city":                    city,
		"state":                   state,
		"postal_code":             postalCode,
		"country":                 country,
		"latitude":                latitude,
		"longitude":               longitude,
		"rating":                  rating,
		"total_ratings":           totalRatings,
		"min_order_amount":        minOrderAmount,
		"delivery_fee":            deliveryFee,
		"estimated_delivery_time": estimatedDeliveryTime,
		"is_active":               isActive,
		"is_open":                 isOpen,
		"created_at":              createdAt,
		"updated_at":              updatedAt,
	}, nil
}

// UpdateStoreStatus updates store active and open status
func (r *PostgresRepository) UpdateStoreStatus(ctx context.Context, storeID string, isActive, isOpen *bool) error {
	query := `UPDATE stores SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}
	argCount := 1

	if isActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argCount)
		args = append(args, *isActive)
		argCount++
	}

	if isOpen != nil {
		query += fmt.Sprintf(", is_open = $%d", argCount)
		args = append(args, *isOpen)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, storeID)

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update store status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("store not found")
	}

	r.logger.Info("Updated store status",
		zap.String("store_id", storeID),
		zap.Any("is_active", isActive),
		zap.Any("is_open", isOpen))

	return nil
}

// GetStoreStatus retrieves store status information
func (r *PostgresRepository) GetStoreStatus(ctx context.Context, storeID string) (map[string]interface{}, error) {
	query := `
		SELECT id, name, is_active, is_open, is_verified,
		       opened_at, closed_at, updated_at
		FROM stores
		WHERE id = $1
	`

	var id, name string
	var isActive, isOpen, isVerified bool
	var openedAt, closedAt, updatedAt interface{}

	err := r.pool.QueryRow(ctx, query, storeID).Scan(
		&id, &name, &isActive, &isOpen, &isVerified,
		&openedAt, &closedAt, &updatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("store not found: %w", err)
	}

	return map[string]interface{}{
		"id":          id,
		"name":        name,
		"is_active":   isActive,
		"is_open":     isOpen,
		"is_verified": isVerified,
		"opened_at":   openedAt,
		"closed_at":   closedAt,
		"updated_at":  updatedAt,
	}, nil
}

// UpdateStoreDetailsInput represents data for updating store details
type UpdateStoreDetailsInput struct {
	Name                  *string  `json:"name"`
	Description           *string  `json:"description"`
	Phone                 *string  `json:"phone"`
	Email                 *string  `json:"email"`
	AddressLine1          *string  `json:"address_line1"`
	AddressLine2          *string  `json:"address_line2"`
	City                  *string  `json:"city"`
	State                 *string  `json:"state"`
	PostalCode            *string  `json:"postal_code"`
	Country               *string  `json:"country"`
	MinOrderAmount        *float64 `json:"min_order_amount"`
	DeliveryFee           *float64 `json:"delivery_fee"`
	EstimatedDeliveryTime *int     `json:"estimated_delivery_time"`
}

// UpdateStoreDetails updates store information
func (r *PostgresRepository) UpdateStoreDetails(ctx context.Context, storeID string, input UpdateStoreDetailsInput) error {
	query := `UPDATE stores SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}
	argCount := 1

	if input.Name != nil {
		query += fmt.Sprintf(", name = $%d, slug = $%d", argCount, argCount+1)
		args = append(args, *input.Name, generateSlug(*input.Name))
		argCount += 2
	}

	if input.Description != nil {
		query += fmt.Sprintf(", description = $%d", argCount)
		args = append(args, *input.Description)
		argCount++
	}

	if input.Phone != nil {
		query += fmt.Sprintf(", phone = $%d", argCount)
		args = append(args, *input.Phone)
		argCount++
	}

	if input.Email != nil {
		query += fmt.Sprintf(", email = $%d", argCount)
		args = append(args, *input.Email)
		argCount++
	}

	if input.AddressLine1 != nil {
		query += fmt.Sprintf(", address_line1 = $%d", argCount)
		args = append(args, *input.AddressLine1)
		argCount++
	}

	if input.AddressLine2 != nil {
		query += fmt.Sprintf(", address_line2 = $%d", argCount)
		args = append(args, *input.AddressLine2)
		argCount++
	}

	if input.City != nil {
		query += fmt.Sprintf(", city = $%d", argCount)
		args = append(args, *input.City)
		argCount++
	}

	if input.State != nil {
		query += fmt.Sprintf(", state = $%d", argCount)
		args = append(args, *input.State)
		argCount++
	}

	if input.PostalCode != nil {
		query += fmt.Sprintf(", postal_code = $%d", argCount)
		args = append(args, *input.PostalCode)
		argCount++
	}

	if input.Country != nil {
		query += fmt.Sprintf(", country = $%d", argCount)
		args = append(args, *input.Country)
		argCount++
	}

	if input.MinOrderAmount != nil {
		query += fmt.Sprintf(", min_order_amount = $%d", argCount)
		args = append(args, *input.MinOrderAmount)
		argCount++
	}

	if input.DeliveryFee != nil {
		query += fmt.Sprintf(", delivery_fee = $%d", argCount)
		args = append(args, *input.DeliveryFee)
		argCount++
	}

	if input.EstimatedDeliveryTime != nil {
		query += fmt.Sprintf(", estimated_delivery_time = $%d", argCount)
		args = append(args, *input.EstimatedDeliveryTime)
		argCount++
	}

	if len(args) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, storeID)

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update store details: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("store not found")
	}

	r.logger.Info("Updated store details",
		zap.String("store_id", storeID),
		zap.Int("fields_updated", len(args)-1))

	return nil
}

// generateSlug creates a URL-friendly slug from a string
func generateSlug(s string) string {
	// Simple slug generation - replace spaces with hyphens and lowercase
	slug := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			slug += string(r)
		} else if r >= 'A' && r <= 'Z' {
			slug += string(r + 32)
		} else if r == ' ' || r == '-' {
			if len(slug) > 0 && slug[len(slug)-1] != '-' {
				slug += "-"
			}
		}
	}
	return slug
}

// UpsertResult contains statistics about an upsert operation
type UpsertResult struct {
	Created                int
	Updated                int
	VariationsProcessed    int
	StoreProductsProcessed int
	TaxesProcessed         int
}

// StoreDetailsInput represents store details for upsert
type StoreDetailsInput struct {
	StoreID  string
	Name     string
	Address  AddressInput
	Location LocationInput
}

type AddressInput struct {
	Line1      string
	City       string
	State      string
	PostalCode string
}

type LocationInput struct {
	Lat float64
	Lng float64
}

// UpsertStore creates or updates a store using external_id as the unique key
func (r *PostgresRepository) UpsertStore(ctx context.Context, storeDetails StoreDetailsInput) error {
	store := storeDetails
	slug := generateSlug(store.Name)

	query := `
		INSERT INTO stores (
			external_id, name, slug, store_type, address_line1, city, state, postal_code, 
			country, latitude, longitude, location, is_active, is_open
		) VALUES (
			$1, $2, $3, 'supermarket', $4, $5, $6, $7, 'India', 
			$8, $9, ST_SetSRID(ST_MakePoint($10, $11), 4326)::geography, 
			true, true
		)
		ON CONFLICT (external_id) DO UPDATE SET
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			address_line1 = EXCLUDED.address_line1,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			postal_code = EXCLUDED.postal_code,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			location = EXCLUDED.location,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.pool.Exec(ctx, query,
		store.StoreID, // This is the external_id
		store.Name,
		slug,
		store.Address.Line1,
		store.Address.City,
		store.Address.State,
		store.Address.PostalCode,
		store.Location.Lat,
		store.Location.Lng,
		store.Location.Lng, // $10 for ST_MakePoint (longitude first)
		store.Location.Lat, // $11 for ST_MakePoint (latitude second)
	)

	if err != nil {
		r.logger.Error("Failed to upsert store", zap.Error(err))
		return fmt.Errorf("failed to upsert store: %w", err)
	}

	r.logger.Info("Upserted store", zap.String("external_id", store.StoreID))
	return nil
}

// CategoryInput represents category data for upsert
type CategoryInput struct {
	ID           string
	ParentID     *string
	Name         string
	Slug         string
	Description  string
	DisplayOrder int
	IsActive     bool
}

// UpsertCategories creates or updates categories using external_id
// Processes parent categories first to ensure proper hierarchy
func (r *PostgresRepository) UpsertCategories(ctx context.Context, categories []CategoryInput) error {
	cats := categories

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Separate root categories (no parent) from child categories
	var rootCats, childCats []CategoryInput
	for _, cat := range cats {
		if cat.ParentID == nil || *cat.ParentID == "" {
			rootCats = append(rootCats, cat)
		} else {
			childCats = append(childCats, cat)
		}
	}

	// Process root categories first
	for _, cat := range rootCats {
		query := `
			INSERT INTO categories (
				external_id, parent_id, name, slug, description, display_order, is_active
			) VALUES ($1, NULL, $2, $3, $4, $5, $6)
			ON CONFLICT (external_id) DO UPDATE SET
				parent_id = NULL,
				name = EXCLUDED.name,
				slug = EXCLUDED.slug,
				description = EXCLUDED.description,
				display_order = EXCLUDED.display_order,
				is_active = EXCLUDED.is_active,
				updated_at = CURRENT_TIMESTAMP
		`
		_, err := tx.Exec(ctx, query,
			cat.ID,
			cat.Name,
			cat.Slug,
			cat.Description,
			cat.DisplayOrder,
			cat.IsActive,
		)
		if err != nil {
			r.logger.Error("Failed to upsert root category", zap.String("external_id", cat.ID), zap.Error(err))
			return fmt.Errorf("failed to upsert root category %s: %w", cat.ID, err)
		}
	}

	// Process child categories
	for _, cat := range childCats {
		query := `
			INSERT INTO categories (
				external_id, parent_id, name, slug, description, display_order, is_active
			) VALUES (
				$1, 
				(SELECT id FROM categories WHERE external_id = $2), 
				$3, $4, $5, $6, $7
			)
			ON CONFLICT (external_id) DO UPDATE SET
				parent_id = (SELECT id FROM categories WHERE external_id = $2),
				name = EXCLUDED.name,
				slug = EXCLUDED.slug,
				description = EXCLUDED.description,
				display_order = EXCLUDED.display_order,
				is_active = EXCLUDED.is_active,
				updated_at = CURRENT_TIMESTAMP
		`
		_, err := tx.Exec(ctx, query,
			cat.ID,
			cat.ParentID,
			cat.Name,
			cat.Slug,
			cat.Description,
			cat.DisplayOrder,
			cat.IsActive,
		)
		if err != nil {
			r.logger.Error("Failed to upsert child category", zap.String("external_id", cat.ID), zap.Error(err))
			return fmt.Errorf("failed to upsert child category %s: %w", cat.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Upserted categories", zap.Int("count", len(cats)))
	return nil
}

// TaxInput represents tax data for upsert
type TaxInput struct {
	ID          string
	StoreID     string
	Name        string
	TaxID       string
	Description string
	Rate        float64
	TaxType     string
	IsInclusive bool
	IsActive    bool
}

// UpsertTaxes creates or updates taxes using (store_id, tax_id) as unique key
func (r *PostgresRepository) UpsertTaxes(ctx context.Context, taxes []TaxInput, storeExternalID string) error {
	txs := taxes

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// First, get the store's internal UUID from external_id
	var storeUUID string
	err = tx.QueryRow(ctx, `SELECT id FROM stores WHERE external_id = $1`, storeExternalID).Scan(&storeUUID)
	if err != nil {
		return fmt.Errorf("failed to find store with external_id %s: %w", storeExternalID, err)
	}

	query := `
		INSERT INTO taxes (
			external_id, store_id, name, tax_id, description, rate, tax_type, is_inclusive, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (store_id, tax_id) DO UPDATE SET
			external_id = EXCLUDED.external_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			rate = EXCLUDED.rate,
			tax_type = EXCLUDED.tax_type,
			is_inclusive = EXCLUDED.is_inclusive,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP
	`

	for _, t := range txs {
		_, err := tx.Exec(ctx, query,
			t.ID, // external_id
			storeUUID,
			t.Name,
			t.TaxID,
			t.Description,
			t.Rate,
			t.TaxType,
			t.IsInclusive,
			t.IsActive,
		)
		if err != nil {
			r.logger.Error("Failed to upsert tax", zap.String("tax_id", t.TaxID), zap.Error(err))
			return fmt.Errorf("failed to upsert tax %s: %w", t.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Upserted taxes", zap.Int("count", len(txs)))
	return nil
}

// ProductInput represents product data for upsert with new schema
type ProductInput struct {
	ExternalProductID string // ERP's product ID
	SKU               string
	Name              string
	Slug              string
	Description       string
	CategoryID        string // External category ID
	BasePrice         float64
	Currency          string
	Unit              string
	UnitQuantity      float64
	PrimaryImageURL   string
	Images            []string
	Brand             string // Brand name (will be normalized)
	Manufacturer      string
	Barcode           string
	EAN               string
	IsActive          bool
	IsFeatured        bool
	IsCustomizable    bool
	IsAddon           bool
}

// VariationInput represents variation data for upsert
type VariationInput struct {
	ExternalID        string // External variation ID from ERP
	ExternalProductID string
	Name              string
	DisplayName       string
	Price             float64
	IsDefault         bool
}

// StoreProductInput represents store product data for upsert
type StoreProductInput struct {
	ExternalProductID    string
	ExternalStoreProduct string // ERP's store-product ID
	StoreID              string
	Price                float64
	StockQuantity        float64
	IsInStock            bool
	Taxes                []string // Tax IDs for this store-product
}

// UpsertProducts is deprecated - use UpsertProductsWithMatching instead
// Kept for backward compatibility
func (r *PostgresRepository) UpsertProducts(ctx context.Context, products []ProductInput, variations []VariationInput, storeProducts []StoreProductInput) (*UpsertResult, error) {
	// This function is deprecated and should not be used
	// Use UpsertProductsWithMatching instead which supports the new schema
	return nil, fmt.Errorf("UpsertProducts is deprecated, use UpsertProductsWithMatching instead")
}

// StockUpdateResult contains statistics about stock update operation
type StockUpdateResult struct {
	Updated          int
	NotFound         int
	VariantsUpdated  int
	VariantsNotFound int
}

// StockProductUpdate represents a product stock update
type StockProductUpdate struct {
	ID            string
	StockQuantity float64
	IsAvailable   bool
	Price         float64
	Variants      []StockVariantUpdate
}

// StockVariantUpdate represents a variation stock update
type StockVariantUpdate struct {
	ID            string
	StockQuantity float64
	IsAvailable   bool
	Price         float64
}

// BulkUpdateStock updates stock for multiple products in a store
func (r *PostgresRepository) BulkUpdateStock(ctx context.Context, storeExternalID string, products []StockProductUpdate) (*StockUpdateResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get store UUID from external_id
	var storeUUID string
	err = tx.QueryRow(ctx, `SELECT id FROM stores WHERE external_id = $1`, storeExternalID).Scan(&storeUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to find store with external_id %s: %w", storeExternalID, err)
	}

	result := &StockUpdateResult{}

	for _, prod := range products {
		// Update store_product by external_id
		query := `
			UPDATE store_products
			SET stock_quantity = $1::numeric,
			    is_in_stock = CASE WHEN $1::numeric > 0 THEN true ELSE false END,
			    is_available = $2,
			    updated_at = CURRENT_TIMESTAMP
			WHERE store_id = $3 AND external_id = $4
		`

		// If price is provided and > 0, include it in the update
		if prod.Price > 0 {
			query = `
				UPDATE store_products
				SET stock_quantity = $1::numeric,
				    is_in_stock = CASE WHEN $1::numeric > 0 THEN true ELSE false END,
				    is_available = $2,
				    price = $5::numeric,
				    updated_at = CURRENT_TIMESTAMP
				WHERE store_id = $3 AND external_id = $4
			`
			cmdTag, err := tx.Exec(ctx, query, prod.StockQuantity, prod.IsAvailable, storeUUID, prod.ID, prod.Price)
			if err != nil {
				r.logger.Error("Failed to update stock with price",
					zap.String("external_id", prod.ID),
					zap.Error(err))
				return nil, fmt.Errorf("failed to update stock for product %s: %w", prod.ID, err)
			}

			if cmdTag.RowsAffected() == 0 {
				result.NotFound++
				r.logger.Warn("Product not found in store",
					zap.String("store_id", storeExternalID),
					zap.String("external_id", prod.ID))
			} else {
				result.Updated++
			}
		} else {
			cmdTag, err := tx.Exec(ctx, query, prod.StockQuantity, prod.IsAvailable, storeUUID, prod.ID)
			if err != nil {
				r.logger.Error("Failed to update stock",
					zap.String("external_id", prod.ID),
					zap.Error(err))
				return nil, fmt.Errorf("failed to update stock for product %s: %w", prod.ID, err)
			}

			if cmdTag.RowsAffected() == 0 {
				result.NotFound++
				r.logger.Warn("Product not found in store",
					zap.String("store_id", storeExternalID),
					zap.String("external_id", prod.ID))
			} else {
				result.Updated++
			}
		}

		// Update variations if provided
		if len(prod.Variants) > 0 {
			for _, variant := range prod.Variants {
				varQuery := `
					UPDATE product_variations
					SET stock_quantity = $1::numeric,
					    is_in_stock = CASE WHEN $1::numeric > 0 THEN true ELSE false END,
					    is_active = $2,
					    updated_at = CURRENT_TIMESTAMP
					WHERE external_id = $3
				`

				// If price is provided and > 0, include it in the update
				if variant.Price > 0 {
					varQuery = `
						UPDATE product_variations
						SET stock_quantity = $1::numeric,
						    is_in_stock = CASE WHEN $1::numeric > 0 THEN true ELSE false END,
						    is_active = $2,
						    price = $4::numeric,
						    updated_at = CURRENT_TIMESTAMP
						WHERE external_id = $3
					`
					cmdTag, err := tx.Exec(ctx, varQuery, variant.StockQuantity, variant.IsAvailable, variant.ID, variant.Price)
					if err != nil {
						r.logger.Error("Failed to update variation stock with price",
							zap.String("external_id", variant.ID),
							zap.Error(err))
						return nil, fmt.Errorf("failed to update variation stock for %s: %w", variant.ID, err)
					}

					if cmdTag.RowsAffected() == 0 {
						result.VariantsNotFound++
						r.logger.Warn("Variation not found",
							zap.String("external_id", variant.ID))
					} else {
						result.VariantsUpdated++
					}
				} else {
					cmdTag, err := tx.Exec(ctx, varQuery, variant.StockQuantity, variant.IsAvailable, variant.ID)
					if err != nil {
						r.logger.Error("Failed to update variation stock",
							zap.String("external_id", variant.ID),
							zap.Error(err))
						return nil, fmt.Errorf("failed to update variation stock for %s: %w", variant.ID, err)
					}

					if cmdTag.RowsAffected() == 0 {
						result.VariantsNotFound++
						r.logger.Warn("Variation not found",
							zap.String("external_id", variant.ID))
					} else {
						result.VariantsUpdated++
					}
				}
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Bulk updated stock",
		zap.String("store_id", storeExternalID),
		zap.Int("updated", result.Updated),
		zap.Int("not_found", result.NotFound),
		zap.Int("variants_updated", result.VariantsUpdated),
		zap.Int("variants_not_found", result.VariantsNotFound))

	return result, nil
}
