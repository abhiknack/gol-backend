# PostgreSQL with pgx Setup Guide

## Overview

Your application is now configured to use PostgreSQL with the `pgx` driver (v5), which is the recommended PostgreSQL driver for Go.

## Configuration

### Environment Variables

The application uses `DATABASE_URL` for PostgreSQL connection:

```env
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/middleware_db?sslmode=disable
```

### Connection String Format

```
postgresql://[user]:[password]@[host]:[port]/[database]?[parameters]
```

**Parameters:**
- `sslmode=disable` - For local development (use `require` in production)
- `pool_max_conns=10` - Maximum number of connections in the pool
- `pool_min_conns=2` - Minimum number of connections in the pool

### From Different Environments

**From Host Machine:**
```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/middleware_db?sslmode=disable
```

**From Docker Containers:**
```env
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/middleware_db?sslmode=disable
```

## Database Initialization

### Option 1: Using PowerShell Script

```powershell
.\scripts\init-db.ps1
```

### Option 2: Using SQL File Directly

```powershell
Get-Content init-postgres.sql | docker exec -i supabase-postgres-temp psql -U postgres -d middleware_db
```

### Option 3: Manual Initialization

```bash
docker exec -it supabase-postgres-temp psql -U postgres -d middleware_db -f /path/to/init-postgres.sql
```

## Testing the Connection

### Run the Test Program

```bash
# In dev container or with Go installed locally
go run cmd/test-db/main.go
```

**Expected Output:**
```
=== PostgreSQL Connection Test ===
DATABASE_URL: postgresql://postgres:postgres@postgres:5432/middleware_db?sslmode=disable

✓ Successfully connected to PostgreSQL!
✓ Database ping successful!

=== Testing Supermarket Products Query ===
Found 5 products:
1. Whole Milk - $3.99 (Category: dairy, Stock: 50)
2. White Bread - $2.49 (Category: bakery, Stock: 100)
...

=== All Tests Passed! ===
```

## Using PostgreSQL Repository

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/yourusername/supabase-redis-middleware/config"
    "github.com/yourusername/supabase-redis-middleware/internal/logger"
    "github.com/yourusername/supabase-redis-middleware/internal/repository"
)

func main() {
    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize logger
    appLogger, err := logger.NewLogger(cfg.Logging.Level)
    if err != nil {
        log.Fatal(err)
    }
    defer appLogger.Sync()
    
    // Create PostgreSQL repository
    pgRepo, err := repository.NewPostgresRepository(cfg.Database.URL, appLogger.Logger)
    if err != nil {
        log.Fatal(err)
    }
    defer pgRepo.Close()
    
    // Use the repository
    ctx := context.Background()
    products, err := pgRepo.QuerySupermarketProducts(ctx, map[string]interface{}{}, 10, 0)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process products...
}
```

### Available Repository Methods

#### Query Supermarket Products

```go
filters := map[string]interface{}{
    "category": "dairy",
    "search": "milk",
}
products, err := pgRepo.QuerySupermarketProducts(ctx, filters, 10, 0)
```

#### Get Product by ID

```go
product, err := pgRepo.GetSupermarketProductByID(ctx, 1)
```

#### Query Movies

```go
filters := map[string]interface{}{
    "genre": "action",
}
movies, err := pgRepo.QueryMovies(ctx, filters, 10, 0)
```

#### Query Medicines

```go
filters := map[string]interface{}{
    "category": "pain-relief",
    "search": "aspirin",
}
medicines, err := pgRepo.QueryMedicines(ctx, filters, 10, 0)
```

#### Execute Custom Query

```go
query := "SELECT * FROM supermarket_products WHERE price < $1"
results, err := pgRepo.ExecuteQuery(ctx, query, 5.00)
```

#### Direct Pool Access

```go
pool := pgRepo.GetPool()
rows, err := pool.Query(ctx, "SELECT * FROM movies")
```

## Database Schema

### Tables Created

1. **supermarket_products**
   - id, name, category, price, stock, description
   - Indexed on: category

2. **movies**
   - id, title, genre, duration, rating, release_date, description
   - Indexed on: genre

3. **showtimes**
   - id, movie_id, theater, showtime, available_seats, price
   - Indexed on: movie_id, showtime

4. **medicines**
   - id, name, category, price, prescription_required, stock, description
   - Indexed on: category

### Views Created

1. **available_movies** - Movies with showtime counts and price ranges
2. **low_stock_products** - Products with low stock across domains

## Connection Pool Configuration

The pgx driver uses connection pooling by default. You can configure it via the connection string:

```env
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/middleware_db?pool_max_conns=20&pool_min_conns=5
```

Or programmatically:

```go
config, err := pgxpool.ParseConfig(databaseURL)
if err != nil {
    return nil, err
}

// Configure pool
config.MaxConns = 20
config.MinConns = 5
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute

