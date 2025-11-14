# Quick Start: Bearer Authentication

## Setup (30 seconds)

### 1. Generate a Token
```bash
# Option 1: Use OpenSSL (recommended)
openssl rand -hex 32

# Option 2: Use any random string
echo "my-super-secret-token-$(date +%s)"
```

### 2. Add to .env File
```bash
# Copy the example file if you haven't already
cp .env.example .env

# Edit .env and add your token
SERVER_BEARER_TOKENS=your-generated-token-here
```

### 3. Start the Server
```bash
go run cmd/server/main.go
```

## Test It

### Test Protected Endpoint (Should Fail)
```bash
curl http://localhost:8080/api/v1/stores/test-id
```

Expected response:
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing authorization header"
  }
}
```

### Test with Token (Should Work)
```bash
curl -H "Authorization: Bearer your-generated-token-here" \
     http://localhost:8080/api/v1/stores/test-id
```

### Test Product Push
```bash
curl -X POST http://localhost:8080/api/v1/products/push \
  -H "Authorization: Bearer your-generated-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "store_details": {
      "store_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Test Store",
      "address": {
        "line1": "123 Main St",
        "city": "New York",
        "state": "NY",
        "postal_code": "10001"
      },
      "location": {
        "lat": 40.7128,
        "lng": -74.0060
      }
    },
    "products": [
      {
        "id": "650e8400-e29b-41d4-a716-446655440001",
        "external_id": "EXT001",
        "sku": "SKU001",
        "name": "Test Product",
        "slug": "test-product",
        "description": "A test product",
        "category_id": "",
        "price": 9.99,
        "currency": "USD",
        "unit": "piece",
        "unit_quantity": 1,
        "primary_image_url": "https://example.com/image.jpg",
        "images": [],
        "brand": "Test Brand",
        "manufacturer": "Test Manufacturer",
        "barcode": "123456789",
        "ean": "1234567890123",
        "taxes": [],
        "is_active": true,
        "is_featured": false,
        "is_customizable": false,
        "is_addon": false
      }
    ],
    "variations": [],
    "store_products": [
      {
        "product_id": "650e8400-e29b-41d4-a716-446655440001",
        "store_id": "550e8400-e29b-41d4-a716-446655440000",
        "price": 9.99,
        "stock_quantity": 100,
        "is_in_stock": true
      }
    ]
  }'
```

## Multiple Tokens

For multiple clients/services:

```bash
# In .env
SERVER_BEARER_TOKENS=token-for-client-1,token-for-client-2,token-for-admin
```

Or in `config.yaml`:
```yaml
server:
  bearer_tokens:
    - "token-for-client-1"
    - "token-for-client-2"
    - "token-for-admin"
```

## Public Endpoints (No Auth Required)

These endpoints work without authentication:
- `GET /health`
- `GET /api/v1/supermarket/*`
- `GET /api/v1/movies/*`
- `GET /api/v1/pharmacy/*`

## Troubleshooting

### "Missing authorization header"
- Add the header: `-H "Authorization: Bearer your-token"`

### "Invalid authorization format"
- Make sure format is: `Bearer <token>` (with space)
- Don't use quotes around the token in the header

### "Invalid bearer token"
- Check your token matches what's in `.env` or `config.yaml`
- No extra spaces or newlines in the token
- Token is case-sensitive

## Next Steps

- Read [AUTHENTICATION.md](./AUTHENTICATION.md) for full details
- Read [PRODUCT-ENDPOINTS.md](./PRODUCT-ENDPOINTS.md) for API documentation
- Read [STORE-UPDATE-FEATURE.md](./STORE-UPDATE-FEATURE.md) for store management
