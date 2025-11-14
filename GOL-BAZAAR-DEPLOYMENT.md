# 🛒 Gol Bazaar - Deployment Configuration

## Application Name: **Gol Bazaar**

A comprehensive grocery superapp platform with multi-store support, location-based features, and real-time inventory management.

## ✅ Updated Configuration

All Docker containers, networks, and volumes have been renamed to use **gol-bazaar** prefix:

### Development Environment

**Containers:**
- `gol-bazaar-app-dev` - Application server
- `gol-bazaar-postgres-dev` - PostgreSQL with PostGIS
- `gol-bazaar-redis-dev` - Redis cache
- `gol-bazaar-pgadmin-dev` - Database management UI
- `gol-bazaar-redis-commander-dev` - Redis management UI

**Network:** `gol-bazaar-dev-network`

**Volumes:**
- `gol-bazaar-postgres-dev-data`
- `gol-bazaar-redis-dev-data`
- `gol-bazaar-pgadmin-dev-data`

### Production Environment

**Containers:**
- `gol-bazaar-app-prod` - Application server (x2 replicas)
- `gol-bazaar-postgres-prod` - PostgreSQL with PostGIS
- `gol-bazaar-redis-prod` - Redis cache
- `gol-bazaar-nginx-prod` - Reverse proxy
- `gol-bazaar-backup-prod` - Automated backup service

**Network:** `gol-bazaar-prod-network`

**Volumes:**
- `gol-bazaar-postgres-prod-data`
- `gol-bazaar-redis-prod-data`
- `gol-bazaar-nginx-cache`

**Image:** `gol-bazaar-app:${BUILD_VERSION}`

## 🚀 Quick Start

### Development Deployment

```powershell
# Deploy Gol Bazaar development environment
.\scripts\deploy-dev.ps1
```

**Access:**
- Application: http://localhost:8080
- pgAdmin: http://localhost:5050 (admin@golbazaar.local/admin)
- Redis Commander: http://localhost:8081

### Production Deployment

```powershell
# 1. Update production credentials in .env.production
# 2. Deploy Gol Bazaar production environment
.\scripts\deploy-prod.ps1
```

**Access:**
- Application: http://localhost (via Nginx)

## 📊 Container Status

### Check Running Containers

```powershell
# Development
docker ps | findstr gol-bazaar

# Production
docker-compose -f docker-compose.prod.yml ps
```

### View Logs

```powershell
# Development
docker logs gol-bazaar-app-dev -f
docker logs gol-bazaar-postgres-dev -f
docker logs gol-bazaar-redis-dev -f

# Production
docker logs gol-bazaar-app-prod -f
docker logs gol-bazaar-postgres-prod -f
docker logs gol-bazaar-redis-prod -f
docker logs gol-bazaar-nginx-prod -f
```

## 🔧 Database Operations

### Development

```powershell
# Connect to PostgreSQL
docker exec -it gol-bazaar-postgres-dev psql -U postgres -d middleware_db

# Create backup
docker exec gol-bazaar-postgres-dev pg_dump -U postgres -d middleware_db > backup.sql

# Restore backup
Get-Content backup.sql | docker exec -i gol-bazaar-postgres-dev psql -U postgres -d middleware_db
```

### Production

```powershell
# Connect to PostgreSQL
docker exec -it gol-bazaar-postgres-prod psql -U gol_bazaar_app -d gol_bazaar_production

# Create backup
docker exec gol-bazaar-postgres-prod pg_dump -U gol_bazaar_app -d gol_bazaar_production -F c -f /backups/manual.dump

# Restore backup
docker exec -i gol-bazaar-postgres-prod pg_restore -U gol_bazaar_app -d gol_bazaar_production -c /backups/backup.dump
```

## 🔄 Redis Operations

### Development

```powershell
# Connect to Redis
docker exec -it gol-bazaar-redis-dev redis-cli

# Test connection
docker exec gol-bazaar-redis-dev redis-cli ping

# View keys
docker exec gol-bazaar-redis-dev redis-cli keys "*"
```

### Production

```powershell
# Connect to Redis (with password)
docker exec -it gol-bazaar-redis-prod redis-cli -a ${REDIS_PASSWORD}

# Test connection
docker exec gol-bazaar-redis-prod redis-cli -a ${REDIS_PASSWORD} ping

# View memory usage
docker exec gol-bazaar-redis-prod redis-cli -a ${REDIS_PASSWORD} INFO memory
```

