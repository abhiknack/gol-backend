# Development vs Production Environment Comparison

## Quick Reference

| Feature | Development | Production |
|---------|-------------|------------|
| **File** | `docker-compose.dev.yml` | `docker-compose.prod.yml` |
| **Command** | `.\scripts\deploy-dev.ps1` | `.\scripts\deploy-prod.ps1` |
| **App Replicas** | 1 | 2 (load balanced) |
| **Exposed Ports** | All (8080, 5432, 6379) | Only 80/443 (via Nginx) |
| **Log Level** | debug | info |
| **Redis Password** | None | Required |
| **SSL/TLS** | No | Yes (Nginx) |
| **Management UIs** | pgAdmin, Redis Commander | None |
| **Auto Backup** | No | Yes (daily) |
| **Resource Limits** | No | Yes |
| **Health Checks** | Basic | Comprehensive |
| **Restart Policy** | unless-stopped | always |
| **Hot Reload** | Yes (volume mount) | No |

## Detailed Comparison

### Application Service

#### Development
```yaml
app:
  container_name: grocery-app-dev
  ports: ["8080:8080"]  # Exposed
  environment:
    LOG_LEVEL: debug
    GIN_MODE: debug
  volumes:
    - .:/app:cached  # Source code mounted
  restart: unless-stopped
```

#### Production
```yaml
app:
  container_name: grocery-app-prod
  # No direct port exposure (via Nginx)
  environment:
    LOG_LEVEL: info
    GIN_MODE: release
  deploy:
    replicas: 2
    resources:
      limits: {cpus: '1.0', memory: 512M}
  restart: always
  read_only: true
  security_opt: [no-new-privileges:true]
```

### PostgreSQL Service

#### Development
```yaml
postgres:
  image: postgis/postgis:16-3.4-alpine
  container_name: grocery-postgres-dev
  ports: ["5432:5432"]  # Exposed for tools
  environment:
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: postgres
  volumes:
    - postgres-data-dev:/var/lib/postgresql/data
    - ./grocery_superapp_schema.sql:/docker-entrypoint-initdb.d/01-schema.sql:ro
    - ./scripts/init-sample-data.sql:/docker-entrypoint-initdb.d/02-sample-data.sql:ro
  command: postgres -c max_connections=100
```

#### Production
```yaml
postgres:
  image: postgis/postgis:16-3.4-alpine
  container_name: grocery-postgres-prod
  # No exposed ports (internal only)
  environment:
    POSTGRES_USER: grocery_app
    POSTGRES_PASSWORD: ${STRONG_PASSWORD}
  volumes:
    - postgres-data-prod:/var/lib/postgresql/data
    - ./grocery_superapp_schema.sql:/docker-entrypoint-initdb.d/01-schema.sql:ro
    - ./backups:/backups
  command: postgres -c max_connections=200 -c shared_buffers=512MB
  deploy:
    resources:
      limits: {cpus: '2.0', memory: 2G}
  security_opt: [no-new-privileges:true]
```

### Redis Service

#### Development
```yaml
redis:
  image: redis:7-alpine
  container_name: grocery-redis-dev
  ports: ["6379:6379"]  # Exposed for tools
  command: redis-server --appendonly yes --maxmemory 256mb
  # No password
```

#### Production
```yaml
redis:
  image: redis:7-alpine
  container_name: grocery-redis-prod
  # No exposed ports (internal only)
  command: >
    redis-server
    --requirepass ${REDIS_PASSWORD}
    --appendonly yes
    --maxmemory 512mb
    --maxmemory-policy allkeys-lru
  deploy:
    resources:
      limits: {cpus: '1.0', memory: 768M}
  security_opt: [no-new-privileges:true]
```

### Additional Services

#### Development Only
- **pgAdmin** (port 5050) - PostgreSQL management UI
- **Redis Commander** (port 8081) - Redis management UI

#### Production Only
- **Nginx** (ports 80, 443) - Reverse proxy with SSL/TLS
- **Backup Service** - Automated daily database backups

## Network Configuration

### Development
```yaml
networks:
  app-network:
    driver: bridge
    name: grocery-dev-network
```

### Production
```yaml
networks:
  app-network:
    driver: bridge
    name: grocery-prod-network
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## Volume Configuration

### Development
```yaml
volumes:
  postgres-data-dev:
    name: grocery-postgres-dev-data
  redis-data-dev:
    name: grocery-redis-dev-data
  pgadmin-data-dev:
    name: grocery-pgadmin-dev-data
