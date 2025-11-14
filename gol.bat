@echo off
REM Gol Bazaar - Development Helper Script
REM Usage: gol <command>

setlocal enabledelayedexpansion

if "%1"=="" goto :help

REM Main command router
if /i "%1"=="start" goto :start
if /i "%1"=="stop" goto :stop
if /i "%1"=="restart" goto :restart
if /i "%1"=="rebuild" goto :rebuild
if /i "%1"=="reload" goto :reload
if /i "%1"=="logs" goto :logs
if /i "%1"=="status" goto :status
if /i "%1"=="air" goto :air
if /i "%1"=="test" goto :test
if /i "%1"=="db" goto :db
if /i "%1"=="redis" goto :redis
if /i "%1"=="clean" goto :clean
if /i "%1"=="deploy" goto :deploy
if /i "%1"=="help" goto :help
if /i "%1"=="-h" goto :help
if /i "%1"=="--help" goto :help

echo Unknown command: %1
echo Run 'gol help' for usage information
exit /b 1

:start
echo [Gol Bazaar] Starting development environment...
docker-compose -f docker-compose.dev.yml up -d
echo [Gol Bazaar] Services started!
echo.
echo Access points:
echo   App:             http://localhost:8080
echo   pgAdmin:         http://localhost:5050
echo   Redis Commander: http://localhost:8081
echo.
echo Hot Reload: ENABLED
echo   Code changes will auto-reload the server
echo   Run 'gol logs' to see reload messages
goto :end

:stop
echo [Gol Bazaar] Stopping development environment...
docker-compose -f docker-compose.dev.yml down
echo [Gol Bazaar] Services stopped!
goto :end

:restart
echo [Gol Bazaar] Restarting application...
docker-compose -f docker-compose.dev.yml restart app
echo [Gol Bazaar] Application restarted!
echo Run 'gol logs' to view logs
goto :end

:rebuild
echo [Gol Bazaar] Rebuilding application...
docker-compose -f docker-compose.dev.yml up -d --build app
echo [Gol Bazaar] Application rebuilt and restarted!
echo.
echo Hot Reload: ENABLED
echo   Code changes will now auto-reload
echo   Run 'gol logs' to view logs
goto :end

:reload
echo [Gol Bazaar] Triggering hot reload...
echo Touching main.go to trigger Air rebuild...
docker exec gol-bazaar-app-dev touch /app/cmd/server/main.go
if errorlevel 1 (
    echo Warning: Container may not be running. Try 'gol start' first.
    exit /b 1
)
echo Hot reload triggered!
echo Run 'gol logs' to see the rebuild process
goto :end

:logs
if "%2"=="" (
    echo [Gol Bazaar] Following application logs... (Ctrl+C to exit)
    docker logs gol-bazaar-app-dev -f
) else if /i "%2"=="app" (
    docker logs gol-bazaar-app-dev -f
) else if /i "%2"=="db" (
    docker logs gol-bazaar-postgres-dev -f
) else if /i "%2"=="redis" (
    docker logs gol-bazaar-redis-dev -f
) else if /i "%2"=="all" (
    docker-compose -f docker-compose.dev.yml logs -f
) else (
    echo Unknown service: %2
    echo Available: app, db, redis, all
    exit /b 1
)
goto :end

:status
echo [Gol Bazaar] Service Status:
echo.
docker ps --filter "name=gol-bazaar" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo.
echo Run 'gol test' to check health
echo Run 'gol air' to check Air hot reload status
goto :end

:air
echo [Gol Bazaar] Air Hot Reload Status:
echo.
docker exec gol-bazaar-app-dev ps aux | findstr air
if errorlevel 1 (
    echo Air is not running or container is not started
    echo Run 'gol start' to start the development environment
    exit /b 1
)
echo.
echo [Gol Bazaar] Recent Air activity:
docker logs gol-bazaar-app-dev --tail 20 | findstr /i "air building watching"
echo.
echo Run 'gol logs' to see full Air output
goto :end

:test
echo [Gol Bazaar] Testing health endpoint...
curl -s http://localhost:8080/health
echo.
echo.
echo [Gol Bazaar] Testing database connection...
docker exec gol-bazaar-postgres-dev pg_isready -U postgres
echo.
echo [Gol Bazaar] Testing Redis connection...
docker exec gol-bazaar-redis-dev redis-cli ping
goto :end

