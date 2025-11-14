# Dev Container Troubleshooting

## Permission Denied Errors

### Problem: "permission denied" when running `go run` or `go mod download`

```
go: writing go.mod cache: mkdir /go/pkg/mod/cache/download/github.com: permission denied
```

### Cause
The `/go` directory is owned by root, but you're running as the `vscode` user.

### Solution 1: Fix Permissions (Quick Fix)

**From your host machine:**
```bash
docker exec -u root devcontainer-app-1 chown -R vscode:vscode /go
```

**Or use the script:**
```bash
# Windows PowerShell
.devcontainer\fix-permissions.sh

# Linux/Mac
bash .devcontainer/fix-permissions.sh
```

### Solution 2: Rebuild Container (Permanent Fix)

The Dockerfile has been updated to fix this automatically. Rebuild the container:

```bash
# Stop the container
docker-compose -f .devcontainer/docker-compose.yml down

# Rebuild with no cache
docker-compose -f .devcontainer/docker-compose.yml build --no-cache

# Start again
docker-compose -f .devcontainer/docker-compose.yml up -d

# Reopen in VS Code
# F1 → "Dev Containers: Rebuild and Reopen in Container"
```

### Solution 3: Run as Root (Temporary Workaround)

```bash
# Run commands as root
docker exec -u root devcontainer-app-1 go mod download

# Then fix permissions
docker exec -u root devcontainer-app-1 chown -R vscode:vscode /go
```

---

## Container Won't Start

### Problem: Container exits immediately or won't start

**Check logs:**
```bash
docker logs devcontainer-app-1
```

**Common causes:**
1. Port conflict (another service using the same port)
2. Docker daemon not running
3. Build failed

**Solution:**
```bash
# Stop all containers
docker-compose -f .devcontainer/docker-compose.yml down

# Remove old containers
docker-compose -f .devcontainer/docker-compose.yml rm -f

# Rebuild
docker-compose -f .devcontainer/docker-compose.yml build --no-cache

# Start
docker-compose -f .devcontainer/docker-compose.yml up -d
```

---

## VS Code Can't Connect to Container

### Problem: "Failed to connect" or "Container not found"

**Check if container is running:**
```bash
docker ps | findstr devcontainer
```

**If not running, start it:**
```bash
docker-compose -f .devcontainer/docker-compose.yml up -d
```

**If still not working:**
1. Close VS Code completely
2. Restart Docker Desktop
3. Start containers again
4. Open VS Code and try "Reopen in Container"

---

## Go Modules Not Downloading

### Problem: `go mod download` fails or hangs

**Check internet connection:**
```bash
docker exec devcontainer-app-1 ping -c 3 proxy.golang.org
```

**Check Go environment:**
```bash
docker exec devcontainer-app-1 go env
```

**Clear module cache and retry:**
```bash
docker exec -u root devcontainer-app-1 rm -rf /go/pkg/mod
docker exec -u root devcontainer-app-1 chown -R vscode:vscode /go
docker exec devcontainer-app-1 go mod download
```

---

## Redis Connection Failed

### Problem: Can't connect to Redis from the app

**Check if Redis is running:**
```bash
docker ps | findstr redis
```

**Test Redis connectivity:**
```bash
docker exec devcontainer-redis-1 redis-cli ping
# Should return: PONG
```

**Check network mode:**
The dev container uses `network_mode: service:redis`, which means:
- Redis should be accessible at `localhost:6379`
- Both containers share the same network namespace

**Test from app container:**
```bash
docker exec devcontainer-app-1 ping localhost
```

**If Redis is not running:**
```bash
docker-compose -f .devcontainer/docker-compose.yml up -d redis
```

---

## Port Forwarding Not Working

### Problem: Can't access the app running in dev container

**VS Code should auto-forward ports.** Check the "Ports" tab in VS Code.

**Manual port forwarding:**
```bash
# Update .devcontainer/docker-compose.yml
services:
  app:
    ports:
      - "8081:8080"  # Add this line
```

