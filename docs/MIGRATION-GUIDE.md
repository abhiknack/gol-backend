# Migration Guide: Simplify External IDs

## Quick Start

```bash
# Apply the migration
psql -d middleware_db -f migrations/simplify_external_ids.sql
```

## What This Does

1. ✅ Adds `external_id` to `store_products` table
2. ✅ Adds `external_id` to `taxes` table
3. ✅ Migrates data from `store_product_mappings` (if exists)
4. ✅ Drops `store_product_mappings` table
5. ✅ Creates necessary indexes

## Verification Steps

### 1. Check Schema Changes
```sql
-- Verify external_id columns
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'store_products' AND column_name = 'external_id';

SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'taxes' AND column_name = 'external_id';
```

Expected output:
```
 column_name  |     data_type     
--------------+-------------------
 external_id  | character varying
```

### 2. Check Indexes
```sql
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename IN ('store_products', 'taxes') 
  AND indexname LIKE '%external_id%';
```

Expected output:
```
idx_store_products_external_id
idx_taxes_external_id
idx_store_products_external_id_lookup
idx_taxes_external_id_lookup
```

### 3. Verify Table Removal
```sql
SELECT * FROM information_schema.tables 
WHERE table_name = 'store_product_mappings';
```

Expected output: (0 rows)

### 4. Test Product Matching
```sql
-- This should work without errors
SELECT * FROM find_matching_product(
    'Coca Cola 1L',
    '8901234567890',
    'CK1L',
    '8901234567890',
    (SELECT id FROM stores LIMIT 1),
    'EXT-PROD-001'
);
```

## Rollback (If Needed)

If you need to rollback:

```sql
-- Recreate store_product_mappings table
CREATE TABLE store_product_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    external_product_id VARCHAR(255) NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    external_sku VARCHAR(255),
    external_barcode VARCHAR(255),
    external_name TEXT,
    last_synced_at TIMESTAMP WITH TIME ZONE,
    sync_source VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, external_product_id)
);

-- Migrate data back
INSERT INTO store_product_mappings (
    store_id, external_product_id, product_id, is_active
)
SELECT 
    store_id, 
    external_id, 
    product_id, 
    is_available
FROM store_products
WHERE external_id IS NOT NULL;

-- Remove external_id columns
ALTER TABLE store_products DROP COLUMN external_id;
ALTER TABLE taxes DROP COLUMN external_id;
```

## Post-Migration Testing

### Test 1: Create Product with External ID
```bash
curl -X POST http://localhost:8080/api/v1/products/push \
  -H "Content-Type: application/json" \
  -d @docs/API-PRODUCTS-PUSH-CORRECTED-EXAMPLE.json
```

Expected response:
```json
{
  "status": "success",
  "data": {
    "products_created": 1,
    "products_updated": 0,
    "variations_processed": 2,
    "store_products_processed": 1,
    "taxes_processed": 1
  }
}
```

### Test 2: Verify External IDs Stored
```sql
-- Check store_products
SELECT external_id, price, stock_quantity
FROM store_products
WHERE external_id IS NOT NULL
LIMIT 5;

-- Check taxes
SELECT external_id, tax_id, name, rate
FROM taxes
WHERE external_id IS NOT NULL
LIMIT 5;
```

### Test 3: Update Existing Product
Send the same payload again - should update instead of create:

```bash
curl -X POST http://localhost:8080/api/v1/products/push \
  -H "Content-Type: application/json" \
  -d @docs/API-PRODUCTS-PUSH-CORRECTED-EXAMPLE.json
```

Expected response:
```json
{
  "status": "success",
  "data": {
    "products_created": 0,
    "products_updated": 1,  // ← Should be 1
    "variations_processed": 2,
    "store_products_processed": 1,
    "taxes_processed": 1
  }
}
```

## Common Issues

### Issue: Migration fails with "column already exists"
**Cause:** Migration was partially applied before
**Solution:** 
```sql
-- Check if columns exist
\d store_products
\d taxes

-- If they exist, skip to dropping store_product_mappings
DROP TABLE IF EXISTS store_product_mappings CASCADE;
```

### Issue: "relation store_product_mappings does not exist"
**Cause:** Table was already removed or never existed
**Solution:** This is fine - the migration handles this gracefully

### Issue: Unique constraint violation on external_id
**Cause:** Duplicate external_ids in your data
**Solution:**
```sql
-- Find duplicates
SELECT store_id, external_id, COUNT(*)
FROM store_products
WHERE external_id IS NOT NULL
GROUP BY store_id, external_id
HAVING COUNT(*) > 1;

-- Fix duplicates (keep the most recent)
-- Manual intervention required based on your data
```

## Performance Impact

- **Positive:** Removed one JOIN from product matching queries
- **Positive:** Direct index lookups on external_id
- **Neutral:** Same number of indexes overall
- **Storage:** Minimal increase (~255 bytes per row with external_id)

## Next Steps

1. ✅ Apply migration
2. ✅ Run verification queries
3. ✅ Test API endpoints
4. ✅ Monitor logs for any issues
5. ✅ Update your ERP integration to use new structure

## Support

If you encounter issues:
1. Check server logs: `tail -f logs/app.log`
2. Check database logs: `tail -f /var/log/postgresql/postgresql.log`
3. Review [Common Mistakes](./COMMON-MISTAKES.md)
4. See [Schema Simplification](./SCHEMA-SIMPLIFICATION.md) for details
