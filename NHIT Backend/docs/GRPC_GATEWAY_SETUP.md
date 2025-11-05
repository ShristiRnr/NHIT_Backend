# gRPC Gateway Setup Guide

## ✅ What We've Accomplished

1. ✅ **Downloaded googleapis proto files** - HTTP annotation support
2. ✅ **Added HTTP annotations** to user_management.proto and auth.proto
3. ✅ **Created API Gateway service** - REST to gRPC proxy
4. ✅ **Created REST API examples** - Complete documentation

## 🔧 Installation Steps

### Step 1: Install protoc-gen-grpc-gateway

```bash
# Install the gateway generator
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Install OpenAPI/Swagger generator
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Verify installation
where protoc-gen-grpc-gateway
```

**Important:** Make sure your Go bin directory is in PATH:
```bash
# Add to PATH (Windows)
$env:PATH += ";$env:USERPROFILE\go\bin"

# Or permanently add: C:\Users\YourUsername\go\bin to System PATH
```

### Step 2: Generate Gateway Code

```bash
# Generate gRPC Gateway for User Service
protoc -I . -I third_party/googleapis \
  --grpc-gateway_out=. \
  --grpc-gateway_opt=paths=source_relative \
  --grpc-gateway_opt=generate_unbound_methods=true \
  api/proto/user_management.proto

# Generate gRPC Gateway for Auth Service
protoc -I . -I third_party/googleapis \
  --grpc-gateway_out=. \
  --grpc-gateway_opt=paths=source_relative \
  --grpc-gateway_opt=generate_unbound_methods=true \
  api/proto/auth.proto

# Optional: Generate OpenAPI/Swagger docs
protoc -I . -I third_party/googleapis \
  --openapiv2_out=. \
  --openapiv2_opt=allow_merge=true \
  --openapiv2_opt=merge_file_name=api_docs \
  api/proto/*.proto
```

### Step 3: Set Up API Gateway Module

```bash
cd services/api-gateway
go mod init github.com/ShristiRnr/NHIT_Backend/services/api-gateway

# Add dependencies
go get github.com/grpc-ecosystem/grpc-gateway/v2/runtime
go get google.golang.org/grpc
go get google.golang.org/protobuf

# Add replace directive for main module
# Edit go.mod and add:
# replace github.com/ShristiRnr/NHIT_Backend => ../..

go mod tidy
```

### Step 4: Build API Gateway

```bash
cd services/api-gateway
go build ./cmd/server
```

## 🚀 Running the Complete System

### Option 1: Manual (Development)

```bash
# Terminal 1 - User Service
cd services/user-service
go run cmd/server/main.go
# Should see: User Service listening on :50051

# Terminal 2 - Auth Service (when ready)
cd services/auth-service
go run cmd/server/main.go
# Should see: Auth Service listening on :50052

# Terminal 3 - API Gateway
cd services/api-gateway
go run cmd/server/main.go
# Should see: API Gateway listening on :8080
```

### Option 2: Docker Compose

Update `docker-compose.yml` to include API Gateway:

```yaml
  api-gateway:
    build:
      context: .
      dockerfile: services/api-gateway/Dockerfile
    container_name: nhit-api-gateway
    environment:
      SERVER_PORT: "8080"
      USER_SERVICE_URL: "user-service:50051"
      AUTH_SERVICE_URL: "auth-service:50052"
      ORG_SERVICE_URL: "organization-service:50053"
    ports:
      - "8080:8080"
    depends_on:
      - user-service
      - auth-service
      - organization-service
    networks:
      - nhit-network
```

Then run:
```bash
docker-compose up -d
```

## 🧪 Testing the Gateway

### Test 1: Health Check
```bash
curl http://localhost:8080/api/v1/users
```

### Test 2: Create User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Test User",
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Test 3: Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "test@example.com",
    "password": "password123",
    "tenant_id": "123e4567-e89b-12d3-a456-426614174000"
  }'
```

## 📊 Architecture Flow

```
┌──────────────┐
│   Client     │ (Browser/Mobile/Postman)
│  REST/JSON   │
└──────┬───────┘
       │ HTTP POST /api/v1/users
       │ Content-Type: application/json
       │ {"name": "John", "email": "john@example.com"}
       ↓
