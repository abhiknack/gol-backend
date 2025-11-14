package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yourusername/supabase-redis-middleware/internal/cache"
	"github.com/yourusername/supabase-redis-middleware/internal/handlers"
	"github.com/yourusername/supabase-redis-middleware/internal/repository"
	"go.uber.org/zap"
)

// HandlerDependencies contains all dependencies needed by handlers
type HandlerDependencies struct {
	Cache        cache.CacheService
	Repository   repository.SupabaseRepository
	PgRepo       *repository.PostgresRepository
	Logger       *zap.Logger
	BearerTokens []string // Valid bearer tokens for authentication
}

// SetupRouter creates and configures the Gin engine with all routes and middleware
func SetupRouter(deps HandlerDependencies, requestTimeout time.Duration) *gin.Engine {
	// Create Gin engine
	router := gin.New()

	// Add recovery middleware (must be first to catch panics from other middleware)
	router.Use(gin.Recovery())

	// Add timeout middleware
	router.Use(TimeoutMiddleware(requestTimeout))

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add logging middleware (after recovery and timeout)
	router.Use(LoggingMiddleware(deps.Logger))

	// Health check endpoint (outside API versioning)
	router.GET("/health", HealthCheckHandler(deps.Cache, deps.Repository, deps.Logger))

	// Initialize handlers
	storeHandler := handlers.NewStoreHandler(deps.PgRepo, deps.Logger)
	productHandler := handlers.NewProductHandler(deps.PgRepo, deps.Logger)
	stockHandler := handlers.NewStockHandler(deps.PgRepo, deps.Logger)

	// API v1 route group - All routes are public (no authentication required)
	v1 := router.Group("/api/v1")
	{
		// Store management
		stores := v1.Group("/stores")
		{
			stores.GET("/:id", storeHandler.GetStoreBasicData)
			stores.PUT("/:id", storeHandler.UpdateStoreDetails)
			stores.PUT("/:id/status", storeHandler.UpdateStoreStatus)
			stores.GET("/:id/status", storeHandler.GetStoreStatus)
		}

		// Product management
		products := v1.Group("/products")
		{
			products.POST("/push", productHandler.PushProducts)
			products.POST("/stock", stockHandler.UpdateStock)
		}

		// Supermarket domain routes
		supermarket := v1.Group("/supermarket")
		{
			supermarket.GET("/products", PlaceholderHandler("supermarket", "products"))
			supermarket.GET("/products/:id", PlaceholderHandler("supermarket", "product"))
			supermarket.GET("/categories", PlaceholderHandler("supermarket", "categories"))
		}

		// Movie domain routes
		movies := v1.Group("/movies")
		{
			movies.GET("", PlaceholderHandler("movies", "list"))
			movies.GET("/:id", PlaceholderHandler("movies", "detail"))
			movies.GET("/showtimes", PlaceholderHandler("movies", "showtimes"))
		}

		// Pharmacy domain routes
		pharmacy := v1.Group("/pharmacy")
		{
			pharmacy.GET("/medicines", PlaceholderHandler("pharmacy", "medicines"))
			pharmacy.GET("/medicines/:id", PlaceholderHandler("pharmacy", "medicine"))
			pharmacy.GET("/categories", PlaceholderHandler("pharmacy", "categories"))
		}
	}

	// 404 handler for unsupported endpoints
	router.NoRoute(NotFoundHandler())

	return router
}
