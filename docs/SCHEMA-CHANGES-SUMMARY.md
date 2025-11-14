# Database Schema Changes Summary

## Overview

This document summarizes all the database schema changes made to support multi-store ERP integration with product matching and flexible tax configuration.

## Changes Made

### 1. Product Matching Engine

#### Added to `products` table:
```sql
-- Normalized fields for product matching
normalized_name TEXT
extracted_volume_ml DECIMAL(10, 3)
extracted_weight_g DECIMAL(10, 3)
```

#### Removed from `products` table:
```sql
external_id VARCHAR(100) UNIQUE  -- Moved to store_product_mappings
```

#### New table: `store_product_mappings`
Maps external ERP product IDs to internal product UUIDs:
```sql
CREATE TABLE store_product_mappings (
    id UUID PRIMARY KEY,
    store_id UUID NOT NULL,
    external_product_id VARCHAR(255) NOT NULL,  -- ERP's product ID
    product_id UUID NOT NULL,                    -- Internal product UUID
    external_sku VARCHAR(255),
    external_barcode VARCHAR(255),
    external_name TEXT,
    sync_source VARCHAR(100),                    -- "Zoho", "Tally", etc.
    last_synced_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(store_id, external_product_id)
);
```

#### New functions:
- `normalize_product_name(TEXT)` - Normalizes product names
- `extract_volume_ml(TEXT)` - Extracts volume in ml
- `extract_weight_g(TEXT)` - Extracts weight in grams
- `find_matching_product(...)` - 3-layer matching strategy

#### New indexes:
```sql
-- Exact matching
CREATE INDEX idx_products_barcode ON products(barcode);
CREATE INDEX idx_products_ean ON products(ean);

-- Fuzzy matching (trigram)
CREATE INDEX idx_products_name_trgm ON products USING gin(name gin_trgm_ops);
CREATE INDEX idx_products_normalized_name_trgm ON products USING gin(normalized_name gin_trgm_ops);

-- Mapping lookups
CREATE INDEX idx_store_product_mappings_store_id ON store_product_mappings(store_id);
CREATE INDEX idx_store_product_mappings_product_id ON store_product_mappings(product_id);
CREATE INDEX idx_store_product_mappings_external_id ON store_product_mappings(external_product_id);
```

### 2. Unique Constraints for Upserts

#### `product_images`:
```sql
ALTER TABLE product_images 
ADD CONSTRAINT product_images_product_id_image_url_key 
UNIQUE (product_id, image_url);
```

#### `product_variations`:
```sql
ALTER TABLE product_variations 
ADD CONSTRAINT product_variations_product_id_name_key 
UNIQUE (product_id, name);
```

### 3. Tax Configuration

#### Removed table:
```sql
DROP TABLE product_taxes;  -- No longer needed
```

#### New table: `store_product_taxes`
Store-specific tax configuration:
```sql
CREATE TABLE store_product_taxes (
    id UUID PRIMARY KEY,
    store_id UUID NOT NULL,
    store_product_id UUID NOT NULL,
    tax_id UUID NOT NULL,
    override_rate DECIMAL(5, 2),  -- Optional rate override
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(store_id, store_product_id, tax_id)
);
```

#### New indexes:
```sql
CREATE INDEX idx_store_product_taxes_store_id ON store_product_taxes(store_id);
CREATE INDEX idx_store_product_taxes_store_product_id ON store_product_taxes(store_product_id);
CREATE INDEX idx_store_product_taxes_tax_id ON store_product_taxes(tax_id);
CREATE INDEX idx_store_product_taxes_is_active ON store_product_taxes(is_active);
```

## Migration Files

All changes are tracked in migration files:

1. `migrations/add_product_images_unique_constraint.sql`
   - Adds unique constraints to product_images and product_variations

2. `migrations/add_store_product_mappings.sql`
   - Creates store_product_mappings table for ERP integration

3. `migrations/add_product_matching_engine.sql`
   - Removes external_id from products
   - Adds normalized fields
   - Creates matching functions
   - Adds trigram indexes

4. `migrations/add_store_product_taxes.sql`
   - Creates store_product_taxes table

5. `migrations/remove_product_taxes.sql`
   - Removes product_taxes table

