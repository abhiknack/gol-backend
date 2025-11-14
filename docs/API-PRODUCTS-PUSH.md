# POST /api/v1/products/push

## Overview

Bulk product upsert endpoint for ERP integration. Supports product matching, brand normalization, store-specific pricing, and tax configuration.

## Features

✅ **Product Matching** - Automatically matches products across ERPs using 3-layer strategy  
✅ **Brand Normalization** - Handles brand name variations (Coca Cola, CocaCola, Coke)  
✅ **Store Mapping** - Creates store_product_mappings for ERP integration  
✅ **Store-specific Pricing** - Different prices per store  
✅ **Store-specific Taxes** - Tax configuration per store+product  
✅ **Variations** - Product variations (sizes, flavors)  
✅ **Categories** - Hierarchical categories  

## Request

### Endpoint
```
POST /api/v1/products/push
Content-Type: application/json
```

### Request Body

```json
{
  "store_details": {
    "store_id": "EXT123456",
    "name": "Main Supermarket",
    "address": {
      "line1": "123 Main St",
      "city": "Mumbai",
      "state": "Maharashtra",
      "postal_code": "400001"
    },
    "location": {
      "lat": 19.0760,
      "lng": 72.8777
    }
  },
  "categories": [
    {
      "id": "CAT-BEVERAGES",
      "parent_id": null,
      "name": "Beverages",
      "slug": "beverages",
      "description": "Soft drinks and beverages",
      "display_order": 1,
      "is_active": true
    }
  ],
  "taxes": [
    {
      "id": "GST-18",
      "name": "GST 18%",
      "tax_id": "GST18",
      "description": "Goods and Services Tax 18%",
      "rate": 18.00,
      "tax_type": "percentage",
      "is_inclusive": false,
      "is_active": true
    }
  ],
  "products": [
    {
      "id": "ZOHO-1001",
      "sku": "CK1L01",
      "name": "Coca Cola 1L",
      "slug": "coca-cola-1l",
      "description": "Coca Cola Soft Drink 1 Litre",
      "category_id": "CAT-BEVERAGES",
      "price": 50.00,
      "currency": "INR",
      "unit": "liter",
      "unit_quantity": 1.0,
      "primary_image_url": "https://example.com/images/coke-1l.jpg",
      "images": [
        "https://example.com/images/coke-1l-front.jpg",
        "https://example.com/images/coke-1l-back.jpg"
      ],
      "brand": "Coca Cola",
      "manufacturer": "Coca-Cola Company",
      "barcode": "8901234567",
      "ean": "8901234567890",
      "is_active": true,
      "is_featured": false,
      "is_customizable": false,
      "is_addon": false
    }
  ],
  "variations": [
    {
      "product_id": "ZOHO-1001",
      "name": "Small",
      "display_name": "250ml",
      "price": 20.00,
      "is_default": false
    },
    {
      "product_id": "ZOHO-1001",
      "name": "Medium",
      "display_name": "500ml",
      "price": 35.00,
      "is_default": false
    },
    {
      "product_id": "ZOHO-1001",
      "name": "Large",
      "display_name": "1L",
      "price": 50.00,
      "is_default": true
    }
  ],
  "store_products": [
    {
      "product_id": "ZOHO-1001",
      "price": 50.00,
      "stock_quantity": 100,
      "is_in_stock": true,
      "taxes": ["GST-18"]
    }
  ]
}
```

### Field Descriptions

#### store_details (required)
- `store_id` - ERP's store identifier (external_id)
- `name` - Store name
- `address` - Store address
- `location` - GPS coordinates

#### categories (optional)
- `id` - External category ID
- `parent_id` - Parent category ID (for hierarchy)
- `name` - Category name
- `slug` - URL-friendly identifier
- `display_order` - Sort order
- `is_active` - Active status

