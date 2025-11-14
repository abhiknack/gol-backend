# Container Setup Explanation

## You Have TWO Separate Setups Running

### 1. Production/Main Application Containers ✅ (EXPOSED)

These are the containers running your actual application:

| Container Name | Image | Ports | Purpose |
|----------------|-------|-------|---------|
| **supabase-redis-middleware** | gol-backend-app | **0.0.0.0:8080→8080** | **Main application (EXPOSED)** |
| **supabase-redis-cache** | redis:7-alpine | **0.0.0.0:6379→6379** | **Redis cache (EXPOSED)** |

**Access Points:**
- 🌐 **Application API**: http://localhost:8080
- 🌐 **Health Check**: http://localhost:8080/health
- 🌐 **Redis**: localhost:6379

**These are LIVE and accessible from your host machine!**

---

### 2. Dev Container (NOT EXPOSED)

These are for VS Code development environment:

| Container Name | Image | Ports | Purpose |
|----------------|-------|-------|---------|
| **devcontainer-app-1** | devcontainer-app | None exposed | Development environment |
| **devcontainer-redis-1** | redis:7-alpine | None exposed | Redis for dev container |

**Access:**
- ❌ **NOT directly accessible** from host machine
- ✅ **Only accessible** when you "Reopen in Container" in VS Code
- ✅ **Internal network** - app can reach Redis at localhost:6379

---

## The Key Difference

### Production Containers (docker-compose.yml)
```yaml
services:
  app:
    ports:
      - "8080:8080"  # ← EXPOSED to host machine
```

### Dev Container (.devcontainer/docker-compose.yml)
```yaml
services:
  app:
    # NO ports section!
    # Uses network_mode: service:redis
    # Not exposed to host
```

---

## Which One Should You Use?

### For Running the Application → Use Production Containers

**Already running and accessible:**
```bash
# Application is live at:
http://localhost:8080

# Test it:
curl http://localhost:8080/health

# View logs:
docker logs supabase-redis-middleware -f

# Restart:
docker-compose restart app
```

### For Development → Use Dev Container

**Open in VS Code:**
1. Press `F1`
2. Select "Dev Containers: Reopen in Container"
3. VS Code connects to `devcontainer-app-1`
4. You get full IDE features inside the container

**Then run the app inside the dev container:**
```bash
# Inside VS Code terminal (connected to dev container)
go run cmd/server/main.go
# Server runs on port 8080 INSIDE the container
# But it's NOT exposed to your host machine
```

---

## How to Expose Dev Container Ports

If you want to run the app in the dev container AND access it from your host machine, you need to expose ports.

### Option 1: Update .devcontainer/docker-compose.yml

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    volumes:
      - ..:/workspace:cached
    command: sleep infinity
    # Remove network_mode and add ports
    ports:
      - "8081:8080"  # Use different port to avoid conflict
    environment:
      - SERVER_PORT=8080
      # ... other env vars
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"  # Use different port to avoid conflict
    # ... rest of config
```

### Option 2: Use VS Code Port Forwarding (Automatic)

When you run the app in the dev container, VS Code automatically detects and forwards ports!

```bash
# Inside dev container terminal
go run cmd/server/main.go

# VS Code will show a notification:
# "Your application running on port 8080 is available"
# Click to open in browser
```

---

## Current Recommended Setup

### Scenario 1: Just Testing the Application

**Use the production containers** (already running):
```bash
# Access the running application
curl http://localhost:8080/health

# View API endpoints
curl http://localhost:8080/api/v1/supermarket/products
```

### Scenario 2: Developing and Testing Code

**Option A: Edit locally, run in production container**
```bash
# 1. Edit code in VS Code (local)
# 2. Restart production container
docker-compose restart app
# 3. Test at http://localhost:8080
```

**Option B: Use dev container with VS Code port forwarding**
```bash
# 1. Open in VS Code dev container
# 2. Run: go run cmd/server/main.go
# 3. VS Code auto-forwards port 8080
# 4. Click the notification to open in browser
```

---

## Port Conflict Resolution

You currently have:
- Production app on port **8080** ✅
- Production Redis on port **6379** ✅
- Dev container app: **Not exposed** (no conflict)
- Dev container Redis: **Not exposed** (no conflict)

If you want to run both simultaneously with exposed ports:

```yaml
# .devcontainer/docker-compose.yml
services:
  app:
    ports:
      - "8081:8080"  # Dev app on 8081
  redis:
    ports:
      - "6380:6379"  # Dev Redis on 6380
```

Then you'd have:
- Production: http://localhost:8080
- Development: http://localhost:8081

---

## Quick Reference

### Production Containers (Currently Exposed)

```bash
# Start
docker-compose up -d

# Stop
docker-compose down

# Logs
docker logs supabase-redis-middleware -f

# Access
curl http://localhost:8080/health
```

**Ports:**
- Application: **8080** ✅
- Redis: **6379** ✅

### Dev Container (Not Exposed)

```bash
# Start
docker-compose -f .devcontainer/docker-compose.yml up -d

# Stop
docker-compose -f .devcontainer/docker-compose.yml down

# Access
# Use VS Code "Reopen in Container"
# Or: docker exec -it devcontainer-app-1 bash
```

**Ports:**
- Application: **Not exposed** ❌
- Redis: **Not exposed** ❌
- Use VS Code port forwarding when running the app

---

## Summary

**Your endpoints are currently exposed in:**

🎯 **supabase-redis-middleware** container (production setup)
- ✅ Port 8080 is exposed and accessible
- ✅ Running the compiled Go application
- ✅ Connected to Redis on port 6379

**The dev container is for:**
- 🛠️ Development environment
- 🛠️ VS Code integration
- 🛠️ Not meant to expose ports (unless you configure it)
- 🛠️ Use VS Code's automatic port forwarding when running the app

**Recommendation:**
- Keep using the production containers for testing the API
- Use the dev container for development with VS Code
- Let VS Code handle port forwarding automatically when you run the app in the dev container
