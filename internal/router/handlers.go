package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/supabase-redis-middleware/internal/cache"
	"github.com/yourusername/supabase-redis-middleware/internal/repository"
	"go.uber.org/zap"
)

// HealthCheckHandler creates a handler for the /health endpoint
// It checks connectivity to Redis and Supabase
func HealthCheckHandler(cacheService cache.CacheService, repo repository.SupabaseRepository, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		health := gin.H{
			"status": "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"dependencies": gin.H{},
		}

		allHealthy := true

		// Check Redis connectivity
		redisStatus := checkRedis(ctx, cacheService, logger)
		health["dependencies"].(gin.H)["redis"] = redisStatus
		if redisStatus["status"] != "healthy" {
			allHealthy = false
		}

		// Check Supabase connectivity
		supabaseStatus := checkSupabase(ctx, repo, logger)
		health["dependencies"].(gin.H)["supabase"] = supabaseStatus
		if supabaseStatus["status"] != "healthy" {
			allHealthy = false
		}

		// Set overall status
		if !allHealthy {
			health["status"] = "degraded"
		}

		// Return appropriate status code
		statusCode := http.StatusOK
		if !allHealthy {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, health)
	}
}

// checkRedis verifies Redis connectivity
func checkRedis(ctx context.Context, cacheService cache.CacheService, logger *zap.Logger) gin.H {
	testKey := "health:check:redis"
	testValue := []byte("ping")

	// Try to set a value
	err := cacheService.Set(ctx, testKey, testValue, 10*time.Second)
	if err != nil {
		logger.Warn("Redis health check failed on SET", zap.Error(err))
		return gin.H{
			"status": "unhealthy",
			"error":  "Failed to write to Redis",
		}
	}

	// Try to get the value
	_, err = cacheService.Get(ctx, testKey)
	if err != nil {
		logger.Warn("Redis health check failed on GET", zap.Error(err))
		return gin.H{
			"status": "unhealthy",
			"error":  "Failed to read from Redis",
		}
	}

	// Clean up
	_ = cacheService.Delete(ctx, testKey)

	return gin.H{
		"status": "healthy",
	}
}

// checkSupabase verifies Supabase connectivity
func checkSupabase(ctx context.Context, repo repository.SupabaseRepository, logger *zap.Logger) gin.H {
	// Try a simple query to verify connectivity
	// We'll query with a limit of 1 to minimize load
	_, err := repo.Query(ctx, "health_check", map[string]interface{}{}, repository.Pagination{Limit: 1})
	
	if err != nil {
		// Check if it's a "table not found" error, which actually means connection is working
		// but the health_check table doesn't exist (which is expected)
		errMsg := err.Error()
		if contains(errMsg, "relation") && contains(errMsg, "does not exist") {
			// Connection is working, table just doesn't exist
			return gin.H{
				"status": "healthy",
			}
		}

		logger.Warn("Supabase health check failed", zap.Error(err))
		return gin.H{
			"status": "unhealthy",
			"error":  "Failed to connect to Supabase",
		}
	}

	return gin.H{
		"status": "healthy",
	}
}

// NotFoundHandler returns a handler for 404 errors
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "The requested endpoint does not exist",
			},
		})
	}
}

// PlaceholderHandler creates a temporary handler for routes that will be implemented later
// This allows the router to be set up before all handlers are complete
func PlaceholderHandler(domain, operation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"status": "error",
			"error": gin.H{
				"code":    "NOT_IMPLEMENTED",
				"message": "This endpoint is not yet implemented",
			},
			"metadata": gin.H{
				"domain":    domain,
				"operation": operation,
			},
		})
	}
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
