# 📮 Postman Testing Guide - NHIT Backend

**Step-by-step guide to test all microservices using Postman**

---

## 🎯 Prerequisites

1. **Postman installed** - Download from https://www.postman.com/downloads/
2. **All services running** - Run `.\test_complete_api.ps1` or start manually
3. **Database ready** - PostgreSQL running with migrations applied

---

## 📋 Step-by-Step Setup

### **Step 1: Open Postman**
1. Launch Postman application
2. Click on **"Workspaces"** in top left
3. Select **"My Workspace"** or create new workspace

### **Step 2: Create New Collection**
1. Click **"Collections"** in left sidebar
2. Click **"+"** or **"Create Collection"**
3. Name it: **"NHIT Backend API"**
4. Click **"Create"**

### **Step 3: Set Environment Variables**
1. Click **"Environments"** in left sidebar (eye icon)
2. Click **"+"** to create new environment
3. Name it: **"NHIT Local"**
4. Add these variables:

| Variable | Initial Value | Current Value |
|----------|---------------|---------------|
| `base_url` | `http://localhost:8081/api/v1` | `http://localhost:8081/api/v1` |
| `tenant_id` | `00000000-0000-0000-0000-000000000001` | `00000000-0000-0000-0000-000000000001` |
| `department_id` | (leave empty) | (leave empty) |
| `designation_id` | (leave empty) | (leave empty) |
| `user_id` | (leave empty) | (leave empty) |
| `session_id` | (leave empty) | (leave empty) |
| `session_token` | (leave empty) | (leave empty) |
| `org_id` | (leave empty) | (leave empty) |
| `vendor_id` | (leave empty) | (leave empty) |

5. Click **"Save"**
6. Select **"NHIT Local"** from environment dropdown (top right)

---

## 🧪 Phase 1: Setup & Registration

### **Test 1: Create Department**

1. **Right-click** on "NHIT Backend API" collection → **"Add Request"**
2. **Name:** `Create Department`
3. **Method:** `POST`
4. **URL:** `{{base_url}}/departments`
5. **Headers Tab:**
   - Key: `Content-Type`, Value: `application/json`
6. **Body Tab:**
   - Select **"raw"**
   - Select **"JSON"** from dropdown
   - Paste:
```json
{
  "name": "Engineering Department",
  "description": "Software Engineering and Development"
}
```
7. Click **"Send"**
8. **Expected Response:** `201 Created`
9. **Copy the `id` from response**
10. Go to **Environments** → **NHIT Local**
11. Set `department_id` = (paste the copied id)
12. Click **"Save"**

---

### **Test 2: Create Designation**

1. **Add new request** to collection
2. **Name:** `Create Designation`
3. **Method:** `POST`
4. **URL:** `{{base_url}}/designations`
5. **Headers:** `Content-Type: application/json`
6. **Body (JSON):**
```json
{
  "name": "Senior Software Engineer",
  "description": "Senior level software engineer position",
  "slug": "senior-software-engineer",
  "is_active": true,
  "level": 3
}
```
7. Click **"Send"**
8. **Expected:** `201 Created`
9. **Copy the `id`** from response
10. Set environment variable `designation_id` = (copied id)

---

### **Test 3: Register User**

1. **Add new request**
2. **Name:** `Register User`
3. **Method:** `POST`
4. **URL:** `{{base_url}}/users`
5. **Headers:** `Content-Type: application/json`
6. **Body (JSON):**
```json
{
  "tenant_id": "{{tenant_id}}",
  "name": "John Doe",
  "email": "john.doe@nhit.com",
  "password": "SecurePassword123!",
  "department_id": "{{department_id}}",
  "designation_id": "{{designation_id}}"
}
```
7. Click **"Send"**
8. **Expected:** `201 Created`
9. **Copy `user_id`** from response
10. Set environment variable `user_id` = (copied user_id)

---

### **Test 4: Login (Create Session)**

1. **Add new request**
2. **Name:** `Login - Create Session`
3. **Method:** `POST`
4. **URL:** `{{base_url}}/auth/sessions`
5. **Headers:** `Content-Type: application/json`
6. **Body (JSON):**
```json
{
  "user_id": "{{user_id}}",
  "session_token": "test_session_token_12345",
  "expires_at": "2025-11-08T10:30:00Z"
}
```
7. Click **"Send"**
8. **Expected:** `201 Created`
9. **Copy `session_id` and `session_token`**
10. Set environment variables:
    - `session_id` = (copied session_id)
    - `session_token` = (copied session_token)

---

## 📊 Phase 2: Department Service

### **Test 5: List Departments**

1. **Add request:** `List Departments`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/departments?page=1&page_size=10`
4. Click **"Send"**
5. **Expected:** `200 OK` with array of departments

### **Test 6: Get Department by ID**

1. **Add request:** `Get Department by ID`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/departments/{{department_id}}`
4. Click **"Send"**
5. **Expected:** `200 OK` with department details

### **Test 7: Update Department**

1. **Add request:** `Update Department`
2. **Method:** `PUT`
3. **URL:** `{{base_url}}/departments/{{department_id}}`
4. **Headers:** `Content-Type: application/json`
5. **Body (JSON):**
```json
{
  "name": "Engineering Department - Updated",
  "description": "Updated: Software Engineering Division"
}
```
6. Click **"Send"**
7. **Expected:** `200 OK`

---

## 🎯 Phase 3: Designation Service

### **Test 8: List Designations**

1. **Add request:** `List Designations`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/designations?page=1&page_size=10`
4. Click **"Send"**

### **Test 9: Get Designation by ID**

1. **Add request:** `Get Designation by ID`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/designations/{{designation_id}}`
4. Click **"Send"**

