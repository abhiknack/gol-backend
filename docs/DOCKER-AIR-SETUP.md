# Docker + Air Hot Reload Setup

## Overview

Your Go application is fully Dockerized with Air for automatic hot reloading during development. Any code changes trigger an instant rebuild without manual intervention.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Host Machine (Windows)                                 │
│  ├─ Source Code (D:\Gol-Backend)                       │
│  └─ gol.bat (Management Script)                        │
└─────────────────────────────────────────────────────────┘
                        ↓ Volume Mount
┌─────────────────────────────────────────────────────────┐
│  Docker Container (gol-bazaar-app-dev)                  │
│  ├─ Air (Hot Reload Tool)                              │
│  ├─ Watches: *.go, *.yaml, *.yml, *.env               │
│  └─ Auto-rebuilds on file changes                      │
└─────────────────────────────────────────────────────────┘
```

## Key Files

### 1. Dockerfile.dev

Development Dockerfile with Air installed:

```dockerfile
FROM golang:1.23-alpine

# Install Air for hot reload
RUN go install github.com/air-verse/air@v1.52.3

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

EXPOSE 8080
CMD ["air", "-c", ".air.toml"]
```

### 2. .air.toml

Air configuration for watching and rebuilding:

```toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  bin = "./tmp/main"
  include_ext = ["go", "yaml", "yml", "env"]
  exclude_dir = ["tmp", "vendor", "node_modules", "migrations"]
  poll = true
  poll_interval = 500
  stop_on_error = true
```

### 3. docker-compose.dev.yml

Development environment with volume mounts:

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app:cached        # Mount source code
      - /app/tmp             # Exclude build artifacts
    ports:
      - "8080:8080"
```

### 4. gol.bat

Management script for common tasks:

```batch
gol start    # Start all services
gol stop     # Stop all services
gol rebuild  # Rebuild and restart
gol reload   # Trigger manual reload
gol logs     # View application logs
gol air      # Check Air status
gol status   # Show service status
```

## How It Works

### 1. Volume Mounting

Your source code is mounted into the container:

```yaml
volumes:
  - .:/app:cached
```

This allows Air inside the container to see file changes on your host machine.

### 2. Air Watching

Air watches for changes to:
- `*.go` files
- `*.yaml` files
- `*.yml` files
- `*.env` files

When a change is detected:
1. Air stops the running application
2. Rebuilds: `go build -o ./tmp/main ./cmd/server`
3. Restarts: `./tmp/main`

### 3. Automatic Rebuild

```
Edit file → Save → Air detects → Rebuild → Restart
  (0s)      (0s)     (~100ms)     (~2s)     (~1s)
```

Total time: ~3 seconds from save to running

## Usage

### Start Development Environment

```bash
gol start
```

This starts:
- Go application with Air (port 8080)
- PostgreSQL with PostGIS (port 5432)
- Redis (port 6379)
- pgAdmin (port 5050)
- Redis Commander (port 8081)

### Make Code Changes

1. Edit any `.go` file
2. Save the file
3. Air automatically detects and rebuilds
4. Check logs: `gol logs`

### View Rebuild Process

```bash
gol logs
```

You'll see:
```
cmd/server/main.go has changed
building...
running...
Server started successfully
```

### Manual Reload

If Air doesn't detect a change:

```bash
gol reload
```

This touches `main.go` to trigger a rebuild.

### Check Air Status

```bash
gol air
```

Shows if Air is running and recent activity.

### Rebuild After Dependency Changes

If you modify `go.mod` or `go.sum`:

```bash
gol rebuild
```

This rebuilds the Docker image with new dependencies.

## Troubleshooting

### Air Not Detecting Changes

**Problem:** File changes don't trigger rebuild

**Solution:**
```bash
# Check if Air is running
gol air

# Manually trigger reload
gol reload

# Restart services
gol restart
```

### Build Errors

**Problem:** Air shows build errors

**Solution:**
```bash
# View full error logs
gol logs

# Fix the code error
# Air will auto-rebuild on next save
```

### Container Not Starting

**Problem:** `gol start` fails

**Solution:**
```bash
# Check Docker Desktop is running
# View container status
gol status

# Rebuild from scratch
gol rebuild
```

### Slow Rebuilds

**Problem:** Rebuilds take too long

**Solution:**
- Reduce `poll_interval` in `.air.toml` (currently 500ms)
- Exclude more directories in `exclude_dir`
- Use faster storage (SSD)

## Performance Tips

### 1. Cached Volume Mount

```yaml
volumes:
  - .:/app:cached
```

The `:cached` flag improves performance on Windows/Mac.

### 2. Exclude Build Artifacts

```yaml
volumes:
  - /app/tmp
```

Prevents tmp directory from syncing back to host.

### 3. Fast Polling

```toml
poll = true
poll_interval = 500
```

Checks for changes every 500ms (fast enough without CPU overhead).

### 4. Stop on Error

```toml
stop_on_error = true
```

Prevents running broken code.

## Database Constraints Fixed

During setup, we fixed several database constraints:

### 1. product_images

Added unique constraint for upsert:
```sql
ALTER TABLE product_images 
ADD CONSTRAINT product_images_product_id_image_url_key 
UNIQUE (product_id, image_url);
```

### 2. product_variations

Added unique constraint for upsert:
```sql
ALTER TABLE product_variations 
ADD CONSTRAINT product_variations_product_id_name_key 
UNIQUE (product_id, name);
```

### 3. store_product_mappings

Created ERP integration table:
```sql
CREATE TABLE store_product_mappings (
    store_id UUID,
    external_product_id VARCHAR(255),
    product_id UUID,
    UNIQUE(store_id, external_product_id)
);
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `gol start` | Start development environment |
| `gol stop` | Stop all services |
| `gol restart` | Restart application only |
| `gol rebuild` | Rebuild Docker image |
| `gol reload` | Trigger manual hot reload |
| `gol logs` | View application logs |
| `gol logs db` | View database logs |
| `gol air` | Check Air status |
| `gol status` | Show all service status |
| `gol test` | Test health endpoints |
| `gol db` | Connect to PostgreSQL |
| `gol redis` | Connect to Redis CLI |

## Access Points

- **Application:** http://localhost:8080
- **Health Check:** http://localhost:8080/health
- **pgAdmin:** http://localhost:5050
- **Redis Commander:** http://localhost:8081

## Next Steps

1. Edit code in your IDE
2. Save files
3. Watch Air rebuild automatically
4. Test at http://localhost:8080
5. Check logs with `gol logs`

Your development environment is fully automated!
