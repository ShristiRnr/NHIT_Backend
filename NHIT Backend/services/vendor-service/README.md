# Vendor Service

A comprehensive vendor management microservice built with Go, following hexagonal architecture principles. This service provides robust vendor and vendor account management capabilities with strong business logic validation.

## 🏗️ Architecture

The service follows **Hexagonal Architecture** (Ports and Adapters) pattern:

```
vendor-service/
├── cmd/server/              # Application entry point
├── internal/
│   ├── core/               # Business logic (Domain layer)
│   │   ├── domain/         # Domain models and business rules
│   │   ├── ports/          # Interfaces (Repository & Service)
│   │   └── services/       # Business logic implementation
│   └── adapters/           # External adapters
│       ├── grpc/           # gRPC handlers
│       └── repository/     # Database implementation
├── proto/                  # Protocol buffer definitions
├── migrations/             # Database migrations
└── go.mod
```

## 🚀 Features

### Vendor Management
- ✅ Create vendors with comprehensive validation
- ✅ Auto-generate vendor codes with customizable patterns
- ✅ Update vendor information with business rule validation
- ✅ List vendors with advanced filtering and pagination
- ✅ Soft delete vendors
- ✅ Duplicate prevention (email, code, mobile)

### Vendor Account Management
- ✅ Multiple banking accounts per vendor
- ✅ Primary account designation with automatic management
- ✅ IFSC code validation
- ✅ Account status management (active/inactive)
- ✅ Banking details retrieval for payment processing
- ✅ Account deletion with primary account reassignment

### Business Logic Features
- 🔒 **Strong Validation**: Email format, IFSC codes, PAN validation
- 🔄 **Transaction Support**: Database transactions for data consistency
- 🏷️ **Auto Code Generation**: Smart vendor code generation based on name and type
- 🔐 **Multi-tenancy**: Tenant-based data isolation
- 📋 **Audit Trail**: Created/updated timestamps and user tracking

## 🛠️ Technology Stack

- **Language**: Go 1.24.2
- **Database**: PostgreSQL with pgx/v5 driver
- **API**: gRPC with grpc-gateway for REST endpoints
- **Architecture**: Hexagonal (Ports & Adapters)
- **Database Queries**: Raw SQL with prepared statements
- **Validation**: Domain-level validation with business rules

## 📡 API Endpoints

### REST API (via grpc-gateway)
```
POST   /api/v1/vendors                    # Create vendor
GET    /api/v1/vendors/{id}               # Get vendor by ID
GET    /api/v1/vendors/code/{code}        # Get vendor by code
PUT    /api/v1/vendors/{id}               # Update vendor
DELETE /api/v1/vendors/{id}               # Delete vendor
GET    /api/v1/vendors                    # List vendors

POST   /api/v1/vendors/generate-code      # Generate vendor code
PUT    /api/v1/vendors/{id}/code          # Update vendor code
POST   /api/v1/vendors/{id}/regenerate-code # Regenerate vendor code

POST   /api/v1/vendors/{id}/accounts      # Create vendor account
GET    /api/v1/vendors/{id}/accounts      # Get vendor accounts
GET    /api/v1/vendors/{id}/banking-details # Get banking details
PUT    /api/v1/vendors/accounts/{id}      # Update vendor account
DELETE /api/v1/vendors/accounts/{id}      # Delete vendor account
POST   /api/v1/vendors/accounts/{id}/toggle-status # Toggle account status
```

### gRPC Service
- Full gRPC interface available on port `:50056`
- Protocol buffer definitions in `proto/vendor_service.proto`

## 🗄️ Database Schema