## 📦 Image Management

### Build Images

```powershell
# Development
docker-compose -f docker-compose.dev.yml build

# Production with version
$env:BUILD_VERSION="1.0.0"
docker-compose -f docker-compose.prod.yml build
```

### Tag Images

```powershell
# Tag production image
docker tag gol-bazaar-app:latest gol-bazaar-app:1.0.0
docker tag gol-bazaar-app:latest gol-bazaar-app:stable
```

### Push to Registry (Optional)

```powershell
# Tag for registry
docker tag gol-bazaar-app:1.0.0 your-registry.com/gol-bazaar-app:1.0.0

# Push to registry
docker push your-registry.com/gol-bazaar-app:1.0.0
```

## 🔍 Health Checks

### Application Health

```powershell
# Development
curl http://localhost:8080/health

# Production
curl http://localhost/health
```

### Service Health

```powershell
# PostgreSQL
docker exec gol-bazaar-postgres-dev pg_isready -U postgres

# Redis
docker exec gol-bazaar-redis-dev redis-cli ping

# All containers
docker ps --filter "name=gol-bazaar" --filter "health=healthy"
```

## 🧹 Cleanup

### Stop Services

```powershell
# Development
docker-compose -f docker-compose.dev.yml down

# Production
docker-compose -f docker-compose.prod.yml down
```

### Remove Volumes (⚠️ Deletes all data)

```powershell
# Development
docker volume rm gol-bazaar-postgres-dev-data gol-bazaar-redis-dev-data gol-bazaar-pgadmin-dev-data

# Production
docker volume rm gol-bazaar-postgres-prod-data gol-bazaar-redis-prod-data gol-bazaar-nginx-cache
```

### Complete Cleanup

```powershell
# Stop and remove everything (development)
docker-compose -f docker-compose.dev.yml down -v --remove-orphans

# Stop and remove everything (production)
docker-compose -f docker-compose.prod.yml down -v --remove-orphans

# Remove images
docker rmi gol-bazaar-app:latest
docker rmi gol-bazaar-app:1.0.0
```

## 📈 Monitoring

### Container Stats

```powershell
# All containers
docker stats

# Specific containers
docker stats gol-bazaar-app-dev gol-bazaar-postgres-dev gol-bazaar-redis-dev
```

### Disk Usage

```powershell
# Docker system
docker system df

# Volumes
docker volume ls | findstr gol-bazaar
```

### Network Inspection

```powershell
# Development network
docker network inspect gol-bazaar-dev-network

# Production network
docker network inspect gol-bazaar-prod-network
```

## 🔐 Security

### Production Checklist

- [ ] Update `POSTGRES_PASSWORD` in `.env.production`
- [ ] Update `REDIS_PASSWORD` in `.env.production`
- [ ] Configure SSL certificates in `nginx/ssl/`
- [ ] Review firewall rules
- [ ] Enable audit logging
- [ ] Set up monitoring alerts
- [ ] Configure backup retention
- [ ] Test disaster recovery

## 📚 Documentation

- **DEPLOYMENT-GUIDE.md** - Complete deployment guide
- **ENVIRONMENT-COMPARISON.md** - Dev vs Prod comparison
- **DEPLOYMENT-READY.md** - Quick reference
- **GOL-BAZAAR-DEPLOYMENT.md** - This file

## 🎯 Next Steps

1. **Deploy Development**
   ```powershell
   .\scripts\deploy-dev.ps1
   ```

2. **Verify Services**
   ```powershell
   docker ps | findstr gol-bazaar
   curl http://localhost:8080/health
   ```

3. **Access Management UIs**
   - pgAdmin: http://localhost:5050
   - Redis Commander: http://localhost:8081

4. **Start Development**
   - Implement API endpoints
   - Add authentication
   - Build features

5. **Prepare Production**
   - Update `.env.production`
   - Configure SSL
   - Set up monitoring

6. **Deploy Production**
   ```powershell
   .\scripts\deploy-prod.ps1
   ```

---

**Gol Bazaar is ready to deploy!** 🛒🚀

All containers, networks, and volumes are properly named and configured.
