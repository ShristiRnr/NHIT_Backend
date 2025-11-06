# 🚀 Start API Gateway & Designation Service

## ✅ **Quick Start (3 Terminals)**

You need **3 PowerShell terminals** open in `d:\Nhit\NHIT Backend`:

---

### **Terminal 1: PostgreSQL (Already Running ✅)**

Your PostgreSQL is already running on port 5432 with:
- User: `postgres`
- Password: `shristi`
- Database: `nhit`

✅ **No action needed - PostgreSQL is ready!**

---

### **Terminal 2: Start Designation Service**

```powershell
# Navigate to project
cd "d:\Nhit\NHIT Backend"

# Set environment variables
$env:DATABASE_URL = "postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
$env:PORT = "50055"

# Run designation service
go run services/designation-service/cmd/server/main.go
```

**Expected Output:**
```
🚀 Starting designation-service on port 50055
📊 Connecting to database...
✅ Connected to database
🎯 Running migrations...
✅ Designation Service listening on port 50055
```

**Keep this terminal running!**

---

### **Terminal 3: Start API Gateway**

```powershell
# Navigate to project
cd "d:\Nhit\NHIT Backend"

# Run API Gateway
go run services/api-gateway/cmd/server/main.go
```

**Expected Output:**
```
✅ Registered User Service gateway -> localhost:50051
✅ Registered Auth Service gateway -> localhost:50052
✅ Registered Department Service gateway -> localhost:50054
✅ Registered Designation Service gateway -> localhost:50055
🚀 API Gateway listening on :8080
📖 REST API available at http://localhost:8080/api/v1/
📝 Examples:
   - Users: curl http://localhost:8080/api/v1/users
   - Departments: curl http://localhost:8080/api/v1/departments
   - Designations: curl http://localhost:8080/api/v1/designations
```

**Note:** You'll see errors for User, Auth, and Department services (they're not running). **That's OK!** The Designation Service will work fine.

**Keep this terminal running!**

---

### **Terminal 4: Test the API**

Open a **4th terminal** to test:

```powershell
cd "d:\Nhit\NHIT Backend"

# Test 1: Create a designation
$body = @{
    name = "Senior Software Engineer"
    description = "Senior level software engineering position"
    is_active = $true
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/designations" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"

# Test 2: List all designations
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/designations" -Method GET

# Test 3: Create child designation (replace {parent-id} with ID from Test 1)
$childBody = @{
    name = "Software Engineer"
    description = "Entry level position"
    is_active = $true
    parent_id = "paste-parent-id-here"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/designations" `
    -Method POST `
    -Body $childBody `
    -ContentType "application/json"

# Test 4: Get hierarchy (replace {id} with child ID)
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/designations/{id}/hierarchy" -Method GET

# Test 5: Search
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/designations?search=engineer" -Method GET

# Test 6: Filter active only
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/designations?active_only=true" -Method GET
```

---

## 🔍 **Troubleshooting**

### **Problem: "Failed to register user service gateway"**

**This is expected!** User, Auth, and Department services are not running. The API Gateway will show errors for those, but **Designation Service will still work**.

The gateway will continue and register the Designation Service successfully.

### **Problem: "Database connection failed"**

Check your PostgreSQL credentials in Terminal 2:
```powershell
$env:DATABASE_URL = "postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
```

Make sure:
- Username: `postgres`
- Password: `shristi`
- Database: `nhit` exists

### **Problem: "Port already in use"**

Check what's using the port:
```powershell
# Check port 50055 (Designation Service)
netstat -ano | findstr :50055

# Check port 8080 (API Gateway)
netstat -ano | findstr :8080
```

Kill the process or change the port.

### **Problem: "go: no required module provides package"**

Run this in the root directory:
```powershell
cd "d:\Nhit\NHIT Backend"
go mod tidy
```

---

## 📊 **Architecture**

```
Browser/Postman
      │
      │ HTTP REST (Port 8080)
      ▼
┌─────────────────────┐
│   API Gateway       │
│   (Port 8080)       │
└──────────┬──────────┘
           │
           │ gRPC (Port 50055)
           ▼
┌─────────────────────┐
│ Designation Service │
│   (Port 50055)      │
└──────────┬──────────┘
           │
           │ SQL
           ▼
┌─────────────────────┐
│   PostgreSQL        │
│   (Port 5432)       │
└─────────────────────┘
```

---

## ✅ **Success Indicators**

You'll know it's working when:
- ✅ Terminal 2 shows "Designation Service listening on port 50055"
- ✅ Terminal 3 shows "Registered Designation Service gateway"
- ✅ Terminal 3 shows "API Gateway listening on :8080"
- ✅ Test commands return JSON responses
- ✅ No errors in Terminal 2 or 3 (except for other services)

---

## 🎯 **Complete Test Flow**

```powershell
# Terminal 2: Start Designation Service
cd "d:\Nhit\NHIT Backend"
$env:DATABASE_URL = "postgres://postgres:shristi@localhost:5432/nhit?sslmode=disable"
$env:PORT = "50055"
go run services/designation-service/cmd/server/main.go

# Terminal 3: Start API Gateway (wait for Terminal 2 to be ready)
cd "d:\Nhit\NHIT Backend"
go run services/api-gateway/cmd/server/main.go

# Terminal 4: Test
cd "d:\Nhit\NHIT Backend"
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/designations" -Method GET
```

---

## 🚀 **You're Ready!**

1. ✅ PostgreSQL is running
2. ✅ Designation Service code is ready
3. ✅ API Gateway is configured
4. ✅ Just run the commands above!

**Start with Terminal 2, then Terminal 3, then test in Terminal 4!** 🎉
