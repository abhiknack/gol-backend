# Gol Bazaar Helper Scripts Guide

## 🚀 Quick Commands

Two helper scripts are available for easy development:
- **gol.bat** - Windows Batch script (works in CMD)
- **gol.ps1** - PowerShell script (recommended for Windows)

## 📋 Usage

### Windows CMD (gol.bat)
```cmd
gol <command> [options]
```

### PowerShell (gol.ps1)
```powershell
.\gol.ps1 <command> [options]
```

## 🎯 Common Commands

### Start/Stop Services

```powershell
# Start all services
gol start
.\gol.ps1 start

# Stop all services
gol stop
.\gol.ps1 stop

# Restart application
gol restart
.\gol.ps1 restart

# Rebuild application (after code changes)
gol rebuild
.\gol.ps1 rebuild
```

### View Logs

```powershell
# View application logs
gol logs
.\gol.ps1 logs

# View database logs
gol logs db
.\gol.ps1 logs db

# View Redis logs
gol logs redis
.\gol.ps1 logs redis

# View all logs
gol logs all
.\gol.ps1 logs all
```

### Check Status

```powershell
# Show service status
gol status
.\gol.ps1 status

# Test all services
gol test
.\gol.ps1 test
```

### Database Operations

```powershell
# Connect to PostgreSQL
gol db
.\gol.ps1 db

# Create backup
gol db backup
.\gol.ps1 db backup

# Restore from backup
gol db restore backup_20241113.sql
.\gol.ps1 db restore backup_20241113.sql

# Reset database (⚠️ deletes all data)
gol db reset
.\gol.ps1 db reset
```

### Redis Operations

```powershell
# Connect to Redis CLI
gol redis
.\gol.ps1 redis

# Clear all cache
gol redis clear
.\gol.ps1 redis clear

# List all keys
gol redis keys
.\gol.ps1 redis keys
```

### Cleanup

```powershell
# Remove all containers and volumes
gol clean
.\gol.ps1 clean

# Full deployment
gol deploy
.\gol.ps1 deploy
```

### Help

```powershell
# Show help
gol help
.\gol.ps1 help
```

## 🔄 Daily Workflow

### Morning Setup
```powershell
# Start development environment
gol start

# Check status
gol status

# Test services
gol test
```

### During Development
```powershell
# 1. Edit code in your IDE

# 2. Rebuild application
gol rebuild

# 3. View logs
gol logs

# 4. Test endpoint
curl http://localhost:8080/api/v1/stores
```

### Debugging
```powershell
# View logs in real-time
gol logs

# Check service status
gol status

# Test health
gol test

# Connect to database
gol db

# Check Redis cache
gol redis keys
```

### End of Day
```powershell
# Stop services (optional)
gol stop
```

## 📊 Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| `start` | Start all services | `gol start` |
| `stop` | Stop all services | `gol stop` |
| `restart` | Restart app | `gol restart` |
| `rebuild` | Rebuild app | `gol rebuild` |
| `logs [service]` | View logs | `gol logs` or `gol logs db` |
| `status` | Show status | `gol status` |
| `test` | Test health | `gol test` |
| `db` | Database CLI | `gol db` |
| `db backup` | Backup database | `gol db backup` |
| `db restore <file>` | Restore database | `gol db restore backup.sql` |
| `db reset` | Reset database | `gol db reset` |
| `redis` | Redis CLI | `gol redis` |
| `redis clear` | Clear cache | `gol redis clear` |
| `redis keys` | List keys | `gol redis keys` |
| `clean` | Remove all | `gol clean` |
| `deploy` | Full deploy | `gol deploy` |
| `help` | Show help | `gol help` |

## 🎨 Examples

### Example 1: Start Fresh
```powershell
# Deploy everything
gol deploy

# Check status
gol status

# Test
gol test
```

### Example 2: Code Change Workflow
```powershell
# Edit code in VS Code

# Rebuild
gol rebuild

# View logs
gol logs

# Test
curl http://localhost:8080/api/v1/stores
```

### Example 3: Database Work
```powershell
# Backup current data
gol db backup

# Connect to database
gol db

# Run queries
SELECT * FROM stores;

# Exit (Ctrl+D or \q)
```

### Example 4: Cache Debugging
```powershell
# View cached keys
gol redis keys

# Clear cache
gol redis clear

# Test again
curl http://localhost:8080/api/v1/stores
```

### Example 5: Full Reset
```powershell
# Stop everything
gol stop

# Clean up
gol clean

# Start fresh
gol deploy
```

## 🔧 Advanced Usage

### Chaining Commands (PowerShell)
```powershell
# Rebuild and view logs
.\gol.ps1 rebuild; .\gol.ps1 logs

# Stop, clean, and deploy
.\gol.ps1 stop; .\gol.ps1 clean; .\gol.ps1 deploy
```

### Creating Aliases (PowerShell Profile)
```powershell
# Add to your PowerShell profile
function gr { .\gol.ps1 rebuild }
function gl { .\gol.ps1 logs }
function gt { .\gol.ps1 test }

# Usage
gr  # Rebuild
gl  # View logs
gt  # Test
```

### Batch Operations
```powershell
# Backup before reset
.\gol.ps1 db backup
.\gol.ps1 db reset

# Clear cache and rebuild
.\gol.ps1 redis clear
.\gol.ps1 rebuild
```

## 🚨 Troubleshooting

### Script Won't Run (PowerShell)
```powershell
# Enable script execution
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Then run
.\gol.ps1 help
```

### Permission Denied
```powershell
# Run as administrator or check Docker is running
docker ps
```

### Services Not Starting
```powershell
# Check Docker Desktop is running
# Check logs
gol logs all

# Try full reset
gol clean
gol deploy
```

## 💡 Pro Tips

1. **Keep logs open**: Run `gol logs` in a separate terminal
2. **Use PowerShell**: Better output formatting and colors
3. **Create aliases**: Add shortcuts to your PowerShell profile
4. **Backup regularly**: Use `gol db backup` before major changes
5. **Test often**: Run `gol test` after changes

## 📱 Quick Reference Card

```
┌─────────────────────────────────────────┐
│  Gol Bazaar Quick Commands              │
├─────────────────────────────────────────┤
│  gol start      Start services          │
│  gol rebuild    Rebuild app             │
│  gol logs       View logs               │
│  gol test       Test health             │
│  gol db         Database CLI            │
│  gol redis      Redis CLI               │
│  gol status     Show status             │
│  gol help       Show help               │
└─────────────────────────────────────────┘
```

---

**Keep this guide handy for quick reference!** 📋

Print it out or bookmark it in your browser.
