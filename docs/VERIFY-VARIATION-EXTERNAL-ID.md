# Verify Variation External ID Mapping

## Changes Made

1. ✅ Added `ID` field to `Variation` struct in handler
2. ✅ Added `ExternalID` field to `VariationInput` in repository
3. ✅ Updated variation insert to save `external_id`
4. ✅ Updated stock endpoint to support variation updates

## Test Payload

### 1. Push Products with Variations

```json
{
  "store_details": {
    "store_id": "EXT123456",
    "name": "Main Supermarket"
  },
  "products": [{
    "id": "ID-PROD-1",
    "sku": "RICE-FORTUNE",
    "name": "Fortune Basmati Rice",
    "price": 455,
    "taxes": ["UUID-GST-5"]
  }],
  "variations": [
    {
      "id": "ID-VAR-5KG",
      "product_id": "ID-PROD-1",
      "name": "5kg",
      "display_name": "5 kg Pack",
      "price": 455,
      "is_default": true
    },
    {
      "id": "ID-VAR-1KG",
      "product_id": "ID-PROD-1",
      "name": "1kg",
      "display_name": "1 kg Pack",
      "price": 110,
      "is_default": false
    }
  ]
}
```

### 2. Verify in Database

```sql
-- Check product external_id
SELECT external_id, name, sku 
FROM store_products sp
JOIN products p ON p.id = sp.product_id
WHERE sp.external_id = 'ID-PROD-1';

-- Check variation external_ids
SELECT 
    pv.external_id,
    pv.name,
    pv.display_name,
    pv.price,
    pv.stock_quantity,
    p.name as product_name
FROM product_variations pv
JOIN products p ON p.id = pv.product_id
WHERE pv.external_id IN ('ID-VAR-5KG', 'ID-VAR-1KG');
```

Expected result:
```
external_id  | name | display_name | price | stock_quantity | product_name
-------------|------|--------------|-------|----------------|------------------
ID-VAR-5KG   | 5kg  | 5 kg Pack    | 455   | NULL           | Fortune Basmati Rice
ID-VAR-1KG   | 1kg  | 1 kg Pack    | 110   | NULL           | Fortune Basmati Rice
```

### 3. Update Stock with Variations

```json
{
  "store_id": "EXT123456",
  "products": [
    {
      "id": "ID-PROD-1",
      "stock_quantity": 120,
      "is_available": true,
      "variants": [
        {
          "id": "ID-VAR-5KG",
          "stock_quantity": 80,
          "is_available": true,
          "price": 455
        },
        {
          "id": "ID-VAR-1KG",
          "stock_quantity": 40,
          "is_available": true,
          "price": 110
        }
      ]
    }
  ]
}
```

Expected response:
```json
{
  "status": "success",
  "data": {
    "products_updated": 1,
    "products_not_found": 0,
    "variants_updated": 2,
    "variants_not_found": 0
  },
  "message": "Stock updated successfully"
}
```

### 4. Verify Stock Updated

```sql
-- Check product stock
SELECT external_id, stock_quantity, is_in_stock, is_available
FROM store_products
WHERE external_id = 'ID-PROD-1';

-- Check variation stock
SELECT external_id, name, stock_quantity, is_in_stock, is_active, price
FROM product_variations
WHERE external_id IN ('ID-VAR-5KG', 'ID-VAR-1KG');
```

Expected result:
```
-- Product
external_id | stock_quantity | is_in_stock | is_available
ID-PROD-1   | 120            | true        | true

-- Variations
external_id | name | stock_quantity | is_in_stock | is_active | price
ID-VAR-5KG  | 5kg  | 80             | true        | true      | 455
ID-VAR-1KG  | 1kg  | 40             | true        | true      | 110
```

## API Endpoints

### Push Products
```bash
POST /api/v1/products/push
```

### Update Stock
```bash
POST /api/v1/products/stock
```

## Complete Flow

1. **Push products with variations**
   - `variations[].id` → saved as `product_variations.external_id`
   - `products[].id` → saved as `store_products.external_id`

2. **Update stock**
   - `products[].id` → matches `store_products.external_id`
   - `variants[].id` → matches `product_variations.external_id`
   - Updates stock for both products and variations

## Test Files

- `test-stock-with-variations.json` - Stock update payload with variations
- Your final payload - Complete product push with variations

## Troubleshooting

### Issue: variants_not_found > 0

**Cause:** Variation external_id doesn't exist

**Solution:**
```sql
-- Check if variation exists
SELECT * FROM product_variations WHERE external_id = 'ID-VAR-5KG';
```

### Issue: Variation not updating

**Cause:** Wrong external_id or variation not created

**Solution:**
1. Ensure variation was pushed with `id` field
2. Check external_id matches exactly
3. Verify variation belongs to correct product

## Summary

✅ **Product External ID**: `products[].id` → `store_products.external_id`  
✅ **Variation External ID**: `variations[].id` → `product_variations.external_id`  
✅ **Stock Update**: Both products and variations can be updated by external_id  
✅ **Complete Flow**: Push → Verify → Update → Verify
