# Payload Fix: Tax Mapping with External IDs

## How Tax Mapping Works

The system now uses **external tax IDs** for mapping:

1. **taxes.id** → Stored as `taxes.external_id` (your ERP's tax ID)
2. **taxes.tax_id** → Tax code for display/reference
3. **store_products.taxes[]** → References `taxes.external_id` (your ERP's tax IDs)
4. **store_product_taxes.tax_id** → Uses internal UUID (auto-mapped)

## Your Current Payload Issues

### ❌ Issue 1: Taxes in Products Array
```json
"products": [{
  "id": "UUID-PROD-1",
  "taxes": ["UUID-GST-5"]  // ❌ WRONG - Remove this field
}]
```

**Why it's wrong:** Products don't have taxes. Taxes are store-specific and belong in `store_products`.

### ❌ Issue 2: Missing Taxes in store_products
```json
"store_products": [{
  "product_id": "UUID-PROD-1",
  "price": 455,
  "stock_quantity": 120,
  "is_in_stock": true
  // ❌ MISSING: "taxes": ["UUID-GST-5"]
}]
```

**Why it's wrong:** This is where taxes should be!

## ✅ Corrected Structure

```json
{
  "taxes": [
    {
      "id": "UUID-GST-5",        // Your ERP's tax ID (stored as external_id)
      "tax_id": "GST_5",          // Tax code for reference
      "name": "GST 5%",
      "rate": 5.0
    }
  ],
  "products": [
    {
      "id": "EXT-PROD-001",
      // NO taxes field here
    }
  ],
  "store_products": [
    {
      "product_id": "EXT-PROD-001",
      "taxes": ["UUID-GST-5"]     // ✅ Use external_id (taxes.id)
    }
  ]
}
```

## How It Works

1. Tax is created with `external_id = "UUID-GST-5"` and internal UUID generated
2. System looks up tax by `external_id = "UUID-GST-5"`
3. System links `store_product_taxes` using internal UUID
4. Your ERP always uses external IDs, system handles UUID mapping

## Expected Result

```json
{
  "data": {
    "products_created": 0,
    "products_updated": 1,
    "variations_processed": 2,
    "store_products_processed": 1,
    "taxes_processed": 1        // ✅ Should be 1
  }
}
```
