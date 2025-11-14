# Grocery Superapp - Deployment Guide

## Overview

This guide covers deploying the Grocery Superapp in both development and production environments using Docker Compose.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Development                          │
├─────────────────────────────────────────────────────────┤
│  App (8080) │ PostgreSQL (5432) │ Redis (6379)         │
│  pgAdmin (5050) │ Redis Commander (8081)                │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                    Production                           │
├─────────────────────────────────────────────────────────┤
│  Nginx (80/443) → App (x2) → PostgreSQL │ Redis         │
│  Backup Service │ Monitoring                            │
└─────────────────────────────────────────────────────────┘
```

## Files Structure

```
.
├── docker-compose.dev.yml          # Development configuration
├── docker-compose.prod.yml         # Production configuration
├── .env.development                # Development environment variables
├── .env.production                 # Production environment variables
├── nginx/
│   ├── nginx.conf                  # Nginx main configuration
│   └── conf.d/
│       └── grocery-app.conf        # App-specific Nginx config
├── scripts/
│   ├── deploy-dev.ps1              # Development deployment script
│   ├── deploy-prod.ps1             # Production deployment script
│   └── backup.sh                   # Database backup script
└── grocery_superapp_schema.sql     # Database schema
```

## Development Deployment

### Prerequisites

- Docker Desktop installed and running
- PowerShell (Windows) or Bash (Linux/Mac)
- At least 4GB RAM available
- Ports 8080, 5432, 6379, 5050, 8081 available

### Quick Start

```powershell
# Deploy development environment
.\scripts\deploy-dev.ps1
```

### Manual Deployment

```powershell
# 1. Copy environment file
cp .env.development .env

# 2. Start services
docker-compose -f docker-compose.dev.yml up -d --build

# 3. Check status
docker-compose -f docker-compose.dev.yml ps

# 4. View logs
docker-compose -f docker-compose.dev.yml logs -f
```

### Development Services

| Service | URL | Credentials |
|---------|-----|-------------|
| **Application** | http://localhost:8080 | - |
| **Health Check** | http://localhost:8080/health | - |
| **PostgreSQL** | localhost:5432 | postgres/postgres |
| **Redis** | localhost:6379 | No password |
| **pgAdmin** | http://localhost:5050 | admin@grocery.local/admin |
| **Redis Commander** | http://localhost:8081 | - |

### Development Features

✅ **Hot Reload**: Source code mounted as volume  
✅ **Debug Logging**: Verbose logs enabled  
✅ **Management UIs**: pgAdmin and Redis Commander  
✅ **Exposed Ports**: All services accessible from host  
✅ **Sample Data**: Auto-loaded on first start  
✅ **No Authentication**: Redis without password  

### Development Commands

```powershell
# Start services
docker-compose -f docker-compose.dev.yml up -d

# Stop services
docker-compose -f docker-compose.dev.yml down

# Restart a service
docker-compose -f docker-compose.dev.yml restart app

# View logs
docker-compose -f docker-compose.dev.yml logs -f app

# Execute command in container
docker-compose -f docker-compose.dev.yml exec app sh

# Access PostgreSQL
docker-compose -f docker-compose.dev.yml exec postgres psql -U postgres -d middleware_db

# Access Redis
docker-compose -f docker-compose.dev.yml exec redis redis-cli

# Rebuild without cache
docker-compose -f docker-compose.dev.yml build --no-cache

# Clean up everything
docker-compose -f docker-compose.dev.yml down -v
```

## Production Deployment

### Prerequisites

- Docker Engine 20.10+ or Docker Desktop
- Docker Compose 2.0+
- SSL certificates (for HTTPS)
- Strong passwords configured
- Firewall rules configured
- Monitoring setup
- Backup strategy in place

### Pre-Deployment Checklist

- [ ] Update `.env.production` with production credentials
- [ ] Generate strong passwords for PostgreSQL and Redis
- [ ] Configure SSL certificates in `nginx/ssl/`
- [ ] Review and adjust resource limits
- [ ] Set up external database backups
- [ ] Configure monitoring and alerting
- [ ] Test deployment in staging environment
- [ ] Prepare rollback plan
- [ ] Document deployment procedure
- [ ] Notify team about deployment

### Production Deployment Steps

```powershell
# Deploy production environment
.\scripts\deploy-prod.ps1
```

### Manual Production Deployment

```powershell
# 1. Copy production environment file
cp .env.production .env

