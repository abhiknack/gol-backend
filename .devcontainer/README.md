# Dev Container Setup

This directory contains the configuration for a VS Code Dev Container that provides a complete development environment for the Supabase Redis Middleware project.

## What's Included

- **Go 1.23**: Latest Go toolchain
- **Redis 7**: Local Redis instance for caching
- **Go Tools**: Pre-installed development tools
  - `gopls` - Go language server
  - `dlv` - Delve debugger
  - `staticcheck` - Static analysis
  - `goimports` - Import formatting
  - `golangci-lint` - Comprehensive linter
- **VS Code Extensions**: Go, Docker, YAML support
- **Redis CLI**: Command-line tools for Redis interaction

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop)
- [Visual Studio Code](https://code.visualstudio.com/)
- [Dev Containers Extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

## Getting Started

1. Open this project in VS Code
2. Press `F1` and select **Dev Containers: Reopen in Container**
3. Wait for the container to build (first time only)
4. The development environment is ready!

## Configuration

### Environment Variables

The dev container uses default development settings. To customize:

1. Create a `.env` file in the project root
2. Add your Supabase credentials:
   ```env
   SUPABASE_URL=https://your-project.supabase.co
   SUPABASE_API_KEY=your-api-key-here
   ```

### Ports

The following ports are automatically forwarded:
- **8080**: Application server
- **6379**: Redis server

## Running the Application

Inside the dev container terminal:

```bash
# Run the application
go run cmd/server/main.go

# Run tests
go test ./...

# Run with hot reload (install air first)
go install github.com/cosmtrek/air@latest
air
```

## Redis Access

Check Redis connectivity:
```bash
redis-cli ping
```

Monitor Redis commands:
```bash
redis-cli monitor
```

View cached keys:
```bash
redis-cli keys "*"
```

## Debugging

The dev container includes Delve debugger. To debug:

1. Set breakpoints in your code
2. Press `F5` or use the Debug panel
3. Select "Launch Package" configuration

## Tips

- All Go tools are pre-installed and configured
- Format on save is enabled by default
- Linting runs automatically on save
- Redis data persists in a Docker volume between container rebuilds

## Troubleshooting

**Container won't start:**
- Ensure Docker Desktop is running
- Check that ports 8080 and 6379 are not in use
- Try rebuilding: `F1` → **Dev Containers: Rebuild Container**

**Go modules issues:**
- Run `go mod download` in the terminal
- Check your internet connection

**Redis connection failed:**
- Redis should start automatically with the container
- Verify with `redis-cli ping`
- Check logs: `docker logs <redis-container-id>`
