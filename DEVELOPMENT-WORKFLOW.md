# Gol Bazaar - Development Workflow Guide

## 🎯 How to Develop

You have **two main options** for development:

### Option 1: Local Development (Recommended)
Edit code locally, run in Docker containers

### Option 2: Dev Container (VS Code)
Develop entirely inside a container with VS Code

---

## 🚀 Option 1: Local Development (Recommended)

### Setup

1. **Deploy Development Environment**
   ```powershell
   .\scripts\deploy-dev.ps1
   ```

2. **Verify Services are Running**
   ```powershell
   docker ps | findstr gol-bazaar
   ```

### Development Workflow

#### 1. Edit Code Locally
Use your favorite IDE (VS Code, GoLand, etc.) to edit code on your local machine:

```
D:\Gol-Backend\
├── cmd/
│   └── server/
│       └── main.go          ← Edit this
├── internal/
│   ├── handlers/            ← Create new handlers
│   ├── repository/          ← Database queries
│   ├── service/             ← Business logic
│   └── router/              ← Routes
├── config/                  ← Configuration
└── go.mod                   ← Dependencies
```

#### 2. Rebuild and Restart Application

**Quick Restart (for code changes):**
```powershell
# Rebuild and restart just the app
docker-compose -f docker-compose.dev.yml up -d --build app

# Or restart without rebuild (if no new dependencies)
docker-compose -f docker-compose.dev.yml restart app
```

**Full Rebuild (for dependency changes):**
```powershell
# When you add new Go packages
docker-compose -f docker-compose.dev.yml down
docker-compose -f docker-compose.dev.yml up -d --build
```

#### 3. View Logs
```powershell
# Follow application logs
docker logs gol-bazaar-app-dev -f

# View last 100 lines
docker logs gol-bazaar-app-dev --tail 100

# All services
docker-compose -f docker-compose.dev.yml logs -f
```

#### 4. Test Your Changes
```powershell
# Test health endpoint
curl http://localhost:8080/health

# Test your new endpoint
curl http://localhost:8080/api/v1/stores

# Test with data
curl -X POST http://localhost:8080/api/v1/stores -H "Content-Type: application/json" -d '{\"name\":\"Test Store\"}'
```

### Typical Development Cycle

```
1. Edit code locally (VS Code, GoLand, etc.)
   ↓
2. Save files
   ↓
3. Rebuild container: docker-compose -f docker-compose.dev.yml up -d --build app
   ↓
4. Test: curl http://localhost:8080/your-endpoint
   ↓
5. Check logs: docker logs gol-bazaar-app-dev -f
   ↓
6. Repeat from step 1
```

### Example: Adding a New Endpoint

**Step 1: Create Handler**
```go
// internal/handlers/store_handler.go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func GetStores(c *gin.Context) {
    // Your logic here
    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data": []string{"Store 1", "Store 2"},
    })
}
```

**Step 2: Add Route**
```go
// internal/router/router.go
func SetupRouter(deps HandlerDependencies, requestTimeout time.Duration) *gin.Engine {
    // ... existing code ...
    
    v1 := router.Group("/api/v1")
    {
        v1.GET("/stores", handlers.GetStores)  // Add this line
    }
    
    return router
}
```

**Step 3: Rebuild and Test**
```powershell
# Rebuild
docker-compose -f docker-compose.dev.yml up -d --build app

# Test
curl http://localhost:8080/api/v1/stores
```

---

## 🐳 Option 2: Dev Container (VS Code)

### Setup

1. **Open in VS Code**
   ```powershell
   code .
   ```

2. **Reopen in Container**
   - Press `F1`
   - Select "Dev Containers: Reopen in Container"
   - Wait for container to build

3. **Start Development**
   - All your code is now inside the container
   - Terminal runs inside the container
   - Go tools are pre-installed

### Development Workflow

#### 1. Edit Code in VS Code
VS Code is now connected to the container. Edit files directly.

#### 2. Run Application
```bash
# In VS Code integrated terminal (inside container)
go run cmd/server/main.go
```

#### 3. Hot Reload (Optional)
```bash
# Install Air for hot reload
go install github.com/cosmtrek/air@latest

# Run with hot reload
air

# Now any code change auto-restarts the server!
```

#### 4. Test
VS Code will automatically forward port 8080. Click the notification or:
```bash
# In another terminal
curl http://localhost:8080/health
```

### Dev Container Features

✅ **Go 1.23** pre-installed  
✅ **All Go tools** (gopls, delve, staticcheck)  
✅ **Redis** running locally  
✅ **PostgreSQL** accessible  
✅ **Source code** mounted  
✅ **VS Code extensions** auto-installed  

