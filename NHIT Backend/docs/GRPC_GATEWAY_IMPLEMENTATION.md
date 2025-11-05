# gRPC Gateway Implementation Guide

## 🎯 Overview

**grpc-gateway** is a plugin that generates a reverse-proxy server which translates RESTful HTTP API into gRPC. This allows clients to interact with your gRPC services using standard REST/JSON APIs.

## 🔄 Architecture Flow

```
┌─────────────┐
│   Client    │ (Mobile/Web/Postman)
│  REST/JSON  │
└──────┬──────┘
       │ HTTP/JSON (POST /api/v1/users)
       ↓
┌──────────────────┐
│  API Gateway     │
│  (grpc-gateway)  │
│  Port: 8080      │
└──────┬───────────┘
       │ gRPC (CreateUser)
       ↓
┌──────────────────┐
│  User Service    │
│  Port: 50051     │
└──────────────────┘
```

## 📦 What grpc-gateway Provides

1. **Automatic REST API Generation** - From proto annotations
2. **JSON ↔ Protobuf Conversion** - Seamless translation
3. **OpenAPI/Swagger Docs** - Auto-generated API documentation
4. **HTTP Method Mapping** - GET, POST, PUT, DELETE support
5. **Path Parameters** - RESTful URL patterns
6. **Query Parameters** - Filter and pagination support

## 🛠️ Implementation Steps

### Step 1: Install Dependencies

```bash
# Install protoc plugins
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Step 2: Update Proto Files with HTTP Annotations

Add HTTP annotations to your proto files:

```protobuf
import "google/api/annotations.proto";

service UserManagement {
  rpc CreateUser(CreateUserRequest) returns (UserResponse) {
    option (google.api.http) = {
      post: "/api/v1/users"
      body: "*"
    };
  }
  
  rpc GetUser(GetUserRequest) returns (UserResponse) {
    option (google.api.http) = {
      get: "/api/v1/users/{user_id}"
    };
  }
  
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {
    option (google.api.http) = {
      get: "/api/v1/users"
    };
  }
  
  rpc UpdateUser(UpdateUserRequest) returns (UserResponse) {
    option (google.api.http) = {
      put: "/api/v1/users/{user_id}"
      body: "*"
    };
  }
  
  rpc DeleteUser(DeleteUserRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/api/v1/users/{user_id}"
    };
  }
}
```

### Step 3: Generate Gateway Code

```bash
# Generate gRPC gateway code
protoc -I . \
  -I third_party/googleapis \
  --grpc-gateway_out=. \
  --grpc-gateway_opt=paths=source_relative \
  --grpc-gateway_opt=generate_unbound_methods=true \
  api/proto/*.proto

# Generate OpenAPI documentation
protoc -I . \
  -I third_party/googleapis \
  --openapiv2_out=. \
  --openapiv2_opt=allow_merge=true \
  --openapiv2_opt=merge_file_name=api \
  api/proto/*.proto
```

### Step 4: Create API Gateway Service

The gateway will:
- Listen on HTTP port (8080)
- Forward requests to gRPC services
- Convert JSON ↔ Protobuf
- Handle CORS
- Provide Swagger UI

## 🎨 Benefits for Your Architecture

### 1. **Client Flexibility**
- Mobile apps can use REST/JSON
- Web apps can use REST/JSON
- Internal services can use gRPC (faster)

### 2. **Single Entry Point**
- All REST requests go through gateway
- Centralized authentication
- Rate limiting
- Logging and monitoring

### 3. **Developer Experience**
- Standard REST APIs
- Auto-generated Swagger docs
- Easy testing with Postman/curl
- No need to learn gRPC for frontend

### 4. **Performance**
- HTTP/2 support
- Streaming support
- Efficient JSON encoding
- Connection pooling to gRPC services

## 📝 Example REST API Calls

Once implemented, you can use standard REST:

```bash
# Create User
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepass123"
  }'

# Get User
curl http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000

# List Users
curl http://localhost:8080/api/v1/users?tenant_id=123e4567-e89b-12d3-a456-426614174000

# Update User
curl -X PUT http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated",
    "email": "john.updated@example.com"
  }'

# Delete User
curl -X DELETE http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000
```

## 🔐 Authentication Flow

```
Client → REST API with JWT
    ↓
API Gateway (validates JWT)
    ↓
Adds user context to gRPC metadata
    ↓
Microservice (processes with user context)
```

## 📊 Comparison

| Feature | Direct gRPC | grpc-gateway |
|---------|-------------|--------------|
| Client Type | gRPC clients only | Any HTTP client |
| Data Format | Protobuf | JSON |
| Learning Curve | Steep | Easy (REST) |
| Performance | Fastest | Fast (small overhead) |
| Browser Support | Limited | Full |
| Tooling | Limited | Excellent (Postman, curl) |
| Documentation | Manual | Auto-generated Swagger |

## 🎯 Recommended Architecture

```
┌─────────────────────────────────────────┐
│          API Gateway (Port 8080)        │
│  ┌─────────────────────────────────┐   │
│  │   HTTP/REST Handler             │   │
│  │   - CORS                        │   │
│  │   - Authentication              │   │
│  │   - Rate Limiting               │   │
│  │   - Logging                     │   │
│  └─────────────┬───────────────────┘   │
│                │                         │
│  ┌─────────────▼───────────────────┐   │
│  │   grpc-gateway Reverse Proxy    │   │
│  │   - JSON ↔ Protobuf             │   │
│  │   - HTTP ↔ gRPC                 │   │
│  └─────────────┬───────────────────┘   │
└────────────────┼───────────────────────┘
                 │
        ┌────────┴────────┬────────────┐
        │                 │            │
   ┌────▼────┐      ┌────▼────┐  ┌───▼────┐
   │  User   │      │  Auth   │  │  Org   │
   │ Service │      │ Service │  │Service │
   │  :50051 │      │  :50052 │  │ :50053 │
   └─────────┘      └─────────┘  └────────┘
```

## 🚀 Next Steps

I'll implement this for you:

1. ✅ Download googleapis proto files
2. ✅ Add HTTP annotations to proto files
3. ✅ Generate gateway code
4. ✅ Create API Gateway service
5. ✅ Add CORS and middleware
6. ✅ Generate Swagger documentation
7. ✅ Update Docker Compose
8. ✅ Create example REST requests

Ready to proceed with implementation?
