# Gol Bazaar - Development Helper Script (PowerShell)
# Usage: .\gol.ps1 <command>

param(
    [Parameter(Position=0)]
    [string]$Command,
    
    [Parameter(Position=1)]
    [string]$SubCommand,
    
    [Parameter(Position=2)]
    [string]$Arg
)

$ComposeFile = "docker-compose.dev.yml"

function Show-Help {
    Write-Host ""
    Write-Host "Gol Bazaar - Development Helper" -ForegroundColor Cyan
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\gol.ps1 <command> [options]" -ForegroundColor White
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Yellow
    Write-Host "  start          " -NoNewline; Write-Host "Start all development services" -ForegroundColor Gray
    Write-Host "  stop           " -NoNewline; Write-Host "Stop all development services" -ForegroundColor Gray
    Write-Host "  restart        " -NoNewline; Write-Host "Restart the application" -ForegroundColor Gray
    Write-Host "  rebuild        " -NoNewline; Write-Host "Rebuild and restart the application" -ForegroundColor Gray
    Write-Host "  logs [service] " -NoNewline; Write-Host "View logs (app, db, redis, all)" -ForegroundColor Gray
    Write-Host "  status         " -NoNewline; Write-Host "Show service status" -ForegroundColor Gray
    Write-Host "  test           " -NoNewline; Write-Host "Test all services health" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  db             " -NoNewline; Write-Host "Connect to PostgreSQL" -ForegroundColor Gray
    Write-Host "  db reset       " -NoNewline; Write-Host "Reset database (deletes all data)" -ForegroundColor Gray
    Write-Host "  db backup      " -NoNewline; Write-Host "Create database backup" -ForegroundColor Gray
    Write-Host "  db restore <file>" -NoNewline; Write-Host "  Restore from backup" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  redis          " -NoNewline; Write-Host "Connect to Redis CLI" -ForegroundColor Gray
    Write-Host "  redis clear    " -NoNewline; Write-Host "Clear all Redis cache" -ForegroundColor Gray
    Write-Host "  redis keys     " -NoNewline; Write-Host "List all Redis keys" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  clean          " -NoNewline; Write-Host "Remove all containers and volumes" -ForegroundColor Gray
    Write-Host "  deploy         " -NoNewline; Write-Host "Run full deployment script" -ForegroundColor Gray
    Write-Host "  help           " -NoNewline; Write-Host "Show this help message" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\gol.ps1 start              " -NoNewline; Write-Host "Start development environment" -ForegroundColor Gray
    Write-Host "  .\gol.ps1 rebuild            " -NoNewline; Write-Host "Rebuild after code changes" -ForegroundColor Gray
    Write-Host "  .\gol.ps1 logs               " -NoNewline; Write-Host "View application logs" -ForegroundColor Gray
    Write-Host "  .\gol.ps1 logs db            " -NoNewline; Write-Host "View database logs" -ForegroundColor Gray
    Write-Host "  .\gol.ps1 test               " -NoNewline; Write-Host "Test all services" -ForegroundColor Gray
    Write-Host "  .\gol.ps1 db                 " -NoNewline; Write-Host "Connect to database" -ForegroundColor Gray
    Write-Host "  .\gol.ps1 redis clear        " -NoNewline; Write-Host "Clear cache" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Quick workflow:" -ForegroundColor Yellow
    Write-Host "  1. .\gol.ps1 start           (first time)" -ForegroundColor Gray
    Write-Host "  2. Edit code" -ForegroundColor Gray
    Write-Host "  3. .\gol.ps1 rebuild         (after changes)" -ForegroundColor Gray
    Write-Host "  4. .\gol.ps1 logs            (check output)" -ForegroundColor Gray
    Write-Host "  5. .\gol.ps1 test            (verify health)" -ForegroundColor Gray
    Write-Host ""
}

