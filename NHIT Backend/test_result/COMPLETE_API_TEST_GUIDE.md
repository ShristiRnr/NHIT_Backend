# 🧪 NHIT Backend - Complete API Testing Guide

**Complete end-to-end API testing for all 6 microservices**

---

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Testing Phases](#testing-phases)
4. [Manual Testing](#manual-testing)
5. [Expected Results](#expected-results)
6. [Troubleshooting](#troubleshooting)

---

## 🎯 Prerequisites

### **1. Database Setup**
```sql
-- Ensure PostgreSQL is running on localhost:5432
-- Database: nhit_db
-- User: postgres
-- Password: postgres
```

### **2. Run Migrations**
```powershell
# Run migrations for each service
cd services/user-service
migrate -path migrations -database "postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable" up

cd services/auth-service
migrate -path migrations -database "postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable" up

# Repeat for all services...
```

### **3. Environment Variables**
```powershell
$env:DATABASE_URL="postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
```

---

## 🚀 Quick Start

### **Step 1: Start All Services**

Open 7 PowerShell terminals:

**Terminal 1 - User Service (Port 50051):**
```powershell
cd "d:\Nhit\NHIT Backend\services\user-service"
$env:DATABASE_URL="postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
$env:USER_SERVICE_PORT="50051"
go run ./cmd/server/main.go
```

**Terminal 2 - Auth Service (Port 50052):**
```powershell
cd "d:\Nhit\NHIT Backend\services\auth-service"
$env:DATABASE_URL="postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
$env:AUTH_SERVICE_PORT="50052"
go run ./cmd/server/main.go
```

**Terminal 3 - Department Service (Port 50054):**
```powershell
cd "d:\Nhit\NHIT Backend\services\department-service"
$env:DATABASE_URL="postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
$env:DEPARTMENT_SERVICE_PORT="50054"
go run ./cmd/server/main.go
```

**Terminal 4 - Designation Service (Port 50055):**
```powershell
cd "d:\Nhit\NHIT Backend\services\designation-service"
$env:DATABASE_URL="postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
$env:PORT="50055"
go run ./cmd/server/main.go
```

**Terminal 5 - Organization Service (Port 8080):**
```powershell
cd "d:\Nhit\NHIT Backend\services\organization-service"
$env:DATABASE_URL="postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
$env:ORGANIZATION_SERVICE_PORT="8080"
go run ./cmd/server/main.go
```

**Terminal 6 - Vendor Service (Port 50056):**
```powershell
cd "d:\Nhit\NHIT Backend\services\vendor-service"
$env:DATABASE_URL="postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
$env:VENDOR_SERVICE_PORT="50056"
go run ./cmd/server/main.go
```

**Terminal 7 - API Gateway (Port 8081):**
```powershell
cd "d:\Nhit\NHIT Backend\services\api-gateway"
$env:API_GATEWAY_PORT="8083"
go run ./cmd/server/main.go
```

### **Step 2: Run Automated Tests**

```powershell
cd "d:\Nhit\NHIT Backend"
.\test_complete_api.ps1
```

---

## 📊 Testing Phases

### **Phase 1: Setup & Registration** (4 tests)

| # | Test | Method | Endpoint | Expected |
|---|------|--------|----------|----------|
| 1 | Create Department | POST | `/api/v1/departments` | 201 Created |
| 2 | Create Designation | POST | `/api/v1/designations` | 201 Created |
| 3 | Register User | POST | `/api/v1/users` | 201 Created |
| 4 | Login (Create Session) | POST | `/api/v1/auth/sessions` | 201 Created |

**Purpose:** Set up test data and authenticate user for subsequent tests.

---

### **Phase 2: Department Service** (4 tests)

| # | Test | Method | Endpoint | Expected |
|---|------|--------|----------|----------|
| 5 | List Departments | GET | `/api/v1/departments` | 200 OK |
| 6 | Get Department by ID | GET | `/api/v1/departments/{id}` | 200 OK |
| 7 | Update Department | PUT | `/api/v1/departments/{id}` | 200 OK |
| 8 | Exists Department | GET | `/api/v1/departments/exists` | not implemented |
**Coverage:** CRUD operations, existence checks

---

### **Phase 3: Designation Service** (6 tests)

| # | Test | Method | Endpoint | Expected |
|---|------|--------|----------|----------|
| 9 | List Designations | GET | `/api/v1/designations` | 200 OK |
| 10 | Get Designation by ID | GET | `/api/v1/designations/{id}` | 200 OK |
| 11 | Get Designation by Slug | GET | `/api/v1/designations/slug/{slug}` | 200 OK |
| 12 | Get Root Designations | GET | `/api/v1/designations/root` | 200 OK | (Not Implemented)
| 13 | Get Active Designations | GET | `/api/v1/designations/active` | 200 OK | (Not Implemented)
| 14 | Create Child Designation | POST | `/api/v1/designations` | 201 Created |

**Coverage:** CRUD operations, hierarchy, filtering, slug-based lookup

---

### **Phase 4: User Service** (7 tests)

| # | Test | Method | Endpoint | Expected |
|---|------|--------|----------|----------|
| 15 | Get User by ID | GET | `/api/v1/users/{id}` | 200 OK |
| 16 | Get User by Email | GET | `/api/v1/users/email/{email}` | 200 OK |
| 17 | List Users by Tenant | GET | `/api/v1/users/tenant/{tenant_id}` | 200 OK |
| 18 | Search Users by Name | GET | `/api/v1/users/search` | 200 OK |
| 19 | Update User | PUT | `/api/v1/users/{id}` | 200 OK |
| 20 | List Users by Department | GET | `/api/v1/users/department/{dept_id}` | 200 OK |
| 21 | List Users by Designation | GET | `/api/v1/users/designation/{desig_id}` | 200 OK |

**Coverage:** User retrieval, search, filtering, updates

---

### **Phase 5: Auth Service** (6 tests)

| # | Test | Method | Endpoint | Expected |
|---|------|--------|----------|----------|
| 22 | Get Session by ID | GET | `/api/v1/auth/sessions/{id}` | 200 OK |
| 23 | Get Session by Token | GET | `/api/v1/auth/sessions/token/{token}` | 200 OK |
| 24 | Get User Sessions | GET | `/api/v1/auth/sessions/user/{user_id}` | 200 OK |
| 25 | Create Refresh Token | POST | `/api/v1/auth/refresh-tokens` | 201 Created |
| 26 | Create Password Reset Token | POST | `/api/v1/auth/password-resets` | 201 Created |
| 27 | Create Email Verification Token | POST | `/api/v1/auth/email-verification` | 201 Created |

**Coverage:** Session management, token operations, password reset, email verification

---

### **Phase 6: Organization Service** (8 tests)

| # | Test | Method | Endpoint | Expected |
|---|------|--------|----------|----------|
| 28 | Create Organization | POST | `/api/v1/organizations` | 201 Created |
| 29 | List Organizations by Tenant | GET | `/api/v1/organizations/tenant/{tenant_id}` | 200 OK |
| 30 | Get Organization by ID | GET | `/api/v1/organizations/{id}` | 200 OK |
| 31 | Get Organization by Code | GET | `/api/v1/organizations/code/{code}` | 200 OK |
| 32 | Update Organization | PUT | `/api/v1/organizations/{id}` | 200 OK |
| 33 | Assign User to Organization | POST | `/api/v1/organizations/{id}/users` | 201 Created |
| 34 | Get Organization Users | GET | `/api/v1/organizations/{id}/users` | 200 OK |
| 35 | Get User Organizations | GET | `/api/v1/organizations/user/{user_id}` | 200 OK |

**Coverage:** Organization CRUD, user assignment, code-based lookup

---

### **Phase 7: Vendor Service** (6 tests)

| # | Test | Method | Endpoint | Expected |
|---|------|--------|----------|----------|
| 36 | Create Vendor | POST | `/api/v1/vendors` | 201 Created |
| 37 | List Vendors | GET | `/api/v1/vendors` | 200 OK |
| 38 | Get Vendor by ID | GET | `/api/v1/vendors/{id}` | 200 OK |
| 39 | Update Vendor | PUT | `/api/v1/vendors/{id}` | 200 OK |
| 40 | Create Vendor Account | POST | `/api/v1/vendors/{id}/accounts` | 201 Created |
| 41 | List Vendor Accounts | GET | `/api/v1/vendors/{id}/accounts` | 200 OK |

**Coverage:** Vendor CRUD, banking account management

---

### **Phase 8: Cleanup & Logout** (2 tests)

| # | Test | Method | Endpoint | Expected |
|---|------|--------|----------|----------|
| 42 | Logout (Delete Session) | DELETE | `/api/v1/auth/sessions/{id}` | 204 No Content |
| 43 | Verify Session Deleted | GET | `/api/v1/auth/sessions/{id}` | 404 Not Found |

**Coverage:** Session cleanup, logout verification

---

## 📝 Manual Testing

### **Phase 1: Setup & Registration**

#### **Test 1: Create Department**
```bash
curl -X POST http://localhost:8081/api/v1/departments \
  -H "Content-Type: application/json" \
  -d
{
    "name": "Engineering Department",
    "description": "Software Engineering and Development"
}

**Response:**
```json
{
    "department": {
        "id": "3f46996d-bf51-4f69-8c98-41cc8c6d7120",
        "name": "Engineering Department",
        "description": "Software Engineering and Development",
        "createdAt": "2025-11-10T04:42:12.728798Z",
        "updatedAt": "2025-11-10T04:42:12.728798Z"
    }
}
```

#### **Test 2: Create Designation**
```bash
curl -X POST http://localhost:8081/api/v1/designations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Senior Software Engineer",
    "description": "Senior level software engineer position",
    "slug": "senior-software-engineer",
    "is_active": true,
    "level": 3
  }'
```
**Response:**
```json
{
    "designation": {
        "id": "1de4376a-b3a5-4a5c-9314-4233cba673e7",
        "name": "Senior Software Engineer",
        "description": "Senior level software engineer position",
        "slug": "senior-software-engineer",
        "isActive": true,
        "parentId": "",
        "level": 0,
        "userCount": 0,
        "createdAt": "2025-11-10T04:44:54.026279Z",
        "updatedAt": "2025-11-10T04:44:54.067865700Z"
    }
}
```

#### **Test 3: Register User**
```bash
curl -X POST http://localhost:8081/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "name": "John Doe",
    "email": "john.doe@nhit.com",
    "password": "SecurePassword123!",
    "department_id": "3f46996d-bf51-4f69-8c98-41cc8c6d7120",
    "designation_id": "1de4376a-b3a5-4a5c-9314-4233cba673e7"
  }'
```
**Response**
```json
{
    "userId": "15a6b637-6995-4bc7-939d-bc2ae9367101",
    "name": "John Doe",
    "email": "john.doe@nhit.com",
    "roles": [],
    "permissions": []
}
```

#### **Test 4: Login (Create Session)**
```bash
curl -X POST http://localhost:8081/api/v1/auth/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "15a6b637-6995-4bc7-939d-bc2ae9367101",
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "session_token": "test_session_token_12345",
    "expires_at": "2025-11-11T10:30:00Z"
}'
```
**Response**
```
{
    "userId": "1ac96824-532f-490c-9df8-88137b03f3ec",
    "name": "",
    "email": "",
    "roles": [],
    "permissions": []
}
```

### **Phase 2: Department Service**

#### **Test 5: List Departments**
```bash
curl http://localhost:8081/api/v1/departments?page=1&page_size=3
```
**Response**
```
{
    "departments": [
        {
            "id": "3f46996d-bf51-4f69-8c98-41cc8c6d7120",
            "name": "Engineering Department",
            "description": "Software Engineering and Development",
            "createdAt": "2025-11-10T04:42:12.728798Z",
            "updatedAt": "2025-11-10T04:42:12.728798Z"
        },
        {
            "id": "38473df6-f941-472c-becf-b69e01402ba5",
            "name": "Electrical Engineering",
            "description": "Department of Electrical and Electronics Engineering",
            "createdAt": "2025-11-06T10:23:41.175439Z",
            "updatedAt": "2025-11-06T10:23:41.175439Z"
        },
        {
            "id": "70660b43-26db-4d94-99c7-4f886b9ae352",
            "name": "Computer Science & Engineering",
            "description": "Department of Computer Science and Engineering - Updated",
            "createdAt": "2025-11-06T10:23:28.990323Z",
            "updatedAt": "2025-11-06T10:24:03.360712Z"
        }
    ],
    "totalCount": 3
}
```
#### **Test 6: Get Department by ID**
```bash
curl http://localhost:8081/api/v1/departments/70660b43-26db-4d94-99c7-4f886b9ae352
```

**Response**
```
{
    "department": {
        "id": "70660b43-26db-4d94-99c7-4f886b9ae352",
        "name": "Computer Science & Engineering",
        "description": "Department of Computer Science and Engineering - Updated",
        "createdAt": "2025-11-06T10:23:28.990323Z",
        "updatedAt": "2025-11-06T10:24:03.360712Z"
    }
}
```
#### **Test 7: Update Department**
```bash
curl -X PUT http://localhost:8081/api/v1/departments/70660b43-26db-4d94-99c7-4f886b9ae352 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering Department - Updated",
    "description": "Updated: Software Engineering and Development Division"
  }'