# 2. Set build version
$env:BUILD_VERSION="1.0.0"

# 3. Build images
docker-compose -f docker-compose.prod.yml build --no-cache

# 4. Start services
docker-compose -f docker-compose.prod.yml up -d

# 5. Verify health
curl http://localhost/health

# 6. Check logs
docker-compose -f docker-compose.prod.yml logs -f
```

### Production Services

| Service | Internal Port | External Access | Replicas |
|---------|---------------|-----------------|----------|
| **Nginx** | 80, 443 | Public | 1 |
| **Application** | 8080 | Via Nginx only | 2 |
| **PostgreSQL** | 5432 | Internal only | 1 |
| **Redis** | 6379 | Internal only | 1 |
| **Backup** | - | Internal only | 1 |

### Production Features

✅ **High Availability**: 2 app replicas with load balancing  
✅ **Reverse Proxy**: Nginx with SSL/TLS termination  
✅ **Security**: No exposed database ports, strong passwords  
✅ **Auto-Restart**: Containers restart on failure  
✅ **Resource Limits**: CPU and memory limits enforced  
✅ **Logging**: Structured logs with rotation  
✅ **Backups**: Automated daily database backups  
✅ **Health Checks**: All services monitored  
✅ **Rate Limiting**: API rate limits configured  

### Production Commands

```powershell
# Start services
docker-compose -f docker-compose.prod.yml up -d

# Stop services (graceful shutdown)
docker-compose -f docker-compose.prod.yml down

# Scale application
docker-compose -f docker-compose.prod.yml up -d --scale app=3

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Monitor resources
docker stats

# Create manual backup
docker exec grocery-postgres-prod pg_dump -U grocery_app -d grocery_production -F c -f /backups/manual_backup.dump

# Restore from backup
docker exec -i grocery-postgres-prod pg_restore -U grocery_app -d grocery_production -c /backups/backup_file.dump

# Update application (zero-downtime)
docker-compose -f docker-compose.prod.yml up -d --no-deps --build app

# Rollback to previous version
docker-compose -f docker-compose.prod.yml up -d --no-deps app:previous-version
```

## Environment Synchronization

### Development to Production Sync

1. **Database Schema**
   ```powershell
   # Export from development
   docker exec grocery-postgres-dev pg_dump -U postgres -d middleware_db -s > schema.sql
   
   # Import to production
   Get-Content schema.sql | docker exec -i grocery-postgres-prod psql -U grocery_app -d grocery_production
   ```

2. **Configuration**
   - Review `.env.development` changes
   - Update `.env.production` accordingly
   - Never copy passwords directly

3. **Code Deployment**
   ```powershell
   # Build with version tag
   $env:BUILD_VERSION="1.1.0"
   docker-compose -f docker-compose.prod.yml build
   
   # Deploy new version
   docker-compose -f docker-compose.prod.yml up -d --no-deps app
   ```

4. **Data Migration**
   ```powershell
   # Export data from development
   docker exec grocery-postgres-dev pg_dump -U postgres -d middleware_db -a > data.sql
   
   # Review and sanitize data.sql (remove test data, update IDs, etc.)
   
   # Import to production
   Get-Content data.sql | docker exec -i grocery-postgres-prod psql -U grocery_app -d grocery_production
   ```

## Monitoring

### Health Checks

```powershell
# Application health
curl http://localhost/health

# PostgreSQL health
docker exec grocery-postgres-prod pg_isready -U grocery_app

# Redis health
docker exec grocery-redis-prod redis-cli --pass $REDIS_PASSWORD ping

# Container health
docker ps --filter "health=unhealthy"
```

### Logs

```powershell
# Application logs
docker logs grocery-app-prod -f --tail 100

# PostgreSQL logs
docker logs grocery-postgres-prod -f --tail 100

# Nginx logs
docker logs grocery-nginx-prod -f --tail 100

# All services
docker-compose -f docker-compose.prod.yml logs -f
```

### Metrics

```powershell
# Container stats
docker stats

# Disk usage
docker system df

