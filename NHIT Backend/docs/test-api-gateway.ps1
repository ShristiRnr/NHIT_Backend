# Test API Gateway with Designation Service
# This script runs the API Gateway locally to test the Designation Service

Write-Host "🚀 Testing API Gateway with Designation Service" -ForegroundColor Cyan
Write-Host "================================================`n" -ForegroundColor Cyan

# Step 1: Check PostgreSQL
Write-Host "1. Checking PostgreSQL..." -ForegroundColor Yellow
$pgRunning = netstat -an | Select-String ":5432.*LISTENING"

if ($pgRunning) {
    Write-Host "   ✅ PostgreSQL is running on port 5432`n" -ForegroundColor Green
} else {
    Write-Host "   ❌ PostgreSQL is not running!" -ForegroundColor Red
    Write-Host "   Please start PostgreSQL first.`n" -ForegroundColor Yellow
    exit 1
}

# Step 2: Start Designation Service in background
Write-Host "2. Starting Designation Service..." -ForegroundColor Yellow
$designationJob = Start-Job -ScriptBlock {
    Set-Location "d:\Nhit\NHIT Backend"
    $env:DATABASE_URL = "postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
    $env:PORT = "50055"
    go run services/designation-service/cmd/server/main.go
}

Write-Host "   Waiting for Designation Service to start..." -ForegroundColor Gray
Start-Sleep -Seconds 8

# Check if designation service is running
$designationRunning = netstat -an | Select-String ":50055.*LISTENING"
if ($designationRunning) {
    Write-Host "   ✅ Designation Service is running on port 50055`n" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  Designation Service may not have started yet" -ForegroundColor Yellow
    Write-Host "   Check job output: Receive-Job $($designationJob.Id)`n" -ForegroundColor Gray
}

# Step 3: Start API Gateway in background
Write-Host "3. Starting API Gateway..." -ForegroundColor Yellow
$gatewayJob = Start-Job -ScriptBlock {
    Set-Location "d:\Nhit\NHIT Backend"
    go run services/api-gateway/cmd/server/main.go
}

Write-Host "   Waiting for API Gateway to start..." -ForegroundColor Gray
Start-Sleep -Seconds 8

# Check if API Gateway is running
$gatewayRunning = netstat -an | Select-String ":8080.*LISTENING"
if ($gatewayRunning) {
    Write-Host "   ✅ API Gateway is running on port 8080`n" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  API Gateway may not have started yet" -ForegroundColor Yellow
    Write-Host "   Check job output: Receive-Job $($gatewayJob.Id)`n" -ForegroundColor Gray
}

# Step 4: Show job outputs
Write-Host "4. Service Logs:" -ForegroundColor Yellow
Write-Host "`n--- Designation Service ---" -ForegroundColor Cyan
Receive-Job $designationJob.Id
Write-Host "`n--- API Gateway ---" -ForegroundColor Cyan
Receive-Job $gatewayJob.Id

Write-Host "`n================================================" -ForegroundColor Cyan
Write-Host "✅ Services Started!" -ForegroundColor Green
Write-Host "================================================`n" -ForegroundColor Cyan

Write-Host "📝 Service Status:" -ForegroundColor Cyan
Write-Host "- Designation Service: http://localhost:50055 (gRPC)" -ForegroundColor White
Write-Host "- API Gateway: http://localhost:8080 (HTTP REST)`n" -ForegroundColor White

Write-Host "🧪 Test Commands:" -ForegroundColor Cyan
Write-Host "`n# Create a designation" -ForegroundColor Yellow
Write-Host 'Invoke-RestMethod -Uri "http://localhost:8080/api/v1/designations" -Method POST -Body ''{"name":"Senior Engineer","description":"Senior level position","is_active":true}'' -ContentType "application/json"' -ForegroundColor Gray

Write-Host "`n# List all designations" -ForegroundColor Yellow
Write-Host 'Invoke-RestMethod -Uri "http://localhost:8080/api/v1/designations" -Method GET' -ForegroundColor Gray

Write-Host "`n`n⚠️  To stop services, run:" -ForegroundColor Yellow
Write-Host "Stop-Job $($designationJob.Id); Stop-Job $($gatewayJob.Id); Remove-Job $($designationJob.Id); Remove-Job $($gatewayJob.Id)" -ForegroundColor Gray

Write-Host "`n📊 Job IDs:" -ForegroundColor Cyan
Write-Host "- Designation Service Job: $($designationJob.Id)" -ForegroundColor White
Write-Host "- API Gateway Job: $($gatewayJob.Id)" -ForegroundColor White

Write-Host "`nPress Ctrl+C to exit (services will keep running in background)" -ForegroundColor Yellow
Write-Host "Or run the test commands above in a new terminal window!`n" -ForegroundColor Green

# Keep script running to show logs
while ($true) {
    Start-Sleep -Seconds 5
    
    # Show new logs
    $newDesignationLogs = Receive-Job $designationJob.Id
    $newGatewayLogs = Receive-Job $gatewayJob.Id
    
    if ($newDesignationLogs) {
        Write-Host "[Designation] $newDesignationLogs" -ForegroundColor Cyan
    }
    
    if ($newGatewayLogs) {
        Write-Host "[Gateway] $newGatewayLogs" -ForegroundColor Green
    }
}
