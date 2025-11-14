# Gol Bazaar API Endpoints

## Base URL
```
http://localhost:8080/api/v1
```

## Response Format

All API responses follow this structure:

**Success Response:**
```json
{
  "status": "success",
  "data": { ... }
}
```

**Error Response:**
```json
{
  "status": "error",
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

## Store Management

### Get Store Basic Data

**Endpoint:** `GET /api/v1/stores/:id`

**Description:** Retrieve basic information about a store.

**Example:**
```bash
curl http://localhost:8080/api/v1/stores/123e4567-e89b-12d3-a456-426614174000
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Fresh Mart Downtown",
    "slug": "fresh-mart-downtown",
    "description": "Your neighborhood grocery store",
    "store_type": "supermarket",
    "phone": "+1-555-0123",
    "email": "contact@freshmart.com",
    "address_line1": "123 Main Street",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "USA",
    "latitude": 40.7128,
    "longitude": -74.0060,
    "rating": 4.5,
    "total_ratings": 150,
    "min_order_amount": 25.00,
    "delivery_fee": 5.99,
    "estimated_delivery_time": 30,
    "is_active": true,
    "is_open": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Update Store Details

**Endpoint:** `PUT /api/v1/stores/:id`

**Description:** Update store information including name, contact details, address, and delivery settings.

**Request Body:**
```json
{
  "name": "Fresh Mart Downtown - Updated",
  "description": "Your premium neighborhood grocery store",
  "phone": "+1-555-9999",
  "email": "newemail@freshmart.com",
  "address_line1": "456 Main Street",
  "address_line2": "Suite 100",
  "city": "New York",
  "state": "NY",
  "postal_code": "10002",
  "country": "USA",
  "min_order_amount": 30.00,
  "delivery_fee": 4.99,
  "estimated_delivery_time": 25
}
```

**Note:** All fields are optional. Only include the fields you want to update.

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/stores/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fresh Mart Downtown - Updated",
    "phone": "+1-555-9999",
    "min_order_amount": 30.00,
    "delivery_fee": 4.99
  }'
```

**Response:**
```json
{
  "status": "success",
  "message": "Store details updated successfully"
}
```

### Update Store Status

**Endpoint:** `PUT /api/v1/stores/:id/status`

**Description:** Update store active and/or open status.

**Request Body:**
```json
{
  "is_active": true,
  "is_open": false
}
```

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/stores/123e4567-e89b-12d3-a456-426614174000/status \
  -H "Content-Type: application/json" \
  -d '{"is_active": true, "is_open": false}'
```

**Response:**
```json
{
  "status": "success",
  "message": "Store status updated successfully"
}
```

### Get Store Status

**Endpoint:** `GET /api/v1/stores/:id/status`

**Description:** Get current store status information.

**Example:**
```bash
curl http://localhost:8080/api/v1/stores/123e4567-e89b-12d3-a456-426614174000/status
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Fresh Mart Downtown",
    "is_active": true,
    "is_open": false,
    "is_verified": true,
    "opened_at": "08:00:00",
    "closed_at": "22:00:00",
    "updated_at": "2024-01-15T15:30:00Z"
  }
}
```

## Product Management

### Bulk Create Products

**Endpoint:** `POST /api/v1/products/bulk`

**Description:** Create multiple products in a single request (max 100).

**Request Body:**
```json
{
  "products": [
    {
      "sku": "MILK-001",
      "name": "Organic Whole Milk",
      "description": "Fresh organic whole milk, 1 gallon",
      "category_id": "dairy-category-uuid",
      "base_price": 4.99,
      "sale_price": 3.99,
      "unit": "gallon",
      "unit_quantity": 1.0,
      "brand": "Organic Valley",
      "is_active": true,
      "requires_prescription": false
    }
  ]
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/products/bulk \
  -H "Content-Type: application/json" \
  -d @products.json
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "created_count": 1,
    "products": [
      {
        "id": "prod-uuid-1",
        "sku": "MILK-001",
        "name": "Organic Whole Milk",
        "base_price": 4.99,
        "is_active": true,
        "created_at": "2024-01-15T16:00:00Z"
      }
    ]
  }
}
```

### Update Product Stock

**Endpoint:** `PUT /api/v1/products/:id/stock`

**Description:** Update stock quantity for a specific product.

**Request Body:**
```json
{
  "stock_quantity": 50.0
}
```

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/products/prod-uuid-1/stock \
  -H "Content-Type: application/json" \
  -d '{"stock_quantity": 50.0}'
```

**Response:**
```json
{
  "status": "success",
  "message": "Product stock updated successfully"
}
```

### Update Product Status

**Endpoint:** `PUT /api/v1/products/:id/status`

**Description:** Update product active status.

**Request Body:**
```json
{
  "is_active": false
}
```

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/products/prod-uuid-1/status \
  -H "Content-Type: application/json" \
  -d '{"is_active": false}'
```

**Response:**
```json
{
  "status": "success",
  "message": "Product status updated successfully"
}
```

### Bulk Update Product Stock

**Endpoint:** `PUT /api/v1/products/stock/bulk`

**Description:** Update stock quantities for multiple products (max 100).

**Request Body:**
```json
{
  "updates": [
    {
      "product_id": "prod-uuid-1",
      "stock_quantity": 25.0
    },
    {
      "product_id": "prod-uuid-2",
      "stock_quantity": 100.0
    }
  ]
}
```

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/products/stock/bulk \
  -H "Content-Type: application/json" \
  -d @stock-updates.json
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "updated_count": 2
  }
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_INPUT` | 400 | Request body validation failed |
| `STORE_NOT_FOUND` | 404 | Store with given ID not found |
| `PRODUCT_NOT_FOUND` | 404 | Product with given ID not found |
| `CREATION_FAILED` | 500 | Failed to create products |
| `UPDATE_FAILED` | 500 | Failed to update resource |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

## Validation Rules

### Store Updates
- `name`: Max 255 characters
- `email`: Must be valid email format
- `phone`: Max 20 characters
- `min_order_amount`: Must be >= 0
- `delivery_fee`: Must be >= 0
- `estimated_delivery_time`: Must be > 0 (in minutes)

### Product Creation
- `sku`: Required, must be unique
- `name`: Required, max 255 characters
- `base_price`: Required, must be >= 0
- `sale_price`: Optional, must be >= 0 if provided
- `unit_quantity`: Must be >= 0
- Maximum 100 products per bulk request

### Stock Updates
- `stock_quantity`: Required, must be >= 0
- Maximum 100 updates per bulk request

## Complete Workflow Examples

### Store Management Workflow

```bash
# 1. Get store information
curl http://localhost:8080/api/v1/stores/STORE_ID

# 2. Update store details
curl -X PUT http://localhost:8080/api/v1/stores/STORE_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Store Name",
    "phone": "+1-555-1234",
    "delivery_fee": 3.99
  }'

# 3. Close store temporarily
curl -X PUT http://localhost:8080/api/v1/stores/STORE_ID/status \
  -H "Content-Type: application/json" \
  -d '{"is_open": false}'

# 4. Check store status
curl http://localhost:8080/api/v1/stores/STORE_ID/status

# 5. Reopen store
curl -X PUT http://localhost:8080/api/v1/stores/STORE_ID/status \
  -H "Content-Type: application/json" \
  -d '{"is_open": true}'
```

---

**All endpoints are ready for testing!** 🚀
