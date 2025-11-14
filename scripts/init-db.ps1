# Initialize PostgreSQL database with sample data (PowerShell version)

Write-Host "Initializing PostgreSQL database..." -ForegroundColor Cyan

# Wait for PostgreSQL to be ready
Write-Host "Waiting for PostgreSQL to be ready..." -ForegroundColor Yellow
$ready = $false
$attempts = 0
$maxAttempts = 30

while (-not $ready -and $attempts -lt $maxAttempts) {
    $result = docker exec supabase-postgres-temp pg_isready -U postgres 2>&1
    if ($LASTEXITCODE -eq 0) {
        $ready = $true
    } else {
        Write-Host "PostgreSQL is unavailable - sleeping" -ForegroundColor Yellow
        Start-Sleep -Seconds 1
        $attempts++
    }
}

if (-not $ready) {
    Write-Host "✗ PostgreSQL failed to start after $maxAttempts attempts" -ForegroundColor Red
    exit 1
}

Write-Host "PostgreSQL is ready!" -ForegroundColor Green

# Run the initialization SQL script
Write-Host "Running initialization script..." -ForegroundColor Cyan
Get-Content init-postgres.sql | docker exec -i supabase-postgres-temp psql -U postgres -d middleware_db

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✓ Database initialized successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Connection details:" -ForegroundColor Cyan
    Write-Host "  Host: localhost"
    Write-Host "  Port: 5432"
    Write-Host "  Database: middleware_db"
    Write-Host "  Username: postgres"
    Write-Host "  Password: postgres"
    Write-Host ""
    Write-Host "DATABASE_URL=postgresql://postgres:postgres@localhost:5432/middleware_db?sslmode=disable" -ForegroundColor Yellow
} else {
    Write-Host "✗ Failed to initialize database" -ForegroundColor Red
    exit 1
}
