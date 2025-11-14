package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/supabase-redis-middleware/internal/repository"
	"go.uber.org/zap"
)

// Mock implementations
type mockCacheService struct {
	getData    map[string][]byte
	setError   error
	getError   error
	shouldFail bool
}

func (m *mockCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	if m.shouldFail {
		return nil, m.getError
	}
	if data, ok := m.getData[key]; ok {
		return data, nil
	}
	return nil, nil
}

func (m *mockCacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.shouldFail {
		return m.setError
	}
	if m.getData == nil {
		m.getData = make(map[string][]byte)
	}
	m.getData[key] = value
	return nil
}

func (m *mockCacheService) Delete(ctx context.Context, key string) error {
	if m.getData != nil {
		delete(m.getData, key)
	}
	return nil
}

func (m *mockCacheService) GenerateKey(domain string, params map[string]string) string {
	if len(params) == 0 {
		return domain
	}
	return domain + ":cached"
}

func (m *mockCacheService) Close() error {
	return nil
}

type mockSupabaseRepository struct {
	queryResult   []map[string]interface{}
	getByIDResult map[string]interface{}
	queryError    error
	getByIDError  error
}

func (m *mockSupabaseRepository) Query(ctx context.Context, table string, filters map[string]interface{}, pagination repository.Pagination) ([]map[string]interface{}, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.queryResult, nil
}

func (m *mockSupabaseRepository) GetByID(ctx context.Context, table string, id string) (map[string]interface{}, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	return m.getByIDResult, nil
}

func setupTestService(cache *mockCacheService, repo *mockSupabaseRepository) DomainService {
	logger, _ := zap.NewDevelopment()
	return NewDomainService(cache, repo, logger, 5*time.Minute)
}

func TestGetItems_CacheHit(t *testing.T) {
	// Prepare cached data
	cachedItems := []map[string]interface{}{
		{"id": "1", "name": "Product 1"},
		{"id": "2", "name": "Product 2"},
	}
	cachedData, _ := json.Marshal(cachedItems)

	mockCache := &mockCacheService{
		getData: map[string][]byte{
			"products:cached": cachedData,
		},
	}
	mockRepo := &mockSupabaseRepository{}

	service := setupTestService(mockCache, mockRepo)

	ctx := context.Background()
	filters := map[string]interface{}{"category": "electronics"}
	pagination := repository.Pagination{Limit: 10, Offset: 0}

	response, err := service.GetItems(ctx, "products", filters, pagination)

	if err != nil {
		t.Errorf("GetItems() error = %v", err)
	}

	if response.Status != "success" {
		t.Errorf("GetItems() status = %v, want success", response.Status)
	}

	if response.Metadata == nil || !response.Metadata.FromCache {
		t.Error("GetItems() should indicate cache hit")
	}

	if response.Metadata.CachedAt == nil {
		t.Error("GetItems() should include cached_at timestamp")
	}

	items, ok := response.Data.([]map[string]interface{})
	if !ok {
		t.Fatal("GetItems() data should be []map[string]interface{}")
	}

	if len(items) != 2 {
		t.Errorf("GetItems() returned %d items, want 2", len(items))
	}
}

func TestGetItems_CacheMiss(t *testing.T) {
	repoItems := []map[string]interface{}{
		{"id": "1", "name": "Product 1"},
		{"id": "2", "name": "Product 2"},
	}

	mockCache := &mockCacheService{
		getData: make(map[string][]byte),
	}
	mockRepo := &mockSupabaseRepository{
		queryResult: repoItems,
	}

	service := setupTestService(mockCache, mockRepo)

	ctx := context.Background()
	filters := map[string]interface{}{"category": "electronics"}
	pagination := repository.Pagination{Limit: 10, Offset: 0}

	response, err := service.GetItems(ctx, "products", filters, pagination)

	if err != nil {
		t.Errorf("GetItems() error = %v", err)
	}

	if response.Status != "success" {
		t.Errorf("GetItems() status = %v, want success", response.Status)
	}

	if response.Metadata == nil || response.Metadata.FromCache {
		t.Error("GetItems() should indicate cache miss")
	}

	if response.Metadata.CachedAt != nil {
		t.Error("GetItems() should not include cached_at for cache miss")
	}

	items, ok := response.Data.([]map[string]interface{})
	if !ok {
		t.Fatal("GetItems() data should be []map[string]interface{}")
	}

	if len(items) != 2 {
		t.Errorf("GetItems() returned %d items, want 2", len(items))
	}

	// Verify cache was updated
	cacheKey := mockCache.GenerateKey("products", map[string]string{"category": "electronics", "limit": "10", "offset": "0"})
	if _, ok := mockCache.getData[cacheKey]; !ok {
		t.Error("GetItems() should update cache after fetching from repository")
	}
}

