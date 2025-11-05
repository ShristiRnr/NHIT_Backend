# NHIT Backend - Microservices Architecture

A multi-tenant backend system built with **Hexagonal Architecture** and **Microservices** pattern using Go and gRPC.

## 🏗️ Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters) with three main microservices:

- **User Service**: User management and role assignments
- **Auth Service**: Authentication, authorization, and session management
- **Organization Service**: Organization and tenant management

### Architecture Diagram

```
┌─────────────────┐
│   API Gateway   │ (HTTP REST)
└────────┬────────┘
         │
    ┌────┴────┬────────┬────────┐
    │         │        │        │
┌───▼───┐ ┌──▼───┐ ┌──▼────┐   │
│ User  │ │ Auth │ │  Org  │   │
│Service│ │Service│ │Service│   │
└───┬───┘ └──┬───┘ └──┬────┘   │
    │        │        │         │
    └────────┴────────┴─────────┘
             │
      ┌──────▼──────┐
      │  PostgreSQL │
      └─────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 15+
- Protocol Buffers compiler (`protoc`)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/ShristiRnr/NHIT_Backend.git
cd NHIT_Backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Generate protobuf code**
```bash
make proto
```

4. **Start services with Docker**
```bash
docker-compose up -d
```

5. **Verify services are running**
```bash
docker-compose ps
```

## 📁 Project Structure

```
NHIT_Backend/
├── api/
│   ├── pb/                    # Generated protobuf code
│   └── proto/                 # Proto definitions
│       ├── auth.proto
│       └── user_management.proto
├── services/
│   ├── user-service/
│   │   ├── cmd/server/        # Entry point
│   │   ├── internal/
│   │   │   ├── core/
│   │   │   │   ├── domain/    # Business entities
│   │   │   │   ├── ports/     # Interfaces
│   │   │   │   └── services/  # Business logic
│   │   │   └── adapters/
│   │   │       ├── grpc/      # gRPC handlers
│   │   │       └── repository/# Database adapters
│   │   └── Dockerfile
│   ├── auth-service/          # Same structure
│   ├── organization-service/  # Same structure
│   └── shared/
│       ├── config/            # Shared configuration
│       └── database/          # Database utilities
├── internal/
│   ├── adapters/database/
│   │   ├── db/                # Generated sqlc code
│   │   └── queries/           # SQL queries
│   └── config/
├── docker-compose.yml
├── Makefile
└── README.md
```

## 🔧 Configuration

Each service uses environment variables for configuration:

```env
# Service Configuration
SERVER_PORT=50051
SERVICE_NAME=user-service

# Database
DB_URL=postgres://user:pass@localhost:5432/nhit?sslmode=disable

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=168h

# Service Discovery
USER_SERVICE_URL=localhost:50051
AUTH_SERVICE_URL=localhost:50052
ORG_SERVICE_URL=localhost:50053
```

## 🛠️ Development

### Run Individual Services

```bash
# User Service
make run-user

# Auth Service
make run-auth

# Organization Service
make run-org
```

### Build Services

```bash
# Build all services
make build

# Build specific service
make build-user
make build-auth
make build-org
```

### Generate Code

```bash
# Generate protobuf code
make proto

# Generate sqlc code
make sqlc
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## 📊 Services

### User Service (Port: 50051)

Manages user data and role assignments.

**Key Operations:**
- Create, Read, Update, Delete users
- Assign roles to users
- List users by tenant

### Auth Service (Port: 50052)

Handles authentication and authorization.

**Key Operations:**
- User login/logout
- JWT token generation
- Refresh tokens
- Password reset
- Email verification

### Organization Service (Port: 50053)

Manages organizations and tenants.

**Key Operations:**
- Create, Read, Update, Delete organizations
- Manage user-organization associations
- Tenant management

## 🗄️ Database

### Schema

The application uses PostgreSQL with the following main tables:

- `users` - User information
- `roles` - Role definitions
- `user_roles` - User-role associations
- `sessions` - Active sessions
- `refresh_tokens` - Refresh tokens
- `password_resets` - Password reset tokens
- `email_verification_tokens` - Email verification tokens
- `organizations` - Organization data
- `user_organizations` - User-organization associations
- `tenants` - Tenant information

### Migrations

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## 🐳 Docker

### Build Images

```bash
make docker-build
```

### Start Services

```bash
make docker-up
```

### View Logs

```bash
make docker-logs
```

### Stop Services

```bash
make docker-down
```

## 🧪 Testing

### Unit Tests

```bash
go test ./services/user-service/...
go test ./services/auth-service/...
go test ./services/organization-service/...
```

### Integration Tests

```bash
go test -tags=integration ./...
```

## 📝 API Documentation

### gRPC Services

- **User Service**: See `api/proto/user_management.proto`
- **Auth Service**: See `api/proto/auth.proto`

### Example gRPC Call

```go
// Create user
conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
client := userpb.NewUserManagementClient(conn)

req := &userpb.CreateUserRequest{
    TenantId: "tenant-uuid",
    Name:     "John Doe",
    Email:    "john@example.com",
    Password: "securepassword",
}

resp, err := client.CreateUser(context.Background(), req)
```

## 🔐 Security

- **Authentication**: JWT-based authentication
- **Authorization**: Role-based access control (RBAC)
- **Password Hashing**: bcrypt
- **TLS**: gRPC connections use TLS in production
- **Input Validation**: All inputs validated at service layer

## 📈 Monitoring

- **Health Checks**: `/health` endpoint on each service
- **Metrics**: Prometheus metrics on `/metrics`
- **Logging**: Structured JSON logging
- **Tracing**: Distributed tracing support

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 👥 Authors

- **Srishti** - Initial work

## 🙏 Acknowledgments

- Hexagonal Architecture pattern
- gRPC for efficient inter-service communication
- sqlc for type-safe SQL
- Docker for containerization

## 📚 Additional Documentation

- [Microservices Migration Guide](MICROSERVICES_MIGRATION.md)
- [API Documentation](docs/API.md)
- [Development Guide](docs/DEVELOPMENT.md)

## 🐛 Known Issues

See the [Issues](https://github.com/ShristiRnr/NHIT_Backend/issues) page for known issues and feature requests.

## 📞 Support

For support, email support@nhit.com or open an issue in the repository.
