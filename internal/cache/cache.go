package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheService defines the interface for cache operations
type CacheService interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	GenerateKey(domain string, params map[string]string) string
	Close() error
}

// RedisCache implements CacheService using Redis
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisCache creates a new Redis cache service with connection pooling
func NewRedisCache(host, port, password string, db int, logger *zap.Logger) (*RedisCache, error) {
	addr := fmt.Sprintf("%s:%s", host, port)
	
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("Failed to connect to Redis, cache will operate in degraded mode",
			zap.String("addr", addr),
			zap.Error(err),
		)
		// Don't return error - allow graceful degradation
	} else {
		logger.Info("Successfully connected to Redis",
			zap.String("addr", addr),
			zap.Int("db", db),
		)
	}

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

// Get retrieves a value from cache by key
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Cache miss - not an error condition
			return nil, nil
		}
		// Redis error - log warning and return nil to allow graceful degradation
		r.logger.Warn("Redis GET operation failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, nil
	}

	return []byte(val), nil
}

// Set stores a value in cache with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		// Log warning but don't fail the operation
		r.logger.Warn("Redis SET operation failed",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return nil // Graceful degradation
	}

	return nil
}

// Delete removes a value from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Warn("Redis DELETE operation failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil // Graceful degradation
	}

	return nil
}

// GenerateKey creates a consistent cache key from domain and parameters
// Uses consistent hashing to ensure parameter order doesn't affect the key
func (r *RedisCache) GenerateKey(domain string, params map[string]string) string {
	if len(params) == 0 {
		return domain
	}

	// Sort parameter keys for consistency
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build parameter string
	var paramParts []string
	for _, k := range keys {
		paramParts = append(paramParts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	paramStr := strings.Join(paramParts, "&")

	// Hash the parameters to keep key length manageable
	hash := sha256.Sum256([]byte(paramStr))
	hashStr := fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes of hash

	return fmt.Sprintf("%s:%s", domain, hashStr)
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// MarshalJSON is a helper function to marshal data to JSON bytes
func MarshalJSON(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// UnmarshalJSON is a helper function to unmarshal JSON bytes to data
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