```

### Production
```yaml
volumes:
  postgres-data-prod:
    name: grocery-postgres-prod-data
  redis-data-prod:
    name: grocery-redis-prod-data
  nginx-cache:
    name: grocery-nginx-cache
```

## Environment Variables

### Development (.env.development)
```env
# Weak passwords OK for development
POSTGRES_PASSWORD=postgres
REDIS_PASSWORD=

# Verbose logging
LOG_LEVEL=debug

# Development URLs
SUPABASE_URL=https://dev-project.supabase.co
```

### Production (.env.production)
```env
# Strong passwords required
POSTGRES_PASSWORD=CHANGE_THIS_TO_STRONG_PASSWORD
REDIS_PASSWORD=CHANGE_THIS_TO_STRONG_REDIS_PASSWORD

# Production logging
LOG_LEVEL=info

# Production URLs
SUPABASE_URL=https://prod-project.supabase.co
```

## Resource Usage

### Development
- **CPU**: No limits
- **Memory**: No limits
- **Disk**: ~2GB (with sample data)
- **Network**: Bridge network

### Production
- **CPU**: Limited per service
  - App: 1.0 CPU per replica
  - PostgreSQL: 2.0 CPU
  - Redis: 1.0 CPU
- **Memory**: Limited per service
  - App: 512MB per replica
  - PostgreSQL: 2GB
  - Redis: 768MB
- **Disk**: ~10GB+ (with backups)
- **Network**: Isolated subnet

## Security Features

### Development
- ❌ No SSL/TLS
- ❌ Weak passwords
- ❌ Exposed database ports
- ❌ No rate limiting
- ✅ Easy debugging
- ✅ Management UIs

### Production
- ✅ SSL/TLS via Nginx
- ✅ Strong passwords required
- ✅ No exposed database ports
- ✅ Rate limiting configured
- ✅ Security headers
- ✅ Read-only containers
- ✅ Resource limits
- ✅ Automated backups
- ✅ Health checks
- ✅ Logging and monitoring

## Performance Tuning

### Development
- Optimized for fast iteration
- No connection pooling limits
- Minimal caching
- Debug logging (slower)

### Production
- Optimized for throughput
- Connection pooling configured
- Aggressive caching
- Minimal logging (faster)
- Load balancing (2+ replicas)
- CDN integration ready

## Deployment Process

### Development
```powershell
# Simple one-command deployment
.\scripts\deploy-dev.ps1

# Or manually
docker-compose -f docker-compose.dev.yml up -d --build
```

### Production
```powershell
# Guided deployment with checks
.\scripts\deploy-prod.ps1

# Or manually with version tagging
$env:BUILD_VERSION="1.0.0"
docker-compose -f docker-compose.prod.yml build
docker-compose -f docker-compose.prod.yml up -d
```

## Monitoring

### Development
- Docker logs
- pgAdmin for database
- Redis Commander for cache
- Manual health checks

### Production
- Structured logging
- Health check endpoints
- Resource monitoring
- Automated alerts
- Backup verification
- Performance metrics

## Cost Considerations

### Development
- **Infrastructure**: Local machine
- **Cost**: $0 (local resources)
- **Scaling**: Limited by local resources

### Production
- **Infrastructure**: Cloud provider (AWS, GCP, Azure)
- **Cost**: ~$50-200/month (depending on scale)
  - Compute: $30-100
  - Database: $10-50
  - Storage: $5-20
  - Network: $5-30
- **Scaling**: Horizontal and vertical

## When to Use Each

### Use Development When:
- ✅ Local development
- ✅ Testing new features
- ✅ Debugging issues
- ✅ Learning the system
- ✅ Running integration tests
- ✅ Database schema changes

### Use Production When:
- ✅ Serving real users
- ✅ Handling production traffic
- ✅ Storing real data
- ✅ Requiring high availability
- ✅ Need security compliance
- ✅ Performance is critical

## Migration Path

### Development → Staging → Production

1. **Develop** in development environment
2. **Test** in staging (production-like)
3. **Deploy** to production
4. **Monitor** and iterate

### Sync Strategy

```
Development (Local)
    ↓ (git push)
CI/CD Pipeline
    ↓ (build & test)
Staging Environment
    ↓ (manual approval)
Production Environment
```

## Summary

| Aspect | Development | Production |
|--------|-------------|------------|
| **Purpose** | Fast iteration | Stable service |
| **Security** | Relaxed | Strict |
| **Performance** | Good enough | Optimized |
| **Monitoring** | Basic | Comprehensive |
| **Cost** | Free (local) | Paid (cloud) |
| **Complexity** | Simple | Complex |
| **Reliability** | Can break | Must not break |

---

**Choose the right environment for your needs!**