┌──────────────────────────┐
│    API Gateway :8080     │
│  (grpc-gateway)          │
│  ┌────────────────────┐  │
│  │ HTTP Handler       │  │
│  │ - CORS             │  │
│  │ - JSON Parsing     │  │
│  └────────┬───────────┘  │
│           │              │
│  ┌────────▼───────────┐  │
│  │ gRPC Gateway Mux   │  │
│  │ - Route matching   │  │
│  │ - JSON→Protobuf    │  │
│  └────────┬───────────┘  │
│           │              │
│  ┌────────▼───────────┐  │
│  │ gRPC Client        │  │
│  │ - Connection pool  │  │
│  └────────┬───────────┘  │
└───────────┼──────────────┘
            │ gRPC CreateUser(CreateUserRequest)
            │ Protobuf binary
            ↓
┌───────────────────────┐
│  User Service :50051  │
│  ┌─────────────────┐  │
│  │ gRPC Handler    │  │
│  └────────┬────────┘  │
│           │           │
│  ┌────────▼────────┐  │
│  │ Business Logic  │  │
│  └────────┬────────┘  │
│           │           │
│  ┌────────▼────────┐  │
│  │ Database        │  │
│  └─────────────────┘  │
└───────────┬───────────┘
            │ gRPC Response (Protobuf)
            ↓
┌──────────────────────────┐
│    API Gateway           │
│  Protobuf→JSON           │
└───────────┬──────────────┘
            │ HTTP 200 OK
            │ Content-Type: application/json
            │ {"user_id": "...", "name": "John"}
            ↓
┌──────────────┐
│   Client     │
└──────────────┘
```

## 🎯 Key Features

### 1. Automatic REST API Generation
- No manual REST handlers needed
- Generated from proto annotations
- Type-safe with protobuf validation

### 2. JSON ↔ Protobuf Conversion
- Automatic serialization/deserialization
- Field name mapping (snake_case ↔ camelCase)
- Proper error handling

### 3. HTTP Method Mapping
```
POST   → Create operations
GET    → Read operations
PUT    → Update operations
DELETE → Delete operations
```

### 4. Path Parameters
```
GET /api/v1/users/{user_id}
→ Extracts user_id from URL path
→ Maps to protobuf field
```

### 5. Query Parameters
```
GET /api/v1/users?tenant_id=123&limit=10
→ Extracts query params
→ Maps to protobuf fields
```

### 6. CORS Support
- Enabled for browser requests
- Configurable origins
- Preflight request handling

## 📝 Generated Files

After running protoc with grpc-gateway plugin:

```
api/pb/userpb/
├── user_management.pb.go          # Protobuf messages
├── user_management_grpc.pb.go     # gRPC service
└── user_management.pb.gw.go       # ✨ Gateway (NEW)

api/pb/authpb/
├── auth.pb.go                     # Protobuf messages
├── auth_grpc.pb.go                # gRPC service
└── auth.pb.gw.go                  # ✨ Gateway (NEW)
```

## 🔐 Adding Authentication

Update API Gateway to validate JWT tokens:

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip auth for login/register endpoints
        if strings.HasPrefix(r.URL.Path, "/api/v1/auth/login") ||
           strings.HasPrefix(r.URL.Path, "/api/v1/auth/register") {
            next.ServeHTTP(w, r)
            return
        }

        // Extract JWT token
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }

        // Validate token (implement JWT validation)
        // ...

        next.ServeHTTP(w, r)
    })
}
```

## 📚 Additional Resources

- [grpc-gateway Documentation](https://grpc-ecosystem.github.io/grpc-gateway/)
- [HTTP Annotations Reference](https://github.com/googleapis/googleapis/blob/master/google/api/http.proto)
- [OpenAPI/Swagger Generation](https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_openapi_output/)

## ✅ Checklist

- [ ] Install protoc-gen-grpc-gateway
- [ ] Add to PATH
- [ ] Generate gateway code
- [ ] Set up API Gateway module
- [ ] Build API Gateway
- [ ] Test REST endpoints
- [ ] Add authentication middleware
- [ ] Generate Swagger docs
- [ ] Update Docker Compose

## 🎉 Benefits

✅ **Clients can use REST/JSON** - No gRPC knowledge needed  
✅ **Single codebase** - Proto files define both gRPC and REST  
✅ **Type safety** - Validated by protobuf  
✅ **Auto-generated** - No manual REST handlers  
✅ **Performance** - Efficient binary gRPC internally  
✅ **Documentation** - Auto-generated Swagger/OpenAPI  

Your microservices now support both gRPC (internal) and REST (external) APIs! 🚀