### Vendors Table
```sql
CREATE TABLE vendors (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    vendor_code VARCHAR(50) NOT NULL,
    vendor_name VARCHAR(255) NOT NULL,
    vendor_email VARCHAR(255) NOT NULL,
    vendor_mobile VARCHAR(20),
    -- ... (comprehensive vendor fields)
    is_active BOOLEAN DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Vendor Accounts Table
```sql
CREATE TABLE vendor_accounts (
    id UUID PRIMARY KEY,
    vendor_id UUID REFERENCES vendors(id),
    account_name VARCHAR(255) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    name_of_bank VARCHAR(255) NOT NULL,
    ifsc_code VARCHAR(11) NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    -- ... (additional banking fields)
);
```

## 🔧 Configuration

### Environment Variables
```bash
DATABASE_URL=postgres://username:password@localhost:5432/nhit_vendor?sslmode=disable
```

### Ports
- **gRPC Server**: `:50056`
- **HTTP Gateway**: `:8086`

## 🚦 Getting Started

### Prerequisites
- Go 1.24.2+
- PostgreSQL 12+
- Protocol Buffers compiler (protoc)

### Installation
1. **Clone and navigate to vendor service**:
   ```bash
   cd services/vendor-service
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run database migrations**:
   ```bash
   # Apply migrations to your PostgreSQL database
   psql -d nhit_vendor -f migrations/001_create_vendors_table.sql
   ```

4. **Generate protobuf files** (if needed):
   ```bash
   # From the root of the project
   protoc --go_out=. --go-grpc_out=. --grpc-gateway_out=. \
     services/vendor-service/proto/vendor_service.proto
   ```

5. **Run the service**:
   ```bash
   go run cmd/server/main.go
   ```

### Testing the API
```bash
# Create a vendor
curl -X POST http://localhost:8086/api/v1/vendors \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
    "vendor_name": "ABC Corp",
    "vendor_email": "contact@abccorp.com",
    "pan": "ABCDE1234F",
    "beneficiary_name": "ABC Corporation",
    "created_by": "123e4567-e89b-12d3-a456-426614174001"
  }'

# List vendors
curl http://localhost:8086/api/v1/vendors?tenant_id=123e4567-e89b-12d3-a456-426614174000
```

## 🔒 Business Rules

### Vendor Code Generation
- **Pattern**: `{TYPE}-{NAME_PREFIX}-{TIMESTAMP}`
- **Types**: `IN` (Internal), `EX` (External), `GN` (General)
- **Auto-collision handling**: Adds timestamp suffix if duplicate

### Account Management
- **Primary Account**: Only one primary account per vendor
- **Account Deletion**: Cannot delete primary account without reassigning
- **IFSC Validation**: Strict format validation (`[A-Z]{4}0[A-Z0-9]{6}`)

### Data Validation
- **Email**: RFC-compliant email validation
- **PAN**: Required for all vendors
- **Multi-tenancy**: All operations are tenant-scoped

## 🧪 Testing

The service includes comprehensive business logic validation:

- **Domain Model Tests**: Vendor and account creation/validation
- **Service Layer Tests**: Business logic and transaction handling
- **Repository Tests**: Database operations and constraints
- **Integration Tests**: End-to-end API testing

## 📈 Performance Considerations

- **Database Indexes**: Optimized indexes on frequently queried fields
- **Connection Pooling**: pgx connection pool for database efficiency
- **Transaction Management**: Proper transaction boundaries for data consistency
- **Pagination**: Built-in pagination for list operations

## 🔄 Integration

### API Gateway Integration
Add to your API gateway configuration:
```go
// Register Vendor Service
err = vendorpb.RegisterVendorServiceHandlerFromEndpoint(ctx, mux, "localhost:50056", opts)
if err != nil {
    log.Fatalf("Failed to register vendor service gateway: %v", err)
}
```

### Service Dependencies
- **Database**: PostgreSQL for data persistence
- **Authentication**: Integrates with your auth service for user context
- **Logging**: Structured logging for audit trails

## 📚 API Documentation

The service automatically exposes REST endpoints via grpc-gateway. Visit:
- **Base URL**: `http://localhost:8086/api/v1/vendors`
- **gRPC Reflection**: Available for tools like grpcurl or Postman

## 🤝 Contributing

1. Follow the hexagonal architecture patterns
2. Add comprehensive tests for new features
3. Update protobuf definitions for API changes
4. Maintain backward compatibility
5. Document business rules and validation logic

## 📄 License

This project is part of the NHIT Backend system.
