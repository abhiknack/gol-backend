# Rebuild Dev Container Script
Write-Host "Rebuilding dev container..." -ForegroundColor Cyan

# Navigate to .devcontainer directory
Set-Location $PSScriptRoot

# Stop any existing containers
Write-Host "Stopping existing containers..." -ForegroundColor Yellow
docker-compose down

# Remove old images
Write-Host "Removing old images..." -ForegroundColor Yellow
docker-compose rm -f

# Build fresh
Write-Host "Building new containers..." -ForegroundColor Green
docker-compose build --no-cache

Write-Host "`nDev container rebuilt successfully!" -ForegroundColor Green
Write-Host "Now in VS Code, press F1 and select 'Dev Containers: Reopen in Container'" -ForegroundColor Cyan
