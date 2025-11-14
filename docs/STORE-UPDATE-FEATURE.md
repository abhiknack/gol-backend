# Store Update Details Feature

## Overview
Added a new endpoint to update store details including name, contact information, address, and delivery settings.

## New Endpoint

**PUT /api/v1/stores/:id**

Updates store information with flexible field selection - only include the fields you want to update.

## Supported Fields

- `name` - Store name (also updates slug automatically)
- `description` - Store description
- `phone` - Contact phone number
- `email` - Contact email
- `address_line1` - Primary address
- `address_line2` - Secondary address (suite, apt, etc.)
- `city` - City
- `state` - State/Province
- `postal_code` - ZIP/Postal code
- `country` - Country
- `min_order_amount` - Minimum order amount (decimal)
- `delivery_fee` - Delivery fee (decimal)
- `estimated_delivery_time` - Estimated delivery time in minutes (integer)

## Implementation Details

### Files Modified

1. **internal/handlers/store_handler.go**
   - Added `UpdateStoreDetails()` handler method

2. **internal/repository/postgres.go**
   - Added `UpdateStoreDetailsInput` struct
   - Added `UpdateStoreDetails()` repository method
   - Dynamically builds SQL query based on provided fields
   - Auto-generates slug when name is updated

3. **internal/router/router.go**
   - Added route: `stores.PUT("/:id", storeHandler.UpdateStoreDetails)`

### Key Features

- **Partial Updates**: Only update the fields you provide
- **Automatic Slug Generation**: When name is updated, slug is regenerated
- **Validation**: Returns error if no fields are provided
- **Transaction Safety**: Uses proper SQL parameter binding
- **Logging**: Logs number of fields updated

## Usage Example

```bash
# Update only name and phone
curl -X PUT http://localhost:8080/api/v1/stores/STORE_ID \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Store Name",
    "phone": "+1-555-9999"
  }'

# Update delivery settings
curl -X PUT http://localhost:8080/api/v1/stores/STORE_ID \
  -H "Content-Type: application/json" \
  -d '{
    "min_order_amount": 25.00,
    "delivery_fee": 3.99,
    "estimated_delivery_time": 30
  }'

# Update full address
curl -X PUT http://localhost:8080/api/v1/stores/STORE_ID \
  -H "Content-Type: application/json" \
  -d '{
    "address_line1": "123 New Street",
    "address_line2": "Suite 200",
    "city": "Boston",
    "state": "MA",
    "postal_code": "02101",
    "country": "USA"
  }'
```

## Response

**Success:**
```json
{
  "status": "success",
  "message": "Store details updated successfully"
}
```

**Error - No fields provided:**
```json
{
  "status": "error",
  "error": {
    "code": "INVALID_INPUT",
    "message": "no fields to update"
  }
}
```

**Error - Store not found:**
```json
{
  "status": "error",
  "error": {
    "code": "UPDATE_FAILED",
    "message": "Failed to update store details"
  }
}
```

## Testing

To test the endpoint:

1. Get a valid store ID from the database
2. Use curl or Postman to send a PUT request
3. Verify the response
4. Check the database to confirm updates

```bash
# Get store details first
curl http://localhost:8080/api/v1/stores/STORE_ID

# Update some fields
curl -X PUT http://localhost:8080/api/v1/stores/STORE_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Name", "phone": "+1-555-1234"}'

# Verify the update
curl http://localhost:8080/api/v1/stores/STORE_ID
```

## Related Endpoints

- `GET /api/v1/stores/:id` - Get store details
- `PUT /api/v1/stores/:id/status` - Update store status (is_active, is_open)
- `GET /api/v1/stores/:id/status` - Get store status

## Notes

- All fields are optional - provide only what you want to update
- The `updated_at` timestamp is automatically set
- When updating `name`, the `slug` is automatically regenerated
- Numeric fields (prices, delivery time) are validated for proper types
- The endpoint returns 500 if the store is not found
