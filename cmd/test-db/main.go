package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourusername/supabase-redis-middleware/config"
	"github.com/yourusername/supabase-redis-middleware/internal/logger"
	"github.com/yourusername/supabase-redis-middleware/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.NewLogger(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	fmt.Println("=== PostgreSQL Connection Test ===")
	fmt.Printf("DATABASE_URL: %s\n\n", cfg.Database.URL)

	// Create PostgreSQL repository
	pgRepo, err := repository.NewPostgresRepository(cfg.Database.URL, appLogger.Logger)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgRepo.Close()

	fmt.Println("✓ Successfully connected to PostgreSQL!")
	fmt.Println()

	// Test ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pgRepo.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("✓ Database ping successful!")
	fmt.Println()

	// Test query: Get supermarket products
	fmt.Println("=== Testing Supermarket Products Query ===")
	products, err := pgRepo.QuerySupermarketProducts(ctx, map[string]interface{}{}, 5, 0)
	if err != nil {
		log.Fatalf("Failed to query products: %v", err)
	}

	fmt.Printf("Found %d products:\n", len(products))
	for i, product := range products {
		fmt.Printf("%d. %s - $%.2f (Category: %s, Stock: %d)\n",
			i+1,
			product["name"],
			product["price"],
			product["category"],
			product["stock"],
		)
	}
	fmt.Println()

	// Test query: Get movies
	fmt.Println("=== Testing Movies Query ===")
	movies, err := pgRepo.QueryMovies(ctx, map[string]interface{}{}, 5, 0)
	if err != nil {
		log.Fatalf("Failed to query movies: %v", err)
	}

	fmt.Printf("Found %d movies:\n", len(movies))
	for i, movie := range movies {
		fmt.Printf("%d. %s (%s) - Rating: %.1f\n",
			i+1,
			movie["title"],
			movie["genre"],
			movie["rating"],
		)
	}
	fmt.Println()

	// Test query: Get medicines
	fmt.Println("=== Testing Medicines Query ===")
	medicines, err := pgRepo.QueryMedicines(ctx, map[string]interface{}{}, 5, 0)
	if err != nil {
		log.Fatalf("Failed to query medicines: %v", err)
	}

	fmt.Printf("Found %d medicines:\n", len(medicines))
	for i, medicine := range medicines {
		rxRequired := "No"
		if medicine["prescription_required"].(bool) {
			rxRequired = "Yes"
		}
		fmt.Printf("%d. %s - $%.2f (Rx Required: %s, Stock: %d)\n",
			i+1,
			medicine["name"],
			medicine["price"],
			rxRequired,
			medicine["stock"],
		)
	}
	fmt.Println()

	// Test query: Get product by ID
	fmt.Println("=== Testing Get Product By ID ===")
	product, err := pgRepo.GetSupermarketProductByID(ctx, 1)
	if err != nil {
		log.Fatalf("Failed to get product by ID: %v", err)
	}

	fmt.Printf("Product ID 1:\n")
	fmt.Printf("  Name: %s\n", product["name"])
	fmt.Printf("  Category: %s\n", product["category"])
	fmt.Printf("  Price: $%.2f\n", product["price"])
	fmt.Printf("  Stock: %d\n", product["stock"])
	fmt.Printf("  Description: %s\n", product["description"])
	fmt.Println()

	// Test custom query
	fmt.Println("=== Testing Custom Query ===")
	results, err := pgRepo.ExecuteQuery(ctx, "SELECT COUNT(*) as total FROM supermarket_products")
	if err != nil {
		log.Fatalf("Failed to execute custom query: %v", err)
	}

	if len(results) > 0 {
		fmt.Printf("Total products in database: %v\n", results[0]["total"])
	}
	fmt.Println()

	fmt.Println("=== All Tests Passed! ===")
	os.Exit(0)
}
