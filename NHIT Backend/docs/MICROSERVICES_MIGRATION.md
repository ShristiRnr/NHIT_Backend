







# Microservices Migration Guide

## Overview

This document describes the migration from monolithic to hexagonal microservices architecture for the NHIT Backend project.

## Architecture

### Hexagonal Architecture (Ports & Adapters)

Each microservice follows the hexagonal architecture pattern:

```
service/
├── cmd/server/          # Application entry point
├── internal/
│   ├── core/
│   │   ├── domain/      # Business entities (pure Go structs)
│   │   ├── ports/       # Interfaces (repository & service contracts)
│   │   └── services/    # Business logic implementation
│   └── adapters/
│       ├── grpc/        # gRPC handlers (input adapter)
│       ├── repository/  # Database implementation (output adapter)
│       └── clients/     # External service clients (output adapter)
```

### Microservices

#### 1. **User Service** (Port: 50051)
- **Responsibilities:**
  - User CRUD operations
  - User profile management
  - User-role assignments
  
- **Domain Models:**
  - User
  - UserRole
  - Role

- **Endpoints:**
  - `CreateUser`
  - `GetUser`
  - `UpdateUser`
  - `DeleteUser`
  - `ListUsers`
  - `AssignRolesToUser`

#### 2. **Auth Service** (Port: 50052)
- **Responsibilities:**
  - Authentication (login/logout)
  - Authorization (JWT tokens)
  - Password reset
  - Email verification
  - Session management
  - Refresh tokens

- **Domain Models:**
  - Session
  - RefreshToken
  - PasswordReset
  - EmailVerificationToken

- **Endpoints:**
  - `Login`
  - `Logout`
  - `RefreshToken`
  - `ForgotPassword`
  - `ResetPassword`
  - `VerifyEmail`
  - `SendVerificationEmail`

#### 3. **Organization Service** (Port: 50053)
- **Responsibilities:**
  - Organization CRUD operations
  - User-organization associations
  - Tenant management

- **Domain Models:**
  - Organization
  - UserOrganization
  - Tenant

- **Endpoints:**
  - `CreateOrganization`
  - `GetOrganization`
  - `UpdateOrganization`
  - `DeleteOrganization`
  - `ListOrganizations`
  - `AddUserToOrganization`
  - `RemoveUserFromOrganization`

### Shared Components

#### `services/shared/`
- **config/**: Centralized configuration management
- **database/**: Database connection utilities
- **proto/**: Protocol buffer definitions (shared across services)

## Database Strategy

All services share the same PostgreSQL database but access different tables:

- **User Service**: `users`, `user_roles`, `roles`
- **Auth Service**: `sessions`, `refresh_tokens`, `password_resets`, `email_verification_tokens`
- **Organization Service**: `organizations`, `user_organizations`, `tenants`

## Communication

- **Inter-service**: gRPC (efficient, type-safe)
- **External clients**: REST API via API Gateway (to be implemented)

## Deployment

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

### Individual Service

```bash
# User Service
cd services/user-service
go run cmd/server/main.go

# Auth Service
cd services/auth-service
go run cmd/server/main.go

# Organization Service
cd services/organization-service
go run cmd/server/main.go
```

## Environment Variables

Each service requires:

```env
SERVER_PORT=<port>
DB_URL=postgres://user:pass@localhost:5432/nhit?sslmode=disable
JWT_SECRET=your-secret-key
JWT_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=168h

# Service discovery
USER_SERVICE_URL=localhost:50051
AUTH_SERVICE_URL=localhost:50052
ORG_SERVICE_URL=localhost:50053
```

## Migration Steps

### 1. Fixed Compilation Errors ✅
- Added `EmailVerificationToken` model
- Fixed `PasswordReset` model (UserID instead of Email)
- Fixed `Session` model (SessionToken field name)
- Removed duplicate `AssignRoleToUser` queries
- Updated repository and service signatures

### 2. Created Microservices Structure ✅
- User Service with hexagonal architecture
- Auth Service with hexagonal architecture
- Organization Service with hexagonal architecture
- Shared configuration and database utilities

### 3. Next Steps

#### A. Complete Service Implementations
- Implement auth service business logic
- Implement organization service business logic
- Add JWT token generation/validation
- Add password hashing utilities

#### B. Update Proto Files
- Ensure all proto files match service interfaces
- Generate gRPC code: `make proto`

#### C. Create API Gateway
- HTTP REST endpoints
- gRPC client connections to services
- Request routing and aggregation
- Authentication middleware

#### D. Add Observability
- Logging (structured logging with zerolog/zap)
- Metrics (Prometheus)
- Tracing (OpenTelemetry/Jaeger)
- Health checks

#### E. Testing
- Unit tests for business logic
- Integration tests for repositories
- E2E tests for gRPC endpoints

## Benefits of This Architecture

### 1. **Separation of Concerns**
- Each service has a single responsibility
- Clear boundaries between services

### 2. **Scalability**
- Services can be scaled independently
- Resource allocation based on load

### 3. **Maintainability**
- Smaller, focused codebases
- Easier to understand and modify

### 4. **Technology Flexibility**
- Can use different technologies per service
- Easy to replace/upgrade individual services

### 5. **Fault Isolation**
- Failure in one service doesn't bring down entire system
- Better resilience

### 6. **Team Autonomy**
- Different teams can own different services
- Parallel development

## Hexagonal Architecture Benefits

### 1. **Testability**
- Business logic isolated from infrastructure
- Easy to mock dependencies

### 2. **Flexibility**
- Easy to swap adapters (e.g., change database)
- Framework-independent core

### 3. **Clear Dependencies**
- Dependencies point inward (toward domain)
- No circular dependencies

## Development Workflow

1. **Define Domain Models** (`internal/core/domain/`)
2. **Define Port Interfaces** (`internal/core/ports/`)
3. **Implement Business Logic** (`internal/core/services/`)
4. **Implement Adapters**:
   - Repository adapters (`internal/adapters/repository/`)
   - gRPC handlers (`internal/adapters/grpc/`)
5. **Wire Everything** (`cmd/server/main.go`)

## Proto Generation

```bash
# Generate Go code from proto files
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/*.proto
```

## Database Migrations

Use existing SQL schema. Services share the database but access different tables based on their domain.

## Monitoring

- **Health Checks**: Each service exposes `/health` endpoint
- **Metrics**: Prometheus metrics on `/metrics`
- **Logs**: Structured JSON logs
- **Tracing**: Distributed tracing with correlation IDs

## Security

- **JWT Authentication**: Handled by Auth Service
- **Authorization**: Role-based access control (RBAC)
- **TLS**: gRPC connections use TLS in production
- **Secrets Management**: Environment variables or secret manager

## Performance Considerations

- **Connection Pooling**: Database connections pooled per service
- **Caching**: Redis for session/token caching (future)
- **Rate Limiting**: Per-service rate limits
- **Circuit Breakers**: Prevent cascade failures

## Troubleshooting

### Service Won't Start
- Check database connection
- Verify environment variables
- Check port availability

### gRPC Connection Errors
- Verify service URLs
- Check network connectivity
- Ensure services are running

### Database Errors
- Check connection string
- Verify database schema
- Check permissions

## References

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
- [Microservices Patterns](https://microservices.io/patterns/)
