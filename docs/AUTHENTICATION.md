# API Authentication

## Overview
The API uses Bearer token authentication to secure protected endpoints. All store and product management endpoints require a valid Bearer token in the Authorization header.

## Configuration

### Environment Variable
Add your bearer tokens to the `.env` file:

```bash
# Single token
SERVER_BEARER_TOKENS=your-secret-token-here

# Multiple tokens (comma-separated)
SERVER_BEARER_TOKENS=token1,token2,token3
```

### YAML Configuration
Alternatively, configure tokens in `config.yaml`:

```yaml
server:
  port: "8080"
  read_timeout: "10s"
  write_timeout: "10s"
  request_timeout: "30s"
  bearer_tokens:
    - "your-secret-token-here"
    - "another-token-for-testing"
```

## Usage

### Making Authenticated Requests

Include the Bearer token in the Authorization header:

```bash
curl -X GET "http://localhost:8080/api/v1/stores/123" \
  -H "Authorization: Bearer your-secret-token-here"
```

### Example with Product Push

```bash
curl -X POST "http://localhost:8080/api/v1/products/push" \
  -H "Authorization: Bearer your-secret-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "store_details": {
      "store_id": "uuid",
      "name": "Store Name",
      "address": {...},
      "location": {...}
    },
    "products": [...]
  }'
```

## Protected Endpoints

The following endpoints require Bearer authentication:

### Store Management
- `GET /api/v1/stores/:id` - Get store details
- `PUT /api/v1/stores/:id` - Update store details
- `PUT /api/v1/stores/:id/status` - Update store status
- `GET /api/v1/stores/:id/status` - Get store status

### Product Management
- `POST /api/v1/products/push` - Bulk push products

## Public Endpoints

The following endpoints do NOT require authentication:

- `GET /health` - Health check
- `GET /api/v1/supermarket/*` - Supermarket domain routes
- `GET /api/v1/movies/*` - Movie domain routes
- `GET /api/v1/pharmacy/*` - Pharmacy domain routes

## Error Responses

### Missing Authorization Header
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing authorization header"
  }
}
```
**HTTP Status:** 401 Unauthorized

### Invalid Authorization Format
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid authorization format. Expected: Bearer <token>"
  }
}
```
**HTTP Status:** 401 Unauthorized

### Empty Bearer Token
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Empty bearer token"
  }
}
```
**HTTP Status:** 401 Unauthorized

### Invalid Bearer Token
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid bearer token"
  }
}
```
**HTTP Status:** 401 Unauthorized

## Security Best Practices

1. **Use Strong Tokens**: Generate cryptographically secure random tokens
   ```bash
   # Generate a secure token (Linux/Mac)
   openssl rand -hex 32
   
   # Or use UUID
   uuidgen
   ```

2. **Keep Tokens Secret**: Never commit tokens to version control
   - Use `.env` files (already in `.gitignore`)
   - Use environment variables in production
   - Rotate tokens regularly

3. **Use HTTPS**: Always use HTTPS in production to prevent token interception

4. **Token Rotation**: Regularly rotate bearer tokens and update clients

5. **Multiple Tokens**: Use different tokens for different clients/services for better tracking and revocation

## Testing

### Test with curl
```bash
# Valid token
curl -X GET "http://localhost:8080/api/v1/stores/123" \
  -H "Authorization: Bearer your-secret-token-here"

# Missing token (should fail)
curl -X GET "http://localhost:8080/api/v1/stores/123"

# Invalid token (should fail)
curl -X GET "http://localhost:8080/api/v1/stores/123" \
  -H "Authorization: Bearer invalid-token"
```

### Test with Postman
1. Create a new request
2. Go to the "Authorization" tab
3. Select "Bearer Token" from the Type dropdown
4. Enter your token in the Token field
5. Send the request

## Logging

Authentication attempts are logged with the following information:
- Client IP address
- Request path
- Authentication result (success/failure)
- Failure reason (if applicable)

Example log entries:
```
WARN  missing authorization header  path=/api/v1/stores/123  client_ip=192.168.1.1
WARN  invalid bearer token  path=/api/v1/stores/123  client_ip=192.168.1.1
DEBUG bearer token validated  path=/api/v1/stores/123  client_ip=192.168.1.1
```
