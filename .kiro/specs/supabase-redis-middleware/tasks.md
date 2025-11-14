# Implementation Plan

- [x] 1. Initialize Go project and install dependencies
  - Create go.mod file with module name
  - Install Gin framework (github.com/gin-gonic/gin)
  - Install Supabase Go client (github.com/supabase-community/supabase-go)
  - Install Redis client (github.com/redis/go-redis/v9)
  - Install Viper for configuration (github.com/spf13/viper)
  - Install Zap for logging (go.uber.org/zap)
  - Create basic project directory structure (cmd, internal, config)
  - _Requirements: 1.1, 5.1_

- [x] 2. Implement configuration module
  - [x] 2.1 Create configuration structs for server, Supabase, and Redis settings
    - Define Config, ServerConfig, SupabaseConfig, and RedisConfig structs
    - Add validation tags for required fields
    - _Requirements: 5.1, 5.2, 5.3, 5.4_
  - [x] 2.2 Implement configuration loading from environment variables and config file
    - Use Viper to load from environment variables
    - Support loading from config.yaml as fallback
    - Implement validation logic that fails on missing required configuration
    - _Requirements: 5.1, 5.5_

- [x] 3. Set up logging infrastructure
  - [x] 3.1 Initialize Zap logger with configurable log levels
    - Create logger factory function
    - Support different log levels (debug, info, warn, error)
    - Configure structured logging format
    - _Requirements: 4.1, 4.2_
  - [x] 3.2 Create logging middleware for Gin
    - Log incoming requests with timestamp, method, path, and client IP
    - Log response status and duration
    - _Requirements: 4.1_

- [x] 4. Implement Redis cache service
  - [x] 4.1 Create CacheService interface and implementation
    - Implement Get, Set, Delete methods with context support
    - Initialize Redis client with connection pooling
    - _Requirements: 3.1, 3.2, 3.3_
  - [x] 4.2 Implement cache key generation logic
    - Create GenerateKey method that structures keys by domain and parameters
    - Use consistent hashing for parameter ordering
    - _Requirements: 3.4_
  - [x] 4.3 Add graceful degradation for Redis failures
    - Implement error handling that allows operation without Redis
    - Log warnings when Redis is unavailable
    - _Requirements: 3.5_
  - [x] 4.4 Write unit tests for cache service
    - Test Get/Set/Delete operations
    - Test cache key generation with various parameters
    - Test graceful degradation when Redis is down
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 5. Implement Supabase repository layer
  - [x] 5.1 Create SupabaseRepository interface and implementation
    - Initialize Supabase client with URL and API key
    - Implement Query method with filtering and pagination support
    - Implement GetByID method for single record retrieval
    - _Requirements: 2.1, 2.2, 2.5_
  - [x] 5.2 Add error handling for Supabase operations
    - Handle connection failures with 503 status
    - Handle query errors with appropriate error messages
    - Implement timeout handling
    - _Requirements: 2.4, 4.4_
  - [x] 5.3 Write unit tests for Supabase repository
    - Mock Supabase client responses
    - Test Query with various filters and pagination
    - Test error scenarios (connection failure, timeout)
    - _Requirements: 2.1, 2.2, 2.4, 2.5_

- [x] 6. Implement service layer with caching logic
  - [x] 6.1 Create DomainService interface and base implementation
    - Define GetItems and GetItemByID methods
    - Implement cache-first logic: check cache, then query Supabase
    - Update cache after Supabase fetch
    - _Requirements: 3.2, 3.3_
  - [x] 6.2 Implement response transformation and metadata
    - Create Response struct with status, data, metadata, and error fields
    - Include cache status (from_cache, cached_at) in metadata
    - Include pagination metadata when applicable
    - _Requirements: 6.1, 6.2, 6.4, 6.5_
  - [x] 6.3 Add cache hit/miss logging
    - Log cache hits with key and domain
    - Log cache misses with key and domain
    - _Requirements: 4.5_
  - [x] 6.4 Write unit tests for service layer
    - Test cache hit scenario
    - Test cache miss scenario with Supabase fetch
    - Test Redis failure fallback
    - Mock both cache and repository dependencies
    - _Requirements: 3.2, 3.3, 3.5_

