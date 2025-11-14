# How Dev Container Code Changes Work

## Live Code Synchronization

When you work in the dev container, your code changes are **automatically synchronized** in real-time. Here's how:

### Volume Mount Magic 🔄

The dev container uses a **bind mount** that connects your local filesystem to the container:

```yaml
volumes:
  - ..:/workspace:cached
```

This means:
- Your local project folder (`D:\Gol-Backend`) is mounted to `/workspace` inside the container
- **Any changes you make are instantly reflected in both places**
- No manual copying or syncing needed!

## How It Works

### When You Edit Code in VS Code

```
Your Local Machine          Dev Container
─────────────────          ─────────────
D:\Gol-Backend\            /workspace/
├── cmd/                   ├── cmd/
│   └── server/       ←────→   └── server/
│       └── main.go            └── main.go  (same file!)
├── internal/              ├── internal/
│   └── router/       ←────→   └── router/
│       └── router.go          └── router.go  (same file!)
└── go.mod            ←────→ └── go.mod  (same file!)
```

### Real-Time Updates

1. **You edit** `internal/router/router.go` in VS Code
2. **Change is written** to `D:\Gol-Backend\internal\router\router.go`
3. **Instantly available** at `/workspace/internal/router/router.go` in container
4. **No rebuild needed** - it's the same file!

## Development Workflow

### Option 1: VS Code Dev Container (Recommended)

When you "Reopen in Container":

```bash
# VS Code connects to the running container
# Your editor is now "inside" the container
# All file operations happen directly in the container

# Edit files in VS Code → Changes are instant
# Run commands in VS Code terminal → Runs in container
# Debug code → Debugger runs in container
```

**Example workflow:**
```bash
# 1. Edit code in VS Code
# (Edit internal/router/handlers.go)

# 2. Run immediately in the integrated terminal
go run cmd/server/main.go

# 3. Test your changes
go test ./internal/router/...

# 4. No rebuild needed - changes are live!
```

### Option 2: Local Editor + Container Execution

If you prefer to edit locally but run in the container:

```bash
# 1. Edit code in your local VS Code (or any editor)
# (Edit D:\Gol-Backend\internal\router\handlers.go)

# 2. Run in the container
docker exec devcontainer-app-1 go run cmd/server/main.go

# 3. Changes are automatically available in the container!
```

## What Requires Rebuilding?

### ✅ NO Rebuild Needed For:

- **Code changes** (`.go` files)
- **Configuration files** (`.yaml`, `.env`)
- **Adding new files** to the project
- **Modifying existing files**
- **Installing Go packages** (`go get`, `go mod download`)

These changes are **instant** because they're on the mounted volume.

### ⚠️ Rebuild Required For:

- **Dockerfile changes** (adding system packages, changing base image)
- **docker-compose.yml changes** (environment variables, ports)
- **devcontainer.json changes** (VS Code extensions, settings)

**How to rebuild:**
```bash
# Stop containers
docker-compose -f .devcontainer/docker-compose.yml down

# Rebuild
docker-compose -f .devcontainer/docker-compose.yml build --no-cache

# Start again
docker-compose -f .devcontainer/docker-compose.yml up -d

# Or use the script
.devcontainer/rebuild.ps1
```

## Hot Reload for Development

For even faster development, you can use **Air** for hot reloading:

### Install Air (one-time setup)

```bash
# Inside the dev container
go install github.com/cosmtrek/air@latest
```

### Create .air.toml configuration

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  bin = "tmp/main"
  include_ext = ["go", "yaml"]
  exclude_dir = ["tmp", "vendor"]
  delay = 1000

[log]
  time = true
```

### Run with hot reload

```bash
# Any code change will automatically rebuild and restart the server
air
```

Now when you edit code, the server automatically restarts!

## Example Development Session

```bash
# 1. Open in VS Code Dev Container
# Press F1 → "Dev Containers: Reopen in Container"

# 2. VS Code opens, you're now "inside" the container