func TestGetItems_RepositoryError(t *testing.T) {
	mockCache := &mockCacheService{
		getData: make(map[string][]byte),
	}
	mockRepo := &mockSupabaseRepository{
		queryError: repository.NewConnectionError(errors.New("connection failed")),
	}

	service := setupTestService(mockCache, mockRepo)

	ctx := context.Background()
	filters := map[string]interface{}{}
	pagination := repository.Pagination{Limit: 10, Offset: 0}

	response, err := service.GetItems(ctx, "products", filters, pagination)

	if err != nil {
		t.Errorf("GetItems() should not return error, got %v", err)
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

func TestGetItems_RedisFallback(t *testing.T) {
	repoItems := []map[string]interface{}{
		{"id": "1", "name": "Product 1"},
	}

	mockCache := &mockCacheService{
		shouldFail: true,
		getError:   errors.New("redis connection failed"),
	}
	mockRepo := &mockSupabaseRepository{
		queryResult: repoItems,
	}

	service := setupTestService(mockCache, mockRepo)

	ctx := context.Background()
	filters := map[string]interface{}{}
	pagination := repository.Pagination{Limit: 10, Offset: 0}

	response, err := service.GetItems(ctx, "products", filters, pagination)

	if err != nil {
		t.Errorf("GetItems() error = %v", err)
	}

	if response.Status != "success" {
		t.Errorf("GetItems() status = %v, want success", response.Status)
	}

	// Should still return data from repository
	items, ok := response.Data.([]map[string]interface{})
	if !ok || len(items) != 1 {
		t.Error("GetItems() should return data from repository when cache fails")
	}
}

func TestGetItemByID_CacheHit(t *testing.T) {
	cachedItem := map[string]interface{}{"id": "123", "name": "Product 123"}
	cachedData, _ := json.Marshal(cachedItem)

	mockCache := &mockCacheService{
		getData: map[string][]byte{
			"products:cached": cachedData,
		},
	}
	mockRepo := &mockSupabaseRepository{}

	service := setupTestService(mockCache, mockRepo)

	ctx := context.Background()
	response, err := service.GetItemByID(ctx, "products", "123")

	if err != nil {
		t.Errorf("GetItemByID() error = %v", err)
	}

	if response.Status != "success" {
		t.Errorf("GetItemByID() status = %v, want success", response.Status)
	}

	if response.Metadata == nil || !response.Metadata.FromCache {
		t.Error("GetItemByID() should indicate cache hit")
	}

	item, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("GetItemByID() data should be map[string]interface{}")
	}

	if item["id"] != "123" {
		t.Errorf("GetItemByID() returned item with id %v, want 123", item["id"])
	}
}

func TestGetItemByID_CacheMiss(t *testing.T) {
	repoItem := map[string]interface{}{"id": "123", "name": "Product 123"}

	mockCache := &mockCacheService{
		getData: make(map[string][]byte),
	}
	mockRepo := &mockSupabaseRepository{
		getByIDResult: repoItem,
	}

	service := setupTestService(mockCache, mockRepo)

	ctx := context.Background()
	response, err := service.GetItemByID(ctx, "products", "123")

	if err != nil {
		t.Errorf("GetItemByID() error = %v", err)
	}

	if response.Status != "success" {
		t.Errorf("GetItemByID() status = %v, want success", response.Status)
	}

	if response.Metadata == nil || response.Metadata.FromCache {
		t.Error("GetItemByID() should indicate cache miss")
	}

	item, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("GetItemByID() data should be map[string]interface{}")
	}

	if item["id"] != "123" {
		t.Errorf("GetItemByID() returned item with id %v, want 123", item["id"])
	}
}

func TestGetItemByID_NotFound(t *testing.T) {
	mockCache := &mockCacheService{
		getData: make(map[string][]byte),
	}
	mockRepo := &mockSupabaseRepository{
		getByIDError: repository.NewNotFoundError("products", "999"),
	}

	service := setupTestService(mockCache, mockRepo)

	ctx := context.Background()
	response, err := service.GetItemByID(ctx, "products", "999")

	if err != nil {
		t.Errorf("GetItemByID() should not return error, got %v", err)
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

func TestGetItemByID_RedisFallback(t *testing.T) {
	repoItem := map[string]interface{}{"id": "123", "name": "Product 123"}

	mockCache := &mockCacheService{
		shouldFail: true,
		getError:   errors.New("redis connection failed"),
	}
	mockRepo := &mockSupabaseRepository{
		getByIDResult: repoItem,
	}

	service := setupTestService(mockCache, mockRepo)

	ctx := context.Background()
	response, err := service.GetItemByID(ctx, "products", "123")

	if err != nil {
		t.Errorf("GetItemByID() error = %v", err)
	}

	if response.Status != "success" {
		t.Errorf("GetItemByID() status = %v, want success", response.Status)
	}

	// Should still return data from repository
	item, ok := response.Data.(map[string]interface{})
	if !ok || item["id"] != "123" {
		t.Error("GetItemByID() should return data from repository when cache fails")
	}
}

func TestBuildCacheParams(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := &domainService{logger: logger}

	filters := map[string]interface{}{
		"category": "electronics",
		"brand":    "Apple",
	}
	pagination := repository.Pagination{Limit: 20, Offset: 10}

	params := service.buildCacheParams(filters, pagination)

	if params["category"] != "electronics" {
		t.Errorf("buildCacheParams() category = %v, want electronics", params["category"])
	}

	if params["brand"] != "Apple" {
		t.Errorf("buildCacheParams() brand = %v, want Apple", params["brand"])
	}

	if params["limit"] != "20" {
		t.Errorf("buildCacheParams() limit = %v, want 20", params["limit"])
	}

	if params["offset"] != "10" {
		t.Errorf("buildCacheParams() offset = %v, want 10", params["offset"])
	}
}

func TestStatusCodeToErrorCode(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := &domainService{logger: logger}

	tests := []struct {
		statusCode int
		want       string
	}{
		{404, "NOT_FOUND"},
		{503, "SERVICE_UNAVAILABLE"},
		{504, "TIMEOUT"},
		{500, "INTERNAL_ERROR"},
		{400, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := service.statusCodeToErrorCode(tt.statusCode)
			if got != tt.want {
				t.Errorf("statusCodeToErrorCode(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}
