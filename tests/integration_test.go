package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/supabase-redis-middleware/internal/cache"
	"github.com/yourusername/supabase-redis-middleware/internal/repository"
	"github.com/yourusername/supabase-redis-middleware/internal/router"
	"github.com/yourusername/supabase-redis-middleware/internal/service"
	"go.uber.org/zap"
)

// MockSupabaseRepository for integration tests
type mockSupabaseRepo struct {
	queryResult   []map[string]interface{}
	getByIDResult map[string]interface{}
	queryError    error
	getByIDError  error
	queryDelay    time.Duration
}

func (m *mockSupabaseRepo) Query(ctx context.Context, table string, filters map[string]interface{}, pagination repository.Pagination) ([]map[string]interface{}, error) {
	if m.queryDelay > 0 {
		time.Sleep(m.queryDelay)
	}
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.queryResult, nil
}

func (m *mockSupabaseRepo) GetByID(ctx context.Context, table string, id string) (map[string]interface{}, error) {
	if m.queryDelay > 0 {
		time.Sleep(m.queryDelay)
	}
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	return m.getByIDResult, nil
}

// setupTestRouter creates a test router with all dependencies
func setupTestRouter(t *testing.T, cacheService cache.CacheService, repo repository.SupabaseRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	deps := router.HandlerDependencies{
		Cache:      cacheService,
		Repository: repo,
		Logger:     logger,
	}

	return router.SetupRouter(deps, 30*time.Second)
}

// setupTestCache creates a real Redis cache for testing
func setupTestCache(t *testing.T) cache.CacheService {
	logger, _ := zap.NewDevelopment()
	
	// Try to connect to Redis
	cacheService, err := cache.NewRedisCache("localhost", "6379", "", 0, logger)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Test connection
	ctx := context.Background()
	if err := cacheService.Set(ctx, "test:connection", []byte("ok"), 10*time.Second); err == nil {
		cacheService.Delete(ctx, "test:connection")
	} else {
		t.Skip("Redis not available, skipping integration test")
	}

	return cacheService
}

// TestHealthCheckEndpoint tests the /health endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	mockRepo := &mockSupabaseRepo{
		queryResult: []map[string]interface{}{},
	}

	r := setupTestRouter(t, cacheService, mockRepo)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" && response["status"] != "degraded" {
		t.Errorf("Expected status 'healthy' or 'degraded', got %v", response["status"])
	}

	if _, ok := response["dependencies"]; !ok {
		t.Error("Expected 'dependencies' field in response")
	}
}

// TestNotFoundEndpoint tests the 404 handler
func TestNotFoundEndpoint(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	mockRepo := &mockSupabaseRepo{}
	r := setupTestRouter(t, cacheService, mockRepo)

	req, _ := http.NewRequest("GET", "/api/v1/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got %v", response["status"])
	}

	errorData, ok := response["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'error' field in response")
	}

	if errorData["code"] != "NOT_FOUND" {
		t.Errorf("Expected error code 'NOT_FOUND', got %v", errorData["code"])
	}
}

// TestServiceCacheHitScenario tests cache hit behavior
func TestServiceCacheHitScenario(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	logger, _ := zap.NewDevelopment()
	
	// Prepare test data
	testData := []map[string]interface{}{
		{"id": "1", "name": "Product 1", "price": 10.99},
		{"id": "2", "name": "Product 2", "price": 20.99},
	}

	mockRepo := &mockSupabaseRepo{
		queryResult: testData,
	}

	domainService := service.NewDomainService(cacheService, mockRepo, logger, 5*time.Minute)

	ctx := context.Background()
	filters := map[string]interface{}{"category": "electronics"}
	pagination := repository.Pagination{Limit: 10, Offset: 0}

	// First call - cache miss, should fetch from repository
	response1, err := domainService.GetItems(ctx, "products", filters, pagination)
	if err != nil {
		t.Fatalf("First GetItems() error = %v", err)
	}

	if response1.Status != "success" {
		t.Errorf("First GetItems() status = %v, want success", response1.Status)
	}

	if response1.Metadata.FromCache {
		t.Error("First GetItems() should be cache miss")
	}

	// Second call - should be cache hit
	response2, err := domainService.GetItems(ctx, "products", filters, pagination)
	if err != nil {
		t.Fatalf("Second GetItems() error = %v", err)
	}

	if response2.Status != "success" {
		t.Errorf("Second GetItems() status = %v, want success", response2.Status)
	}

	if !response2.Metadata.FromCache {
		t.Error("Second GetItems() should be cache hit")
	}

	if response2.Metadata.CachedAt == nil {
		t.Error("Cache hit should include cached_at timestamp")
	}

	// Verify data consistency
	data1, _ := json.Marshal(response1.Data)
	data2, _ := json.Marshal(response2.Data)
	if string(data1) != string(data2) {
		t.Error("Cache hit should return same data as cache miss")
	}
}

