# Gol Bazaar - Development Cheat Sheet

## 🚀 Quick Start

```powershell
# Deploy development environment
.\scripts\deploy-dev.ps1

# Check status
docker ps | findstr gol-bazaar
```

## 📝 Daily Development

### Edit → Build → Test Cycle

```powershell
# 1. Edit code in your IDE (VS Code, GoLand, etc.)

# 2. Rebuild app
docker-compose -f docker-compose.dev.yml up -d --build app

# 3. Test
curl http://localhost:8080/health

# 4. View logs
docker logs gol-bazaar-app-dev -f
```

## 🔧 Container Management

```powershell
# Start all services
docker-compose -f docker-compose.dev.yml up -d

# Stop all services
docker-compose -f docker-compose.dev.yml down

# Restart app only
docker-compose -f docker-compose.dev.yml restart app

# Rebuild everything
docker-compose -f docker-compose.dev.yml up -d --build

# View all logs
docker-compose -f docker-compose.dev.yml logs -f

# View app logs only
docker logs gol-bazaar-app-dev -f
```

## 🗄️ Database Commands

```powershell
# Connect to PostgreSQL
docker exec -it gol-bazaar-postgres-dev psql -U postgres -d middleware_db

# Run SQL file
Get-Content your-script.sql | docker exec -i gol-bazaar-postgres-dev psql -U postgres -d middleware_db

# Backup database
docker exec gol-bazaar-postgres-dev pg_dump -U postgres -d middleware_db > backup.sql

# Restore database
Get-Content backup.sql | docker exec -i gol-bazaar-postgres-dev psql -U postgres -d middleware_db

# Reset database (⚠️ deletes all data)
docker-compose -f docker-compose.dev.yml down
docker volume rm gol-bazaar-postgres-dev-data
docker-compose -f docker-compose.dev.yml up -d
```

## 🔴 Redis Commands

```powershell
# Connect to Redis
docker exec -it gol-bazaar-redis-dev redis-cli

# Test connection
docker exec gol-bazaar-redis-dev redis-cli ping

# View all keys
docker exec gol-bazaar-redis-dev redis-cli keys "*"

# Clear all cache
docker exec gol-bazaar-redis-dev redis-cli FLUSHALL

# Get specific key
docker exec gol-bazaar-redis-dev redis-cli GET "your-key"
```

## 🧪 Testing

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

## 📦 Go Dependencies

```powershell
# Add new package
docker exec gol-bazaar-app-dev go get github.com/some/package

# Update dependencies
docker exec gol-bazaar-app-dev go get -u ./...

# Tidy up
docker exec gol-bazaar-app-dev go mod tidy

# After adding dependencies, rebuild
docker-compose -f docker-compose.dev.yml up -d --build app
```

## 🌐 API Testing

```powershell
# Health check
curl http://localhost:8080/health

# GET request
curl http://localhost:8080/api/v1/stores

# POST request
curl -X POST http://localhost:8080/api/v1/stores -H "Content-Type: application/json" -d '{\"name\":\"Test Store\"}'

# With query parameters
curl "http://localhost:8080/api/v1/products?category=dairy&limit=10"
```

## 🔍 Debugging

```powershell
# View environment variables
docker exec gol-bazaar-app-dev env

# Check Go version
docker exec gol-bazaar-app-dev go version

# Execute shell in container
docker exec -it gol-bazaar-app-dev sh

# View container stats
docker stats gol-bazaar-app-dev
```

## 🌐 Access URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| **Application** | http://localhost:8080 | - |
| **Health Check** | http://localhost:8080/health | - |
| **pgAdmin** | http://localhost:5050 | admin@golbazaar.local / admin |
| **Redis Commander** | http://localhost:8081 | - |
| **PostgreSQL** | localhost:5432 | postgres / postgres |
| **Redis** | localhost:6379 | (no password) |

## 🧹 Cleanup

```powershell
# Stop services
docker-compose -f docker-compose.dev.yml down

# Stop and remove volumes (⚠️ deletes data)
docker-compose -f docker-compose.dev.yml down -v

# Remove specific volume
docker volume rm gol-bazaar-postgres-dev-data

# Clean up Docker system
docker system prune -a
```

## 🔄 Common Workflows

### Adding a New Endpoint

```powershell
# 1. Create handler in internal/handlers/
# 2. Add route in internal/router/router.go
# 3. Rebuild
docker-compose -f docker-compose.dev.yml up -d --build app
# 4. Test
curl http://localhost:8080/api/v1/your-endpoint
```

### Database Schema Change

```powershell
# 1. Update grocery_superapp_schema.sql
# 2. Reset database
docker-compose -f docker-compose.dev.yml down
docker volume rm gol-bazaar-postgres-dev-data
docker-compose -f docker-compose.dev.yml up -d
# 3. Verify
docker exec -it gol-bazaar-postgres-dev psql -U postgres -d middleware_db -c "\dt"
```

### Debugging Connection Issues

```powershell
# Check all containers
docker ps | findstr gol-bazaar

# Check app logs
docker logs gol-bazaar-app-dev --tail 50

# Check database
docker exec gol-bazaar-postgres-dev pg_isready -U postgres

# Check Redis
docker exec gol-bazaar-redis-dev redis-cli ping

# Restart everything
docker-compose -f docker-compose.dev.yml restart
```

## 📊 Monitoring

```powershell
# Container stats
docker stats

# Disk usage
docker system df

# Network info
docker network inspect gol-bazaar-dev-network

# Volume info
docker volume ls | findstr gol-bazaar
```

## 🎯 Pro Tips

```powershell
# Keep logs open in separate terminal
docker logs gol-bazaar-app-dev -f

# Quick rebuild alias (add to PowerShell profile)
function Rebuild-GolBazaar { docker-compose -f docker-compose.dev.yml up -d --build app }

# Quick test alias
function Test-GolBazaar { curl http://localhost:8080/health }

# Watch logs with grep
docker logs gol-bazaar-app-dev -f | Select-String "ERROR"
```

## 🚨 Emergency Commands

```powershell
# Everything is broken - nuclear option
docker-compose -f docker-compose.dev.yml down -v
docker system prune -a -f
.\scripts\deploy-dev.ps1

# Port already in use
netstat -ano | findstr "8080"
# Kill the process using the port

# Container won't stop
docker kill gol-bazaar-app-dev
docker rm gol-bazaar-app-dev
```

## 📱 Mobile/Postman Testing

### Import to Postman

Create collection with these endpoints:

```
GET  http://localhost:8080/health
GET  http://localhost:8080/api/v1/stores
GET  http://localhost:8080/api/v1/stores/:id
POST http://localhost:8080/api/v1/stores
GET  http://localhost:8080/api/v1/products
GET  http://localhost:8080/api/v1/products/:id
```

---

**Keep this cheat sheet handy!** 📋

Print it out or keep it open in a browser tab for quick reference.
