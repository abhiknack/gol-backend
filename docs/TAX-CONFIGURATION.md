# Tax Configuration Guide

## Overview

The system supports two tax models to handle different business scenarios:

1. **Product-level taxes** (`product_taxes`) - Same tax for all stores
2. **Store-specific taxes** (`store_product_taxes`) - Different taxes per store

## Tax Model

The system uses **store-specific taxes** (`store_product_taxes`) exclusively. This provides maximum flexibility for multi-store operations.

### Why Store-specific Taxes?

✅ **Handles all scenarios:**

#### 1. Different States/Regions
Different states have different tax rates:
- Karnataka: 18% GST
- Delhi: 12% GST
- Maharashtra: 15% GST + local cess

#### 2. Store-specific GST Registration
- **GST Registered Store**: Charges full GST
- **Composition Scheme Store**: Charges reduced rate (1-5%)
- **Non-GST Seller**: No GST charged

#### 3. Product Category Variations
- **Alcohol**: Varies by state (20-70%)
- **Tobacco**: State-specific taxes
- **Fuel**: Central + state taxes
- **Food items**: Some states exempt, others charge

#### 4. Local Taxes & Cess
- Municipal taxes
- Service charges (restaurants)
- Environmental cess
- Luxury tax

#### 5. Franchise Models
Each franchise outlet may have:
- Different GST numbers
- Different tax schemes
- Different state registrations

## Database Schema

### Store Product Taxes

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

## Usage Examples

### Example 1: Same Tax for All Stores

**Scenario:** Laptop with 18% GST nationwide

```sql
-- Create tax
INSERT INTO taxes (id, name, tax_id, rate, tax_type, is_inclusive)
VALUES ('gst-18-uuid', 'GST 18%', 'GST18', 18.00, 'percentage', false);

-- Apply to all stores (insert for each store)
INSERT INTO store_product_taxes (store_id, store_product_id, tax_id)
SELECT 
    sp.store_id,
    sp.id as store_product_id,
    'gst-18-uuid'
FROM store_products sp
WHERE sp.product_id = 'laptop-uuid';
```

### Example 2: Different Tax Rates by State

**Scenario:** Alcohol with state-specific taxes

```sql
-- Karnataka store (70% tax)
INSERT INTO store_product_taxes (
    store_id, store_product_id, tax_id
) VALUES (
    'karnataka-store-uuid',
    'karnataka-whiskey-store-product-uuid',
    'alcohol-tax-karnataka-uuid'
);

-- Delhi store (50% tax)
INSERT INTO store_product_taxes (
    store_id, store_product_id, tax_id
) VALUES (
    'delhi-store-uuid',
    'delhi-whiskey-store-product-uuid',
    'alcohol-tax-delhi-uuid'
);
```

### Example 3: GST Registered vs Composition Scheme

**Scenario:** Same product, different stores with different GST schemes

```sql
-- Store A: GST Registered (18% GST)
INSERT INTO store_product_taxes (
    store_id, store_product_id, tax_id
) VALUES (
    'store-a-uuid',
    'store-a-product-uuid',
    'gst-18-uuid'
);

-- Store B: Composition Scheme (1% tax)
INSERT INTO store_product_taxes (
    store_id, store_product_id, tax_id
) VALUES (
    'store-b-uuid',
    'store-b-product-uuid',
    'composition-1-uuid'
);

-- Store C: Non-GST (no tax)
-- Don't insert any tax record
```

### Example 4: Tax Rate Override

**Scenario:** Special promotional rate for specific store

```sql
-- Normal rate is 18%, but this store gets 12%
INSERT INTO store_product_taxes (
    store_id, store_product_id, tax_id, override_rate
) VALUES (
    'promotional-store-uuid',
    'store-product-uuid',
    'gst-tax-uuid',
    12.00  -- Override to 12% instead of default 18%
);
```

### Example 5: Multiple Taxes per Product

**Scenario:** Product with GST + Service Charge (restaurant)

```sql
-- GST 5%
INSERT INTO store_product_taxes (
    store_id, store_product_id, tax_id
) VALUES (
    'restaurant-uuid',
    'food-item-uuid',
    'gst-5-uuid'
);

-- Service Charge 10%
INSERT INTO store_product_taxes (
    store_id, store_product_id, tax_id
) VALUES (
    'restaurant-uuid',
    'food-item-uuid',
    'service-charge-uuid'
);
```

## Tax Calculation Logic

### Tax Resolution

```
1. Query store_product_taxes for the specific store + product
2. For each tax:
   ├─ If override_rate is set → Use override_rate
   └─ If override_rate is NULL → Use tax.rate from taxes table
3. If no taxes found → Product has no tax
```

### Calculation Example

**Product:** Laptop (₹50,000)
**Store:** Karnataka Store