**Then rebuild:**
```bash
docker-compose -f .devcontainer/docker-compose.yml down
docker-compose -f .devcontainer/docker-compose.yml up -d
```

---

## Go Tools Not Working

### Problem: gopls, dlv, or other tools not found

**Check if tools are installed:**
```bash
docker exec devcontainer-app-1 which gopls
docker exec devcontainer-app-1 which dlv
```

**Reinstall tools:**
```bash
docker exec devcontainer-app-1 go install golang.org/x/tools/gopls@v0.16.2
docker exec devcontainer-app-1 go install github.com/go-delve/delve/cmd/dlv@latest
```

**Or rebuild container:**
```bash
docker-compose -f .devcontainer/docker-compose.yml build --no-cache
```

---

## Slow Performance

### Problem: Container is slow or unresponsive

**Check Docker resources:**
- Open Docker Desktop → Settings → Resources
- Increase CPU and Memory allocation

**Check volume mount performance:**
The dev container uses `:cached` flag for better performance:
```yaml
volumes:
  - ..:/workspace:cached
```

**For Windows users:**
- Ensure WSL 2 is enabled
- Store project files in WSL filesystem for better performance

---

## Build Errors

### Problem: "go build" fails with errors

**Check Go version:**
```bash
docker exec devcontainer-app-1 go version
# Should be: go version go1.23.12
```

**Clean build cache:**
```bash
docker exec devcontainer-app-1 go clean -cache -modcache -testcache
docker exec -u root devcontainer-app-1 chown -R vscode:vscode /go
docker exec devcontainer-app-1 go mod download
```

**Verify go.mod:**
```bash
docker exec devcontainer-app-1 go mod verify
docker exec devcontainer-app-1 go mod tidy
```

---

## Common Commands

### Check Container Status
```bash
docker-compose -f .devcontainer/docker-compose.yml ps
```

### View Logs
```bash
docker logs devcontainer-app-1 -f
docker logs devcontainer-redis-1 -f
```

### Access Container Shell
```bash
docker exec -it devcontainer-app-1 bash
```

### Restart Containers
```bash
docker-compose -f .devcontainer/docker-compose.yml restart
```

### Stop Containers
```bash
docker-compose -f .devcontainer/docker-compose.yml down
```

### Remove Everything and Start Fresh
```bash
# Stop and remove containers
docker-compose -f .devcontainer/docker-compose.yml down -v

# Remove images
docker rmi devcontainer-app

# Rebuild from scratch
docker-compose -f .devcontainer/docker-compose.yml build --no-cache

# Start
docker-compose -f .devcontainer/docker-compose.yml up -d
```

---

## Getting Help

If you're still having issues:

1. **Check Docker logs:**
   ```bash
   docker logs devcontainer-app-1
   ```

2. **Check VS Code Dev Container logs:**
   - View → Output → Select "Dev Containers" from dropdown

3. **Verify Docker is running:**
   ```bash
   docker ps
   docker version
   ```

4. **Check system resources:**
   - Ensure Docker has enough CPU/Memory
   - Check disk space

5. **Try a clean rebuild:**
   ```bash
   docker-compose -f .devcontainer/docker-compose.yml down -v
   docker-compose -f .devcontainer/docker-compose.yml build --no-cache
   docker-compose -f .devcontainer/docker-compose.yml up -d
   ```

---

## Quick Fix Checklist

- [ ] Docker Desktop is running
- [ ] Containers are running: `docker ps`
- [ ] Permissions are correct: `docker exec -u root devcontainer-app-1 chown -R vscode:vscode /go`
- [ ] Go modules downloaded: `docker exec devcontainer-app-1 go mod download`
- [ ] Redis is accessible: `docker exec devcontainer-redis-1 redis-cli ping`
- [ ] VS Code is connected to container
- [ ] Ports are forwarded (check VS Code "Ports" tab)

If all checks pass, you should be good to go! 🚀