pool, err := pgxpool.NewWithConfig(context.Background(), config)
```

## Integrating with Main Application

### Update cmd/server/main.go

```go
// After initializing Supabase repository
pgRepo, err := repository.NewPostgresRepository(cfg.Database.URL, log.Logger)
if err != nil {
    log.Error("Failed to initialize PostgreSQL repository", zap.Error(err))
    os.Exit(1)
}
defer pgRepo.Close()

log.Info("Successfully initialized PostgreSQL repository")

// Pass to handlers
routerDeps := router.HandlerDependencies{
    Cache:      cacheService,
    Repository: supabaseRepo,
    PgRepo:     pgRepo,  // Add this
    Logger:     log.Logger,
}
```

## Migrations

### Using golang-migrate

**Install:**
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**Create Migration:**
```bash
migrate create -ext sql -dir migrations -seq add_users_table
```

**Run Migrations:**
```bash
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/middleware_db?sslmode=disable" up
```

**Rollback:**
```bash
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/middleware_db?sslmode=disable" down 1
```

## Performance Tips

### 1. Use Prepared Statements

```go
// pgx automatically prepares frequently used statements
// Just use the same query multiple times
for _, id := range ids {
    product, err := pgRepo.GetSupermarketProductByID(ctx, id)
    // pgx will prepare and cache this query
}
```

### 2. Batch Operations

```go
pool := pgRepo.GetPool()
batch := &pgx.Batch{}

for _, product := range products {
    batch.Queue("INSERT INTO supermarket_products (name, price) VALUES ($1, $2)", 
        product.Name, product.Price)
}

results := pool.SendBatch(ctx, batch)
defer results.Close()
```

### 3. Use Transactions

```go
pool := pgRepo.GetPool()
tx, err := pool.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx)

// Do multiple operations
_, err = tx.Exec(ctx, "INSERT INTO ...")
if err != nil {
    return err
}

_, err = tx.Exec(ctx, "UPDATE ...")
if err != nil {
    return err
}

// Commit
return tx.Commit(ctx)
```

### 4. Connection Pooling

```go
// Configure pool for your workload
config.MaxConns = 25  // Max concurrent connections
config.MinConns = 5   // Keep 5 connections warm
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute
config.HealthCheckPeriod = time.Minute
```

## Monitoring

### Check Pool Stats

```go
pool := pgRepo.GetPool()
stats := pool.Stat()

fmt.Printf("Total connections: %d\n", stats.TotalConns())
fmt.Printf("Idle connections: %d\n", stats.IdleConns())
fmt.Printf("Acquired connections: %d\n", stats.AcquiredConns())
```

### Query Logging

Enable query logging in development:

```go
config.ConnConfig.Logger = logger
config.ConnConfig.LogLevel = pgx.LogLevelDebug
```

## Troubleshooting

### Connection Refused

**Check if PostgreSQL is running:**
```bash
docker ps | findstr postgres
```

**Test connection:**
```bash
docker exec supabase-postgres-temp pg_isready -U postgres
```

### Too Many Connections

**Check current connections:**
```sql
SELECT count(*) FROM pg_stat_activity;
```

**Kill idle connections:**
```sql
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle' 
AND state_change < current_timestamp - INTERVAL '5 minutes';
```

### Slow Queries

**Enable query logging:**
```sql
ALTER DATABASE middleware_db SET log_statement = 'all';
ALTER DATABASE middleware_db SET log_duration = on;
```

**Check slow queries:**
```sql
SELECT query, mean_exec_time, calls 
FROM pg_stat_statements 
ORDER BY mean_exec_time DESC 
LIMIT 10;
```

## Security Best Practices

### For Production

1. **Use SSL/TLS:**
   ```env
   DATABASE_URL=postgresql://user:pass@host:5432/db?sslmode=require
   ```

2. **Use Strong Passwords:**
   ```env
   POSTGRES_PASSWORD=$(openssl rand -base64 32)
   ```

3. **Limit Permissions:**
   ```sql
   CREATE USER app_user WITH PASSWORD 'strong_password';
   GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
   ```

4. **Use Connection Pooling:**
   - Limit max connections
   - Set appropriate timeouts
   - Monitor pool usage

5. **Enable Audit Logging:**
   ```sql
   ALTER DATABASE middleware_db SET log_connections = on;
   ALTER DATABASE middleware_db SET log_disconnections = on;
   ```

## Next Steps

1. ✅ PostgreSQL is running with sample data
2. ✅ pgx driver is configured
3. ✅ Repository methods are available
4. Run the test: `go run cmd/test-db/main.go`
5. Integrate with your handlers
6. Add caching layer for frequently accessed data
7. Set up migrations for schema changes

---

**PostgreSQL with pgx is ready to use!** 🐘

Connection: `postgresql://postgres:postgres@localhost:5432/middleware_db?sslmode=disable`
