# Schema Simplification - External IDs

## Overview

Simplified the external ID management by removing the `store_product_mappings` table and adding `external_id` directly to the relevant tables.

## Changes Made

### 1. Removed Table
- **`store_product_mappings`** - No longer needed since we don't support multiple ERPs per store

### 2. Added Columns

#### `store_products.external_id`
```sql
ALTER TABLE store_products ADD COLUMN external_id VARCHAR(255);
CREATE UNIQUE INDEX idx_store_products_external_id 
ON store_products(store_id, external_id) 
WHERE external_id IS NOT NULL;
```

**Purpose:** Store the ERP's product identifier directly
**Example:** `"ZOHO-1001"`, `"TALLY-991"`, `"SAP-PROD-123"`

#### `taxes.external_id`
```sql
ALTER TABLE taxes ADD COLUMN external_id VARCHAR(255);
CREATE UNIQUE INDEX idx_taxes_external_id 
ON taxes(store_id, external_id) 
WHERE external_id IS NOT NULL;
```

**Purpose:** Store the ERP's tax identifier
**Example:** `"TAX-001"`, `"GST-5-EXT"`, `"ZOHO-TAX-18"`

### 3. Updated Functions

#### `find_matching_product()`
**Before:**
```sql
SELECT product_id FROM store_product_mappings
WHERE store_id = p_store_id
  AND external_product_id = p_external_product_id
```

**After:**
```sql
SELECT product_id FROM store_products
WHERE store_id = p_store_id
  AND external_id = p_external_product_id
```

### 4. Updated Repository Code

#### `UpsertTaxes()`
Now includes `external_id` when creating/updating taxes:
```go
INSERT INTO taxes (
    external_id, store_id, name, tax_id, ...
) VALUES ($1, $2, $3, $4, ...)
```

#### `UpsertProductsWithMatching()`
- Removed `store_product_mappings` insert
- `store_products` already had `external_id` support

## Migration

Run the migration to apply these changes:

```bash
psql -d middleware_db -f migrations/simplify_external_ids.sql
```

The migration will:
1. Add `external_id` columns to `store_products` and `taxes`
2. Migrate existing data from `store_product_mappings` (if exists)
3. Drop the `store_product_mappings` table
4. Create necessary indexes

## API Changes

### Taxes Payload

**Before (still works):**
```json
{
  "taxes": [
    {
      "id": "UUID-GST-5",
      "tax_id": "GST_5",
      "name": "GST 5%"
    }
  ]
}
```

**After (recommended):**
```json
{
  "taxes": [
    {
      "id": "ZOHO-TAX-001",  // Your ERP's tax ID (stored as external_id)
      "tax_id": "GST_5",      // Tax code (used for linking)
      "name": "GST 5%"
    }
  ]
}
```

### Store Products Payload

No changes needed - already supported:
```json
{
  "store_products": [
    {
      "product_id": "EXT-PROD-001",
      "price": 455,
      "taxes": ["GST_5"]  // Still use tax_id for linking
    }
  ]
}
```

## Benefits

1. **Simpler Schema** - One less table to manage
2. **Better Performance** - Direct lookups instead of joins
3. **Clearer Intent** - External IDs are where they belong
4. **Single ERP Model** - Reflects the actual use case (one ERP per store)

## Database Queries

### Find product by external_id
```sql
SELECT p.* 
FROM products p
JOIN store_products sp ON sp.product_id = p.id
WHERE sp.store_id = '<store-uuid>'
  AND sp.external_id = 'ZOHO-1001';
```

### Find tax by external_id
```sql
SELECT * FROM taxes
WHERE store_id = '<store-uuid>'
  AND external_id = 'ZOHO-TAX-001';
```

### List all external mappings for a store
```sql
-- Products
SELECT 
    sp.external_id,
    p.name,
    p.sku,
    sp.price
FROM store_products sp
JOIN products p ON p.id = sp.product_id
WHERE sp.store_id = '<store-uuid>'
  AND sp.external_id IS NOT NULL;

-- Taxes
SELECT 
    external_id,
    tax_id,
    name,
    rate
FROM taxes
WHERE store_id = '<store-uuid>'
  AND external_id IS NOT NULL;
```

## Backward Compatibility

The migration automatically handles existing data:
- Data from `store_product_mappings` is migrated to `store_products.external_id`
- Existing code continues to work
- No API changes required

## Files Modified

1. `grocery_superapp_schema.sql` - Removed `store_product_mappings`, added `external_id` columns
2. `migrations/simplify_external_ids.sql` - Migration script
3. `migrations/add_product_matching_engine.sql` - Updated `find_matching_product()` function
4. `internal/repository/postgres.go` - Updated `UpsertTaxes()` to include `external_id`
5. `internal/repository/product_matching.go` - Removed `store_product_mappings` insert

## Testing

After migration, verify:

```sql
-- Check external_id columns exist
\d store_products
\d taxes

-- Check indexes
\di idx_store_products_external_id
\di idx_taxes_external_id

-- Verify store_product_mappings is gone
\dt store_product_mappings  -- Should show "Did not find any relation"
```

## See Also

- [API Documentation](./API-PRODUCTS-PUSH.md)
- [Common Mistakes](./COMMON-MISTAKES.md)
- [Corrected Example](./API-PRODUCTS-PUSH-CORRECTED-EXAMPLE.json)