### **Test 10: Get Designation by Slug**

1. **Add request:** `Get Designation by Slug`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/designations/slug/senior-software-engineer`
4. Click **"Send"**

### **Test 11: Get Root Designations**

1. **Add request:** `Get Root Designations`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/designations/root`
4. Click **"Send"**

---

## 👥 Phase 4: User Service

### **Test 12: Get User by ID**

1. **Add request:** `Get User by ID`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/users/{{user_id}}`
4. Click **"Send"**

### **Test 13: Get User by Email**

1. **Add request:** `Get User by Email`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/users/email/john.doe@nhit.com`
4. Click **"Send"**

### **Test 14: List Users by Tenant**

1. **Add request:** `List Users by Tenant`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/users/tenant/{{tenant_id}}?page=1&page_size=10`
4. Click **"Send"**

### **Test 15: Search Users**

1. **Add request:** `Search Users`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/users/search?name=John&page=1&page_size=10`
4. Click **"Send"**

---

## 🔐 Phase 5: Auth Service

### **Test 16: Get Session by ID**

1. **Add request:** `Get Session by ID`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/auth/sessions/{{session_id}}`
4. Click **"Send"**

### **Test 17: Get User Sessions**

1. **Add request:** `Get User Sessions`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/auth/sessions/user/{{user_id}}`
4. Click **"Send"**

### **Test 18: Create Refresh Token**

1. **Add request:** `Create Refresh Token`
2. **Method:** `POST`
3. **URL:** `{{base_url}}/auth/refresh-tokens`
4. **Headers:** `Content-Type: application/json`
5. **Body (JSON):**
```json
{
  "user_id": "{{user_id}}",
  "token": "refresh_token_12345",
  "expires_at": "2025-12-07T10:30:00Z"
}
```
6. Click **"Send"**

---

## 🏢 Phase 6: Organization Service

### **Test 19: Create Organization**

1. **Add request:** `Create Organization`
2. **Method:** `POST`
3. **URL:** `{{base_url}}/organizations`
4. **Headers:** `Content-Type: application/json`
5. **Body (JSON):**
```json
{
  "tenant_id": "{{tenant_id}}",
  "name": "NHIT Corporation",
  "code": "NHIT001",
  "is_active": true
}
```
6. Click **"Send"**
7. **Copy `org_id`** and save to environment

### **Test 20: List Organizations**

1. **Add request:** `List Organizations`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/organizations/tenant/{{tenant_id}}?page=1&page_size=10`
4. Click **"Send"**

### **Test 21: Get Organization by ID**

1. **Add request:** `Get Organization by ID`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/organizations/{{org_id}}`
4. Click **"Send"**

---

## 🏪 Phase 7: Vendor Service

### **Test 22: Create Vendor**

1. **Add request:** `Create Vendor`
2. **Method:** `POST`
3. **URL:** `{{base_url}}/vendors`
4. **Headers:** `Content-Type: application/json`
5. **Body (JSON):**
```json
{
  "name": "Tech Supplies Inc",
  "email": "contact@techsupplies.com",
  "phone": "+1234567890",
  "address": "123 Tech Street, Silicon Valley, CA",
  "is_active": true
}
```
6. Click **"Send"**
7. **Copy `vendor_id`** and save to environment

### **Test 23: List Vendors**

1. **Add request:** `List Vendors`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/vendors?page=1&page_size=10`
4. Click **"Send"**

### **Test 24: Get Vendor by ID**

1. **Add request:** `Get Vendor by ID`
2. **Method:** `GET`
3. **URL:** `{{base_url}}/vendors/{{vendor_id}}`
4. Click **"Send"**

---

## 🧹 Phase 8: Cleanup

### **Test 25: Logout (Delete Session)**

1. **Add request:** `Logout - Delete Session`
2. **Method:** `DELETE`
3. **URL:** `{{base_url}}/auth/sessions/{{session_id}}`
4. Click **"Send"**
5. **Expected:** `204 No Content`

---

## 💡 Pro Tips

### **Organize Your Collection**

1. Create **folders** for each phase:
   - Right-click collection → **"Add Folder"**
   - Name: "Phase 1 - Setup"
   - Drag requests into folders

### **Use Tests Tab for Auto-Save**

In each request, go to **"Tests"** tab and add:

```javascript
// Auto-save department_id
if (pm.response.code === 201 && pm.response.json().id) {
    pm.environment.set("department_id", pm.response.json().id);
}
```

### **Create Collection Runner**

1. Click collection → **"Run"**
2. Select all requests
3. Click **"Run NHIT Backend API"**
4. Watch all tests execute automatically

### **Export Collection**

1. Right-click collection → **"Export"**
2. Choose **"Collection v2.1"**
3. Save as `NHIT_Backend_API.postman_collection.json`
4. Share with team

---

## ✅ Quick Checklist

- [ ] Postman installed
- [ ] Environment created with variables
- [ ] Collection created
- [ ] All services running
- [ ] Phase 1 completed (4 tests)
- [ ] Phase 2 completed (4 tests)
- [ ] Phase 3 completed (6 tests)
- [ ] Phase 4 completed (7 tests)
- [ ] Phase 5 completed (6 tests)
- [ ] Phase 6 completed (8 tests)
- [ ] Phase 7 completed (6 tests)
- [ ] Phase 8 completed (2 tests)

**Total: 43 tests** ✅

---

## 🎯 Summary

You've learned how to:
- ✅ Set up Postman environment
- ✅ Create and organize collections
- ✅ Use environment variables
- ✅ Test all 43 API endpoints
- ✅ Follow the complete testing flow

**Happy Testing!** 🎉
