package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TimeoutMiddleware creates a middleware that enforces request timeout
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context with timeout context
		c.Request = c.Request.WithContext(ctx)

		// Channel to signal when request processing is done
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// Request completed successfully
			return
		case <-ctx.Done():
			// Timeout occurred
			if ctx.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusGatewayTimeout, gin.H{
					"status": "error",
					"error": gin.H{
						"code":    "TIMEOUT",
						"message": "Request timeout exceeded",
					},
				})
				c.Abort()
			}
		}
	}
}

// LoggingMiddleware creates a Gin middleware that logs all incoming requests
// and their responses with structured logging
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get request details before processing
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		// Log incoming request
		logger.Info("incoming request",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.Time("timestamp", start),
		)

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Get response status
		status := c.Writer.Status()

		// Log response with duration
		logger.Info("request completed",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.Time("timestamp", time.Now()),
		)

		// Log errors if any occurred
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("request error",
					zap.String("method", method),
					zap.String("path", path),
					zap.String("error", err.Error()),
				)
			}
		}
	}
}

// BearerAuthMiddleware creates a middleware that validates Bearer tokens
func BearerAuthMiddleware(validTokens []string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			logger.Warn("missing authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()))

			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing authorization header",
				},
			})
			c.Abort()
			return
		}

		// Check if it starts with "Bearer "
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			logger.Warn("invalid authorization format",
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()))

			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization format. Expected: Bearer <token>",
				},
			})
			c.Abort()
			return
		}

		// Extract token
		token := authHeader[len(bearerPrefix):]

		if token == "" {
			logger.Warn("empty bearer token",
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()))

			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Empty bearer token",
				},
			})
			c.Abort()
			return
		}

		// Validate token against valid tokens list
		isValid := false
		for _, validToken := range validTokens {
			if token == validToken {
				isValid = true
				break
			}
		}

		if !isValid {
			logger.Warn("invalid bearer token",
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()))

			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid bearer token",
				},
			})
			c.Abort()
			return
		}

		// Token is valid, continue
		logger.Debug("bearer token validated",
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()))

		c.Next()
	}
}
