# Memory Optimization Guide for 1 GB RAM

## Problem: docker-compose.prod.yml is TOO LARGE

Your `docker-compose.prod.yml` requires **~4.5 GB RAM** but you only have **1 GB RAM** on VM.Standard.E2.1.Micro.

### Memory Breakdown (docker-compose.prod.yml)

| Service | Memory Limit | Issue |
|---------|--------------|-------|
| App (2 replicas) | 2 × 512 MB = 1024 MB | ❌ Exceeds total RAM |
| PostgreSQL | 2048 MB | ❌ Way too much |
| Redis | 768 MB | ❌ Too much |
| Nginx | 256 MB | ⚠️ Not needed initially |
| Backup | 100 MB | ⚠️ Can run manually |
| System | ~300 MB | Required |
| **Total** | **~4.5 GB** | ❌ **4.5x over budget!** |

## Solution: docker-compose.lowmem.yml

I've created an optimized configuration for 1 GB RAM:

### Memory Breakdown (docker-compose.lowmem.yml)

| Service | Memory Limit | Percentage |
|---------|--------------|------------|
| App | 200 MB | 20% |
| PostgreSQL | 384 MB | 38% |
| Redis | 96 MB | 10% |
| System | ~300 MB | 30% |
| **Total** | **~980 MB** | ✅ **Fits in 1 GB!** |

## Key Optimizations

### 1. Application (Go)
```yaml
mem_limit: 200m        # Down from 512 MB
mem_reservation: 128m
cpus: 0.5
# Removed: Second replica (was using 512 MB extra)
```

**Changes:**
- Single instance instead of 2 replicas
- Reduced memory limit
- Healthcheck interval increased to 60s

### 2. PostgreSQL
```yaml
mem_limit: 384m        # Down from 2048 MB
shared_buffers: 64MB   # Down from 512 MB
effective_cache_size: 192MB  # Down from 2 GB
max_connections: 50    # Down from 200
```

**Changes:**
- Drastically reduced buffer sizes
- Fewer connections allowed
- Minimal WAL settings
- Reduced worker processes

### 3. Redis
```yaml
mem_limit: 96m         # Down from 768 MB
maxmemory: 64mb        # Down from 512 MB
save: ""               # Disabled persistence
appendonly: no         # Disabled AOF
```

**Changes:**
- Minimal memory allocation
- No persistence (cache only)
- Reduced max clients to 100

### 4. Removed Services
- ❌ Nginx (use direct access initially)
- ❌ Backup service (run manually)

## Usage

### Deploy with Low-Memory Config

```bash
# On your OCI instance
cd /opt/gol-backend
docker-compose -f docker-compose.lowmem.yml up -d
```

### Check Memory Usage

```bash
# Monitor container memory
docker stats

# Check system memory
free -h

# Check swap usage
swapon --show
```

## Performance Expectations

### What Works Well ✅
- Development and testing
- Low to moderate traffic (10-50 concurrent users)
- Simple API queries
- Basic CRUD operations
- Cache hits (Redis)

### What May Struggle ⚠️
- High concurrent connections (>50 users)
- Complex database queries
- Large data imports
- Heavy cache misses
- Multiple simultaneous operations

### What Won't Work ❌
- Production high-traffic workloads
- Large file uploads
- Complex analytics queries
- Multiple replicas
- Heavy background jobs

## Monitoring

### Watch for These Signs

```bash
# High memory usage (>90%)
free -h | grep Mem

# Swap being used heavily
vmstat 1 5

# Containers being OOM killed
dmesg | grep -i "out of memory"

# Container restarts
docker ps -a | grep "Restarting"
```

### If Memory Issues Occur

1. **Check container stats:**
   ```bash
   docker stats --no-stream
   ```

2. **Restart services:**
   ```bash
   docker-compose -f docker-compose.lowmem.yml restart
   ```

3. **Clear caches:**
   ```bash
   # Clear Redis
   docker exec gol-backend-redis redis-cli FLUSHALL
   
   # Clear system cache
   sudo sync && sudo sysctl -w vm.drop_caches=3
   ```

## Alternative Strategies

