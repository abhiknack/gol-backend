# Simplified API Structure

## Overview

The API now supports a **simplified payload structure** where you only need to send `products` with taxes, and the system automatically creates `store_products` entries.

## Simplified Payload (Recommended)

```json
{
  "store_details": {
    "store_id": "EXT123456",
    "name": "Main Supermarket",
    "address": {
      "line1": "Main Road",
      "city": "Mumbai",
      "state": "MH",
      "postal_code": "400001"
    },
    "location": {
      "lat": 19.076,
      "lng": 72.8777
    }
  },
  "categories": [
    {
      "id": "UUID-CAT-RICE",
      "name": "Rice & Grains",
      "slug": "rice-and-grains"
    }
  ],
  "taxes": [
    {
      "id": "UUID-GST-5",
      "tax_id": "GST_5",
      "name": "GST 5%",
      "rate": 5.0,
      "tax_type": "percentage"
    }
  ],
  "products": [
    {
      "id": "UUID-PROD-1",
      "sku": "RICE-FORTUNE-5KG",
      "name": "Fortune Basmati Rice 5kg",
      "price": 455,
      "brand": "Fortune",
      "barcode": "1234567890123",
      "taxes": ["UUID-GST-5"],
      "is_active": true
    }
  ],
  "variations": [
    {
      "product_id": "UUID-PROD-1",
      "name": "5kg",
      "display_name": "5 kg Pack",
      "price": 455,
      "is_default": true
    }
  ]
}
```

**Note:** No `store_products` array needed! The system automatically creates store_products from products.

## What Happens Automatically

When you send products with taxes:

1. **Product Created/Updated**: Product is matched or created
2. **Store Product Auto-Created**: System creates `store_products` entry with:
   - `external_id` = product.id
   - `price` = product.price
   - `stock_quantity` = 0 (default)
   - `is_in_stock` = true (default)
3. **Taxes Linked**: System links taxes using `product.taxes[]` array
4. **Result**: `taxes_processed` count reflects linked taxes

## Advanced: Explicit store_products (Optional)

If you need more control (custom stock, pricing per store), you can still provide `store_products`:

```json
{
  "products": [
    {
      "id": "UUID-PROD-1",
      "sku": "RICE-5KG",
      "name": "Rice 5kg",
      "price": 455
    }
  ],
  "store_products": [
    {
      "product_id": "UUID-PROD-1",
      "price": 450,
      "stock_quantity": 100,
      "is_in_stock": true,
      "taxes": ["UUID-GST-5"]
    }
  ]
}
```

## Tax Mapping

Taxes in `products.taxes[]` are mapped by `external_id`:

```json
{
  "taxes": [
    {
      "id": "UUID-GST-5",        // External ID (your ERP's ID)
      "tax_id": "GST_5",          // Tax code
      "name": "GST 5%",
      "rate": 5.0
    }
  ],
  "products": [
    {
      "id": "PROD-001",
      "taxes": ["UUID-GST-5"]     // References taxes[].id
    }
  ]
}
```

**System Process:**
1. Looks up: `SELECT id FROM taxes WHERE external_id = 'UUID-GST-5'`
2. Gets internal UUID
3. Links to `store_product_taxes` using internal UUID

## Complete Example

```json
{
  "store_details": {
    "store_id": "EXT123456",
    "name": "Main Supermarket",
    "address": {
      "line1": "Main Road",
      "city": "Mumbai",
      "state": "MH",
      "postal_code": "400001"
    },
    "location": {
      "lat": 19.076,
      "lng": 72.8777
    }
  },
  "categories": [
    {
      "id": "UUID-CAT-GROCERY",
      "parent_id": null,
      "name": "Grocery",
      "slug": "grocery"
    },
    {
      "id": "UUID-CAT-RICE",
      "parent_id": "UUID-CAT-GROCERY",
      "name": "Rice & Grains",
      "slug": "rice-and-grains"
    }
  ],
  "taxes": [
    {
      "id": "UUID-GST-5",
      "tax_id": "GST_5",
      "name": "GST 5%",
      "rate": 5.0,
      "tax_type": "percentage",
      "is_inclusive": false,
      "is_active": true
    }
  ],
  "products": [
    {
      "id": "UUID-PROD-1",
      "sku": "RICE-FORTUNE-5KG",
      "name": "Fortune Basmati Rice 5kg",
      "slug": "fortune-basmati-rice-5kg",
      "description": "Premium basmati rice.",
      "category_id": "UUID-CAT-RICE",
      "price": 455,
      "currency": "INR",
      "unit": "kg",
      "unit_quantity": 5,
      "primary_image_url": "https://cdn.example.com/rice.png",
      "images": [
        "https://cdn.example.com/rice_1.png",
        "https://cdn.example.com/rice_2.png"
      ],
      "brand": "Fortune",
      "manufacturer": "Adani Wilmar",
      "barcode": "1234567890123",
      "ean": "8901234567890",
      "taxes": ["UUID-GST-5"],
      "is_active": true,
      "is_featured": false,
      "is_customizable": false,
      "is_addon": false
    }
  ],
  "variations": [
    {
      "product_id": "UUID-PROD-1",
      "name": "5kg",
      "display_name": "5 kg Pack",
      "price": 455,
      "is_default": true
    },
    {
      "product_id": "UUID-PROD-1",
      "name": "1kg",
      "display_name": "1 kg Pack",
      "price": 110,
      "is_default": false
    }
  ]
}
```

## Expected Response

```json
{
  "status": "success",
  "data": {
    "products_created": 1,
    "products_updated": 0,
    "variations_processed": 2,
    "store_products_processed": 1,
    "taxes_processed": 1
  },
  "message": "Products pushed successfully"
}
```

## Benefits

1. **Simpler Payload**: No need to duplicate product data in store_products
2. **Less Redundancy**: Price and taxes defined once in products
3. **ERP Friendly**: Matches typical ERP data structure
4. **Backward Compatible**: Still supports explicit store_products if needed

## Migration from Old Format

### Old Format (Still Supported)
```json
{
  "products": [...],
  "store_products": [
    {
      "product_id": "PROD-001",
      "price": 455,
      "taxes": ["UUID-GST-5"]
    }
  ]
}
```

### New Format (Recommended)
```json
{
  "products": [
    {
      "id": "PROD-001",
      "price": 455,
      "taxes": ["UUID-GST-5"]
    }
  ]
}
```

Both formats work! Use whichever is more convenient for your ERP.