#### taxes (optional)
- `id` - External tax ID
- `name` - Tax name
- `tax_id` - Tax identifier code
- `rate` - Tax rate (percentage or fixed)
- `tax_type` - "percentage" or "fixed"
- `is_inclusive` - Whether tax is included in price

#### products (required)
- `id` - **ERP's product ID** (used for matching and mapping)
- `sku` - Stock Keeping Unit
- `name` - Product name
- `slug` - URL-friendly identifier (optional, defaults to SKU)
- `category_id` - External category ID
- `price` - Base price
- `brand` - Brand name (will be normalized)
- `barcode` - Product barcode (used for matching)
- `ean` - EAN code (used for matching)
- `images` - Array of image URLs
- `is_addon` - Whether product can be used as addon

#### variations (optional)
- `product_id` - External product ID
- `name` - Variation name (e.g., "Small", "Medium", "Large")
- `display_name` - Display name (e.g., "250ml", "500ml", "1L")
- `price` - Variation price
- `is_default` - Whether this is the default selection

#### store_products (required)
- `product_id` - Links to products.id
- `price` - Store-specific price
- `stock_quantity` - Current stock level
- `is_in_stock` - Stock availability
- `taxes` - Array of tax IDs for this store-product

## Response

### Success Response (200 OK)

```json
{
  "status": "success",
  "data": {
    "products_created": 5,
    "products_updated": 3,
    "variations_processed": 15,
    "store_products_processed": 8,
    "taxes_processed": 12
  },
  "message": "Products pushed successfully"
}
```

### Error Responses

#### 400 Bad Request
```json
{
  "status": "error",
  "error": {
    "code": "INVALID_INPUT",
    "message": "Validation error details"
  }
}
```

#### 500 Internal Server Error
```json
{
  "status": "error",
  "error": {
    "code": "PRODUCT_UPSERT_FAILED",
    "message": "Failed to create or update products"
  }
}
```

## Product Matching Logic

The API uses a 3-layer matching strategy to prevent duplicate products:

### Layer 1: Exact Match (100% confidence)
1. Check existing mapping (store + product id)
2. Match by barcode
3. Match by EAN
4. Match by SKU

### Layer 2: Normalized Match (95% confidence)
1. Normalize product name
2. Extract volume/weight
3. Match by normalized name + size

### Layer 3: Fuzzy Match (45-90% confidence)
1. Use trigram similarity
2. Match if similarity > 0.45

### No Match Found
- Create new product
- Create store_product_mapping

## Brand Normalization

Brand names are automatically normalized:

```
Input: "Coca-Cola" → Normalized: "coca cola"
Input: "CocaCola"  → Normalized: "cocacola"
Input: "Coke"      → Normalized: "coke"
```

The system will:
1. Try to find existing brand by exact name
2. Try to find by normalized name
3. Create new brand if not found

## Store Product Mapping

For each product, the system creates a mapping:

```
store_product_mappings:
- store_id: store UUID
- external_product_id: "ZOHO-1001" (from products.id in payload)
- product_id: internal product UUID
- external_sku: "CK1L01"
- external_barcode: "8901234567"
- external_name: "Coca Cola 1L"
- sync_source: "API"
```

This allows:
- Different ERPs to use different product IDs
- Tracking sync history
- Stable product matching across syncs

## Tax Configuration

Taxes are configured per store-product:

```
store_product_taxes:
- store_id: store UUID
- store_product_id: store_product UUID
- tax_id: tax UUID
- is_active: true
```

This allows:
- Different tax rates per store (state-specific GST)
- Multiple taxes per product (GST + Service Charge)
- Tax rate overrides

## Examples

### Example 1: Simple Product Push

