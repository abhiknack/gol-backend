# Verify External ID Mapping

## Issue Fixed

The `products[].id` was not being saved as `store_products.external_id`. This has been fixed.

## What Changed

**Before:**
```go
// Used sp.ExternalStoreProduct (which was empty)
`, sp.ExternalStoreProduct, storeUUID, productUUID, ...
```

**After:**
```go
// Now uses sp.ExternalProductID (the product's external ID)
`, sp.ExternalProductID, storeUUID, productUUID, ...
```

## Verification Query

After pushing products, verify the external_id is saved:

```sql
SELECT 
    sp.external_id,
    p.name,
    p.sku,
    sp.price,
    sp.stock_quantity
FROM store_products sp
JOIN products p ON p.id = sp.product_id
WHERE sp.store_id = (SELECT id FROM stores WHERE external_id = 'EXT123456');
```

Expected result:
```
external_id  | name                        | sku              | price | stock_quantity
-------------|-----------------------------|--------------------|-------|---------------
ID-PROD-1    | Fortune Basmati Rice 5kg    | RICE-FORTUNE-5KG   | 455   | 0
```

## Test Payload

```json
{
  "store_details": {
    "store_id": "EXT123456",
    "name": "Main Supermarket"
  },
  "products": [
    {
      "id": "ID-PROD-1",
      "sku": "RICE-FORTUNE-5KG",
      "name": "Fortune Basmati Rice 5kg",
      "price": 455,
      "taxes": ["UUID-GST-5"]
    }
  ]
}
```

## Stock Update Test

After the fix, you can update stock using the external_id:

```json
{
  "store_id": "EXT123456",
  "products": [
    {
      "id": "ID-PROD-1",
      "stock_quantity": 120,
      "is_available": true
    }
  ]
}
```

This should now work because `store_products.external_id = "ID-PROD-1"`.

## Database Check

```sql
-- Check if external_id is saved
SELECT external_id, product_id, price 
FROM store_products 
WHERE external_id = 'ID-PROD-1';

-- Should return:
-- external_id | product_id                           | price
-- ID-PROD-1   | <uuid>                               | 455
```

## Complete Flow

1. **Push Product**: `POST /api/v1/products/push`
   - `products[].id = "ID-PROD-1"`
   - Saves to `store_products.external_id = "ID-PROD-1"`

2. **Update Stock**: `POST /api/v1/products/stock`
   - `products[].id = "ID-PROD-1"`
   - Matches by `store_products.external_id = "ID-PROD-1"`
   - Updates stock successfully ✅

## Restart Required

After this code change, restart the application:

```bash
docker restart gol-bazaar-app-dev
```

Then test with your payload.
