# Product Management Endpoints

## Overview
Product management endpoints for bulk product operations including categories, taxes, products, variations, and store inventory.

**Authentication Required:** All product endpoints require Bearer token authentication. See [AUTHENTICATION.md](./AUTHENTICATION.md) for details.

## Endpoints

### POST /api/v1/products/push
Bulk upsert products with related data (categories, taxes, variations, store products).

**Authentication:** Required (Bearer token)

**Request Body:**
```json
{
  "store_details": {
    "store_id": "uuid",
    "name": "Store Name",
    "address": {
      "line1": "123 Main St",
      "city": "City",
      "state": "State",
      "postal_code": "12345"
    },
    "location": {
      "lat": 40.7128,
      "lng": -74.0060
    }
  },
  "categories": [
    {
      "id": "uuid",
      "parent_id": "uuid or null",
      "name": "Category Name",
      "slug": "category-slug",
      "description": "Description",
      "display_order": 1,
      "is_active": true
    }
  ],
  "taxes": [
    {
      "id": "uuid",
      "name": "Tax Name",
      "tax_id": "TAX001",
      "description": "Tax description",
      "rate": 0.08,
      "tax_type": "percentage",
      "is_inclusive": false,
      "is_active": true
    }
  ],
  "products": [
    {
      "id": "uuid",
      "external_id": "EXT123",
      "sku": "SKU123",
      "name": "Product Name",
      "slug": "product-slug",
      "description": "Product description",
      "category_id": "uuid",
      "price": 9.99,
      "currency": "USD",
      "unit": "kg",
      "unit_quantity": 1.0,
      "primary_image_url": "https://...",
      "images": ["https://...", "https://..."],
      "brand": "Brand Name",
      "manufacturer": "Manufacturer",
      "barcode": "123456789",
      "ean": "1234567890123",
      "taxes": ["tax-uuid-1", "tax-uuid-2"],
      "is_active": true,
      "is_featured": false,
      "is_customizable": true,
      "is_addon": false
    }
  ],
  "variations": [
    {
      "product_id": "uuid",
      "name": "small",
      "display_name": "Small",
      "price": 8.99,
      "is_default": false
    }
  ],
  "store_products": [
    {
      "product_id": "uuid",
      "store_id": "uuid",
      "price": 9.99,
      "stock_quantity": 100,
      "is_in_stock": true
    }
  ]
}
```

**Response (Success):**
```json
{
  "status": "success",
  "data": {
    "products_created": 5,
    "products_updated": 3,
    "variations_processed": 12,
    "store_products_processed": 8
  },
  "message": "Products pushed successfully"
}
```

**Response (Error):**
```json
{
  "status": "error",
  "error": {
    "code": "PRODUCT_UPSERT_FAILED",
    "message": "Failed to create or update products"
  }
}
```

## Features

### Upsert Operations
- All operations use INSERT ... ON CONFLICT DO UPDATE
- Creates new records or updates existing ones based on ID
- Maintains referential integrity

### Transaction Safety
- All operations within a single request are wrapped in a transaction
- If any operation fails, all changes are rolled back

### Store Management
- Automatically creates or updates store information
- Supports PostGIS location data for geographic queries

### Product Images
- Supports multiple images per product
- Automatically sets display order
- Marks first image as primary

### Variations
- Product variations (sizes, flavors, etc.)
- Each variation has its own price
- Can mark one as default

### Store Products
- Links products to specific stores
- Store-specific pricing and inventory
- Stock tracking per store

## Error Codes

- `INVALID_INPUT` - Request validation failed
- `STORE_UPSERT_FAILED` - Failed to create/update store
- `CATEGORY_UPSERT_FAILED` - Failed to create/update categories
- `TAX_UPSERT_FAILED` - Failed to create/update taxes
- `PRODUCT_UPSERT_FAILED` - Failed to create/update products

## Notes

- All IDs should be UUIDs
- Store location uses PostGIS GEOGRAPHY type
- Products can be both regular products and addons (is_addon flag)
- Variations are optional but recommended for products with multiple sizes/options
- Store products are required to make products available in stores