## Key Benefits

### 1. Multi-ERP Support
- Each store can use different ERP systems (Zoho, Tally, SAP, etc.)
- External product IDs are mapped to internal UUIDs
- Prevents duplicate products across stores

### 2. Intelligent Product Matching
- **Layer 1:** Exact match (barcode, SKU, EAN) - 100% confidence
- **Layer 2:** Normalized match (name + size) - 95% confidence
- **Layer 3:** Fuzzy match (similarity) - 45-90% confidence

### 3. Flexible Tax Configuration
- Different tax rates per store (state-specific GST)
- Support for GST registered vs composition scheme
- Tax rate overrides per store+product
- Multiple taxes per product (GST + Service Charge)

### 4. Data Integrity
- Unique constraints prevent duplicates
- Foreign keys with cascade deletes
- Automatic timestamp updates
- Normalized fields auto-populated via triggers

## Breaking Changes

### ⚠️ Removed: `products.external_id`
**Impact:** Code using `products.external_id` must be updated

**Migration:**
```sql
-- Old way
SELECT * FROM products WHERE external_id = 'EXT-123';

-- New way
SELECT p.* 
FROM products p
JOIN store_product_mappings spm ON p.id = spm.product_id
WHERE spm.external_product_id = 'EXT-123'
  AND spm.store_id = 'store-uuid';
```

### ⚠️ Removed: `product_taxes` table
**Impact:** Tax configuration must use `store_product_taxes`

**Migration:**
```sql
-- Old way
INSERT INTO product_taxes (product_id, tax_id)
VALUES ('product-uuid', 'tax-uuid');

-- New way (for each store)
INSERT INTO store_product_taxes (store_id, store_product_id, tax_id)
SELECT 
    sp.store_id,
    sp.id,
    'tax-uuid'
FROM store_products sp
WHERE sp.product_id = 'product-uuid';
```

## Documentation

Comprehensive guides available:

- `docs/PRODUCT-MATCHING-ENGINE.md` - Product matching strategy
- `docs/TAX-CONFIGURATION.md` - Tax configuration guide
- `docs/DOCKER-AIR-SETUP.md` - Development environment setup

## Database State

### Current Tables (Tax-related):
- ✅ `taxes` - Tax definitions
- ✅ `store_product_taxes` - Store-specific tax assignments
- ❌ `product_taxes` - REMOVED

### Current Tables (Product-related):
- ✅ `products` - Master product catalog
- ✅ `store_products` - Store inventory
- ✅ `store_product_mappings` - ERP integration
- ✅ `product_images` - Product images
- ✅ `product_variations` - Product variations

### Extensions Enabled:
- ✅ `uuid-ossp` - UUID generation
- ✅ `postgis` - Location data
- ✅ `pg_trgm` - Fuzzy text matching

## Testing

### Verify Product Matching:
```sql
SELECT * FROM find_matching_product(
    'Coca Cola 1L',
    '8901234567',  -- barcode
    NULL,          -- sku
    NULL,          -- ean
    'store-uuid',
    'EXT-123'
);
```

### Verify Tax Configuration:
```sql
SELECT 
    s.name as store_name,
    p.name as product_name,
    t.name as tax_name,
    COALESCE(spt.override_rate, t.rate) as effective_rate
FROM store_product_taxes spt
JOIN stores s ON spt.store_id = s.id
JOIN store_products sp ON spt.store_product_id = sp.id
JOIN products p ON sp.product_id = p.id
JOIN taxes t ON spt.tax_id = t.id
WHERE spt.is_active = true;
```

## Performance Impact

### Positive:
- ✅ Trigram indexes speed up fuzzy matching
- ✅ Normalized fields avoid repeated computation
- ✅ Proper indexes on all foreign keys

### Considerations:
- Fuzzy matching queries may take 10-20ms (acceptable)
- Normalized fields add ~24 bytes per product (negligible)
- Store-specific taxes require joins (optimized with indexes)

## Next Steps

1. Update application code to use new schema
2. Test product matching with real ERP data
3. Configure taxes for each store
4. Monitor query performance
5. Add more matching rules if needed

---

**Schema Version:** 2.0  
**Last Updated:** 2025-11-14  
**Status:** ✅ Applied to database
