# Design Document

## Overview

The Supabase-Redis Middleware is a Go-based application server built with the Gin framework that serves as a unified API gateway for multiple business domains. It implements a caching layer using Redis to optimize performance and reduce load on the Supabase backend. The architecture follows clean architecture principles with clear separation between HTTP handlers, business logic, data access, and caching layers.

## Architecture

### High-Level Architecture

```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │ HTTP/REST
       ▼
┌─────────────────────────────────────┐
│     Gin HTTP Server (Router)       │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│         Handler Layer               │
│  (supermarket, movie, pharmacy)     │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────┐
│        Service Layer                │
│   (Business Logic & Caching)        │
└──────┬──────────────────────────────┘
       │
       ├──────────────┬─────────────────┐
       ▼              ▼                 ▼
┌──────────┐   ┌──────────┐    ┌──────────┐
│  Redis   │   │ Supabase │    │  Logger  │
│  Cache   │   │  Client  │    │          │
└──────────┘   └──────────┘    └──────────┘
```

### Technology Stack

- **Framework**: Gin (github.com/gin-gonic/gin)
- **Supabase Client**: supabase-go (github.com/supabase-community/supabase-go)
- **Redis Client**: go-redis (github.com/redis/go-redis/v9)
- **Configuration**: viper (github.com/spf13/viper)
- **Logging**: zap (go.uber.org/zap)

## Components and Interfaces

### 1. Configuration Module

**Purpose**: Load and validate application configuration from environment variables or config files

**Structure**:
```go
type Config struct {
    Server   ServerConfig
    Supabase SupabaseConfig
    Redis    RedisConfig
}

type ServerConfig struct {
    Port           string
    ReadTimeout    time.Duration
    WriteTimeout   time.Duration
    RequestTimeout time.Duration
}

type SupabaseConfig struct {
    URL    string
    APIKey string
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
    TTL      time.Duration
}
```

### 2. Cache Layer

**Purpose**: Provide Redis caching functionality with automatic fallback

**Interface**:
```go
type CacheService interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    GenerateKey(domain string, params map[string]string) string
}
```

**Implementation Details**:
- Use Redis GET/SET commands with TTL
- Implement graceful degradation when Redis is unavailable
- Generate cache keys using format: `{domain}:{param1}:{param2}:...`
- Marshal/unmarshal data as JSON

### 3. Supabase Repository Layer

**Purpose**: Handle all interactions with Supabase database

**Interface**:
```go
type SupabaseRepository interface {
    Query(ctx context.Context, table string, filters map[string]interface{}, pagination Pagination) ([]map[string]interface{}, error)
    GetByID(ctx context.Context, table string, id string) (map[string]interface{}, error)
}

type Pagination struct {
    Limit  int
    Offset int
}
```

**Implementation Details**:
- Initialize Supabase client with URL and API key
- Support filtering with WHERE clauses
- Support pagination with LIMIT and OFFSET
- Handle connection errors and timeouts

### 4. Service Layer

**Purpose**: Implement business logic and orchestrate caching + data fetching

**Interface**:
```go
type DomainService interface {
    GetItems(ctx context.Context, filters map[string]string, pagination Pagination) (*Response, error)
    GetItemByID(ctx context.Context, id string) (*Response, error)
}
```

**Implementation Details**:
- Check cache first before querying Supabase
- On cache miss, fetch from Supabase and update cache
- Transform Supabase responses to standardized format
- Include cache status in response metadata

### 5. Handler Layer

**Purpose**: Handle HTTP requests and responses

**Handlers**:
- SupermarketHandler
- MovieHandler
- PharmacyHandler

**Common Handler Pattern**:
```go
func (h *Handler) GetItems(c *gin.Context) {
    // Parse query parameters
    // Call service layer
    // Return JSON response
}

func (h *Handler) GetItemByID(c *gin.Context) {
    // Parse path parameters
    // Call service layer
    // Return JSON response
}
```

### 6. Middleware Components

**Logging Middleware**: Log all requests with method, path, status, duration

**Error Recovery Middleware**: Catch panics and return 500 errors

**Timeout Middleware**: Enforce request timeout limits

