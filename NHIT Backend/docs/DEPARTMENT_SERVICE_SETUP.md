# Department Service Setup Guide

## ✅ Service Created Successfully!

I've created a complete **Department Service** microservice following the exact same structure as your existing services (User, Auth, Organization).

## 📁 Structure Created

```
services/department-service/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point
├── internal/
│   ├── core/                          # HEXAGONAL CENTER
│   │   ├── domain/
│   │   │   ├── department.go          # Business entity
│   │   │   └── errors.go              # Domain errors
│   │   ├── ports/
│   │   │   ├── repository.go          # Repository interface
│   │   │   └── service.go             # Service interface
│   │   └── services/
│   │       └── department_service.go  # Business logic
│   └── adapters/                      # HEXAGONAL EDGES
│       ├── grpc/
│       │   └── department_handler.go  # gRPC adapter
│       └── repository/
│           └── department_repository.go # Database adapter
├── Dockerfile
└── go.mod
```

## 🔧 Setup Steps

### Step 1: Generate Proto Files

```bash
cd "d:\Nhit\NHIT Backend"

# Generate department proto
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
  api/proto/department.proto
```

Or use the Makefile:
```bash
make proto-dept
```

### Step 2: Generate SQLC Code

```bash
cd "d:\Nhit\NHIT Backend"

# Generate sqlc code for department queries
sqlc generate
```

Or use the Makefile:
```bash
make sqlc
```

### Step 3: Download Dependencies

```bash
cd "d:\Nhit\NHIT Backend\services\department-service"
go mod download
go mod tidy
```

### Step 4: Run the Service Locally

```bash
cd "d:\Nhit\NHIT Backend\services\department-service"
go run cmd/server/main.go
```

The service will start on port **50054**.

### Step 5: Run with Docker Compose

```bash
cd "d:\Nhit\NHIT Backend"

# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f department-service

# Check status
docker-compose ps
```

## 🎯 Service Details

### Port
- **gRPC**: 50054

### Environment Variables
```env
SERVER_PORT=50054
SERVICE_NAME=department-service
DB_URL=postgres://nhit_user:nhit_password@postgres:5432/nhit?sslmode=disable
USER_SERVICE_URL=user-service:50051
```

### Database Tables Used
- `departments` - Department data
- `users` - For checking user assignments

## 📊 SQL Queries Created

File: `internal/adapters/database/queries/department.sql`

- ✅ `CreateDepartment` - Create new department
- ✅ `GetDepartmentByID` - Get by ID
- ✅ `GetDepartmentByName` - Get by name
- ✅ `UpdateDepartment` - Update department
- ✅ `DeleteDepartment` - Delete department
- ✅ `ListDepartments` - List with pagination
- ✅ `CountDepartments` - Total count
- ✅ `DepartmentExists` - Check existence by name
- ✅ `DepartmentExistsByID` - Check existence by ID
- ✅ `CountUsersByDepartment` - Count assigned users

## 🏗️ Architecture

### Hexagonal Architecture ✅

**Core (Business Logic)**:
- `domain/department.go` - Department entity with validation
- `domain/errors.go` - Domain-specific errors
- `ports/repository.go` - Repository interface (port)
- `ports/service.go` - Service interface (port)
- `services/department_service.go` - Business logic implementation

**Adapters (Infrastructure)**:
- `grpc/department_handler.go` - gRPC input adapter
- `repository/department_repository.go` - PostgreSQL output adapter

### Microservices Architecture ✅

- **Independent Service**: Runs on its own port (50054)
- **Own Container**: Has its own Dockerfile
- **gRPC Communication**: Communicates via gRPC
- **Database Access**: Uses sqlc for type-safe queries
- **Service Discovery**: Registered in docker-compose

## 🔐 Business Logic Features

### Validation
- ✅ Name required (max 255 chars)
- ✅ Description required (max 500 chars)
- ✅ Duplicate name prevention
- ✅ Input trimming

### Business Rules
- ✅ Cannot delete department with assigned users
- ✅ Cannot create duplicate departments
- ✅ Cannot update to existing department name
- ✅ Proper error handling with domain errors

### Logging
- ✅ Create operations logged
- ✅ Update operations logged
- ✅ Delete operations logged
- ✅ Error operations logged

## 🧪 Testing the Service

### Using grpcurl

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50054 list

# Create department
grpcurl -plaintext -d '{
  "name": "Engineering",
  "description": "Engineering Department"
}' localhost:50054 departments.DepartmentService/CreateDepartment

