# POST /api/v1/products/stock

## Overview

Bulk stock update endpoint for updating inventory levels across multiple products in a store.

## Features

✅ **Bulk Updates** - Update multiple products in one request  
✅ **Stock Management** - Update quantity and availability  
✅ **Price Updates** - Optionally update prices  
✅ **External ID Mapping** - Uses your ERP's product IDs  

## Request

### Endpoint
```
POST /api/v1/products/stock
Content-Type: application/json
```

### Request Body

```json
{
  "store_id": "UUID-STORE-1",
  "products": [
    {
      "id": "UUID-P1",
      "stock_quantity": 120,
      "is_available": true,
      "price": 455
    },
    {
      "id": "UUID-P2",
      "stock_quantity": 0,
      "is_available": false
    }
  ]
}
```

### Field Descriptions

#### store_id (required)
- **Type:** string
- **Description:** External store ID (from your ERP)
- **Example:** `"EXT123456"`, `"STORE-001"`

#### products (required)
Array of product stock updates

##### products[].id (required)
- **Type:** string
- **Description:** External product ID (matches `store_products.external_id`)
- **Example:** `"UUID-PROD-1"`, `"EXT-PROD-001"`

##### products[].stock_quantity (required)
- **Type:** number
- **Description:** Current stock quantity
- **Example:** `120`, `0`, `50.5`

##### products[].is_available (required)
- **Type:** boolean
- **Description:** Whether product is available for sale
- **Example:** `true`, `false`

##### products[].price (optional)
- **Type:** number
- **Description:** Update product price (optional)
- **Example:** `455`, `99.99`
- **Note:** If not provided or 0, price is not updated

## Response

### Success Response (200 OK)

```json
{
  "status": "success",
  "data": {
    "products_updated": 2,
    "products_not_found": 0
  },
  "message": "Stock updated successfully"
}
```

### Partial Success Response (200 OK)

When some products are not found:

```json
{
  "status": "success",
  "data": {
    "products_updated": 1,
    "products_not_found": 1
  },
  "message": "Stock updated successfully"
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
    "code": "STOCK_UPDATE_FAILED",
    "message": "Failed to update stock"
  }
}
```

## Behavior

### Stock Quantity
- Updates `store_products.stock_quantity`
- Automatically sets `is_in_stock = true` if quantity > 0
- Automatically sets `is_in_stock = false` if quantity = 0

### Availability
- Updates `store_products.is_available`
- Independent of stock quantity
- Use to temporarily disable products

### Price Updates
- Only updates if `price` field is provided and > 0
- Omit field or set to 0 to skip price update

### Product Matching
- Matches products by `store_products.external_id`
- Must match the product ID used in `/products/push`
- Products not found are counted but don't cause errors

## Examples

### Example 1: Simple Stock Update

```bash
curl -X POST http://localhost:8080/api/v1/products/stock \
  -H "Content-Type: application/json" \
  -d '{
    "store_id": "STORE-001",
    "products": [
      {
        "id": "PROD-001",
        "stock_quantity": 100,
        "is_available": true
      }
    ]
  }'
```

### Example 2: Update Stock and Price

```json
{
  "store_id": "STORE-001",
  "products": [
    {
      "id": "PROD-001",
      "stock_quantity": 50,
      "is_available": true,
      "price": 499
    }
  ]
}
```

### Example 3: Mark Out of Stock

```json
{
  "store_id": "STORE-001",
  "products": [
    {
      "id": "PROD-001",
      "stock_quantity": 0,
      "is_available": false
    }
  ]
}
```

### Example 4: Bulk Update

```json
{
  "store_id": "STORE-001",
  "products": [
    {
      "id": "PROD-001",
      "stock_quantity": 120,
      "is_available": true
    },
    {
      "id": "PROD-002",
      "stock_quantity": 0,
      "is_available": false
    },
    {
      "id": "PROD-003",
      "stock_quantity": 75,
      "is_available": true,
      "price": 299
    }
  ]
}
```

## Use Cases

### 1. Inventory Sync
Sync stock levels from your ERP to the platform:
```json
{
  "store_id": "STORE-001",
  "products": [
    {"id": "P1", "stock_quantity": 100, "is_available": true},
    {"id": "P2", "stock_quantity": 50, "is_available": true},
    {"id": "P3", "stock_quantity": 0, "is_available": false}
  ]
}
```

### 2. Price Updates
Update prices during promotions:
```json
{
  "store_id": "STORE-001",
  "products": [
    {"id": "P1", "stock_quantity": 100, "is_available": true, "price": 399}
  ]
}
```

### 3. Disable Products
Temporarily disable products without changing stock:
```json
{
  "store_id": "STORE-001",
  "products": [
    {"id": "P1", "stock_quantity": 100, "is_available": false}
  ]
}
```

## Best Practices

1. **Batch Updates**: Send multiple products in one request (up to 1000)
2. **Regular Sync**: Schedule periodic stock syncs (e.g., every 5 minutes)
3. **Error Handling**: Check `products_not_found` count
4. **Price Updates**: Only include `price` when needed to reduce payload size
5. **Availability**: Use `is_available=false` for temporary disabling

## Limitations

- Maximum 1000 products per request
- Products must exist in `store_products` table
- Uses `external_id` for matching (set during product push)

## Database Updates

The endpoint updates the following fields in `store_products`:

```sql
UPDATE store_products
SET stock_quantity = ?,
    is_in_stock = CASE WHEN ? > 0 THEN true ELSE false END,
    is_available = ?,
    price = ?,  -- Optional
    updated_at = CURRENT_TIMESTAMP
WHERE store_id = ? AND external_id = ?
```

## Integration Example

### Node.js
```javascript
const updateStock = async (storeId, products) => {
  const response = await fetch('http://localhost:8080/api/v1/products/stock', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ store_id: storeId, products })
  });
  return response.json();
};

// Usage
await updateStock('STORE-001', [
  { id: 'PROD-001', stock_quantity: 100, is_available: true }
]);
```

### Python
```python
import requests

def update_stock(store_id, products):
    response = requests.post(
        'http://localhost:8080/api/v1/products/stock',
        json={'store_id': store_id, 'products': products}
    )
    return response.json()

# Usage
update_stock('STORE-001', [
    {'id': 'PROD-001', 'stock_quantity': 100, 'is_available': True}
])
```

## Troubleshooting

### Issue: products_not_found > 0

**Cause:** Product external_id doesn't exist in store_products

**Solution:** 
1. Ensure product was pushed via `/products/push` first
2. Check external_id matches exactly
3. Verify store_id is correct

### Issue: Stock not updating

**Cause:** Wrong store_id or external_id

**Solution:**
```sql
-- Check if product exists
SELECT * FROM store_products 
WHERE external_id = 'PROD-001';

-- Check store mapping
SELECT * FROM stores WHERE external_id = 'STORE-001';
```

## See Also

- [Products Push API](./API-PRODUCTS-PUSH.md)
- [Simplified API Structure](./SIMPLIFIED-API-STRUCTURE.md)
- [Store Update Feature](./STORE-UPDATE-FEATURE.md)
