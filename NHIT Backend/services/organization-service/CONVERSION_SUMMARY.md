# PHP to Go Conversion Summary - Organization Service

## Overview
This document summarizes the conversion of the PHP Laravel organization management code to a Go-based microservice using hexagonal architecture.

## PHP Code Analyzed

### Original PHP Components
1. **OrganizationSeeder.php** - Database seeding logic
2. **OrganizationContext.php** - Middleware for organization context switching
3. **OrganizationController.php** - HTTP controller with CRUD operations
4. **OrganizationPolicy.php** - Authorization policies

### Key PHP Features Identified
- Multi-database switching per organization
- Organization CRUD with image upload
- User-organization relationships
- Current organization context management
- Dashboard cache clearing
- Database cloning for new organizations
- Authorization via policies

## Go Implementation Structure

### Hexagonal Architecture Layers

```
services/organization-service/
├── cmd/
│   └── server/
│       └── main.go                          # Dependency injection & startup
├── internal/
│   ├── core/                                # Domain/Business Logic (Pure Go)
│   │   ├── domain/
│   │   │   └── organization.go             # Domain models with validation
│   │   ├── ports/
│   │   │   ├── repository.go               # Repository interfaces
│   │   │   └── service.go                  # Service interfaces
│   │   └── services/
│   │       ├── organization_service.go     # Organization business logic
│   │       └── user_organization_service.go # User-org relationship logic
│   └── adapters/                            # External dependencies
│       ├── grpc/
│       │   └── organization_handler.go     # gRPC request handlers
│       └── repository/
│           ├── organization_repository.go   # PostgreSQL implementation
│           └── user_organization_repository.go
├── migrations/
│   └── 001_create_organizations_tables.sql
└── README.md
```

## Conversion Decisions & Rationale

### 1. Separation of Concerns

**PHP (Monolithic)**
```php
class OrganizationController {
    public function store(Request $request) {
        // Validation
        // Business logic
        // Database operations
        // File handling
        // Background jobs
        // Cache management
    }
}
```

**Go (Hexagonal)**
```go
// Domain Layer - Pure business logic
type Organization struct {
    // Domain model with self-validation
}

// Service Layer - Business operations
type OrganizationService interface {
    CreateOrganization(ctx, tenantID, name, code, ...) (*Organization, error)
}

// Adapter Layer - Infrastructure
type organizationRepository struct {
    db *sql.DB
}
```

### 2. Organization vs UserOrganization Separation

**PHP** - Mixed concerns in controller and middleware

**Go** - Two separate services:
- `OrganizationService`: Manages organization lifecycle
- `UserOrganizationService`: Manages user-organization relationships

This provides better:
- Testability
- Maintainability
- Single Responsibility Principle adherence

### 3. Strong Business Validation

**PHP**
```php
$request->validate([
    'name' => 'required|string|max:255',
    'code' => 'required|string|max:10|unique:organizations,code',
]);
```

**Go (Domain Model)**
```go
func (o *Organization) Validate() error {
    if len(o.Name) < 3 || len(o.Name) > 255 {
        return ErrInvalidOrganizationName
    }
    if !o.IsValidCode() {
        return ErrInvalidOrganizationCode
    }
    return nil
}
```

Validation is:
- Part of the domain model
- Executed on creation and update
- Type-safe and explicit
- Reusable across all entry points

### 4. Database Name Generation

**PHP**
```php
public static function generateDatabaseName($code) {
    return config('database.connections.mysql.database') . '_' . strtolower($code);
}
```

**Go**
```go
func (o *Organization) GenerateDatabaseName() string {
    dbName := strings.ToLower(o.Code)
    dbName = strings.ReplaceAll(dbName, " ", "_")
    return fmt.Sprintf("org_%s", dbName)
}
```

Improvements:
- Part of domain model
- Validated automatically
- No configuration dependencies
- Consistent naming strategy

### 5. Context Switching (Critical Business Logic)

**PHP (Middleware)**
```php
public function handle(Request $request, Closure $next): Response {
    if (Auth::check()) {
        $user = Auth::user();
        $currentOrg = $user->currentOrganization();
        if ($currentOrg && $currentOrg->is_active) {
            $this->switchDatabaseContext($currentOrg->database_name);
        }
    }
    return $next($request);
}
```

**Go (Service Layer)**
```go
func (s *userOrganizationService) SwitchUserOrganization(
    ctx context.Context,
    userID, orgID uuid.UUID,
) error {
    // Validate access
    // Validate organization is active
    // Atomic transaction to switch context
    // Return database name for connection switching
}
```

Improvements:
- Explicit business operation
- Atomic transaction guarantees
- Better error handling
- Testable in isolation
- No middleware coupling

### 6. Authorization Removed from Service

**PHP**
```php
class OrganizationPolicy {
    public function view(User $user, Organization $organization): bool {
        return $user->hasRole('Super Admin') || $organization->created_by === $user->id;
    }
}
```

**Go** - Authorization moved to API Gateway or separate Auth service
- Microservices should focus on business logic
- Authorization is cross-cutting concern
- Handled by API Gateway with JWT/OAuth
- Service validates business rules, not permissions

### 7. Features NOT Implemented (Intentionally)

#### Removed from Service Scope
1. **File Upload** - Should be separate file storage service
2. **Cache Management** - Should be separate caching layer/service
3. **Database Cloning** - Infrastructure concern, not business logic
4. **Background Jobs** - Should use job queue service
5. **Email Notifications** - Should use notification service
6. **View/UI Logic** - Frontend responsibility

#### Rationale
Each microservice should have a single, well-defined responsibility. Cross-cutting concerns should be handled by dedicated services.

## Business Logic Comparison

### Organization Creation