// TestServiceCacheMissScenario tests cache miss behavior
func TestServiceCacheMissScenario(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	logger, _ := zap.NewDevelopment()

	testData := []map[string]interface{}{
		{"id": "1", "name": "Product 1"},
	}

	mockRepo := &mockSupabaseRepo{
		queryResult: testData,
	}

	domainService := service.NewDomainService(cacheService, mockRepo, logger, 5*time.Minute)

	ctx := context.Background()
	filters := map[string]interface{}{"category": "new-category"}
	pagination := repository.Pagination{Limit: 5, Offset: 0}

	response, err := domainService.GetItems(ctx, "products", filters, pagination)
	if err != nil {
		t.Fatalf("GetItems() error = %v", err)
	}

	if response.Status != "success" {
		t.Errorf("GetItems() status = %v, want success", response.Status)
	}

	if response.Metadata.FromCache {
		t.Error("GetItems() should be cache miss for new query")
	}

	items, ok := response.Data.([]map[string]interface{})
	if !ok {
		t.Fatal("Response data should be []map[string]interface{}")
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}
}

// TestServiceGetByIDCacheHit tests GetByID with cache hit
func TestServiceGetByIDCacheHit(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	logger, _ := zap.NewDevelopment()

	testItem := map[string]interface{}{
		"id":   "123",
		"name": "Test Product",
	}

	mockRepo := &mockSupabaseRepo{
		getByIDResult: testItem,
	}

	domainService := service.NewDomainService(cacheService, mockRepo, logger, 5*time.Minute)

	ctx := context.Background()

	// First call - cache miss
	response1, err := domainService.GetItemByID(ctx, "products", "123")
	if err != nil {
		t.Fatalf("First GetItemByID() error = %v", err)
	}

	if response1.Metadata.FromCache {
		t.Error("First GetItemByID() should be cache miss")
	}

	// Second call - cache hit
	response2, err := domainService.GetItemByID(ctx, "products", "123")
	if err != nil {
		t.Fatalf("Second GetItemByID() error = %v", err)
	}

	if !response2.Metadata.FromCache {
		t.Error("Second GetItemByID() should be cache hit")
	}
}

