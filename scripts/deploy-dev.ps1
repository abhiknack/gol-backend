# Development Deployment Script
# Deploys the grocery superapp in development mode

Write-Host "=== Gol Bazaar - Development Deployment ===" -ForegroundColor Cyan
Write-Host ""

# Check if .env exists, if not copy from .env.development
if (-not (Test-Path ".env")) {
    Write-Host "Creating .env from .env.development..." -ForegroundColor Yellow
    Copy-Item ".env.development" ".env"
}

# Stop existing containers
Write-Host "Stopping existing containers..." -ForegroundColor Yellow
docker-compose -f docker-compose.dev.yml down

# Remove old volumes (optional - uncomment if you want fresh start)
# Write-Host "Removing old volumes..." -ForegroundColor Yellow
# docker volume rm grocery-postgres-dev-data grocery-redis-dev-data -f

# Build and start services
Write-Host "Building and starting services..." -ForegroundColor Green
docker-compose -f docker-compose.dev.yml up -d --build

# Wait for services to be healthy
Write-Host ""
Write-Host "Waiting for services to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Check service status
Write-Host ""
Write-Host "Service Status:" -ForegroundColor Cyan
docker-compose -f docker-compose.dev.yml ps

# Test health endpoint
Write-Host ""
Write-Host "Testing health endpoint..." -ForegroundColor Cyan
Start-Sleep -Seconds 5
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing
    Write-Host "✓ Application is healthy!" -ForegroundColor Green
} catch {
    Write-Host "✗ Application health check failed" -ForegroundColor Red
}

# Display access information
Write-Host ""
Write-Host "=== Development Environment Ready ===" -ForegroundColor Green
Write-Host ""
Write-Host "Application:      http://localhost:8080" -ForegroundColor White
Write-Host "Health Check:     http://localhost:8080/health" -ForegroundColor White
Write-Host "PostgreSQL:       localhost:5432" -ForegroundColor White
Write-Host "Redis:            localhost:6379" -ForegroundColor White
Write-Host "pgAdmin:          http://localhost:5050" -ForegroundColor White
Write-Host "Redis Commander:  http://localhost:8081" -ForegroundColor White
Write-Host ""
Write-Host "View logs: docker-compose -f docker-compose.dev.yml logs -f" -ForegroundColor Yellow
Write-Host "Stop services: docker-compose -f docker-compose.dev.yml down" -ForegroundColor Yellow
Write-Host ""