function Start-Services {
    Write-Host "[Gol Bazaar] Starting development environment..." -ForegroundColor Cyan
    docker-compose -f $ComposeFile up -d
    Write-Host "[Gol Bazaar] Services started!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Access points:" -ForegroundColor Yellow
    Write-Host "  App:             http://localhost:8080" -ForegroundColor White
    Write-Host "  pgAdmin:         http://localhost:5050" -ForegroundColor White
    Write-Host "  Redis Commander: http://localhost:8081" -ForegroundColor White
}

function Stop-Services {
    Write-Host "[Gol Bazaar] Stopping development environment..." -ForegroundColor Cyan
    docker-compose -f $ComposeFile down
    Write-Host "[Gol Bazaar] Services stopped!" -ForegroundColor Green
}

function Restart-App {
    Write-Host "[Gol Bazaar] Restarting application..." -ForegroundColor Cyan
    docker-compose -f $ComposeFile restart app
    Write-Host "[Gol Bazaar] Application restarted!" -ForegroundColor Green
    Write-Host "Run '.\gol.ps1 logs' to view logs" -ForegroundColor Yellow
}

function Rebuild-App {
    Write-Host "[Gol Bazaar] Rebuilding application..." -ForegroundColor Cyan
    docker-compose -f $ComposeFile up -d --build app
    Write-Host "[Gol Bazaar] Application rebuilt and restarted!" -ForegroundColor Green
    Write-Host "Run '.\gol.ps1 logs' to view logs" -ForegroundColor Yellow
}

function Show-Logs {
    param([string]$Service)
    
    if ([string]::IsNullOrEmpty($Service)) {
        Write-Host "[Gol Bazaar] Following application logs... (Ctrl+C to exit)" -ForegroundColor Cyan
        docker logs gol-bazaar-app-dev -f
    } elseif ($Service -eq "app") {
        docker logs gol-bazaar-app-dev -f
    } elseif ($Service -eq "db") {
        docker logs gol-bazaar-postgres-dev -f
    } elseif ($Service -eq "redis") {
        docker logs gol-bazaar-redis-dev -f
    } elseif ($Service -eq "all") {
        docker-compose -f $ComposeFile logs -f
    } else {
        Write-Host "Unknown service: $Service" -ForegroundColor Red
        Write-Host "Available: app, db, redis, all" -ForegroundColor Yellow
        exit 1
    }
}

function Show-Status {
    Write-Host "[Gol Bazaar] Service Status:" -ForegroundColor Cyan
    Write-Host ""
    docker ps --filter "name=gol-bazaar" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    Write-Host ""
    Write-Host "Run '.\gol.ps1 test' to check health" -ForegroundColor Yellow
}

function Test-Services {
    Write-Host "[Gol Bazaar] Testing health endpoint..." -ForegroundColor Cyan
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing
        Write-Host "✓ Application is healthy!" -ForegroundColor Green
        $response.Content
    } catch {
        Write-Host "✗ Application health check failed" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "[Gol Bazaar] Testing database connection..." -ForegroundColor Cyan
    docker exec gol-bazaar-postgres-dev pg_isready -U postgres
    
    Write-Host ""
    Write-Host "[Gol Bazaar] Testing Redis connection..." -ForegroundColor Cyan
    docker exec gol-bazaar-redis-dev redis-cli ping
}

