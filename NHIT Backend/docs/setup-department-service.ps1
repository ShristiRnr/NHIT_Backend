# Department Service Setup Script
# Run this script to generate proto and sqlc code

Write-Host "🚀 Setting up Department Service..." -ForegroundColor Green

# Change to NHIT Backend directory
Set-Location "d:\Nhit\NHIT Backend"

Write-Host "`n📦 Step 1: Generating Proto files..." -ForegroundColor Cyan
protoc --go_out=. --go_opt=paths=source_relative `
  --go-grpc_out=. --go-grpc_opt=paths=source_relative `
  --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative `
  api/proto/department.proto

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Proto files generated successfully!" -ForegroundColor Green
} else {
    Write-Host "❌ Proto generation failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n📦 Step 2: Generating SQLC code..." -ForegroundColor Cyan
sqlc generate

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ SQLC code generated successfully!" -ForegroundColor Green
} else {
    Write-Host "❌ SQLC generation failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n📦 Step 3: Downloading dependencies..." -ForegroundColor Cyan
Set-Location "services\department-service"
go mod download
go mod tidy

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Dependencies downloaded successfully!" -ForegroundColor Green
} else {
    Write-Host "❌ Dependency download failed!" -ForegroundColor Red
    exit 1
}

Set-Location "..\..\"

Write-Host "`n✅ Department Service setup complete!" -ForegroundColor Green
Write-Host "`nNext steps:" -ForegroundColor Yellow
Write-Host "1. Run the service: cd services\department-service && go run cmd\server\main.go" -ForegroundColor White
Write-Host "2. Or use Docker: docker-compose up -d department-service" -ForegroundColor White
Write-Host "3. Test with grpcurl: grpcurl -plaintext localhost:50054 list" -ForegroundColor White
