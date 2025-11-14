# Quick Start: Products Push API

## Prerequisites

1. PostgreSQL database running with the schema applied
2. Go server running on port 8080 (or your configured port)
3. Product matching engine migration applied

## Step 1: Apply Database Migration

```bash
psql -d middleware_db -f migrations/add_product_matching_engine.sql
```

## Step 2: Start the Server

```bash
go run cmd/server/main.go
```

## Step 3: Test the Endpoint

### Windows (PowerShell)
```powershell
.\docs\test-products-push.ps1
```

### Linux/Mac (Bash)
```bash
chmod +x docs/test-products-push.sh
./docs/test-products-push.sh
```

### Manual curl
```bash
curl -X POST http://localhost:8080/api/v1/products/push \
  -H "Content-Type: application/json" \
  -d @docs/API-PRODUCTS-PUSH-EXAMPLE.json
```

## Expected Response

```json
{
  "status": "success",
  "data": {
    "products_created": 2,
    "products_updated": 0,
    "variations_processed": 3,
    "store_products_processed": 2
  },
  "message": "Products pushed successfully"
}
```

## Verify in Database

```sql
-- Check stores
SELECT id, external_id, name FROM stores WHERE external_id = 'STORE-001';

-- Check products
SELECT id, sku, name, brand_id FROM products WHERE sku IN ('CK1L01', 'PP2L01');

-- Check store_product_mappings
SELECT * FROM store_product_mappings WHERE external_product_id IN ('ZOHO-1001', 'ZOHO-1002');

-- Check variations
SELECT pv.* FROM product_variations pv
JOIN products p ON p.id = pv.product_id
WHERE p.sku = 'CK1L01';

-- Check store_products
SELECT sp.* FROM store_products sp
JOIN products p ON p.id = sp.product_id
WHERE p.sku IN ('CK1L01', 'PP2L01');
```

## Common Issues

### Issue: "store not found"
**Solution**: Ensure store_details.store_id is provided in the request

### Issue: "category not found"
**Solution**: Either include categories in the request or use existing category IDs

### Issue: "tax not found"
**Solution**: Include taxes in the request before referencing them in store_products

### Issue: "duplicate key violation"
**Solution**: The product already exists - the API will update it automatically

## Next Steps

1. Review the full API documentation: `docs/API-PRODUCTS-PUSH.md`
2. Customize the example payload: `docs/API-PRODUCTS-PUSH-EXAMPLE.json`
3. Integrate with your ERP system
4. Monitor logs for matching confidence scores

## Support

For detailed information:
- API Reference: `docs/API-PRODUCTS-PUSH.md`
- Implementation Details: `docs/IMPLEMENTATION-SUMMARY.md`
- Database Schema: `grocery_superapp_schema.sql`