```sql
-- Query taxes for this store+product
SELECT 
    t.name,
    COALESCE(spt.override_rate, t.rate) as effective_rate,
    t.is_inclusive
FROM store_product_taxes spt
JOIN taxes t ON spt.tax_id = t.id
WHERE spt.store_id = 'karnataka-store-uuid'
  AND spt.store_product_id = 'laptop-store-product-uuid'
  AND spt.is_active = true;
```

**Result:**
- GST 18%: ₹9,000
- **Total:** ₹59,000

## API Integration

### When Pushing Products from ERP

```json
{
  "store_details": {
    "store_id": "EXT123456"
  },
  "products": [...],
  "store_products": [
    {
      "product_id": "PROD-001",
      "price": 50000,
      "taxes": [
        {
          "tax_id": "GST-18",
          "override_rate": null
        }
      ]
    }
  ]
}
```

### Tax Resolution Flow

```
1. Receive product from ERP with tax_id
2. Find or create tax in taxes table
3. Create store_product entry
4. Create store_product_taxes entry linking:
   - store_id
   - store_product_id
   - tax_id
   - override_rate (if provided)
```

## Best Practices

### 1. Create Taxes for All Stores

When a product has the same tax across all stores, create `store_product_taxes` entries for each store to maintain consistency.

### 2. Use Tax IDs Consistently

Maintain consistent tax IDs across ERPs:
- `GST-18` for 18% GST
- `GST-12` for 12% GST
- `CGST-9` for 9% CGST
- `SGST-9` for 9% SGST

### 3. Handle Tax Exemptions

For tax-exempt products/stores, simply don't create any tax records.

### 4. Track Tax Changes

Use `updated_at` to track when tax configurations change for compliance.

### 5. Validate Tax Rates

Ensure tax rates are valid:
- Percentage: 0-100
- Fixed amount: > 0
- Inclusive vs Exclusive properly set

## Query Examples

### Get All Taxes for a Store Product

```sql
SELECT 
    t.name,
    t.tax_id,
    COALESCE(spt.override_rate, t.rate) as rate,
    t.tax_type,
    t.is_inclusive
FROM store_product_taxes spt
JOIN taxes t ON spt.tax_id = t.id
WHERE spt.store_id = $1
  AND spt.store_product_id = $2
  AND spt.is_active = true;
```

### Get Tax Summary by Store

```sql
SELECT 
    s.name as store_name,
    COUNT(DISTINCT spt.store_product_id) as products_with_taxes,
    COUNT(DISTINCT spt.tax_id) as unique_taxes,
    AVG(COALESCE(spt.override_rate, t.rate)) as avg_tax_rate
FROM store_product_taxes spt
JOIN stores s ON spt.store_id = s.id
JOIN taxes t ON spt.tax_id = t.id
WHERE spt.is_active = true
GROUP BY s.id, s.name;
```

### Find Products with Different Taxes Across Stores

```sql
SELECT 
    p.name as product_name,
    COUNT(DISTINCT spt.tax_id) as different_tax_configs
FROM products p
JOIN store_products sp ON p.id = sp.product_id
JOIN store_product_taxes spt ON sp.id = spt.store_product_id
WHERE spt.is_active = true
GROUP BY p.id, p.name
HAVING COUNT(DISTINCT spt.tax_id) > 1;
```

## Bulk Tax Assignment

### Apply Same Tax to All Stores

```sql
-- Apply a tax to a product across all stores
INSERT INTO store_product_taxes (
    store_id, store_product_id, tax_id
)
SELECT 
    sp.store_id,
    sp.id as store_product_id,
    'tax-uuid'
FROM store_products sp
WHERE sp.product_id = 'product-uuid'
ON CONFLICT (store_id, store_product_id, tax_id) DO NOTHING;
```

### Apply Tax to All Products in a Store

```sql
-- Apply a tax to all products in a specific store
INSERT INTO store_product_taxes (
    store_id, store_product_id, tax_id
)
SELECT 
    'store-uuid',
    sp.id as store_product_id,
    'tax-uuid'
FROM store_products sp
WHERE sp.store_id = 'store-uuid'
ON CONFLICT (store_id, store_product_id, tax_id) DO NOTHING;
```

## Compliance & Reporting

### Tax Report by Store

```sql
SELECT 
    s.name as store_name,
    s.city,
    s.state,
    t.name as tax_name,
    t.tax_id,
    COUNT(*) as products_count,
    SUM(sp.price * COALESCE(spt.override_rate, t.rate) / 100) as total_tax_amount
FROM store_product_taxes spt
JOIN stores s ON spt.store_id = s.id
JOIN store_products sp ON spt.store_product_id = sp.id
JOIN taxes t ON spt.tax_id = t.id
WHERE spt.is_active = true
GROUP BY s.id, s.name, s.city, s.state, t.id, t.name, t.tax_id
ORDER BY s.state, s.city, t.name;
```

This flexible tax model supports all scenarios from simple single-store operations to complex multi-state, multi-franchise deployments!
