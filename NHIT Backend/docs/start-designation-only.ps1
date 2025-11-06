# Start ONLY Designation Service for Testing
# This bypasses broken services

Write-Host "🚀 Starting Designation Service (Standalone)" -ForegroundColor Cyan
Write-Host "============================================`n" -ForegroundColor Cyan

# Check if PostgreSQL is running
Write-Host "1. Checking PostgreSQL..." -ForegroundColor Yellow
$pgRunning = docker ps --filter "name=postgres" --format "{{.Names}}" | Select-String "postgres"

if (-not $pgRunning) {
    Write-Host "   Starting PostgreSQL..." -ForegroundColor Yellow
    docker run -d `
        --name nhit-postgres `
        --network nhitbackend_nhit-network `
        -e POSTGRES_USER=nhit_user `
        -e POSTGRES_PASSWORD=nhit_password `
        -e POSTGRES_DB=nhit `
        -p 5432:5432 `
        postgres:16-alpine
    
    Write-Host "   Waiting for PostgreSQL to be ready..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
}
Write-Host "   ✅ PostgreSQL is running`n" -ForegroundColor Green

# Start Designation Service
Write-Host "2. Starting Designation Service..." -ForegroundColor Yellow
docker run -d `
    --name nhit-designation-service `
    --network nhitbackend_nhit-network `
    -e DATABASE_URL="postgres://nhit_user:nhit_password@nhit-postgres:5432/nhit?sslmode=disable" `
    -e PORT="50055" `
    -p 50055:50055 `
    nhitbackend-designation-service

Start-Sleep -Seconds 5

# Check if it's running
$designationRunning = docker ps --filter "name=designation" --format "{{.Names}}" | Select-String "designation"

if ($designationRunning) {
    Write-Host "   ✅ Designation Service is running on port 50055`n" -ForegroundColor Green
    
    # Show logs
    Write-Host "3. Service Logs:" -ForegroundColor Yellow
    docker logs nhit-designation-service
    
    Write-Host "`n============================================" -ForegroundColor Cyan
    Write-Host "✅ Designation Service Started!" -ForegroundColor Green
    Write-Host "============================================`n" -ForegroundColor Cyan
    
    Write-Host "📝 Service Details:" -ForegroundColor Cyan
    Write-Host "- gRPC Port: 50055" -ForegroundColor White
    Write-Host "- Database: PostgreSQL (nhit)" -ForegroundColor White
    Write-Host "- Network: nhitbackend_nhit-network`n" -ForegroundColor White
    
    Write-Host "🧪 Test with grpcurl:" -ForegroundColor Cyan
    Write-Host 'grpcurl -plaintext localhost:50055 list' -ForegroundColor Gray
    Write-Host 'grpcurl -plaintext -d ''{"name":"Test","description":"Test desc","is_active":true}'' localhost:50055 designations.DesignationService/CreateDesignation' -ForegroundColor Gray
    
} else {
    Write-Host "   ❌ Failed to start Designation Service" -ForegroundColor Red
    Write-Host "   Check logs with: docker logs nhit-designation-service" -ForegroundColor Yellow
}
