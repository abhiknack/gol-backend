# Integration Tests

This directory contains integration tests for the Supabase-Redis Middleware application. These tests verify the complete request flow from HTTP handlers through the service layer to the repository and cache layers.

## Overview

The integration tests cover:

- Complete HTTP request/response flow
- Cache hit and miss scenarios with real Redis
- Error handling for Supabase connection failures
- Graceful degradation when Redis is unavailable
- Timeout handling
- Concurrent request handling
- Cache key consistency

## Prerequisites

### Required
- Go 1.23 or higher
- Redis server (optional - tests will skip if unavailable)

### Optional
- Docker and Docker Compose (for running Redis in a container)

## Running the Tests

### Run All Integration Tests

```bash
go test -v ./tests/...
```

### Run Specific Tests

```bash
# Run a single test
go test -v ./tests/... -run TestHealthCheckEndpoint

# Run multiple specific tests
go test -v ./tests/... -run "TestCacheHit|TestCacheMiss"
```

### Run with Redis

If Redis is not running locally, you can start it using Docker:

```bash
# Start Redis using docker-compose
docker-compose up -d redis

# Run the tests
go test -v ./tests/...

# Stop Redis
docker-compose down
```

Alternatively, start Redis directly:

```bash
docker run -d -p 6379:6379 redis:latest
```

### Run without Redis

The tests are designed to gracefully handle Redis being unavailable. Tests that require Redis will be skipped automatically if Redis is not accessible.

```bash
# Tests will skip Redis-dependent scenarios if Redis is unavailable
go test -v ./tests/...
```

## Test Coverage

### HTTP Endpoint Tests

- `TestHealthCheckEndpoint` - Tests the `/health` endpoint
- `TestNotFoundEndpoint` - Tests 404 error handling
- `TestCompleteRequestFlow` - Tests complete HTTP request flow through all layers

### Cache Tests

- `TestServiceCacheHitScenario` - Verifies cache hit behavior
- `TestServiceCacheMissScenario` - Verifies cache miss behavior
- `TestServiceGetByIDCacheHit` - Tests GetByID with cache hit
- `TestCacheKeyConsistency` - Ensures cache keys are generated consistently

### Error Handling Tests

- `TestServiceRedisFallback` - Tests graceful degradation when Redis fails
- `TestServiceSupabaseConnectionError` - Tests Supabase connection error handling
- `TestServiceSupabaseTimeout` - Tests timeout error handling
- `TestServiceNotFoundError` - Tests 404 error handling

### Concurrency Tests

- `TestConcurrentRequests` - Tests handling of concurrent requests

## Test Structure

Each integration test follows this pattern:

1. **Setup** - Create test dependencies (cache, repository, router)
2. **Execute** - Perform the operation being tested
3. **Verify** - Assert expected behavior and responses
4. **Cleanup** - Close connections and clean up resources

## Mock Components

The tests use mock implementations for:

- **mockSupabaseRepo** - Simulates Supabase database responses
  - Configurable query results
  - Configurable errors
  - Configurable delays for timeout testing

Real components used:

- **Redis Cache** - Uses actual Redis connection when available
- **Gin Router** - Uses real Gin HTTP router
- **Service Layer** - Uses actual service implementation

## Environment Variables

No environment variables are required for running the integration tests. The tests use sensible defaults:

- Redis: `localhost:6379`
- Test timeout: 30 seconds
- Cache TTL: 5 minutes

## Troubleshooting

### Redis Connection Issues

If you see warnings about Redis connection failures:

```
WARN cache/cache.go:52 Failed to connect to Redis, cache will operate in degraded mode
```

This is expected if Redis is not running. Tests will continue and skip Redis-dependent scenarios.

### Test Timeouts

If tests are timing out, ensure:

1. Redis is running and accessible
2. No firewall is blocking port 6379
3. System has sufficient resources

### Port Conflicts

If port 6379 is already in use:

```bash
# Check what's using the port
netstat -ano | findstr :6379  # Windows
lsof -i :6379                 # Linux/Mac

# Stop the conflicting service or use a different port
```

## Best Practices

1. **Run tests before committing** - Ensure all tests pass
2. **Keep tests isolated** - Each test should be independent
3. **Clean up resources** - Always close connections in defer statements
4. **Use meaningful test names** - Test names should describe what they test
5. **Test error paths** - Don't just test the happy path

## Adding New Tests

When adding new integration tests:

1. Follow the existing test structure
2. Use the `setupTestRouter` and `setupTestCache` helper functions
3. Add appropriate cleanup with `defer`
4. Document what the test covers
5. Ensure tests can run with or without Redis

Example:

```go
func TestNewFeature(t *testing.T) {
    // Setup
    cacheService := setupTestCache(t)
    defer cacheService.Close()
    
    mockRepo := &mockSupabaseRepo{
        queryResult: testData,
    }
    
    // Execute
    result := performOperation()
    
    // Verify
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

## CI/CD Integration

These tests are designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Integration Tests
  run: |
    docker run -d -p 6379:6379 redis:latest
    go test -v ./tests/...
```

## Performance Considerations

- Tests use real Redis connections which may be slower than mocks
- Concurrent tests may impact system resources
- Consider running tests in parallel with `-parallel` flag for faster execution

```bash
go test -v -parallel 4 ./tests/...
```