---

## 🗄️ Database Development

### Access PostgreSQL

**Using pgAdmin (Recommended):**
1. Open http://localhost:5050
2. Login: `admin@golbazaar.local` / `admin`
3. Add server:
   - Host: `gol-bazaar-postgres-dev`
   - Port: `5432`
   - Database: `middleware_db`
   - Username: `postgres`
   - Password: `postgres`

**Using Command Line:**
```powershell
# Connect to PostgreSQL
docker exec -it gol-bazaar-postgres-dev psql -U postgres -d middleware_db

# Run queries
SELECT * FROM stores LIMIT 10;
SELECT * FROM products WHERE category_id = 'some-uuid';
```

**Using VS Code Extension:**
1. Install "PostgreSQL" extension
2. Connect to `localhost:5432`
3. Database: `middleware_db`
4. User: `postgres` / `postgres`

### Run Migrations

```powershell
# Create migration
docker exec gol-bazaar-app-dev migrate create -ext sql -dir migrations -seq add_new_table

# Run migrations
docker exec gol-bazaar-app-dev migrate -path migrations -database "postgresql://postgres:postgres@postgres:5432/middleware_db?sslmode=disable" up

# Rollback
docker exec gol-bazaar-app-dev migrate -path migrations -database "postgresql://postgres:postgres@postgres:5432/middleware_db?sslmode=disable" down 1
```

### Sample Data

```powershell
# Load sample data
Get-Content scripts/sample-data.sql | docker exec -i gol-bazaar-postgres-dev psql -U postgres -d middleware_db

# Or create your own
docker exec -it gol-bazaar-postgres-dev psql -U postgres -d middleware_db
```

---

## 🔴 Redis Development

### Access Redis

**Using Redis Commander (Recommended):**
1. Open http://localhost:8081
2. Browse keys, view data, execute commands

**Using Command Line:**
```powershell
# Connect to Redis
docker exec -it gol-bazaar-redis-dev redis-cli

# Test commands
PING
SET test "Hello Gol Bazaar"
GET test
KEYS *
```

### Test Caching

```go
// Example: Test caching in your code
func GetStoreWithCache(c *gin.Context, cache cache.CacheService) {
    storeID := c.Param("id")
    cacheKey := fmt.Sprintf("store:%s", storeID)
    
    // Try cache first
    cached, err := cache.Get(c, cacheKey)
    if err == nil {
        c.JSON(200, gin.H{"data": cached, "from_cache": true})
        return
    }
    
    // Fetch from database
    store := fetchStoreFromDB(storeID)
    
    // Cache it
    cache.Set(c, cacheKey, store, 5*time.Minute)
    
    c.JSON(200, gin.H{"data": store, "from_cache": false})
}
```

---

## 🧪 Testing

### Run Tests

```powershell
# Run all tests
docker exec gol-bazaar-app-dev go test ./...

# Run with coverage
docker exec gol-bazaar-app-dev go test -cover ./...

# Run specific package
docker exec gol-bazaar-app-dev go test ./internal/handlers/...

# Verbose output
docker exec gol-bazaar-app-dev go test -v ./...
```

### Integration Tests

```powershell
# Run integration tests (requires database)
docker exec gol-bazaar-app-dev go test ./tests/integration/...
```

---

## 📦 Dependency Management

### Add New Package

```powershell
# Add package
docker exec gol-bazaar-app-dev go get github.com/some/package

# Update go.mod and go.sum
docker exec gol-bazaar-app-dev go mod tidy

# Rebuild container
docker-compose -f docker-compose.dev.yml up -d --build app
```

### Update Dependencies

```powershell
# Update all dependencies
docker exec gol-bazaar-app-dev go get -u ./...

# Update specific package
docker exec gol-bazaar-app-dev go get -u github.com/gin-gonic/gin

# Tidy up
docker exec gol-bazaar-app-dev go mod tidy
```

---

## 🐛 Debugging

### View Logs

```powershell
# Application logs
docker logs gol-bazaar-app-dev -f

# PostgreSQL logs
docker logs gol-bazaar-postgres-dev -f

# Redis logs
docker logs gol-bazaar-redis-dev -f

# All logs
docker-compose -f docker-compose.dev.yml logs -f
```

### Debug with Delve

**In Dev Container:**
```bash
# Install delve (already installed in dev container)
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger
dlv debug cmd/server/main.go

# Or attach to running process
dlv attach $(pgrep server)
```

