# Project Management Microservice

Complete project management service with tasks, members, and document management.

## Features

✅ **Project CRUD Operations**
- Create, Read, Update, Delete projects
- Search projects by name/code
- List projects by organization/manager
- Budget tracking and utilization

✅ **Project Members**
- Assign/remove members to projects
- Define roles within projects
- Track member assignments

✅ **Project Tasks**
- Create and manage tasks
- Assign tasks to members
- Track task status (TODO, IN_PROGRESS, COMPLETED, BLOCKED)
- Set due dates and priorities

✅ **Project Documents**
- Upload/download documents
- Document metadata tracking
- File size and type management

✅ **Analytics**
- Project statistics
- Budget utilization reports
- Status distribution

## Project Structure

```
project-service/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── core/
│   │   ├── domain/                 # Domain models
│   │   │   └── project.go
│   │   ├── ports/                  # Interfaces
│   │   │   ├── project_repository.go
│   │   │   └── project_service.go
│   │   └── services/               # Business logic
│   │       └── project_service.go
│   └── adapters/
│       ├── grpc/
│       │   └── handler/            # gRPC handlers
│       │       └── project_handler.go
│       └── repository/             # Database adapters
│           └── project_repository.go
├── migrations/                     # Database migrations
│   ├── 000001_create_projects_table.up.sql
│   └── 000001_create_projects_table.down.sql
├── go.mod
├── Dockerfile
└── README.md
```

## API Endpoints

### gRPC Server
- Port: `50057`

### HTTP Gateway (REST)
- Port: `8057`
- Base URL: `/api/v1/projects`

### Endpoints

**Projects:**
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/{id}` - Get project
- `PUT /api/v1/projects/{id}` - Update project
- `DELETE /api/v1/projects/{id}` - Delete project
- `GET /api/v1/projects/organization/{org_id}` - List by organization
- `GET /api/v1/projects/manager/{manager_id}` - List by manager
- `GET /api/v1/projects/search?search_term=...` - Search projects

**Members:**
- `POST /api/v1/projects/{id}/members` - Assign member
- `DELETE /api/v1/projects/{id}/members/{member_id}` - Remove member
- `GET /api/v1/projects/{id}/members` - List members

**Tasks:**
- `POST /api/v1/projects/{id}/tasks` - Create task
- `PUT /api/v1/projects/{id}/tasks/{task_id}` - Update task
- `GET /api/v1/projects/{id}/tasks/{task_id}` - Get task
- `GET /api/v1/projects/{id}/tasks` - List tasks
- `DELETE /api/v1/projects/{id}/tasks/{task_id}` - Delete task

**Documents:**
- `POST /api/v1/projects/{id}/documents` - Upload document
- `GET /api/v1/projects/{id}/documents/{doc_id}` - Get document
- `GET /api/v1/projects/{id}/documents` - List documents
- `DELETE /api/v1/projects/{id}/documents/{doc_id}` - Delete document

**Analytics:**
- `GET /api/v1/projects/statistics/{org_id}` - Get statistics

## Database Schema

### projects
- project_id (UUID, PK)
- tenant_id (UUID)
- org_id (UUID)
- project_code (VARCHAR)
- project_name (VARCHAR)
- description (TEXT)
- status (VARCHAR) - PLANNING, ACTIVE, ON_HOLD, COMPLETED, CANCELLED
- priority (VARCHAR) - LOW, MEDIUM, HIGH, CRITICAL
- start_date (TIMESTAMP)
- end_date (TIMESTAMP)
- actual_end_date (TIMESTAMP)
- budget (NUMERIC)
- actual_cost (NUMERIC)
- currency (VARCHAR)
- manager_id (UUID)
- client_name (VARCHAR)
- location (VARCHAR)
- is_active (BOOLEAN)
- created_by (UUID)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)

### project_members
- project_member_id (UUID, PK)
- project_id (UUID, FK)
- user_id (UUID)
- role_in_project (VARCHAR)
- assigned_date (TIMESTAMP)
- removal_date (TIMESTAMP)
- is_active (BOOLEAN)
- assigned_by (UUID)

### project_tasks
- task_id (UUID, PK)
- project_id (UUID, FK)
- task_name (VARCHAR)
- description (TEXT)
- assigned_to (UUID)
- status (VARCHAR) - TODO, IN_PROGRESS, COMPLETED, BLOCKED
- priority (VARCHAR)
- due_date (TIMESTAMP)
- completed_at (TIMESTAMP)
- created_by (UUID)

### project_documents
- document_id (UUID, PK)
- project_id (UUID, FK)
- document_name (VARCHAR)
- document_type (VARCHAR)
- file_path (VARCHAR)
- file_size (BIGINT)
- uploaded_by (UUID)
- uploaded_at (TIMESTAMP)

## Environment Variables

```bash
GRPC_PORT=50057                    # gRPC server port
HTTP_PORT=8057                     # HTTP gateway port
DATABASE_URL=postgres://...        # PostgreSQL connection string
```

## Running the Service

### Using Go
```bash
cd services/project-service
go run cmd/server/main.go
```

### Using Docker
```bash
docker build -t project-service .
docker run -p 50057:50057 -p 8057:8057 project-service
```

### With Docker Compose
```bash
docker-compose up project-service
```

## Generate Proto Files

```bash
# From project root
protoc -I . -I third_party/googleapis \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
  api/proto/project.proto
```

## Run Migrations

```bash
migrate -path migrations \
  -database "postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable" up
```

## Testing

Create a project:
```bash
curl -X POST http://localhost:8057/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "...",
    "org_id": "...",
    "project_code": "PROJ001",
    "project_name": "New Project",
    "description": "Project description",
    "status": "ACTIVE",
    "priority": "HIGH",
    "start_date": "2025-01-01T00:00:00Z",
    "budget": 100000,
    "currency": "INR",
    "manager_id": "...",
    "created_by": "..."
  }'
```

## Architecture

Follows **Hexagonal Architecture (Ports & Adapters)**:
- **Domain Layer**: Business logic and entities
- **Ports**: Interfaces for dependencies
- **Adapters**: External implementations (gRPC, DB, etc.)
- **Service Layer**: Orchestrates business operations

## Integration

- **Kafka Events**: Published on project creation/updates
- **Activity Logs**: All operations logged
- **Multi-tenant**: Isolated by tenant_id and org_id