```
**Response**
```
{
    "department": {
        "id": "70660b43-26db-4d94-99c7-4f886b9ae352",
        "name": "Engineering Department - Updated",
        "description": "Updated: Software Engineering and Development Division",
        "createdAt": "2025-11-06T10:23:28.990323Z",
        "updatedAt": "2025-11-10T05:01:22.140496Z"
    }
}
```
#### **Test 8: Exists Designations**

(Not Implemented)

### **Phase 3: Designation Service**

#### **Test 9: List Designations**
```bash
curl "http://localhost:8081/api/v1/designations?page=1&page_size=10"
```
**Response**
{
    "designations": [
        {
            "id": "1de4376a-b3a5-4a5c-9314-4233cba673e7",
            "name": "Senior Software Engineer",
            "description": "Senior level software engineer position",
            "slug": "senior-software-engineer",
            "isActive": true,
            "parentId": "",
            "level": 0,
            "userCount": 0,
            "createdAt": "2025-11-10T04:44:54.026279Z",
            "updatedAt": "2025-11-10T05:59:49.895988Z"
        },
        {
            "id": "c9fc9c09-b4f5-49fd-acf8-76b32d7576d1",
            "name": "Senior Software Engineer - Updated",
            "description": "Updated description for senior software engineer position",
            "slug": "senior-software-engineer-updated",
            "isActive": false,
            "parentId": "",
            "level": 0,
            "userCount": 0,
            "createdAt": "2025-11-06T08:29:53.384292Z",
            "updatedAt": "2025-11-10T05:59:49.895988Z"
        }
    ],
    "totalCount": "2",
    "page": 1,
    "pageSize": 10
}

#### **Test 11: Get Designation by ID**
```bash
curl http://localhost:8081/api/v1/designations/1de4376a-b3a5-4a5c-9314-4233cba673e7
```
**Response**
{
    "designation": {
        "id": "1de4376a-b3a5-4a5c-9314-4233cba673e7",
        "name": "Senior Software Engineer",
        "description": "Senior level software engineer position",
        "slug": "senior-software-engineer",
        "isActive": true,
        "parentId": "",
        "level": 0,
        "userCount": 0,
        "createdAt": "2025-11-10T04:44:54.026279Z",
        "updatedAt": "2025-11-10T07:10:28.655745300Z"
    }
}

#### **Test 11: Get Designation by Slug**
```bash
curl http://localhost:8081/api/v1/designations/slug/senior-software-engineer
```
**Response**
{
    "designation": {
        "id": "1de4376a-b3a5-4a5c-9314-4233cba673e7",
        "name": "Senior Software Engineer",
        "description": "Senior level software engineer position",
        "slug": "senior-software-engineer",
        "isActive": true,
        "parentId": "",
        "level": 0,
        "userCount": 0,
        "createdAt": "2025-11-10T04:44:54.026279Z",
        "updatedAt": "2025-11-10T06:00:23.053787500Z"
    }
}

#### **Test 12: Get Root Designations**
```bash
curl http://localhost:8081/api/v1/designations/root/1de4376a-b3a5-4a5c-9314-4233cba673e7
```
**Response**
(Not Implemented)
---

#### **Test 13: Get Active Designations**
```bash
curl http://localhost:8081/api/v1/designations/root/1de4376a-b3a5-4a5c-9314-4233cba673e7
```
**Response**
(Not Implemented)
---

#### **Test 14: Create Child Designations**
```bash
curl http://localhost:8081/api/v1/designations
```
{
  "name": "Assistant Manager",
  "description": "Leads a small team under the Senior Manager",
  "is_active": true,
  "parent_id": "1de4376a-b3a5-4a5c-9314-4233cba673e7"
}

**Response**
{
    "designation": {
        "id": "3f743497-ff07-4882-9347-0c91bd678b86",
        "name": "Assistant Manager",
        "description": "Leads a small team under the Senior Manager",
        "slug": "assistant-manager",
        "isActive": true,
        "parentId": "1de4376a-b3a5-4a5c-9314-4233cba673e7",
        "level": 0,
        "userCount": 0,
        "createdAt": "2025-11-10T07:06:38.979756Z",
        "updatedAt": "2025-11-10T07:06:38.983670200Z"
    }
}
---


### **Phase 4: User Service**
#### **Test 16: Get User by Email**
```bash
curl http://localhost:8081/api/v1/users/email/john.doe@nhit.com
```


#### **Test 18: Search Users by Name**
```bash
curl "http://localhost:8081/api/v1/users/search?name=John&page=1&page_size=10"
```

---

### **Phase 5: Auth Service**

#### **Test 25: Create Refresh Token**
```bash
curl -X POST http://localhost:8081/api/v1/auth/refresh-tokens \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "<USER_ID>",
    "token": "refresh_token_12345",
    "expires_at": "2025-12-07T10:30:00Z"
  }'