```bash
curl -X POST http://localhost:8080/api/v1/products/push \
  -H "Content-Type: application/json" \
  -d '{
    "store_details": {
      "store_id": "STORE-001",
      "name": "Mumbai Store",
      "address": {
        "line1": "123 Main St",
        "city": "Mumbai",
        "state": "Maharashtra",
        "postal_code": "400001"
      },
      "location": {
        "lat": 19.0760,
        "lng": 72.8777
      }
    },
    "products": [
      {
        "external_product_id": "PROD-001",
        "sku": "CK1L",
        "name": "Coca Cola 1L",
        "price": 50.00,
        "brand": "Coca Cola",
        "barcode": "8901234567",
        "is_active": true
      }
    ],
    "store_products": [
      {
        "external_product_id": "PROD-001",
        "price": 50.00,
        "stock_quantity": 100,
        "is_in_stock": true
      }
    ]
  }'
```

### Example 2: Product with Variations

```json
{
  "products": [
    {
      "id": "PIZZA-001",
      "sku": "PIZZA-MARG",
      "name": "Margherita Pizza",
      "price": 299.00,
      "is_customizable": true
    }
  ],
  "variations": [
    {
      "product_id": "PIZZA-001",
      "name": "Small",
      "display_name": "8 inch",
      "price": 199.00
    },
    {
      "product_id": "PIZZA-001",
      "name": "Medium",
      "display_name": "10 inch",
      "price": 299.00,
      "is_default": true
    },
    {
      "product_id": "PIZZA-001",
      "name": "Large",
      "display_name": "12 inch",
      "price": 399.00
    }
  ]
}
```

### Example 3: Product with Store-specific Taxes

```json
{
  "taxes": [
    {
      "id": "GST-5",
      "name": "GST 5%",
      "tax_id": "GST5",
      "rate": 5.00,
      "tax_type": "percentage",
      "is_inclusive": false
    },
    {
      "id": "SERVICE-10",
      "name": "Service Charge 10%",
      "tax_id": "SERVICE10",
      "rate": 10.00,
      "tax_type": "percentage",
      "is_inclusive": false
    }
  ],
  "store_products": [
    {
      "product_id": "FOOD-001",
      "price": 250.00,
      "taxes": ["GST-5", "SERVICE-10"]
    }
  ]
}
```

## Best Practices

1. **Use Consistent IDs** - Keep external_product_id stable across syncs
2. **Include Barcodes** - Helps with product matching
3. **Normalize Brand Names** - System handles variations automatically
4. **Set Stock Levels** - Keep inventory accurate
5. **Configure Taxes** - Set store-specific taxes correctly
6. **Use Variations** - For size/flavor options
7. **Batch Updates** - Send multiple products in one request

## Limitations

- Maximum 1000 products per request
- Maximum 10 images per product
- Maximum 20 variations per product
- Maximum 5 taxes per store-product

## Migration from Old API

### Old Format (Deprecated)
```json
{
  "products": [
    {
      "id": "internal-uuid",
      "external_id": "ZOHO-1001",
      ...
    }
  ]
}
```

### New Format
```json
{
  "products": [
    {
      "external_product_id": "ZOHO-1001",
      ...
    }
  ]
}
```

**Key Changes:**
- `id` → removed (system generates UUIDs)
- `external_id` → `external_product_id`
- `brand` → normalized automatically
- `taxes` → moved to store_products level
- Product matching engine handles duplicates

## Troubleshooting

### Issue: Duplicate Products Created

**Cause:** Different external_product_id for same product

**Solution:** Use consistent IDs or include barcode for matching

### Issue: Brand Not Matching

**Cause:** Brand name variation too different

**Solution:** Use standard brand names or check normalization

### Issue: Tax Not Applied

**Cause:** Tax ID not found or not linked to store_product

**Solution:** Ensure tax is created and ID matches

### Issue: Product Not Found for Variation

**Cause:** external_product_id mismatch

**Solution:** Ensure variation.product_id matches product.external_product_id

## See Also

- [Product Matching Engine](./PRODUCT-MATCHING-ENGINE.md)
- [Brand Normalization](./BRAND-NORMALIZATION.md)
- [Tax Configuration](./TAX-CONFIGURATION.md)
- [Addon Architecture](./ADDON-ARCHITECTURE.md)
