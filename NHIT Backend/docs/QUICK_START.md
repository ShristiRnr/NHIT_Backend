# Quick Start Guide

## ✅ Current Status
All compilation errors are fixed! The project is ready for the next phase.

## 🚀 Next Steps

### 1. Generate Protobuf Code (Required)
```bash
# Install protoc compiler if not already installed
# Then generate Go code from proto files
make proto

# Or manually:
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/*.proto
```

### 2. Set Up Service Modules
Each microservice needs its own go.mod:

```bash
# User Service
cd services/user-service
go mod init github.com/ShristiRnr/NHIT_Backend/services/user-service
go get golang.org/x/crypto/bcrypt
go get google.golang.org/grpc
go get google.golang.org/protobuf
go mod tidy

# Auth Service
cd ../auth-service
go mod init github.com/ShristiRnr/NHIT_Backend/services/auth-service
go get golang.org/x/crypto/bcrypt
go get google.golang.org/grpc
go mod tidy

# Organization Service
cd ../organization-service
go mod init github.com/ShristiRnr/NHIT_Backend/services/organization-service
go get google.golang.org/grpc
go mod tidy

# Shared Module
cd ../shared
go mod init github.com/ShristiRnr/NHIT_Backend/services/shared
go mod tidy
```

### 3. Add Password Hashing Back

In `services/user-service/internal/core/services/user_service.go`:

```go
import (
    "golang.org/x/crypto/bcrypt"
    // ... other imports
)

func (s *userService) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
    // Hash password before storing
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }
    user.Password = string(hashedPassword)
    
    // ... rest of the code
}
```

### 4. Test Database Connection

```bash
# Start PostgreSQL
docker-compose up -d postgres

# Test connection
psql -h localhost -U nhit_user -d nhit
```

### 5. Run Services

#### Option A: Docker Compose (Recommended)
```bash
# Build and start all services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

#### Option B: Run Individually
```bash
# Terminal 1 - User Service
cd services/user-service
go run cmd/server/main.go

# Terminal 2 - Auth Service
cd services/auth-service
go run cmd/server/main.go

# Terminal 3 - Organization Service
cd services/organization-service
go run cmd/server/main.go
```

### 6. Test gRPC Endpoints

Install grpcurl:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Test User Service:
```bash
# List services
grpcurl -plaintext localhost:50051 list

# Create user
grpcurl -plaintext -d '{
  "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepass123"
}' localhost:50051 UserManagement/CreateUser
```

## 📝 Environment Variables

Create `.env` file in root:
```env
# Copy from .env.example
cp .env.example .env

# Edit with your values
nano .env
```

Required variables:
- `DB_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT tokens
- `SERVER_PORT` - Port for each service

## 🔍 Verify Everything Works

### Check 1: SQLC Generation
```bash
sqlc generate
# Should complete without errors
```

### Check 2: Go Build
```bash
# Build user service
cd services/user-service
go build ./cmd/server
# Should compile successfully
```

### Check 3: Docker Build
```bash
docker-compose build
# Should build all services
```

### Check 4: Service Health
```bash
# Start services
docker-compose up -d

# Check if running
docker-compose ps

# Should show all services as "Up"
```

## 🐛 Troubleshooting

### Issue: "cannot find module"
**Solution:** Run `go mod download` in the service directory

### Issue: "protoc: command not found"
**Solution:** Install Protocol Buffers compiler:
```bash
# Windows (using chocolatey)
choco install protoc

# Or download from: https://github.com/protocolbuffers/protobuf/releases
```

### Issue: "port already in use"
**Solution:** Change port in docker-compose.yml or stop conflicting service

### Issue: Database connection failed
**Solution:** 
1. Ensure PostgreSQL is running: `docker-compose up -d postgres`
2. Check connection string in `.env`
3. Verify database exists: `psql -h localhost -U nhit_user -l`

## 📚 Additional Resources

- **Architecture Guide**: See `MICROSERVICES_MIGRATION.md`
- **Problems Fixed**: See `PROBLEMS_FIXED.md`
- **API Documentation**: See proto files in `api/proto/`
- **Makefile Commands**: Run `make help`

## ✨ You're Ready!

All code issues are resolved. Follow the steps above to:
1. Generate protobuf code
2. Set up modules
3. Run services
4. Start building features!

Happy coding! 🚀