### Option 1: Use External Services (Recommended)

Instead of running everything locally, use cloud services:

#### A. Use Supabase PostgreSQL (Already Have It!)
```yaml
# Remove postgres service from docker-compose
# Update app environment:
environment:
  - DATABASE_URL=postgresql://user:pass@db.supabase.co:5432/postgres
```

**Saves:** ~384 MB

#### B. Use Redis Cloud (Free Tier)
```yaml
# Remove redis service
# Update app environment:
environment:
  - REDIS_HOST=redis-xxxxx.cloud.redislabs.com
  - REDIS_PORT=12345
  - REDIS_PASSWORD=your-password
```

**Saves:** ~96 MB

#### C. Both External Services
```yaml
# Only run the app container
services:
  app:
    # ... app config only
```

**Total Memory:** ~200 MB for app + 300 MB system = **500 MB total!**

### Option 2: Upgrade to A1.Flex (Still Free!)

Switch to VM.Standard.A1.Flex with 6 GB RAM:

```hcl
# In terraform main.tf
shape = "VM.Standard.A1.Flex"
shape_config {
  ocpus         = 1
  memory_in_gbs = 6
}
```

**Benefits:**
- 6x more RAM
- Can use docker-compose.prod.yml
- Better performance
- Still completely FREE!

### Option 3: Minimal Setup

Run only the app, use Supabase for everything:

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SUPABASE_URL=${SUPABASE_URL}
      - SUPABASE_API_KEY=${SUPABASE_API_KEY}
      # No Redis, no local PostgreSQL
    mem_limit: 256m
    restart: unless-stopped
```

**Total Memory:** ~256 MB + 300 MB system = **556 MB total**

## Recommended Approach

### For Development/Testing (Current Setup)
✅ Use `docker-compose.lowmem.yml`
- All services local
- Good for testing
- Acceptable performance

### For Production (Recommended)
✅ Use external services:
1. **PostgreSQL**: Use Supabase (you already have it!)
2. **Redis**: Use Redis Cloud free tier (30 MB)
3. **App**: Run locally

**Or** upgrade to A1.Flex (6 GB RAM, still free)

## Configuration Files

### docker-compose.lowmem.yml (Created)
- Optimized for 1 GB RAM
- All services included
- Minimal memory footprint
- Use this for E2.1.Micro

### docker-compose.prod.yml (Original)
- Requires 4-6 GB RAM
- Production-grade settings
- Use with A1.Flex or larger instances

### docker-compose.yml (Original)
- Development settings
- Moderate memory usage (~2 GB)
- Good for local development

## Memory Limits Comparison

| Configuration | App | PostgreSQL | Redis | Total |
|---------------|-----|------------|-------|-------|
| **prod.yml** | 1024 MB | 2048 MB | 768 MB | ~4.5 GB |
| **lowmem.yml** | 200 MB | 384 MB | 96 MB | ~980 MB |
| **External Services** | 200 MB | 0 MB | 0 MB | ~500 MB |

## Testing Your Setup

After deployment, verify memory usage:

```bash
# SSH into instance
ssh opc@YOUR_IP

# Check total memory usage
free -h

# Check container memory
docker stats --no-stream

# Check for OOM kills
dmesg | grep -i oom

# Monitor for 5 minutes
watch -n 10 'free -h && echo "---" && docker stats --no-stream'
```

### Expected Results

With `docker-compose.lowmem.yml`:
```
              total        used        free
Mem:          1.0Gi       850Mi       150Mi
Swap:         2.0Gi        50Mi       1.9Gi
```

If swap usage is high (>500 MB), consider using external services.

## Conclusion

**For VM.Standard.E2.1.Micro (1 GB RAM):**

1. ✅ **Use docker-compose.lowmem.yml** (created for you)
2. ⚠️ **DO NOT use docker-compose.prod.yml** (requires 4+ GB)
3. 💡 **Consider external services** for better performance
4. 🚀 **Or upgrade to A1.Flex** (6 GB RAM, still free!)

The low-memory configuration will work, but performance will be limited. For production workloads, external services or A1.Flex are recommended.
