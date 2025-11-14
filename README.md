# Supabase-Redis Middleware

A high-performance Go-based middleware application server built with the Gin framework that serves as a unified API gateway for multiple business domains. It implements intelligent caching using Redis to optimize performance and reduce load on the Supabase backend.

## Features

- **Unified API Gateway**: Single interface for multiple business domains (supermarket, movies, pharmacy)
- **Intelligent Caching**: Redis-based caching layer with automatic fallback
- **Clean Architecture**: Clear separation of concerns with handler, service, repository, and cache layers
- **Graceful Degradation**: Continues operation even when Redis is unavailable
- **Comprehensive Logging**: Structured logging with configurable levels using Zap
- **Health Monitoring**: Built-in health check endpoints for Redis and Supabase connectivity
- **Docker Support**: Containerized deployment with Docker Compose
- **Configurable**: Environment variables and YAML configuration support

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.23+**: [Download Go](https://golang.org/dl/)
- **Redis**: [Install Redis](https://redis.io/download) or use Docker
- **Supabase Account**: [Sign up for Supabase](https://supabase.com)
- **Docker & Docker Compose** (optional, for containerized deployment): [Install Docker](https://docs.docker.com/get-docker/)

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/supabase-redis-middleware.git
cd supabase-redis-middleware
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment Variables

Create a `.env` file in the project root by copying the example:

```bash
cp .env.example .env
```

Edit the `.env` file with your configuration:

```env
# Server Configuration
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
REQUEST_TIMEOUT=30s

# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_API_KEY=your-supabase-api-key-here

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_TTL=300s

# Logging Configuration
LOG_LEVEL=info
```

**Important**: Replace `SUPABASE_URL` and `SUPABASE_API_KEY` with your actual Supabase project credentials.

### 4. Alternative: YAML Configuration

You can also use a YAML configuration file:

```bash
cp config.yaml.example config.yaml
```

Edit `config.yaml` with your settings. Note that environment variables take precedence over YAML configuration.

## Running the Application

### Local Development

#### Option 1: Using Go directly

1. Ensure Redis is running locally:
```bash
redis-server
```

2. Run the application:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080` (or the port specified in your configuration).

#### Option 2: Using Docker Compose (Recommended)

This method automatically starts both the application and Redis:

```bash
docker-compose up --build
```

To run in detached mode:
```bash
docker-compose up -d --build
```

To stop the services:
```bash
docker-compose down
```

### Building for Production

Build the binary:

```bash
go build -o bin/server cmd/server/main.go
```

Run the binary:

```bash
./bin/server
```

## API Documentation

### Base URL

```
http://localhost:8080/api/v1
```

### Response Format

All API responses follow a consistent structure:

**Success Response:**
```json
{
  "status": "success",
  "data": { ... },
  "metadata": {
    "from_cache": true,
    "cached_at": "2024-01-15T10:30:00Z",
    "pagination": {
      "limit": 10,
      "offset": 0
    }
  }
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

### Endpoints

#### Health Check

Check the health status of the application and its dependencies.

**Endpoint:** `GET /health`

**Example:**
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "dependencies": {
    "redis": {
      "status": "healthy"
    },
    "supabase": {
      "status": "healthy"
    }
  }
}
```

---

#### Supermarket Domain

##### Get Products

Retrieve a list of supermarket products with optional filtering and pagination.

**Endpoint:** `GET /api/v1/supermarket/products`

**Query Parameters:**
- `category` (optional): Filter by product category
- `limit` (optional): Number of items to return (default: 10)
- `offset` (optional): Number of items to skip (default: 0)

**Example:**
```bash
curl "http://localhost:8080/api/v1/supermarket/products?category=dairy&limit=20"
```

**Response:**
```json
{
  "status": "success",
  "data": [
    {
      "id": "1",
      "name": "Milk",
      "category": "dairy",
      "price": 3.99
    }
  ],
  "metadata": {
    "from_cache": false,
    "pagination": {
      "limit": 20,
      "offset": 0
    }
  }
}
```

##### Get Product by ID

Retrieve a specific product by its ID.

**Endpoint:** `GET /api/v1/supermarket/products/:id`

**Example:**
```bash
curl http://localhost:8080/api/v1/supermarket/products/123
```

##### Get Categories

Retrieve all supermarket product categories.

**Endpoint:** `GET /api/v1/supermarket/categories`

**Example:**
```bash
curl http://localhost:8080/api/v1/supermarket/categories
```

---

#### Movie Domain

##### Get Movies

Retrieve a list of movies with optional filtering and pagination.

**Endpoint:** `GET /api/v1/movies`

**Query Parameters:**
- `genre` (optional): Filter by movie genre
- `limit` (optional): Number of items to return
- `offset` (optional): Number of items to skip

**Example:**
```bash
curl "http://localhost:8080/api/v1/movies?genre=action&limit=10"
```

##### Get Movie by ID

Retrieve a specific movie by its ID.

**Endpoint:** `GET /api/v1/movies/:id`

**Example:**
```bash
curl http://localhost:8080/api/v1/movies/456
```

##### Get Showtimes

Retrieve movie showtimes.

**Endpoint:** `GET /api/v1/movies/showtimes`

**Query Parameters:**
- `date` (optional): Filter by date
- `theater` (optional): Filter by theater

**Example:**
```bash
curl "http://localhost:8080/api/v1/movies/showtimes?date=2024-01-15"
```

---

#### Pharmacy Domain

##### Get Medicines

Retrieve a list of medicines with optional filtering and pagination.

**Endpoint:** `GET /api/v1/pharmacy/medicines`

**Query Parameters:**
- `category` (optional): Filter by medicine category
- `search` (optional): Search by medicine name
- `limit` (optional): Number of items to return
- `offset` (optional): Number of items to skip

**Example:**
```bash
curl "http://localhost:8080/api/v1/pharmacy/medicines?category=pain-relief&limit=15"
```

##### Get Medicine by ID

Retrieve a specific medicine by its ID.

**Endpoint:** `GET /api/v1/pharmacy/medicines/:id`

**Example:**
```bash
curl http://localhost:8080/api/v1/pharmacy/medicines/789
```

##### Get Pharmacy Categories

Retrieve all pharmacy medicine categories.

**Endpoint:** `GET /api/v1/pharmacy/categories`

**Example:**
```bash
curl http://localhost:8080/api/v1/pharmacy/categories
```

---

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NOT_FOUND` | 404 | The requested resource or endpoint does not exist |
| `NOT_IMPLEMENTED` | 501 | The endpoint exists but is not yet implemented |
| `SUPABASE_CONNECTION_ERROR` | 503 | Failed to connect to Supabase |
| `REDIS_ERROR` | 500 | Redis operation failed (with fallback) |
| `TIMEOUT` | 504 | Request exceeded timeout duration |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

## Configuration Reference

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `SERVER_PORT` | No | `8080` | Port on which the server listens |
| `SERVER_READ_TIMEOUT` | No | `10s` | Maximum duration for reading the entire request |
| `SERVER_WRITE_TIMEOUT` | No | `10s` | Maximum duration before timing out writes |
| `REQUEST_TIMEOUT` | No | `30s` | Maximum duration for processing a request |
| `SUPABASE_URL` | **Yes** | - | Your Supabase project URL |
| `SUPABASE_API_KEY` | **Yes** | - | Your Supabase API key (anon/public key) |
| `REDIS_HOST` | No | `localhost` | Redis server hostname |
| `REDIS_PORT` | No | `6379` | Redis server port |
| `REDIS_PASSWORD` | No | - | Redis password (if authentication is enabled) |
| `REDIS_DB` | No | `0` | Redis database number (0-15) |
| `REDIS_TTL` | No | `300s` | Cache time-to-live duration |
| `LOG_LEVEL` | No | `info` | Logging level: `debug`, `info`, `warn`, `error` |

### Cache TTL Guidelines

Recommended TTL values based on data volatility:

- **Product lists**: 5 minutes (`300s`)
- **Individual items**: 15 minutes (`900s`)
- **Categories**: 30 minutes (`1800s`)
- **Search results**: 2 minutes (`120s`)

## Docker Deployment

### Using Docker Compose

The easiest way to deploy is using Docker Compose, which handles both the application and Redis:

1. Ensure your `.env` file is configured with your Supabase credentials

2. Build and start the services:
```bash
docker-compose up -d --build
```

3. Check the logs:
```bash
docker-compose logs -f app
```

4. Stop the services:
```bash
docker-compose down
```

### Using Docker Only

Build the Docker image:

```bash
docker build -t supabase-redis-middleware .
```

Run the container:

```bash
docker run -d \
  --name middleware \
  -p 8080:8080 \
  -e SUPABASE_URL=https://your-project.supabase.co \
  -e SUPABASE_API_KEY=your-api-key \
  -e REDIS_HOST=redis \
  supabase-redis-middleware
```

**Note**: You'll need to run a separate Redis container and configure networking between containers.

## Testing

### Run Unit Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

### Run Integration Tests

```bash
go test ./tests/...
```

Integration tests require Redis to be running. You can use Docker:

```bash
docker run -d -p 6379:6379 redis:7-alpine
go test ./tests/...
```

## Troubleshooting

### Application won't start

**Problem**: Error message "SUPABASE_URL and SUPABASE_API_KEY must be set"

**Solution**: Ensure you've set the required Supabase environment variables in your `.env` file or as system environment variables.

---

**Problem**: Error connecting to Redis

**Solution**: 
- Verify Redis is running: `redis-cli ping` (should return `PONG`)
- Check `REDIS_HOST` and `REDIS_PORT` configuration
- If Redis is unavailable, the application will continue in degraded mode (without caching)

---

### Cache not working

**Problem**: All requests show `"from_cache": false` in metadata

**Solution**:
- Check Redis connectivity using the `/health` endpoint
- Verify `REDIS_TTL` is set to a reasonable value (e.g., `300s`)
- Check Redis logs for errors
- Ensure Redis has sufficient memory

---

### Slow response times

**Problem**: API responses are slower than expected

**Solution**:
- Check if Redis is running and healthy via `/health` endpoint
- Verify network latency to Supabase
- Increase `REDIS_TTL` for frequently accessed data
- Check `REQUEST_TIMEOUT` setting
- Review application logs for errors or warnings

---

### Docker Compose issues

**Problem**: Services fail to start with Docker Compose

**Solution**:
- Ensure Docker and Docker Compose are installed and running
- Check that port 8080 and 6379 are not already in use
- Verify your `.env` file exists and contains valid values
- Check logs: `docker-compose logs`

---

### Health check fails

**Problem**: `/health` endpoint returns status 503

**Solution**:
- Check the `dependencies` section in the health response to identify which service is unhealthy
- For Redis issues: Verify Redis is running and accessible
- For Supabase issues: Verify your `SUPABASE_URL` and `SUPABASE_API_KEY` are correct
- Check network connectivity to external services

## Architecture

The application follows clean architecture principles:

```
cmd/
  server/          # Application entry point
internal/
  cache/           # Redis caching layer
  logger/          # Structured logging
  middleware/      # HTTP middleware (logging, timeout, CORS)
  repository/      # Supabase data access layer
  router/          # HTTP routing and handlers
  service/         # Business logic and caching orchestration
config/            # Configuration loading and validation
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues, questions, or contributions, please open an issue on GitHub.

## Acknowledgments

- Built with [Gin Web Framework](https://github.com/gin-gonic/gin)
- Powered by [Supabase](https://supabase.com)
- Caching by [Redis](https://redis.io)
- Logging with [Zap](https://github.com/uber-go/zap)
#   g o l - b a c k e n d  
 