**In VS Code:**
1. Set breakpoints in your code
2. Press `F5` or use Debug panel
3. Select "Launch Package"

### Check Environment

```powershell
# View environment variables
docker exec gol-bazaar-app-dev env

# Check Go version
docker exec gol-bazaar-app-dev go version

# Check database connection
docker exec gol-bazaar-app-dev psql -h postgres -U postgres -d middleware_db -c "SELECT 1"
```

---

## 🔄 Common Development Tasks

### Restart Services

```powershell
# Restart app only
docker-compose -f docker-compose.dev.yml restart app

# Restart all services
docker-compose -f docker-compose.dev.yml restart

# Restart specific service
docker-compose -f docker-compose.dev.yml restart postgres
```

### Reset Database

```powershell
# Stop services
docker-compose -f docker-compose.dev.yml down

# Remove database volume
docker volume rm gol-bazaar-postgres-dev-data

# Start fresh
docker-compose -f docker-compose.dev.yml up -d

# Schema will be auto-applied from grocery_superapp_schema.sql
```

### Clear Redis Cache

```powershell
# Clear all keys
docker exec gol-bazaar-redis-dev redis-cli FLUSHALL

# Clear specific pattern
docker exec gol-bazaar-redis-dev redis-cli --scan --pattern "store:*" | xargs docker exec -i gol-bazaar-redis-dev redis-cli DEL
```

### View Container Stats

```powershell
# All containers
docker stats

# Specific containers
docker stats gol-bazaar-app-dev gol-bazaar-postgres-dev gol-bazaar-redis-dev
```

---

## 📝 Development Best Practices

### 1. Use Debug Logging
```go
// Set LOG_LEVEL=debug in .env.development
logger.Debug("Processing store request", 
    zap.String("store_id", storeID),
    zap.Any("filters", filters))
```

### 2. Test Locally First
```powershell
# Always test locally before committing
curl http://localhost:8080/api/v1/your-endpoint
```

### 3. Use pgAdmin for Database Work
- Visual query builder
- Easy data browsing
- Schema visualization

### 4. Monitor Redis with Redis Commander
- See cached data
- Monitor memory usage
- Debug cache issues

### 5. Keep Logs Open
```powershell
# Always have logs running in a separate terminal
docker logs gol-bazaar-app-dev -f
```

---

## 🚨 Troubleshooting

### Application Won't Start

```powershell
# Check logs
docker logs gol-bazaar-app-dev

# Check if port is in use
netstat -ano | findstr "8080"

# Restart container
docker-compose -f docker-compose.dev.yml restart app
```

### Database Connection Failed

```powershell
# Check if PostgreSQL is running
docker ps | findstr postgres

# Test connection
docker exec gol-bazaar-postgres-dev pg_isready -U postgres

# Check logs
docker logs gol-bazaar-postgres-dev
```

### Redis Connection Failed

```powershell
# Check if Redis is running
docker ps | findstr redis

# Test connection
docker exec gol-bazaar-redis-dev redis-cli ping

# Check logs
docker logs gol-bazaar-redis-dev
```

### Code Changes Not Reflected

```powershell
# Rebuild container
docker-compose -f docker-compose.dev.yml up -d --build app

# Or full rebuild
docker-compose -f docker-compose.dev.yml down
docker-compose -f docker-compose.dev.yml up -d --build
```

---

## 📚 Quick Reference

### Start Development
```powershell
.\scripts\deploy-dev.ps1
```

### Edit Code
```
Use your local IDE → Edit files → Save
```

### Rebuild & Test
```powershell
docker-compose -f docker-compose.dev.yml up -d --build app
curl http://localhost:8080/your-endpoint
```

### View Logs
```powershell
docker logs gol-bazaar-app-dev -f
```

### Access Services
- **App**: http://localhost:8080
- **pgAdmin**: http://localhost:5050
- **Redis Commander**: http://localhost:8081

---

## 🎉 You're Ready to Develop!

**Recommended Workflow:**
1. Deploy dev environment: `.\scripts\deploy-dev.ps1`
2. Open project in your IDE
3. Edit code locally
4. Rebuild: `docker-compose -f docker-compose.dev.yml up -d --build app`
5. Test: `curl http://localhost:8080/...`
6. Check logs: `docker logs gol-bazaar-app-dev -f`
7. Repeat!

**Pro Tips:**
- Keep logs open in a separate terminal
- Use pgAdmin for database work
- Use Redis Commander to monitor cache
- Test frequently with curl or Postman
- Commit often to Git

Happy coding! 🚀