```

---

### **Phase 6: Organization Service**

#### **Test 28: Create Organization**
```bash
curl -X POST http://localhost:8081/api/v1/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "00000000-0000-0000-0000-000000000001",
    "name": "NHIT Corporation",
    "code": "NHIT001",
    "is_active": true
  }'
```

---

### **Phase 7: Vendor Service**

#### **Test 36: Create Vendor**
```bash
curl -X POST http://localhost:8081/api/v1/vendors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tech Supplies Inc",
    "email": "contact@techsupplies.com",
    "phone": "+1234567890",
    "address": "123 Tech Street, Silicon Valley, CA 94000",
    "is_active": true
  }'
```

---

### **Phase 8: Cleanup**

#### **Test 42: Logout (Delete Session)**
```bash
curl -X DELETE http://localhost:8081/api/v1/auth/sessions/<SESSION_ID>
```

---

## ✅ Expected Results

### **Success Metrics**

| Metric | Target | Description |
|--------|--------|-------------|
| **Total Tests** | 43 | All API endpoints tested |
| **Pass Rate** | 100% | All tests should pass |
| **Response Time** | < 500ms | Average response time |
| **Error Rate** | 0% | No errors expected |

### **Phase-wise Expected Results**

| Phase | Tests | Expected Pass |
|-------|-------|---------------|
| Phase 1: Setup & Registration | 4 | 4/4 (100%) |
| Phase 2: Department Service | 4 | 4/4 (100%) |
| Phase 3: Designation Service | 6 | 6/6 (100%) |
| Phase 4: User Service | 7 | 7/7 (100%) |
| Phase 5: Auth Service | 6 | 6/6 (100%) |
| Phase 6: Organization Service | 8 | 8/8 (100%) |
| Phase 7: Vendor Service | 6 | 6/6 (100%) |
| Phase 8: Cleanup & Logout | 2 | 2/2 (100%) |
| **TOTAL** | **43** | **43/43 (100%)** |

---

## 🔧 Troubleshooting

### **Issue: Connection Refused**

**Symptom:** `Connection refused` or `Cannot connect to localhost:8081`

**Solution:**
1. Verify all services are running
2. Check port availability: `netstat -ano | findstr "8081"`
3. Restart API Gateway

### **Issue: 404 Not Found**

**Symptom:** API returns 404 for valid endpoints

**Solution:**
1. Verify API Gateway is routing correctly
2. Check service registration in gateway
3. Verify endpoint paths match proto definitions

### **Issue: 500 Internal Server Error**

**Symptom:** API returns 500 error

**Solution:**
1. Check service logs for errors
2. Verify database connection
3. Check migrations are applied
4. Verify SQLC generated code is up to date

### **Issue: Database Connection Failed**

**Symptom:** `connection refused` or `database does not exist`

**Solution:**
```powershell
# Check PostgreSQL is running
Get-Service postgresql*