# Network usage
docker network inspect grocery-prod-network
```

## Backup and Restore

### Automated Backups

Backups run daily at midnight (configured in production compose file).

**Backup Location**: `./backups/`  
**Retention**: 7 days, 4 weeks, 6 months  

### Manual Backup

```powershell
# Full database backup
docker exec grocery-postgres-prod pg_dump -U grocery_app -d grocery_production -F c -f /backups/manual_$(date +%Y%m%d).dump

# Schema only
docker exec grocery-postgres-prod pg_dump -U grocery_app -d grocery_production -s > schema_backup.sql

# Data only
docker exec grocery-postgres-prod pg_dump -U grocery_app -d grocery_production -a > data_backup.sql

# Specific table
docker exec grocery-postgres-prod pg_dump -U grocery_app -d grocery_production -t users > users_backup.sql
```

### Restore

```powershell
# Restore full backup
docker exec -i grocery-postgres-prod pg_restore -U grocery_app -d grocery_production -c /backups/backup_file.dump

# Restore from SQL file
Get-Content backup.sql | docker exec -i grocery-postgres-prod psql -U grocery_app -d grocery_production
```

## Troubleshooting

### Application Won't Start

```powershell
# Check logs
docker logs grocery-app-prod --tail 100

# Check environment variables
docker exec grocery-app-prod env

# Verify database connection
docker exec grocery-app-prod wget -O- http://localhost:8080/health
```

### Database Connection Issues

```powershell
# Check if PostgreSQL is running
docker ps | findstr postgres

# Test connection
docker exec grocery-postgres-prod pg_isready -U grocery_app

# Check logs
docker logs grocery-postgres-prod --tail 100

# Verify network
docker network inspect grocery-prod-network
```

### High Memory Usage

```powershell
# Check container stats
docker stats

# Adjust PostgreSQL memory
# Edit docker-compose.prod.yml: shared_buffers, effective_cache_size

# Adjust Redis memory
# Edit docker-compose.prod.yml: maxmemory

# Restart services
docker-compose -f docker-compose.prod.yml restart
```

### Slow Performance

```powershell
# Check database queries
docker exec grocery-postgres-prod psql -U grocery_app -d grocery_production -c "SELECT * FROM pg_stat_activity;"

# Check Redis memory
docker exec grocery-redis-prod redis-cli --pass $REDIS_PASSWORD INFO memory

# Check Nginx logs for slow requests
docker logs grocery-nginx-prod | grep "request_time"

# Analyze database performance
docker exec grocery-postgres-prod psql -U grocery_app -d grocery_production -c "SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"
```

## Security Best Practices

### Production Security Checklist

- [ ] Strong passwords for all services
- [ ] SSL/TLS certificates configured
- [ ] Database ports not exposed externally
- [ ] Redis password authentication enabled
- [ ] Firewall rules configured
- [ ] Regular security updates
- [ ] Audit logging enabled
- [ ] Secrets management (not in .env files)
- [ ] Network segmentation
- [ ] Regular backups tested
- [ ] Monitoring and alerting configured
- [ ] Incident response plan documented

### Recommended Tools

- **Secrets Management**: HashiCorp Vault, AWS Secrets Manager
- **Monitoring**: Prometheus + Grafana, Datadog
- **Logging**: ELK Stack, Splunk
- **Security Scanning**: Trivy, Clair
- **Load Balancing**: AWS ALB, Nginx Plus
- **CDN**: CloudFlare, AWS CloudFront

## Scaling

### Horizontal Scaling

```powershell
# Scale application instances
docker-compose -f docker-compose.prod.yml up -d --scale app=4

# Verify scaling
docker-compose -f docker-compose.prod.yml ps
```

### Vertical Scaling

Edit `docker-compose.prod.yml` resource limits:

```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 1G
```

## Migration from Development to Production

1. **Test in Staging**: Always test in staging first
2. **Backup Production**: Create full backup before migration
3. **Schema Migration**: Apply schema changes
4. **Data Migration**: Migrate data if needed
5. **Configuration**: Update production config
6. **Deploy**: Deploy new version
7. **Verify**: Run health checks and tests
8. **Monitor**: Watch logs and metrics
9. **Rollback Plan**: Be ready to rollback if issues occur

## Support

For issues or questions:
- Check logs: `docker-compose logs -f`
- Review documentation
- Check GitHub issues
- Contact DevOps team

---

**Remember**: Always test deployments in a staging environment before production!
