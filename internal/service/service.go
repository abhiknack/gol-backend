package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourusername/supabase-redis-middleware/internal/cache"
	"github.com/yourusername/supabase-redis-middleware/internal/repository"
	"go.uber.org/zap"
)

// Response represents the standard API response format
type Response struct {
	Status   string            `json:"status"`
	Data     interface{}       `json:"data,omitempty"`
	Metadata *ResponseMetadata `json:"metadata,omitempty"`
	Error    *ErrorDetail      `json:"error,omitempty"`
}

// ResponseMetadata contains metadata about the response
type ResponseMetadata struct {
	CachedAt   *time.Time              `json:"cached_at,omitempty"`
	FromCache  bool                    `json:"from_cache"`
	Pagination *repository.Pagination  `json:"pagination,omitempty"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DomainService defines the interface for domain-specific operations
type DomainService interface {
	GetItems(ctx context.Context, table string, filters map[string]interface{}, pagination repository.Pagination) (*Response, error)
	GetItemByID(ctx context.Context, table string, id string) (*Response, error)
}

// domainService implements DomainService with caching logic
type domainService struct {
	cache      cache.CacheService
	repository repository.SupabaseRepository
	logger     *zap.Logger
	cacheTTL   time.Duration
}

// NewDomainService creates a new domain service instance
func NewDomainService(
	cache cache.CacheService,
	repository repository.SupabaseRepository,
	logger *zap.Logger,
	cacheTTL time.Duration,
) DomainService {
	return &domainService{
		cache:      cache,
		repository: repository,
		logger:     logger,
		cacheTTL:   cacheTTL,
	}
}

// GetItems retrieves items with cache-first logic
func (s *domainService) GetItems(ctx context.Context, table string, filters map[string]interface{}, pagination repository.Pagination) (*Response, error) {
	// Generate cache key
	cacheParams := s.buildCacheParams(filters, pagination)
	cacheKey := s.cache.GenerateKey(table, cacheParams)

	// Check cache first
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		// Cache hit
		var items []map[string]interface{}
		if err := json.Unmarshal(cachedData, &items); err == nil {
			s.logger.Info("Cache hit",
				zap.String("key", cacheKey),
				zap.String("domain", table),
			)

			cachedAt := time.Now()
			return &Response{
				Status: "success",
				Data:   items,
				Metadata: &ResponseMetadata{
					FromCache:  true,
					CachedAt:   &cachedAt,
					Pagination: &pagination,
				},
			}, nil
		}
	}

	// Cache miss - fetch from Supabase
	s.logger.Info("Cache miss",
		zap.String("key", cacheKey),
		zap.String("domain", table),
	)

	items, err := s.repository.Query(ctx, table, filters, pagination)
	if err != nil {
		return s.errorResponse(err), nil
	}

	// Update cache
	if data, err := json.Marshal(items); err == nil {
		_ = s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
	}

	return &Response{
		Status: "success",
		Data:   items,
		Metadata: &ResponseMetadata{
			FromCache:  false,
			Pagination: &pagination,
		},
	}, nil
}

// GetItemByID retrieves a single item by ID with cache-first logic
func (s *domainService) GetItemByID(ctx context.Context, table string, id string) (*Response, error) {
	// Generate cache key
	cacheParams := map[string]string{"id": id}
	cacheKey := s.cache.GenerateKey(table, cacheParams)

	// Check cache first
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		// Cache hit
		var item map[string]interface{}
		if err := json.Unmarshal(cachedData, &item); err == nil {
			s.logger.Info("Cache hit",
				zap.String("key", cacheKey),
				zap.String("domain", table),
			)

			cachedAt := time.Now()
			return &Response{
				Status: "success",
				Data:   item,
				Metadata: &ResponseMetadata{
					FromCache: true,
					CachedAt:  &cachedAt,
				},
			}, nil
		}
	}

	// Cache miss - fetch from Supabase
	s.logger.Info("Cache miss",
		zap.String("key", cacheKey),
		zap.String("domain", table),
	)

	item, err := s.repository.GetByID(ctx, table, id)
	if err != nil {
		return s.errorResponse(err), nil
	}

	// Update cache
	if data, err := json.Marshal(item); err == nil {
		_ = s.cache.Set(ctx, cacheKey, data, s.cacheTTL)
	}

	return &Response{
		Status: "success",
		Data:   item,
		Metadata: &ResponseMetadata{
			FromCache: false,
		},
	}, nil
}

// buildCacheParams converts filters and pagination to cache parameters
func (s *domainService) buildCacheParams(filters map[string]interface{}, pagination repository.Pagination) map[string]string {
	params := make(map[string]string)

	// Add filters
	for key, value := range filters {
		params[key] = fmt.Sprintf("%v", value)
	}

	// Add pagination
	if pagination.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", pagination.Limit)
	}
	if pagination.Offset > 0 {
		params["offset"] = fmt.Sprintf("%d", pagination.Offset)
	}

	return params
}

// errorResponse converts repository errors to Response format
func (s *domainService) errorResponse(err error) *Response {
	if repoErr, ok := err.(*repository.RepositoryError); ok {
		code := s.statusCodeToErrorCode(repoErr.StatusCode)
		return &Response{
			Status: "error",
			Error: &ErrorDetail{
				Code:    code,
				Message: repoErr.Message,
			},
		}
	}

	return &Response{
		Status: "error",
		Error: &ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		},
	}
}

// statusCodeToErrorCode converts HTTP status codes to error codes
func (s *domainService) statusCodeToErrorCode(statusCode int) string {
	switch statusCode {
	case 404:
		return "NOT_FOUND"
	case 503:
		return "SERVICE_UNAVAILABLE"
	case 504:
		return "TIMEOUT"
	default:
		return "INTERNAL_ERROR"
	}
}