- [x] 7. Implement HTTP handlers and routing



  - [ ] 7.1 Create handler package with common utilities
    - Create internal/handler package
    - Implement parseQueryParams helper for filters and pagination
    - Implement parsePathParam helper for ID extraction
    - Implement respondWithJSON and respondWithError helpers
    - _Requirements: 1.1, 1.3, 4.3_
  - [ ] 7.2 Create SupermarketHandler with endpoints
    - Implement GetProducts endpoint (GET /api/v1/supermarket/products)
    - Implement GetProductByID endpoint (GET /api/v1/supermarket/products/:id)
    - Implement GetCategories endpoint (GET /api/v1/supermarket/categories)
    - _Requirements: 1.1, 1.2, 1.3, 4.3_
  - [ ] 7.3 Create MovieHandler with endpoints
    - Implement GetMovies endpoint (GET /api/v1/movies)
    - Implement GetMovieByID endpoint (GET /api/v1/movies/:id)
    - Implement GetShowtimes endpoint (GET /api/v1/movies/showtimes)
    - _Requirements: 1.1, 1.2, 1.3_
  - [ ] 7.4 Create PharmacyHandler with endpoints
    - Implement GetMedicines endpoint (GET /api/v1/pharmacy/medicines)
    - Implement GetMedicineByID endpoint (GET /api/v1/pharmacy/medicines/:id)
    - Implement GetPharmacyCategories endpoint (GET /api/v1/pharmacy/categories)
    - _Requirements: 1.1, 1.2, 1.3_

-


- [x] 8. Set up Gin router and middleware chain


  - [x] 8.1 Create router package with setup function

    - Create internal/router package
    - Implement SetupRouter function that accepts handlers and returns configured Gin engine
    - Set up /api/v1 route group
    - Register all domain handler routes
    - _Requirements: 1.1, 1.2_
  - [x] 8.2 Add additional middleware


    - Add error recovery middleware
    - Add timeout middleware with configurable duration
    - Add CORS middleware
    - _Requirements: 4.1, 4.4_
  - [x] 8.3 Implement health check endpoint


    - Create /health endpoint that checks Redis and Supabase connectivity
    - Return JSON with status for each dependency
    - _Requirements: 1.1_
  - [x] 8.4 Add 404 handler for unsupported endpoints

    - Return structured error response with 404 status
    - _Requirements: 1.5_

- [x] 9. Create main application entry point




  - [x] 9.1 Create main.go in cmd/server directory


    - Load configuration using config.Load()
    - Initialize logger with configured log level
    - Initialize Redis cache service
    - Initialize Supabase repository
    - Create domain service instance
    - Create handler instances
    - Set up router with all handlers
    - Start HTTP server on configured port
    - _Requirements: 2.1, 3.1, 5.1_
  - [x] 9.2 Implement graceful shutdown

    - Handle OS signals (SIGINT, SIGTERM)
    - Close Redis connections
    - Shutdown HTTP server with timeout
    - Flush logger
    - _Requirements: 1.4_
  - [x] 9.3 Add startup validation and connectivity checks

    - Validate configuration is loaded successfully
    - Log startup information (port, log level, etc.)
    - Test Redis connectivity on startup (warn if unavailable)
    - Fail fast with descriptive errors if Supabase credentials missing
    - _Requirements: 5.5_

- [x] 10. Create Docker configuration and deployment files






  - [x] 10.1 Write Dockerfile for the application

    - Use multi-stage build (build stage with Go, runtime stage with minimal image)
    - Copy binary to runtime stage
    - Expose server port (8080)
    - Set up non-root user for security
    - _Requirements: 1.1_
  - [x] 10.2 Create docker-compose.yml for local development


    - Define service for the Go application
    - Define Redis service with persistence
    - Set up environment variables
    - Configure networking between services
    - _Requirements: 3.1, 5.3_


  - [ ] 10.3 Create .env.example file
    - Document all required environment variables (SERVER_PORT, SUPABASE_URL, SUPABASE_API_KEY, REDIS_HOST, etc.)
    - Provide example values for local development
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 11. Write integration tests





  - Create integration test suite in tests directory
  - Test complete request flow from handler to service to repository
  - Test cache hit and miss scenarios with real Redis (using testcontainers)
  - Test error handling scenarios
  - _Requirements: 1.4, 3.2, 3.3, 3.5, 4.3_

- [x] 12. Create README documentation






  - [x] 12.1 Write setup and installation instructions

    - Document prerequisites (Go 1.23+, Redis, Supabase account)
    - Provide step-by-step setup guide
    - Include configuration instructions
    - _Requirements: 5.1_

  - [ ] 12.2 Document API endpoints
    - List all available endpoints with curl examples
    - Show request/response formats
    - Document query parameters and filters

    - _Requirements: 1.1, 1.2, 6.1, 6.2_
  - [ ] 12.3 Add deployment and usage guide
    - Document Docker deployment steps
    - Provide environment variable reference
    - Include troubleshooting section
    - _Requirements: 5.1, 5.5_
