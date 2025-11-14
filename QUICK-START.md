# Quick Start Guide

## Your Development Environment is Ready! 🚀

You now have a complete development setup with:
- ✅ Go application with Gin framework
- ✅ Redis caching layer
- ✅ PostgreSQL database with sample data
- ✅ Supabase integration
- ✅ Dev container for VS Code

## 1. Test PostgreSQL Connection

```bash
# In dev container terminal
go run cmd/test-db/main.go
```

Expected output: Connection successful with sample data displayed.

## 2. Run the Application

### Option A: In Dev Container (Recommended for Development)

```bash
# Open in VS Code
# Press F1 → "Dev Containers: Reopen in Container"

# In the integrated terminal
go run cmd/server/main.go
```

VS Code will automatically forward port 8080.

### Option B: In Production Container (For Testing)

```bash
# Rebuild with latest code
docker-compose down
docker-compose up -d --build

# Access at http://localhost:8080
```

## 3. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Get supermarket products (placeholder)
curl http://localhost:8080/api/v1/supermarket/products

# Get movies (placeholder)
curl http://localhost:8080/api/v1/movies

# Get medicines (placeholder)
curl http://localhost:8080/api/v1/pharmacy/medicines
```

## 4. Access PostgreSQL

### Using psql

```bash
docker exec -it supabase-postgres-temp psql -U postgres -d middleware_db
```

### Using GUI Tool

- **Host:** localhost
- **Port:** 5432
- **Database:** middleware_db
- **Username:** postgres
- **Password:** postgres

### Sample Queries

```sql
-- View all products
SELECT * FROM supermarket_products;

-- View all movies
SELECT * FROM movies;

-- View medicines
SELECT * FROM medicines;

-- Check low stock items
SELECT * FROM low_stock_products;
```

## 5. Access Redis

```bash
# Connect to Redis CLI
docker exec -it supabase-redis-cache redis-cli

# Test commands
PING
KEYS *
```

## 6. View Logs

```bash
# Application logs
docker logs supabase-redis-middleware -f

# PostgreSQL logs
docker logs supabase-postgres-temp -f

# Redis logs
docker logs supabase-redis-cache -f
```

## 7. Stop Everything

```bash
# Stop all containers
docker-compose down

# Stop and remove volumes (deletes data)
docker-compose down -v
```

## Common Tasks

### Rebuild Application

```bash
docker-compose up -d --build app
```

### Restart Services

```bash
docker-compose restart
```

### Reset Database

```bash
# Reinitialize with sample data
.\scripts\init-db.ps1
```

### Update Dependencies

```bash
# In dev container
go mod download
go mod tidy
```

## Environment Variables

Edit `.env` file to configure:

```env
# Server
SERVER_PORT=8080

# Supabase (replace with your credentials)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_API_KEY=your-api-key

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_TTL=300s

# PostgreSQL
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/middleware_db?sslmode=disable

# Logging
LOG_LEVEL=debug
```

## Development Workflow

### 1. Make Code Changes

Edit files in VS Code (either locally or in dev container).

### 2. Test Changes

**In Dev Container:**
```bash
go run cmd/server/main.go
# Changes are instant!
```

**In Production Container:**
```bash
docker-compose up -d --build
# Rebuilds with your changes
```

### 3. Run Tests

```bash
go test ./...
```

### 4. Check for Issues

```bash
# Format code
go fmt ./...

# Run linter (if installed)
golangci-lint run
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using the port
netstat -ano | findstr "8080"

# Change port in .env
SERVER_PORT=8081
```

### Permission Denied (Dev Container)

```bash
# Fix Go module permissions
docker exec -u root devcontainer-app-1 chown -R vscode:vscode /go
```

### Database Connection Failed

```bash
# Check if PostgreSQL is running
docker ps | findstr postgres

# Test connection
docker exec supabase-postgres-temp pg_isready -U postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Redis Connection Failed

```bash
# Check if Redis is running
docker ps | findstr redis

# Test connection
docker exec supabase-redis-cache redis-cli ping

# Restart Redis
docker-compose restart redis
```

## Next Steps

1. **Implement Real Handlers**
   - Replace placeholder handlers with actual PostgreSQL queries
   - Add caching layer for frequently accessed data

2. **Add Authentication**
   - Implement JWT authentication
   - Add user management

3. **Add More Endpoints**
   - CRUD operations for products, movies, medicines
   - Search and filtering
   - Pagination

4. **Add Tests**
   - Unit tests for handlers
   - Integration tests with test database
   - Load testing

5. **Deploy to Production**
   - Use managed PostgreSQL (Supabase, AWS RDS, etc.)
   - Use managed Redis (AWS ElastiCache, Redis Cloud, etc.)
   - Deploy application to cloud (AWS, GCP, Azure, etc.)

## Useful Commands

```bash
# View all containers
docker ps -a

# View all images
docker images

# Clean up unused resources
docker system prune

# View container resource usage
docker stats

# Execute command in container
docker exec -it <container-name> <command>

# Copy file from container
docker cp <container-name>:/path/to/file ./local/path

# View container environment variables
docker exec <container-name> env
```

## Documentation

- `README.md` - Main project documentation
- `POSTGRES-SETUP.md` - PostgreSQL setup guide
- `PGX-SETUP-GUIDE.md` - pgx driver guide
- `DEVELOPMENT-WORKFLOW.md` - Development workflow
- `.devcontainer/HOW-IT-WORKS.md` - Dev container guide
- `.devcontainer/TROUBLESHOOTING.md` - Troubleshooting guide

## Support

If you encounter issues:

1. Check the logs: `docker logs <container-name>`
2. Verify containers are running: `docker ps`
3. Check the troubleshooting guides
4. Restart containers: `docker-compose restart`
5. Rebuild from scratch: `docker-compose down -v && docker-compose up -d --build`

---

**Everything is ready! Start coding!** 🎉

Your API is running at: http://localhost:8080
