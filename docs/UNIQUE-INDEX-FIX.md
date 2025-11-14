# Unique Index Fix for ON CONFLICT

## Issue

The unique index was created with a `WHERE` clause (partial index):
```sql
CREATE UNIQUE INDEX idx_variations_store_product_name 
ON product_variations(store_product_id, name) 
WHERE store_product_id IS NOT NULL;  -- ❌ Partial index
```

PostgreSQL's `ON CONFLICT` clause doesn't work with partial indexes, causing this error:
```
ERROR: there is no unique or exclusion constraint matching the ON CONFLICT specification
```

## Solution

Recreated the index without the `WHERE` clause:
```sql
DROP INDEX IF EXISTS idx_variations_store_product_name;

CREATE UNIQUE INDEX idx_variations_store_product_name 
ON product_variations(store_product_id, name);  -- ✅ Full index
```

## Why This Works

### Partial Index (Doesn't Work with ON CONFLICT)
```sql
-- Only indexes rows where store_product_id IS NOT NULL
CREATE UNIQUE INDEX ... WHERE store_product_id IS NOT NULL;

-- ON CONFLICT can't use this index
INSERT ... ON CONFLICT (store_product_id, name) DO UPDATE ...
-- ERROR: no matching constraint
```

### Full Index (Works with ON CONFLICT)
```sql
-- Indexes all rows
CREATE UNIQUE INDEX ... ON (store_product_id, name);

-- ON CONFLICT can use this index
INSERT ... ON CONFLICT (store_product_id, name) DO UPDATE ...
-- ✅ Works!
```

## Applied Fix

```bash
docker exec -i gol-bazaar-postgres-dev psql -U postgres -d middleware_db -c "
  DROP INDEX IF EXISTS idx_variations_store_product_name;
  CREATE UNIQUE INDEX idx_variations_store_product_name 
  ON product_variations(store_product_id, name);
"
```

## Verification

```sql
-- Check index exists
\d product_variations

-- Should show:
-- "idx_variations_store_product_name" UNIQUE, btree (store_product_id, name)
-- (without WHERE clause)
```

## Test

Now your payload should work:

```json
{
  "variations": [
    {"id": "ID-VAR-5KG", "product_id": "ID-PROD-1", "name": "5kg"}
  ]
}
```

The `ON CONFLICT (store_product_id, name)` will now work correctly! ✅

## Summary

✅ **Issue**: Partial index doesn't work with ON CONFLICT  
✅ **Fix**: Recreated as full index  
✅ **Result**: Variations can now be upserted correctly  
✅ **Migration Updated**: Future deployments will use correct index
