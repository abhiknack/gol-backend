# Tax Mapping Guide: External IDs to Internal UUIDs

## Overview

The system uses **external tax IDs** from your ERP and automatically maps them to internal UUIDs for database relationships.

## Tax Flow

```
Your ERP Tax ID → taxes.external_id → Internal UUID → store_product_taxes.tax_id
   "UUID-GST-5"         "UUID-GST-5"      (generated)         (UUID reference)
```

## Payload Structure

### 1. Define Taxes
```json
{
  "taxes": [
    {
      "id": "UUID-GST-5",           // Your ERP's tax ID
      "tax_id": "GST_5",             // Tax code (for display/reference)
      "name": "GST 5%",
      "rate": 5.0,
      "tax_type": "percentage"
    }
  ]
}
```

**What happens:**
- `id` → stored as `taxes.external_id`
- System generates internal UUID for `taxes.id`
- `tax_id` stored as `taxes.tax_id`

### 2. Reference Taxes in store_products
```json
{
  "store_products": [
    {
      "product_id": "EXT-PROD-001",
      "price": 455,
      "taxes": ["UUID-GST-5"]       // Reference external_id
    }
  ]
}
```

**What happens:**
- System looks up: `SELECT id FROM taxes WHERE external_id = 'UUID-GST-5'`
- Gets internal UUID (e.g., `a1b2c3d4-...`)
- Creates link: `store_product_taxes.tax_id = a1b2c3d4-...`

## Database Tables

### taxes table
```sql
id           | UUID (generated)      | a1b2c3d4-5678-...
external_id  | VARCHAR(255)          | UUID-GST-5
tax_id       | VARCHAR(50)           | GST_5
name         | VARCHAR(100)          | GST 5%
rate         | DECIMAL(5,2)          | 5.00
```

### store_product_taxes table
```sql
id                | UUID (generated)
store_id          | UUID (reference)
store_product_id  | UUID (reference)
tax_id            | UUID (reference) → taxes.id (internal UUID)
```

## Complete Example

### Request Payload
```json
{
  "store_details": {
    "store_id": "EXT123456",
    "name": "Main Store"
  },
  "taxes": [
    {
      "id": "ZOHO-TAX-001",
      "tax_id": "GST_5",
      "name": "GST 5%",
      "rate": 5.0,
      "tax_type": "percentage"
    },
    {
      "id": "ZOHO-TAX-002",
      "tax_id": "GST_18",
      "name": "GST 18%",
      "rate": 18.0,
      "tax_type": "percentage"
    }
  ],
  "products": [
    {
      "id": "ZOHO-PROD-001",
      "sku": "RICE-5KG",
      "name": "Rice 5kg",
      "price": 455
    }
  ],
  "store_products": [
    {
      "product_id": "ZOHO-PROD-001",
      "price": 455,
      "taxes": ["ZOHO-TAX-001"]
    }
  ]
}
```

### Database Result

**taxes table:**
```
id                                   | external_id    | tax_id | name    | rate
-------------------------------------|----------------|--------|---------|------
a1b2c3d4-5678-90ab-cdef-123456789abc | ZOHO-TAX-001   | GST_5  | GST 5%  | 5.00
b2c3d4e5-6789-01bc-def0-234567890bcd | ZOHO-TAX-002   | GST_18 | GST 18% | 18.00
```

**store_product_taxes table:**
```
id       | store_id | store_product_id | tax_id (internal UUID)
---------|----------|------------------|----------------------------------
xyz123...| abc456...| def789...        | a1b2c3d4-5678-90ab-cdef-123456789abc
```

## Multiple Taxes Per Product

```json
{
  "taxes": [
    {
      "id": "TAX-GST-5",
      "tax_id": "GST_5",
      "rate": 5.0
    },
    {
      "id": "TAX-SERVICE-10",
      "tax_id": "SERVICE_10",
      "rate": 10.0
    }
  ],
  "store_products": [
    {
      "product_id": "PROD-001",
      "price": 100,
      "taxes": ["TAX-GST-5", "TAX-SERVICE-10"]  // Multiple taxes
    }
  ]
}
```

## Verification Queries

### Check tax mapping
```sql
SELECT 
    t.external_id,
    t.tax_id,
    t.name,
    t.rate,
    t.id as internal_uuid
FROM taxes t
WHERE t.external_id = 'UUID-GST-5';
```

### Check product taxes
```sql
SELECT 
    p.name as product_name,
    t.external_id as tax_external_id,
    t.tax_id,
    t.name as tax_name,
    t.rate
FROM store_product_taxes spt
JOIN store_products sp ON sp.id = spt.store_product_id
JOIN products p ON p.id = sp.product_id
JOIN taxes t ON t.id = spt.tax_id
WHERE sp.external_id = 'EXT-PROD-001';
```

## Common Mistakes

### ❌ Wrong: Using tax_id in store_products
```json
{
  "store_products": [{
    "taxes": ["GST_5"]  // ❌ This is tax_id, not external_id
  }]
}
```

### ✅ Correct: Using external_id in store_products
```json
{
  "store_products": [{
    "taxes": ["UUID-GST-5"]  // ✅ This is external_id
  }]
}
```

### ❌ Wrong: Taxes in products array
```json
{
  "products": [{
    "taxes": ["UUID-GST-5"]  // ❌ Products don't have taxes
  }]
}
```

### ✅ Correct: Taxes in store_products array
```json
{
  "store_products": [{
    "taxes": ["UUID-GST-5"]  // ✅ Taxes are store-specific
  }]
}
```

## Benefits

1. **ERP Independence**: Your ERP uses its own tax IDs
2. **No UUID Management**: System generates UUIDs automatically
3. **Stable References**: External IDs remain constant
4. **Multi-ERP Support**: Different ERPs can use different tax IDs
5. **Clean Separation**: External IDs for API, UUIDs for database

## Troubleshooting

### Issue: taxes_processed = 0

**Check 1:** Are taxes in store_products?
```json
"store_products": [{"taxes": ["UUID-GST-5"]}]  // Must be here
```

**Check 2:** Does external_id match?
```sql
SELECT * FROM taxes WHERE external_id = 'UUID-GST-5';
```

**Check 3:** Check logs
```
Tax not found by external_id: UUID-GST-5
```

### Issue: Tax not found

**Cause:** Tax wasn't created or external_id doesn't match

**Solution:** Ensure tax is in the `taxes` array with matching `id`
