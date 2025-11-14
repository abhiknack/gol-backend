# Test script for POST /api/v1/products/push endpoint
# Usage: .\test-products-push.ps1 [base_url]
# Example: .\test-products-push.ps1 http://localhost:8080

param(
    [string]$BaseUrl = "http://localhost:8080"
)

$Endpoint = "$BaseUrl/api/v1/products/push"
$JsonFile = "docs/API-PRODUCTS-PUSH-EXAMPLE.json"

Write-Host "Testing POST $Endpoint" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

# Read JSON file
$Body = Get-Content $JsonFile -Raw

# Make request
try {
    $Response = Invoke-RestMethod -Uri $Endpoint -Method Post -Body $Body -ContentType "application/json"
    
    Write-Host "Response:" -ForegroundColor Green
    $Response | ConvertTo-Json -Depth 10
    
    Write-Host ""
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host "Test completed successfully" -ForegroundColor Green
}
catch {
    Write-Host "Error occurred:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
    Write-Host "Response:" -ForegroundColor Yellow
    $_.Exception.Response | ConvertTo-Json -Depth 10
}
