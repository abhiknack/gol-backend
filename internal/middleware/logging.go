package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
