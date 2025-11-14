package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/supabase-redis-middleware/config"
	"github.com/yourusername/supabase-redis-middleware/internal/cache"
	"github.com/yourusername/supabase-redis-middleware/internal/logger"
	"github.com/yourusername/supabase-redis-middleware/internal/repository"
	"github.com/yourusername/supabase-redis-middleware/internal/router"
	"github.com/yourusername/supabase-redis-middleware/internal/service"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.Logging.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Log startup information
	log.Info("Starting Supabase-Redis Middleware",
		zap.String("port", cfg.Server.Port),
		zap.String("log_level", cfg.Logging.Level),
		zap.Duration("request_timeout", cfg.Server.RequestTimeout),
	)

	// Validate Supabase credentials
	if cfg.Supabase.URL == "" || cfg.Supabase.APIKey == "" {
		log.Error("Supabase credentials are missing",
			zap.String("url", cfg.Supabase.URL),
			zap.Bool("api_key_set", cfg.Supabase.APIKey != ""),
		)
		fmt.Fprintf(os.Stderr, "SUPABASE_URL and SUPABASE_API_KEY must be set\n")
		os.Exit(1)
	}

	// Initialize Redis cache service
	cacheService, err := cache.NewRedisCache(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
		log.Logger,
	)
	if err != nil {
		log.Error("Failed to initialize Redis cache", zap.Error(err))
		os.Exit(1)
	}
	defer cacheService.Close()

	// Test Redis connectivity on startup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	testKey := "startup:health:check"
	if err := cacheService.Set(ctx, testKey, []byte("ok"), 10*time.Second); err != nil {
		log.Warn("Redis connectivity test failed - cache will operate in degraded mode", zap.Error(err))
	} else {
		log.Info("Redis connectivity test passed")
		cacheService.Delete(ctx, testKey)
	}
	cancel()

	// Initialize Supabase repository
	supabaseRepo, err := repository.NewSupabaseRepository(cfg.Supabase.URL, cfg.Supabase.APIKey)
	if err != nil {
		log.Error("Failed to initialize Supabase repository", zap.Error(err))
		os.Exit(1)
	}

	log.Info("Successfully initialized Supabase repository",
		zap.String("url", cfg.Supabase.URL),
	)

	// Create domain service instance
	_ = service.NewDomainService(
		cacheService,
		supabaseRepo,
		log.Logger,
		cfg.Redis.TTL,
	)

	log.Info("Domain service initialized",
		zap.Duration("cache_ttl", cfg.Redis.TTL),
	)

	// Initialize PostgreSQL repository
	pgRepo, err := repository.NewPostgresRepository(cfg.Database.URL, log.Logger)
	if err != nil {
		log.Error("Failed to initialize PostgreSQL repository", zap.Error(err))
		os.Exit(1)
	}
	defer pgRepo.Close()

	log.Info("Successfully initialized PostgreSQL repository")

	// Set up router with all handlers
	routerDeps := router.HandlerDependencies{
		Cache:        cacheService,
		Repository:   supabaseRepo,
		PgRepo:       pgRepo,
		Logger:       log.Logger,
		BearerTokens: cfg.Server.BearerTokens,
	}
	ginRouter := router.SetupRouter(routerDeps, cfg.Server.RequestTimeout)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      ginRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info("HTTP server starting",
			zap.String("address", server.Addr),
			zap.Duration("read_timeout", cfg.Server.ReadTimeout),
			zap.Duration("write_timeout", cfg.Server.WriteTimeout),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server failed", zap.Error(err))
			os.Exit(1)
		}
	}()

	log.Info("Server started successfully", zap.String("port", cfg.Server.Port))

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	} else {
		log.Info("HTTP server shutdown complete")
	}

	// Close Redis connections
	if err := cacheService.Close(); err != nil {
		log.Error("Error closing Redis connection", zap.Error(err))
	} else {
		log.Info("Redis connection closed")
	}

	// Flush logger
	if err := log.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Error flushing logger: %v\n", err)
	}

	log.Info("Shutdown complete")
}