# 3. Open integrated terminal (Ctrl+`)
vscode@container:/workspace$ go run cmd/server/main.go
# Server starts...

# 4. Edit internal/router/handlers.go in VS Code
# Add a new endpoint

# 5. Stop the server (Ctrl+C)

# 6. Run again
vscode@container:/workspace$ go run cmd/server/main.go
# Your changes are live!

# 7. Or use Air for auto-reload
vscode@container:/workspace$ air
# Now just save files and server auto-restarts
```

## File Permissions

The dev container runs as user `vscode` (UID 1000), which should match your local user on most systems. This means:

- ✅ Files created in the container are owned by you locally
- ✅ Files created locally are accessible in the container
- ✅ No permission issues when editing

## Caching Strategy

The volume mount uses `:cached` flag:

```yaml
- ..:/workspace:cached
```

This means:
- **Writes are fast** - Changes from the container are batched
- **Reads are consistent** - You always see the latest changes
- **Performance optimized** for development

## Testing Your Setup

### Test 1: Create a file in VS Code

```bash
# In VS Code (connected to container)
echo "test" > /workspace/test.txt

# Check locally
# You should see D:\Gol-Backend\test.txt
```

### Test 2: Edit locally, run in container

```bash
# 1. Edit a file locally in any editor
# Add a comment to cmd/server/main.go

# 2. Check in container
docker exec devcontainer-app-1 cat /workspace/cmd/server/main.go
# You should see your comment!
```

### Test 3: Install a package

```bash
# In container
go get github.com/some/package

# Check locally
# D:\Gol-Backend\go.mod should be updated
# D:\Gol-Backend\go.sum should be updated
```

## Common Scenarios

### Scenario 1: Adding a new API endpoint

```bash
# 1. Edit internal/router/router.go (add new route)
# 2. Edit internal/router/handlers.go (add handler)
# 3. Save files
# 4. Restart server: go run cmd/server/main.go
# ✅ Changes are live immediately
```

### Scenario 2: Updating dependencies

```bash
# 1. Edit go.mod (add new dependency)
# 2. Run: go mod download
# 3. Use the new package in your code
# ✅ No container rebuild needed
```

### Scenario 3: Changing environment variables

```bash
# 1. Edit .env file locally
# 2. Restart the application
# ✅ New env vars are loaded
```

### Scenario 4: Adding a new package to Dockerfile

```bash
# 1. Edit .devcontainer/Dockerfile
# 2. Add: RUN apt-get install -y some-package
# 3. Rebuild container:
docker-compose -f .devcontainer/docker-compose.yml build --no-cache
docker-compose -f .devcontainer/docker-compose.yml up -d
# 4. Reconnect VS Code to container
# ⚠️ Rebuild required for Dockerfile changes
```

## Performance Tips

1. **Use Air for hot reload** - Saves time during development
2. **Keep the container running** - No startup delay
3. **Use VS Code's integrated terminal** - Faster than docker exec
4. **Enable Go language server** - Already configured for you
5. **Use cached volume mount** - Already configured for optimal performance

## Debugging

Your code changes are live, and so is debugging:

```json
// .vscode/launch.json (auto-created in dev container)
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Package",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/server"
    }
  ]
}
```

- Set breakpoints in VS Code
- Press F5 to debug
- Code changes are reflected immediately on next debug session

## Summary

✅ **Code changes are instant** - No rebuild needed  
✅ **Files are synchronized** - Edit anywhere, run in container  
✅ **Dependencies update live** - Just run `go mod download`  
✅ **Hot reload available** - Use Air for automatic restarts  
✅ **Debugging works** - Set breakpoints and debug in VS Code  
⚠️ **Only Dockerfile changes** require rebuild  

---

**The dev container gives you the best of both worlds:**
- Consistent environment (everyone has the same setup)
- Live code updates (no manual syncing or copying)
- Full development tools (Go, Redis, debugger, linters)
- Fast iteration (edit → run → test cycle)

Happy coding! 🚀