function Manage-Database {
    param([string]$Action, [string]$File)
    
    if ([string]::IsNullOrEmpty($Action)) {
        Write-Host "[Gol Bazaar] Connecting to PostgreSQL..." -ForegroundColor Cyan
        docker exec -it gol-bazaar-postgres-dev psql -U postgres -d middleware_db
    } elseif ($Action -eq "reset") {
        Write-Host "[Gol Bazaar] WARNING: This will delete all database data!" -ForegroundColor Red
        $confirm = Read-Host "Are you sure? (yes/no)"
        if ($confirm -eq "yes") {
            Write-Host "Resetting database..." -ForegroundColor Yellow
            docker-compose -f $ComposeFile down
            docker volume rm gol-bazaar-postgres-dev-data
            docker-compose -f $ComposeFile up -d
            Write-Host "Database reset complete!" -ForegroundColor Green
        } else {
            Write-Host "Reset cancelled." -ForegroundColor Yellow
        }
    } elseif ($Action -eq "backup") {
        $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
        $filename = "backup_$timestamp.sql"
        Write-Host "[Gol Bazaar] Creating backup: $filename" -ForegroundColor Cyan
        docker exec gol-bazaar-postgres-dev pg_dump -U postgres -d middleware_db > $filename
        Write-Host "Backup saved to $filename" -ForegroundColor Green
    } elseif ($Action -eq "restore") {
        if ([string]::IsNullOrEmpty($File)) {
            Write-Host "Usage: .\gol.ps1 db restore filename.sql" -ForegroundColor Red
            exit 1
        }
        Write-Host "[Gol Bazaar] Restoring from $File..." -ForegroundColor Cyan
        Get-Content $File | docker exec -i gol-bazaar-postgres-dev psql -U postgres -d middleware_db
        Write-Host "Restore complete!" -ForegroundColor Green
    } else {
        Write-Host "Unknown db command: $Action" -ForegroundColor Red
        Write-Host "Available: connect (default), reset, backup, restore" -ForegroundColor Yellow
        exit 1
    }
}

function Manage-Redis {
    param([string]$Action)
    
    if ([string]::IsNullOrEmpty($Action)) {
        Write-Host "[Gol Bazaar] Connecting to Redis..." -ForegroundColor Cyan
        docker exec -it gol-bazaar-redis-dev redis-cli
    } elseif ($Action -eq "clear") {
        Write-Host "[Gol Bazaar] Clearing Redis cache..." -ForegroundColor Cyan
        docker exec gol-bazaar-redis-dev redis-cli FLUSHALL
        Write-Host "Cache cleared!" -ForegroundColor Green
    } elseif ($Action -eq "keys") {
        Write-Host "[Gol Bazaar] Redis keys:" -ForegroundColor Cyan
        docker exec gol-bazaar-redis-dev redis-cli keys "*"
    } else {
        Write-Host "Unknown redis command: $Action" -ForegroundColor Red
        Write-Host "Available: connect (default), clear, keys" -ForegroundColor Yellow
        exit 1
    }
}

function Clean-All {
    Write-Host "[Gol Bazaar] Cleaning up..." -ForegroundColor Cyan
    $confirm = Read-Host "This will remove all containers and volumes. Continue? (yes/no)"
    if ($confirm -eq "yes") {
        docker-compose -f $ComposeFile down -v
        Write-Host "Cleanup complete!" -ForegroundColor Green
    } else {
        Write-Host "Cleanup cancelled." -ForegroundColor Yellow
    }
}

function Deploy-Dev {
    Write-Host "[Gol Bazaar] Deploying development environment..." -ForegroundColor Cyan
    & ".\scripts\deploy-dev.ps1"
}

# Main command router
switch ($Command.ToLower()) {
    "start"   { Start-Services }
    "stop"    { Stop-Services }
    "restart" { Restart-App }
    "rebuild" { Rebuild-App }
    "logs"    { Show-Logs -Service $SubCommand }
    "status"  { Show-Status }
    "test"    { Test-Services }
    "db"      { Manage-Database -Action $SubCommand -File $Arg }
    "redis"   { Manage-Redis -Action $SubCommand }
    "clean"   { Clean-All }
    "deploy"  { Deploy-Dev }
    "help"    { Show-Help }
    "-h"      { Show-Help }
    "--help"  { Show-Help }
    default   {
        if ([string]::IsNullOrEmpty($Command)) {
            Show-Help
        } else {
            Write-Host "Unknown command: $Command" -ForegroundColor Red
            Write-Host "Run '.\gol.ps1 help' for usage information" -ForegroundColor Yellow
            exit 1
        }
    }
}
