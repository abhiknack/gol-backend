package cache

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func setupTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func TestGenerateKey(t *testing.T) {
	logger := setupTestLogger()
	cache := &RedisCache{logger: logger}

	tests := []struct {
		name     string
		domain   string
		params   map[string]string
		expected string
	}{
		{
			name:     "domain only",
			domain:   "supermarket",
			params:   map[string]string{},
			expected: "supermarket",
		},
		{
			name:   "domain with single param",
			domain: "movies",
			params: map[string]string{"id": "123"},
		},
		{
			name:   "domain with multiple params",
			domain: "pharmacy",
			params: map[string]string{
				"category": "medicine",
				"limit":    "10",
				"offset":   "0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := cache.GenerateKey(tt.domain, tt.params)
			
			if tt.expected != "" && key != tt.expected {
				t.Errorf("GenerateKey() = %v, want %v", key, tt.expected)
			}
			
			if len(tt.params) > 0 && key == tt.domain {
				t.Errorf("GenerateKey() should include params hash, got %v", key)
			}
		})
	}
}

func TestGenerateKeyConsistency(t *testing.T) {
	logger := setupTestLogger()
	cache := &RedisCache{logger: logger}

	// Same params in different order should produce same key
	params1 := map[string]string{
		"category": "dairy",
		"limit":    "10",
		"offset":   "0",
	}
	
	params2 := map[string]string{
		"offset":   "0",
		"limit":    "10",
		"category": "dairy",
	}

	key1 := cache.GenerateKey("supermarket", params1)
	key2 := cache.GenerateKey("supermarket", params2)

	if key1 != key2 {
		t.Errorf("GenerateKey() should be consistent regardless of param order: %v != %v", key1, key2)
	}
}

func TestRedisCache_GetSetDelete(t *testing.T) {
	logger := setupTestLogger()
	
	// Try to connect to Redis
	cache, err := NewRedisCache("localhost", "6379", "", 0, logger)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	
	// Test connection
	if err := cache.client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	testKey := "test:key:123"
	testValue := []byte(`{"test": "data"}`)

	// Test Set
	err = cache.Set(ctx, testKey, testValue, 10*time.Second)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// Test Get
	result, err := cache.Get(ctx, testKey)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if string(result) != string(testValue) {
		t.Errorf("Get() = %v, want %v", string(result), string(testValue))
	}

	// Test Delete
	err = cache.Delete(ctx, testKey)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify deletion
	result, err = cache.Get(ctx, testKey)
	if err != nil {
		t.Errorf("Get() after Delete() error = %v", err)
	}
	if result != nil {
		t.Errorf("Get() after Delete() should return nil, got %v", result)
	}
}

func TestRedisCache_GetNonExistent(t *testing.T) {
	logger := setupTestLogger()
	
	cache, err := NewRedisCache("localhost", "6379", "", 0, logger)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	
	if err := cache.client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	// Get non-existent key should return nil without error
	result, err := cache.Get(ctx, "nonexistent:key")
	if err != nil {
		t.Errorf("Get() for non-existent key error = %v, want nil", err)
	}
	if result != nil {
		t.Errorf("Get() for non-existent key = %v, want nil", result)
	}
}

func TestRedisCache_GracefulDegradation(t *testing.T) {
	logger := setupTestLogger()
	
	// Connect to invalid Redis instance
	cache, err := NewRedisCache("invalid-host", "9999", "", 0, logger)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	testKey := "test:key"
	testValue := []byte("test")

	// Operations should not fail even when Redis is unavailable
	err = cache.Set(ctx, testKey, testValue, 10*time.Second)
	if err != nil {
		t.Errorf("Set() should not fail with unavailable Redis, got error: %v", err)
	}

	result, err := cache.Get(ctx, testKey)
	if err != nil {
		t.Errorf("Get() should not fail with unavailable Redis, got error: %v", err)
	}
	if result != nil {
		t.Errorf("Get() with unavailable Redis should return nil, got: %v", result)
	}

	err = cache.Delete(ctx, testKey)
	if err != nil {
		t.Errorf("Delete() should not fail with unavailable Redis, got error: %v", err)
	}
}

func TestRedisCache_TTL(t *testing.T) {
	logger := setupTestLogger()
	
	cache, err := NewRedisCache("localhost", "6379", "", 0, logger)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	
	if err := cache.client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping integration test")
	}

	testKey := "test:ttl:key"
	testValue := []byte("test")

	// Set with short TTL
	err = cache.Set(ctx, testKey, testValue, 1*time.Second)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	// Verify it exists
	result, err := cache.Get(ctx, testKey)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if result == nil {
		t.Error("Get() should return value before TTL expires")
	}

	// Wait for TTL to expire
	time.Sleep(2 * time.Second)

	// Verify it's gone
	result, err = cache.Get(ctx, testKey)
	if err != nil {
		t.Errorf("Get() after TTL error = %v", err)
	}
	if result != nil {
		t.Errorf("Get() after TTL should return nil, got %v", result)
	}
}

func TestMarshalUnmarshalJSON(t *testing.T) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := TestData{Name: "test", Value: 42}

	// Test Marshal
	data, err := MarshalJSON(original)
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
	}

	// Test Unmarshal
	var result TestData
	err = UnmarshalJSON(data, &result)
	if err != nil {
		t.Errorf("UnmarshalJSON() error = %v", err)
	}

	if result.Name != original.Name || result.Value != original.Value {
		t.Errorf("UnmarshalJSON() = %v, want %v", result, original)
	}
}