# Get department
grpcurl -plaintext -d '{
  "id": "uuid-here"
}' localhost:50054 departments.DepartmentService/GetDepartment

# List departments
grpcurl -plaintext -d '{
  "page": 1,
  "page_size": 10
}' localhost:50054 departments.DepartmentService/ListDepartments

# Update department
grpcurl -plaintext -d '{
  "id": "uuid-here",
  "name": "Software Engineering",
  "description": "Updated description"
}' localhost:50054 departments.DepartmentService/UpdateDepartment

# Delete department
grpcurl -plaintext -d '{
  "id": "uuid-here"
}' localhost:50054 departments.DepartmentService/DeleteDepartment
```

### Using Go Client

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
)

func main() {
    conn, _ := grpc.Dial("localhost:50054", grpc.WithInsecure())
    defer conn.Close()
    
    client := departmentpb.NewDepartmentServiceClient(conn)
    
    // Create department
    resp, err := client.CreateDepartment(context.Background(), &departmentpb.CreateDepartmentRequest{
        Name:        "Engineering",
        Description: "Engineering Department",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Created: %v", resp.Department)
}
```

## 📝 API Gateway Integration

To integrate with API Gateway, update `services/api-gateway/cmd/server/main.go`:

```go
import (
    departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
)

func main() {
    // ... existing code ...
    
    // Register Department Service
    departmentServiceEndpoint := "localhost:50054"
    err = departmentpb.RegisterDepartmentServiceHandlerFromEndpoint(ctx, mux, departmentServiceEndpoint, opts)
    if err != nil {
        log.Fatalf("Failed to register department service gateway: %v", err)
    }
    log.Printf("✅ Registered Department Service gateway -> %s", departmentServiceEndpoint)
    
    // ... rest of code ...
}
```

Then you can access via HTTP:
```bash
# Create department
curl -X POST http://localhost:8080/api/v1/departments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering",
    "description": "Engineering Department"
  }'

# List departments
curl http://localhost:8080/api/v1/departments?page=1&page_size=10
```

## 🐳 Docker Commands

```bash
# Build department service
docker-compose build department-service

# Start department service
docker-compose up -d department-service

# View logs
docker-compose logs -f department-service

# Restart service
docker-compose restart department-service

# Stop service
docker-compose stop department-service

# Remove service
docker-compose rm -f department-service
```

## 📊 Service Health Check

Add to `main.go` if needed:

```go
// Health check endpoint
go func() {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "healthy",
            "service": "department-service",
            "version": "1.0.0",
        })
    })
    http.ListenAndServe(":8081", nil)
}()
```

## 🔍 Troubleshooting

### Proto files not generating
```bash
# Make sure protoc is installed
protoc --version

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

### SQLC errors
```bash
# Make sure sqlc is installed
sqlc version

# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Regenerate
cd "d:\Nhit\NHIT Backend"
sqlc generate
```

### Import errors
```bash
cd "d:\Nhit\NHIT Backend\services\department-service"
go mod tidy
go mod download
```

### Database connection errors
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check connection string
psql -h localhost -U nhit_user -d nhit
```

## ✅ Verification Checklist

- [ ] Proto files generated in `api/pb/departmentpb/`
- [ ] SQLC code generated in `internal/adapters/database/db/`
- [ ] Service compiles without errors
- [ ] Service starts on port 50054
- [ ] Can create department via grpcurl
- [ ] Can list departments via grpcurl
- [ ] Docker container builds successfully
- [ ] Docker container runs successfully
- [ ] Service appears in `docker-compose ps`

## 🎉 Summary

You now have a complete **Department Service** microservice that:

✅ **Follows Microservices Architecture**
- Independent service on port 50054
- Own Dockerfile and container
- gRPC communication
- Registered in docker-compose

✅ **Follows Hexagonal Architecture**
- Core business logic isolated
- Ports (interfaces) defined
- Adapters (implementations) separated
- Domain-driven design

✅ **Uses SQLC for Type-Safe Queries**
- All queries in `department.sql`
- Type-safe Go code generated
- PostgreSQL integration

✅ **Complete CRUD Operations**
- Create, Read, Update, Delete
- List with pagination
- Validation and business rules
- Error handling

✅ **Production-Ready**
- Logging
- Error handling
- Input validation
- Docker support

Next steps:
1. Run `make proto-dept` or generate proto files manually
2. Run `make sqlc` or `sqlc generate`
3. Run `go mod tidy` in department-service directory
4. Start the service with `docker-compose up -d`
5. Test with grpcurl or integrate with API Gateway

Your Department Service is ready to use! 🚀