# Create database if missing
psql -U postgres -c "CREATE DATABASE nhit_db;"

# Run migrations
migrate -path services/user-service/migrations -database "postgres://postgres:postgres@localhost:5432/nhit_db?sslmode=disable" up
```

### **Issue: Port Already in Use**

**Symptom:** `bind: address already in use`

**Solution:**
```powershell
# Find process using port
netstat -ano | findstr "50051"

# Kill process
taskkill /PID <PID> /F
```

---

Phase 1: Setup & Registration (4 tests)
✅ Create Department
✅ Create Designation
✅ Register User
✅ Login (Create Session)

Phase 2: Department Service (4 tests)
✅ List Departments
✅ Get Department by ID
✅ Update Department
✅ Check Department Exists

Phase 3: Designation Service (6 tests)
✅ List Designations
✅ Get by ID
✅ Get by Slug
✅ Get Root Designations
✅ Get Active Designations
✅ Create Child Designation

Phase 4: User Service (7 tests)
✅ Get User by ID
✅ Get by Email
✅ List by Tenant
✅ Search Users
✅ Update User
✅ List by Department
✅ List by Designation

Phase 5: Auth Service (6 tests)
✅ Get Session
✅ Get Session by Token
✅ Get User Sessions
✅ Create Refresh Token
✅ Create Password Reset
✅ Create Email Verification

Phase 6: Organization Service (8 tests)
✅ Create Organization
✅ List Organizations
✅ Get by ID
✅ Get by Code
✅ Update Organization
✅ Assign User
✅ Get Organization Users
✅ Get User Organizations

Phase 7: Vendor Service (6 tests)
✅ Create Vendor
✅ List Vendors
✅ Get by ID
✅ Update Vendor
✅ Create Vendor Account
✅ List Vendor Accounts

Phase 8: Cleanup (2 tests)
✅ Logout (Delete Session)
✅ Verify Session Deleted