**CORS Middleware**: Handle cross-origin requests

## Data Models

### Standard Response Format

```go
type Response struct {
    Status   string                 `json:"status"`
    Data     interface{}            `json:"data,omitempty"`
    Metadata ResponseMetadata       `json:"metadata,omitempty"`
    Error    *ErrorDetail           `json:"error,omitempty"`
}

type ResponseMetadata struct {
    CachedAt   *time.Time `json:"cached_at,omitempty"`
    FromCache  bool       `json:"from_cache"`
    Pagination *Pagination `json:"pagination,omitempty"`
}

type ErrorDetail struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

### Domain Models

Each business domain (supermarket, movie, pharmacy) will have its own data structures, but all will be returned through the standard Response format.

## API Endpoints

### Supermarket Domain
- `GET /api/v1/supermarket/products` - List products with filters
- `GET /api/v1/supermarket/products/:id` - Get product by ID
- `GET /api/v1/supermarket/categories` - List categories

### Movie Domain
- `GET /api/v1/movies` - List movies with filters
- `GET /api/v1/movies/:id` - Get movie by ID
- `GET /api/v1/movies/showtimes` - Get showtimes

### Pharmacy Domain
- `GET /api/v1/pharmacy/medicines` - List medicines with filters
- `GET /api/v1/pharmacy/medicines/:id` - Get medicine by ID
- `GET /api/v1/pharmacy/categories` - List categories

### Health Check
- `GET /health` - Health check endpoint

## Error Handling

### Error Categories

1. **Client Errors (4xx)**
   - 400 Bad Request: Invalid parameters
   - 404 Not Found: Resource not found
   - 422 Unprocessable Entity: Validation errors

2. **Server Errors (5xx)**
   - 500 Internal Server Error: Unexpected errors
   - 503 Service Unavailable: Supabase connection failure
   - 504 Gateway Timeout: Request timeout

### Error Response Format

All errors follow the standard Response format with populated Error field:

```json
{
  "status": "error",
  "error": {
    "code": "SUPABASE_CONNECTION_ERROR",
    "message": "Failed to connect to Supabase"
  }
}
```

### Logging Strategy

- **Info Level**: Successful requests, cache hits/misses
- **Warn Level**: Fallback scenarios (Redis unavailable)
- **Error Level**: Failed requests, connection errors
- **Debug Level**: Detailed request/response data (development only)

## Caching Strategy

### Cache Key Structure

Format: `{domain}:{operation}:{params_hash}`

Examples:
- `supermarket:products:category=dairy&limit=10`
- `movies:detail:id=123`
- `pharmacy:medicines:search=aspirin`

### TTL Configuration

- Product lists: 5 minutes
- Individual items: 15 minutes
- Categories: 30 minutes
- Search results: 2 minutes

### Cache Invalidation

- Automatic expiration via TTL
- Manual invalidation endpoints (future enhancement)
- Cache warming on startup (optional)

## Testing Strategy

### Unit Tests

- Test each service method independently
- Mock Redis and Supabase clients
- Test error handling paths
- Test cache key generation

### Integration Tests

- Test handler-to-service integration
- Test with real Redis (using testcontainers)
- Test with mock Supabase responses
- Test middleware chain

### End-to-End Tests

- Test complete request flow
- Test cache hit/miss scenarios
- Test error scenarios (Redis down, Supabase down)
- Test concurrent requests

## Configuration Management

### Environment Variables

```
# Server
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
REQUEST_TIMEOUT=30s

# Supabase
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_API_KEY=your-api-key

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_TTL=300s

# Logging
LOG_LEVEL=info
```

### Configuration File (config.yaml)

Alternative to environment variables for local development:

```yaml
server:
  port: "8080"
  read_timeout: "10s"
  write_timeout: "10s"
  request_timeout: "30s"

supabase:
  url: "https://your-project.supabase.co"
  api_key: "your-api-key"

redis:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0
  ttl: "300s"

logging:
  level: "info"
```

## Deployment Considerations

- Use Docker for containerization
- Support for health checks and readiness probes
- Graceful shutdown handling
- Connection pooling for Redis and Supabase
- Horizontal scaling support (stateless design)