**PHP Flow**
1. Validate request
2. Upload logo
3. Generate database name
4. Create organization record
5. Create database
6. Clone database structure
7. Handle errors

**Go Flow**
1. Create domain object (validates automatically)
2. Check code uniqueness
3. Persist to database
4. Return organization

Database creation/cloning is infrastructure concern, handled separately.

### User Organization Switching

**PHP Flow**
1. Check authentication (middleware)
2. Query organization
3. Cache check for migration status
4. Switch database connection
5. Clear dashboard cache
6. Return success

**Go Flow**
1. Validate user has access
2. Validate organization is active
3. Atomic transaction to update context
4. Return organization details

Database connection switching happens at API Gateway/application level based on returned organization info.

## Database Schema Changes

### Enhanced Fields
- `code`: Added unique constraint and validation
- `database_name`: Auto-generated from code
- `is_active`: Boolean for status management
- `created_by`: UUID reference to creator

### UserOrganization Table
- `is_current_context`: Replaces `current_organization_id`
- Atomic switching via transaction
- Indexed for performance

## Error Handling Improvements

**PHP**
```php
try {
    // operation
} catch (\Exception $e) {
    Log::error('Failed: ' . $e->getMessage());
    return back()->with('error', 'Failed');
}
```

**Go**
```go
// Custom error types
var (
    ErrOrganizationNotFound = errors.New("organization not found")
    ErrDuplicateCode = errors.New("organization code already exists")
    ErrUnauthorizedAccess = errors.New("unauthorized access")
)

// Service returns specific errors
if exists {
    return nil, domain.ErrDuplicateOrganizationCode
}

// gRPC handler maps to status codes
if err == domain.ErrDuplicateCode {
    return nil, status.Errorf(codes.AlreadyExists, "code exists")
}
```

Benefits:
- Type-safe error handling
- Explicit error types
- Better error context
- Easier to test

## Performance Considerations

### PHP
- Laravel ORM overhead
- Middleware runs on every request
- Cache queries with Redis
- Eager loading for relationships

### Go
- Compiled binary (faster startup)
- Direct SQL queries (no ORM overhead)
- Connection pooling built-in
- Prepared statements for security
- Pagination at database level

## Testing Strategy

### PHP
```php
public function test_create_organization() {
    $response = $this->post('/organizations', $data);
    $response->assertStatus(200);
}
```

### Go
```go
// Unit test - Domain
func TestOrganization_Validate(t *testing.T) {
    org := &Organization{Name: "AB"} // Too short
    err := org.Validate()
    assert.Equal(t, ErrInvalidOrganizationName, err)
}

// Unit test - Service (with mocks)
func TestOrganizationService_CreateOrganization(t *testing.T) {
    mockRepo := &MockOrganizationRepository{}
    service := NewOrganizationService(mockRepo, mockUserOrgRepo)
    // Test business logic in isolation
}

// Integration test - Repository
func TestOrganizationRepository_Create(t *testing.T) {
    // Test actual database operations
}
```

Benefits:
- Fast unit tests (no database)
- Integration tests separate
- Mock-friendly architecture
- High test coverage possible

## Migration Path

### Running Both Systems
1. Keep PHP system as primary
2. Deploy Go service alongside
3. Route specific operations to Go service
4. Gradually migrate more operations
5. Eventually deprecate PHP endpoints

### Data Migration
- Organizations table compatible
- User organizations needs migration
- Run migration script to convert

## Conclusion

### Key Improvements
✅ **Clean Architecture**: Hexagonal pattern with clear boundaries
✅ **Strong Typing**: Compile-time safety
✅ **Better Testability**: Isolated business logic
✅ **Performance**: Compiled, concurrent, efficient
✅ **Maintainability**: Clear separation of concerns
✅ **Scalability**: Microservice ready

### Trade-offs
- More initial code (boilerplate)
- Learning curve for hexagonal architecture
- Need to build missing infrastructure services
- More services to manage

### Recommended Next Steps
1. Run database migrations
2. Generate protobuf files: `protoc --go_out=...`
3. Run `go mod tidy` to resolve dependencies
4. Build and test the service
5. Deploy alongside existing PHP system
6. Implement infrastructure services (file storage, cache, etc.)
7. Set up API Gateway with authentication
8. Gradually migrate traffic from PHP to Go

## PHP vs Go Feature Matrix

| Feature | PHP Implementation | Go Implementation | Status |
|---------|-------------------|-------------------|--------|
| Create Organization | ✅ Controller | ✅ Service + gRPC | ✅ Converted |
| Update Organization | ✅ Controller | ✅ Service + gRPC | ✅ Converted |
| Delete Organization | ✅ Controller | ✅ Service + gRPC | ✅ Converted |
| Toggle Status | ✅ Controller | ✅ Service + gRPC | ✅ Converted |
| Check Code Availability | ❌ Not present | ✅ Service + gRPC | ✅ Added |
| List by Tenant | ✅ Controller | ✅ Service + gRPC | ✅ Converted |
| List Accessible | ✅ Controller | ✅ Service + gRPC | ✅ Converted |
| Add User to Org | ❌ Separate concern | ✅ UserOrgService | ✅ Added |
| Remove User | ❌ Separate concern | ✅ UserOrgService | ✅ Added |
| Switch Context | ✅ Middleware | ✅ UserOrgService | ✅ Converted |
| File Upload | ✅ Controller | ❌ External service | 🔄 Separated |
| Database Cloning | ✅ Model | ❌ Infrastructure | 🔄 Separated |
| Cache Management | ✅ Controller | ❌ External service | 🔄 Separated |
| Authorization | ✅ Policy | ❌ API Gateway | 🔄 Separated |

**Legend:**
- ✅ = Implemented
- ❌ = Intentionally not implemented
- 🔄 = Moved to separate concern
