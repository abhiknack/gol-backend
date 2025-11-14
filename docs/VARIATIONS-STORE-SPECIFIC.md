# Variations Now Store-Specific

## Overview

Product variations now reference `store_products` instead of `products`, making them store-specific. This allows different stores to have different stock levels and pricing for the same variation.

## Changes Made

### 1. Database Schema (`grocery_superapp_schema.sql`)
- Changed `product_variations.product_id` to reference `store_products` instead of `products`
- Added `store_product_id` column
- Updated unique constraint to `(store_product_id, name)`
- Kept `product_id` for backward compatibility (nullable)

### 2. Migration (`migrations/update_variations_reference_store_products.sql`)
- Adds `store_product_id` column
- Migrates existing data
- Updates constraints and indexes
- Maintains backward compatibility

### 3. Code (`internal/repository/product_matching.go`)
- Added `storeProductIDMap` to track store_product UUIDs
- Updated variation insert to use `store_product_id`
- Updated conflict resolution to use `(store_product_id, name)`

## Benefits

### Before (Product-Level Variations)
```
Product: Rice
├─ Variation: 5kg (price: 455)
└─ Variation: 1kg (price: 110)

Problem: All stores share same stock for variations
```

### After (Store-Specific Variations)
```
Store A:
  Product: Rice
  ├─ Variation: 5kg (price: 455, stock: 80)
  └─ Variation: 1kg (price: 110, stock: 40)

Store B:
  Product: Rice
  ├─ Variation: 5kg (price: 460, stock: 50)
  └─ Variation: 1kg (price: 115, stock: 30)

Benefit: Each store has independent stock and pricing
```

## Database Structure

### Old Structure
```sql
product_variations
├─ id (UUID)
├─ product_id (UUID) → products.id
├─ name (VARCHAR)
└─ stock_quantity (DECIMAL)

UNIQUE (product_id, name)
```

### New Structure
```sql
product_variations
├─ id (UUID)
├─ store_product_id (UUID) → store_products.id
├─ product_id (UUID) → products.id (deprecated, nullable)
├─ name (VARCHAR)
└─ stock_quantity (DECIMAL)

UNIQUE (store_product_id, name)
```

## API Flow

### 1. Push Products with Variations
```json
{
  "store_details": {"store_id": "STORE-A"},
  "products": [{"id": "PROD-1", "price": 455}],
  "variations": [
    {"id": "VAR-5KG", "product_id": "PROD-1", "name": "5kg", "price": 455}
  ]
}
```

**What happens:**
1. Product created/matched → `product_uuid`
2. Store product created → `store_product_uuid` (links store + product)
3. Variation created → references `store_product_uuid`

### 2. Update Stock
```json
{
  "store_id": "STORE-A",
  "products": [{
    "id": "PROD-1",
    "variants": [
      {"id": "VAR-5KG", "stock_quantity": 80}
    ]
  }]
}
```

**What happens:**
1. Finds variation by `external_id = "VAR-5KG"`
2. Updates stock for that specific store's variation

## Migration Steps

### 1. Apply Migration
```bash
Get-Content migrations/update_variations_reference_store_products.sql | docker exec -i gol-bazaar-postgres-dev psql -U postgres -d middleware_db
```

### 2. Verify Migration
```sql
-- Check new column exists
\d product_variations

-- Check data migrated
SELECT 
    pv.external_id,
    pv.name,
    pv.store_product_id,
    sp.external_id as store_product_external_id
FROM product_variations pv
JOIN store_products sp ON sp.id = pv.store_product_id
LIMIT 5;
```

### 3. Restart Application
```bash
docker restart gol-bazaar-app-dev
```

### 4. Test
Push products with variations and verify they're created correctly.

## Verification Queries

### Check Variation Structure
```sql
SELECT 
    pv.external_id as variation_id,
    pv.name as variation_name,
    pv.price as variation_price,
    pv.stock_quantity,
    sp.external_id as store_product_id,
    p.name as product_name,
    s.external_id as store_id
FROM product_variations pv
JOIN store_products sp ON sp.id = pv.store_product_id
JOIN products p ON p.id = sp.product_id
JOIN stores s ON s.id = sp.store_id
WHERE pv.external_id = 'ID-VAR-5KG';
```

### Check Store-Specific Variations
```sql
-- Same product, different stores
SELECT 
    s.external_id as store,
    p.name as product,
    pv.name as variation,
    pv.price,
    pv.stock_quantity
FROM product_variations pv
JOIN store_products sp ON sp.id = pv.store_product_id
JOIN products p ON p.id = sp.product_id
JOIN stores s ON s.id = sp.store_id
WHERE p.sku = 'RICE-FORTUNE-5KG'
ORDER BY s.external_id, pv.name;
```

## Use Cases

### 1. Multi-Store Inventory
Different stores have different stock for same variation:
```
Store A: Rice 5kg - 80 units
Store B: Rice 5kg - 50 units
Store C: Rice 5kg - 0 units (out of stock)
```

### 2. Regional Pricing
Different stores have different prices:
```
Store A (Mumbai): Rice 5kg - ₹455
Store B (Delhi): Rice 5kg - ₹460
Store C (Bangalore): Rice 5kg - ₹450
```

### 3. Store-Specific Availability
Enable/disable variations per store:
```
Store A: Rice 5kg - Available
Store B: Rice 5kg - Not available (discontinued)
Store C: Rice 5kg - Available
```

## Backward Compatibility

- `product_id` column kept (nullable) for backward compatibility
- Old queries using `product_id` will still work
- New code uses `store_product_id`
- Migration handles existing data automatically

## Summary

✅ **Variations are now store-specific**  
✅ **Each store has independent stock/pricing**  
✅ **Migration handles existing data**  
✅ **Backward compatible**  
✅ **API unchanged (works transparently)**  

This change enables true multi-store inventory management with store-specific variations!
