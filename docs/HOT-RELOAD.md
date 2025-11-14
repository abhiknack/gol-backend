# Hot Reload Development Setup

## What is Hot Reload?

Hot reload automatically rebuilds and restarts your Go application when you save code changes. No more manual `gol rebuild` after every change!

## How It Works

We use [Air](https://github.com/air-verse/air) - a live reload tool for Go applications.

When you save a `.go` file:
1. Air detects the change
2. Rebuilds the application (takes ~2-5 seconds)
3. Restarts the server automatically
4. You see the changes immediately

## Setup (One Time)

```bash
# Rebuild with the new hot reload Dockerfile
gol rebuild
```

That's it! Hot reload is now enabled.

## Usage

### Start Development
```bash
gol start
```

### Make Code Changes
1. Edit any `.go` file
2. Save the file (Ctrl+S)
3. Watch the logs to see the reload:
   ```bash
   gol logs
   ```
4. Test your changes immediately

### Watch Reload in Action
```bash
# In one terminal, watch logs
gol logs

# In another terminal, make changes to code
# You'll see Air rebuild and restart automatically
```

## What Gets Watched?

Air watches these file types:
- `.go` - Go source files
- `.yaml`, `.yml` - Config files
- `.env` - Environment files
- `.html`, `.tpl`, `.tmpl` - Templates

## What Gets Ignored?

- `tmp/` - Build artifacts
- `vendor/` - Dependencies
- `*_test.go` - Test files
- `docs/` - Documentation
- `.git/` - Git files

## Typical Workflow

```bash
# 1. Start services (first time or after stopping)
gol start

# 2. Watch logs in one terminal
gol logs

# 3. Edit code in your IDE
#    - Save changes
#    - Air automatically rebuilds
#    - Server restarts with new code

# 4. Test immediately
curl http://localhost:8080/api/v1/products/push -X POST ...

# 5. Repeat steps 3-4 as needed
```

## Rebuild Times

- **Initial build**: ~10-15 seconds
- **Hot reload**: ~2-5 seconds
- **No changes**: Instant (Air skips rebuild)

## Troubleshooting

### Changes Not Detected
```bash
# Check if Air is running
gol logs

# You should see: "watching .go files..."
```

### Build Errors
```bash
# Check build-errors.log
cat build-errors.log

# Or watch logs
gol logs
```

### Slow Rebuilds
Air is already optimized, but you can:
- Close unused applications
- Exclude more directories in `.air.toml`

### Manual Rebuild Needed
Sometimes you need a full rebuild:
```bash
gol rebuild
```

Use this when:
- Changing Dockerfile
- Updating dependencies (go.mod)
- Adding new packages

## Configuration

Hot reload settings are in `.air.toml`:

```toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000  # Wait 1s before rebuild
  exclude_dir = ["tmp", "vendor", "docs"]
  include_ext = ["go", "yaml", "env"]
```

## Disable Hot Reload

If you want to disable hot reload:

```bash
# Use production Dockerfile
# Edit docker-compose.dev.yml:
# Change: dockerfile: Dockerfile.dev
# To:     dockerfile: Dockerfile

gol rebuild
```

## Benefits

✅ **Faster Development** - No manual rebuilds
✅ **Immediate Feedback** - See changes in 2-5 seconds
✅ **Better Workflow** - Stay in your IDE
✅ **Less Context Switching** - No terminal commands needed
✅ **Automatic** - Just save and test

## Comparison

### Before (Manual Rebuild)
```bash
# Edit code
# Save file
gol rebuild          # 10-15 seconds
# Wait...
# Test changes
```

### After (Hot Reload)
```bash
# Edit code
# Save file
# Wait 2-5 seconds (automatic)
# Test changes
```

## Production

Hot reload is **only for development**. Production uses the optimized `Dockerfile`:
- Smaller image size
- No dev tools
- Faster startup
- More secure

```bash
# Production deployment
docker-compose -f docker-compose.prod.yml up -d
```

## Tips

1. **Keep logs open** - Watch rebuilds happen
2. **Save frequently** - Each save triggers rebuild
3. **Wait for rebuild** - Don't test until "Server started" appears
4. **Check errors** - Air shows build errors immediately
5. **Use health check** - `curl http://localhost:8080/health`

## Example Session

```bash
# Terminal 1: Start and watch logs
gol start
gol logs

# Terminal 2: Test endpoint
curl http://localhost:8080/health

# IDE: Edit internal/handlers/product_handler.go
# Save file (Ctrl+S)

# Terminal 1: See output
# [Air] building...
# [Air] Build finished
# [Air] Server started

# Terminal 2: Test with new changes
curl http://localhost:8080/api/v1/products/push ...
```

Enjoy faster development! 🚀
