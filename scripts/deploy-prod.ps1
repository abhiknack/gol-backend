# Production Deployment Script
# Deploys the grocery superapp in production mode

Write-Host "=== Gol Bazaar - Production Deployment ===" -ForegroundColor Cyan
Write-Host ""

# Check if .env.production exists
if (-not (Test-Path ".env.production")) {
    Write-Host "✗ Error: .env.production file not found!" -ForegroundColor Red
    Write-Host "Please create .env.production with production credentials" -ForegroundColor Yellow
    exit 1
}

# Warning about production deployment
Write-Host "WARNING: You are about to deploy to PRODUCTION!" -ForegroundColor Red
Write-Host ""
$confirmation = Read-Host "Are you sure you want to continue? (yes/no)"
if ($confirmation -ne "yes") {
    Write-Host "Deployment cancelled." -ForegroundColor Yellow
    exit 0
}

# Copy production env to .env
Write-Host ""
Write-Host "Using production environment configuration..." -ForegroundColor Yellow
Copy-Item ".env.production" ".env" -Force

# Build version
$buildVersion = Read-Host "Enter build version (default: latest)"
if ([string]::IsNullOrWhiteSpace($buildVersion)) {
    $buildVersion = "latest"
}
$env:BUILD_VERSION = $buildVersion

# Pull latest code (if using git)
Write-Host ""
Write-Host "Pulling latest code..." -ForegroundColor Yellow
# git pull origin main

# Build images
Write-Host ""
Write-Host "Building production images..." -ForegroundColor Green
docker-compose -f docker-compose.prod.yml build --no-cache

# Tag images
Write-Host "Tagging images as version $buildVersion..." -ForegroundColor Yellow
docker tag gol-bazaar-app:latest gol-bazaar-app:$buildVersion

# Stop existing containers (zero-downtime deployment)
Write-Host ""
Write-Host "Deploying new version..." -ForegroundColor Green
docker-compose -f docker-compose.prod.yml up -d --no-deps --build app

# Wait for new containers to be healthy
Write-Host ""
Write-Host "Waiting for services to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Check service status
Write-Host ""
Write-Host "Service Status:" -ForegroundColor Cyan
docker-compose -f docker-compose.prod.yml ps

# Test health endpoint
Write-Host ""
Write-Host "Testing health endpoint..." -ForegroundColor Cyan
Start-Sleep -Seconds 5
try {
    $response = Invoke-WebRequest -Uri "http://localhost/health" -UseBasicParsing
    Write-Host "✓ Application is healthy!" -ForegroundColor Green
} catch {
    Write-Host "✗ Application health check failed" -ForegroundColor Red
    Write-Host "Rolling back..." -ForegroundColor Yellow
    docker-compose -f docker-compose.prod.yml down
    exit 1
}

# Create backup
Write-Host ""
Write-Host "Creating database backup..." -ForegroundColor Yellow
$backupDate = Get-Date -Format "yyyyMMdd_HHmmss"
docker exec gol-bazaar-postgres-prod pg_dump -U gol_bazaar_app -d gol_bazaar_production -F c -f /backups/backup_$backupDate.dump

# Display access information
Write-Host ""
Write-Host "=== Production Deployment Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Application:      http://localhost (via Nginx)" -ForegroundColor White
Write-Host "Health Check:     http://localhost/health" -ForegroundColor White
Write-Host "Build Version:    $buildVersion" -ForegroundColor White
Write-Host ""
Write-Host "View logs: docker-compose -f docker-compose.prod.yml logs -f" -ForegroundColor Yellow
Write-Host "Monitor: docker stats" -ForegroundColor Yellow
Write-Host ""
Write-Host "IMPORTANT: Verify all services are working correctly!" -ForegroundColor Red
Write-Host ""
