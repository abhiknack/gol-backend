# Common Mistakes - Products Push API

## Issue: taxes_processed = 0

### Problem
Your response shows `"taxes_processed": 0` even though you have taxes in your payload.

### Root Causes

#### 1. Taxes in Wrong Location ❌
```json
{
  "products": [
    {
      "id": "EXT-PROD-001",
      "taxes": ["UUID-GST-5"]  // ❌ WRONG - taxes don't belong here
    }
  ]
}
```

**Fix:** Remove `taxes` from products array. Taxes are store-specific, not product-specific.

#### 2. Missing Taxes in store_products ❌
```json
{
  "store_products": [
    {
      "product_id": "EXT-PROD-001",
      "price": 455,
      "stock_quantity": 120,
      "is_in_stock": true
      // ❌ MISSING: "taxes" field
    }
  ]
}
```

**Fix:** Add `taxes` array to each store_product:
```json
{
  "store_products": [
    {
      "product_id": "EXT-PROD-001",
      "price": 455,
      "stock_quantity": 120,
      "is_in_stock": true,
      "taxes": ["GST_5"]  // ✅ CORRECT
    }
  ]
}
```

#### 3. Using Wrong Tax Identifier ❌
```json
{
  "store_products": [
    {
      "taxes": ["UUID-GST-5"]  // ❌ WRONG - using UUID instead of tax_id
    }
  ]
}
```

**Fix:** Use the `tax_id` field value, not the UUID:
```json
{
  "taxes": [
    {
      "id": "UUID-GST-5",
      "tax_id": "GST_5",  // ← This is what you reference
      "name": "GST 5%",
      "rate": 5.0
    }
  ],
  "store_products": [
    {
      "taxes": ["GST_5"]  // ✅ CORRECT - reference tax_id
    }
  ]
}
```

## Correct Payload Structure

```json
{
  "store_details": { ... },
  
  "taxes": [
    {
      "id": "EXTERNAL-TAX-ID",      // Your ERP's tax ID
      "tax_id": "GST_5",             // Tax code (used for linking)
      "name": "GST 5%",
      "rate": 5.0,
      "tax_type": "percentage"
    }
  ],
  
  "products": [
    {
      "id": "EXT-PROD-001",          // Your ERP's product ID
      "sku": "RICE-FORTUNE-5KG",
      "name": "Fortune Basmati Rice 5kg",
      "price": 455
      // NO taxes field here
    }
  ],
  
  "store_products": [
    {
      "product_id": "EXT-PROD-001",  // Links to products[].id
      "price": 455,
      "stock_quantity": 120,
      "is_in_stock": true,
      "taxes": ["GST_5"]             // ✅ Taxes go here, using tax_id
    }
  ]
}
```

## Field Reference Guide

### products[].id
- **Purpose:** Your ERP's product identifier
- **Used for:** Product matching and mapping
- **Example:** `"EXT-PROD-001"`, `"ZOHO-1001"`

### taxes[].id
- **Purpose:** Your ERP's tax identifier (for reference only)
- **Not used for linking**
- **Example:** `"UUID-GST-5"`, `"TAX-001"`

### taxes[].tax_id
- **Purpose:** Tax code used for linking
- **Used in:** `store_products[].taxes` array
- **Example:** `"GST_5"`, `"GST18"`, `"SERVICE10"`

### store_products[].product_id
- **Purpose:** Links to `products[].id`
- **Must match:** An existing product in the same payload
- **Example:** If product has `"id": "EXT-PROD-001"`, use `"product_id": "EXT-PROD-001"`

### store_products[].taxes
- **Purpose:** Array of tax codes to apply
- **Must match:** `taxes[].tax_id` values
- **Example:** `["GST_5", "SERVICE10"]`

## Validation Checklist

Before sending your payload, verify:

- [ ] `products` array does NOT have a `taxes` field
- [ ] `store_products` array DOES have a `taxes` field
- [ ] `store_products[].taxes` references `taxes[].tax_id` (not `taxes[].id`)
- [ ] `store_products[].product_id` matches `products[].id`
- [ ] `variations[].product_id` matches `products[].id`

## Expected Response

With correct payload:
```json
{
  "status": "success",
  "data": {
    "products_created": 1,
    "products_updated": 0,
    "variations_processed": 2,
    "store_products_processed": 1,
    "taxes_processed": 1  // ✅ Should be > 0
  }
}
```

## Debugging

If `taxes_processed` is still 0:

1. **Check server logs** for warnings:
   ```
   "Tax not found" - tax_id doesn't exist in taxes table
   ```

2. **Verify tax was created:**
   ```sql
   SELECT id, tax_id, name FROM taxes WHERE tax_id = 'GST_5';
   ```

3. **Check store_product_taxes table:**
   ```sql
   SELECT * FROM store_product_taxes 
   WHERE store_product_id = '<your-store-product-id>';
   ```

## See Also

- [API Documentation](./API-PRODUCTS-PUSH.md)
- [Corrected Example](./API-PRODUCTS-PUSH-CORRECTED-EXAMPLE.json)
- [Original Example](./API-PRODUCTS-PUSH-EXAMPLE.json)