:db
if "%2"=="" (
    echo [Gol Bazaar] Connecting to PostgreSQL...
    docker exec -it gol-bazaar-postgres-dev psql -U postgres -d middleware_db
) else if /i "%2"=="reset" (
    echo [Gol Bazaar] WARNING: This will delete all database data!
    set /p confirm="Are you sure? (yes/no): "
    if /i "!confirm!"=="yes" (
        echo Resetting database...
        docker-compose -f docker-compose.dev.yml down
        docker volume rm gol-bazaar-postgres-dev-data
        docker-compose -f docker-compose.dev.yml up -d
        echo Database reset complete!
    ) else (
        echo Reset cancelled.
    )
) else if /i "%2"=="backup" (
    set filename=backup_%date:~-4,4%%date:~-10,2%%date:~-7,2%_%time:~0,2%%time:~3,2%%time:~6,2%.sql
    set filename=!filename: =0!
    echo [Gol Bazaar] Creating backup: !filename!
    docker exec gol-bazaar-postgres-dev pg_dump -U postgres -d middleware_db > !filename!
    echo Backup saved to !filename!
) else if /i "%2"=="restore" (
    if "%3"=="" (
        echo Usage: gol db restore filename.sql
        exit /b 1
    )
    echo [Gol Bazaar] Restoring from %3...
    type %3 | docker exec -i gol-bazaar-postgres-dev psql -U postgres -d middleware_db
    echo Restore complete!
) else (
    echo Unknown db command: %2
    echo Available: connect (default), reset, backup, restore
    exit /b 1
)
goto :end

:redis
if "%2"=="" (
    echo [Gol Bazaar] Connecting to Redis...
    docker exec -it gol-bazaar-redis-dev redis-cli
) else if /i "%2"=="clear" (
    echo [Gol Bazaar] Clearing Redis cache...
    docker exec gol-bazaar-redis-dev redis-cli FLUSHALL
    echo Cache cleared!
) else if /i "%2"=="keys" (
    echo [Gol Bazaar] Redis keys:
    docker exec gol-bazaar-redis-dev redis-cli keys "*"
) else (
    echo Unknown redis command: %2
    echo Available: connect (default), clear, keys
    exit /b 1
)
goto :end

:clean
echo [Gol Bazaar] Cleaning up...
set /p confirm="This will remove all containers and volumes. Continue? (yes/no): "
if /i "%confirm%"=="yes" (
    docker-compose -f docker-compose.dev.yml down -v
    echo Cleanup complete!
) else (
    echo Cleanup cancelled.
)
goto :end

:deploy
echo [Gol Bazaar] Deploying development environment...
call scripts\deploy-dev.ps1
goto :end

:help
echo.
echo Gol Bazaar - Development Helper
echo ================================
echo.
echo Usage: gol ^<command^> [options]
echo.
echo Commands:
echo   start          Start all development services
echo   stop           Stop all development services
echo   restart        Restart the application
echo   rebuild        Rebuild and restart the application
echo   reload         Trigger hot reload manually
echo   logs [service] View logs (app, db, redis, all)
echo   status         Show service status
echo   air            Check Air hot reload status
echo   test           Test all services health
echo.
echo   db             Connect to PostgreSQL
echo   db reset       Reset database (deletes all data)
echo   db backup      Create database backup
echo   db restore ^<file^>  Restore from backup
echo.
echo   redis          Connect to Redis CLI
echo   redis clear    Clear all Redis cache
echo   redis keys     List all Redis keys
echo.
echo   clean          Remove all containers and volumes
echo   deploy         Run full deployment script
echo   help           Show this help message
echo.
echo Examples:
echo   gol start              Start development environment
echo   gol reload             Trigger hot reload manually
echo   gol rebuild            Rebuild after code changes
echo   gol logs               View application logs
echo   gol logs db            View database logs
echo   gol test               Test all services
echo   gol db                 Connect to database
echo   gol redis clear        Clear cache
echo.
echo Quick workflow:
echo   1. gol start           (first time)
echo   2. Edit code
echo   3. gol rebuild         (after changes)
echo   4. gol logs            (check output)
echo   5. gol test            (verify health)
echo.
goto :end

:end
endlocal