// TestServiceRedisFallback tests graceful degradation when Redis fails
func TestServiceRedisFallback(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create cache with invalid Redis connection
	cacheService, err := cache.NewRedisCache("invalid-host", "9999", "", 0, logger)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cacheService.Close()

	testData := []map[string]interface{}{
		{"id": "1", "name": "Product 1"},
	}

	mockRepo := &mockSupabaseRepo{
		queryResult: testData,
	}

	domainService := service.NewDomainService(cacheService, mockRepo, logger, 5*time.Minute)

	ctx := context.Background()
	filters := map[string]interface{}{}
	pagination := repository.Pagination{Limit: 10, Offset: 0}

	// Should still work even with Redis unavailable
	response, err := domainService.GetItems(ctx, "products", filters, pagination)
	if err != nil {
		t.Fatalf("GetItems() should not fail when Redis is unavailable, got error: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("GetItems() status = %v, want success", response.Status)
	}

	items, ok := response.Data.([]map[string]interface{})
	if !ok || len(items) != 1 {
		t.Error("GetItems() should return data from repository when Redis fails")
	}
}

// TestServiceSupabaseConnectionError tests error handling for Supabase connection failures
func TestServiceSupabaseConnectionError(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	logger, _ := zap.NewDevelopment()

	mockRepo := &mockSupabaseRepo{
		queryError: repository.NewConnectionError(nil),
	}

	domainService := service.NewDomainService(cacheService, mockRepo, logger, 5*time.Minute)

	ctx := context.Background()
	filters := map[string]interface{}{}
	pagination := repository.Pagination{Limit: 10, Offset: 0}

	response, err := domainService.GetItems(ctx, "products", filters, pagination)
	if err != nil {
		t.Fatalf("GetItems() should not return error, got %v", err)
	}

	if response.Status != "error" {
		t.Errorf("GetItems() status = %v, want error", response.Status)
	}

	if response.Error == nil {
		t.Fatal("GetItems() should include error details")
	}

	if response.Error.Code != "SERVICE_UNAVAILABLE" {
		t.Errorf("GetItems() error code = %v, want SERVICE_UNAVAILABLE", response.Error.Code)
	}
}

// TestServiceSupabaseTimeout tests timeout handling
func TestServiceSupabaseTimeout(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	logger, _ := zap.NewDevelopment()

	mockRepo := &mockSupabaseRepo{
		queryError: repository.NewTimeoutError(nil),
	}

	domainService := service.NewDomainService(cacheService, mockRepo, logger, 5*time.Minute)

	ctx := context.Background()
	filters := map[string]interface{}{}
	pagination := repository.Pagination{Limit: 10, Offset: 0}

	response, err := domainService.GetItems(ctx, "products", filters, pagination)
	if err != nil {
		t.Fatalf("GetItems() should not return error, got %v", err)
	}

	if response.Status != "error" {
		t.Errorf("GetItems() status = %v, want error", response.Status)
	}

	if response.Error == nil {
		t.Fatal("GetItems() should include error details")
	}

	if response.Error.Code != "TIMEOUT" {
		t.Errorf("GetItems() error code = %v, want TIMEOUT", response.Error.Code)
	}
}

// TestServiceNotFoundError tests 404 error handling
func TestServiceNotFoundError(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	logger, _ := zap.NewDevelopment()

	mockRepo := &mockSupabaseRepo{
		getByIDError: repository.NewNotFoundError("products", "999"),
	}

	domainService := service.NewDomainService(cacheService, mockRepo, logger, 5*time.Minute)

	ctx := context.Background()

	response, err := domainService.GetItemByID(ctx, "products", "999")
	if err != nil {
		t.Fatalf("GetItemByID() should not return error, got %v", err)
	}

	if response.Status != "error" {
		t.Errorf("GetItemByID() status = %v, want error", response.Status)
	}

	if response.Error == nil {
		t.Fatal("GetItemByID() should include error details")
	}

	if response.Error.Code != "NOT_FOUND" {
		t.Errorf("GetItemByID() error code = %v, want NOT_FOUND", response.Error.Code)
	}
}

// TestCompleteRequestFlow tests the complete flow from HTTP request to response
func TestCompleteRequestFlow(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	testData := []map[string]interface{}{
		{"id": "1", "name": "Product 1"},
		{"id": "2", "name": "Product 2"},
	}

	mockRepo := &mockSupabaseRepo{
		queryResult: testData,
	}

	r := setupTestRouter(t, cacheService, mockRepo)

	// Test placeholder endpoints (they should return 501 Not Implemented)
	endpoints := []string{
		"/api/v1/supermarket/products",
		"/api/v1/supermarket/categories",
		"/api/v1/movies",
		"/api/v1/pharmacy/medicines",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusNotImplemented {
				t.Errorf("Expected status 501, got %d", w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if response["status"] != "error" {
				t.Errorf("Expected status 'error', got %v", response["status"])
			}
		})
	}
}

// TestCacheKeyConsistency tests that cache keys are generated consistently
func TestCacheKeyConsistency(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	logger, _ := zap.NewDevelopment()

	testData := []map[string]interface{}{
		{"id": "1", "name": "Product 1"},
	}

	mockRepo := &mockSupabaseRepo{
		queryResult: testData,
	}

	domainService := service.NewDomainService(cacheService, mockRepo, logger, 5*time.Minute)

	ctx := context.Background()

	// Same filters in different order should hit the same cache
	filters1 := map[string]interface{}{
		"category": "electronics",
		"brand":    "Apple",
	}

	filters2 := map[string]interface{}{
		"brand":    "Apple",
		"category": "electronics",
	}

	pagination := repository.Pagination{Limit: 10, Offset: 0}

	// First call with filters1
	response1, _ := domainService.GetItems(ctx, "products", filters1, pagination)
	if response1.Metadata.FromCache {
		t.Error("First call should be cache miss")
	}

	// Second call with filters2 (different order) should hit cache
	response2, _ := domainService.GetItems(ctx, "products", filters2, pagination)
	if !response2.Metadata.FromCache {
		t.Error("Second call with same filters (different order) should be cache hit")
	}
}

// TestConcurrentRequests tests handling of concurrent requests
func TestConcurrentRequests(t *testing.T) {
	cacheService := setupTestCache(t)
	defer cacheService.Close()

	logger, _ := zap.NewDevelopment()

	testData := []map[string]interface{}{
		{"id": "1", "name": "Product 1"},
	}

	mockRepo := &mockSupabaseRepo{
		queryResult: testData,
	}

	domainService := service.NewDomainService(cacheService, mockRepo, logger, 5*time.Minute)

	ctx := context.Background()
	filters := map[string]interface{}{"category": "test"}
	pagination := repository.Pagination{Limit: 10, Offset: 0}

	// Make concurrent requests
	numRequests := 10
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			response, err := domainService.GetItems(ctx, "products", filters, pagination)
			if err != nil {
				errors <- err
			} else if response.Status != "success" {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}
}
