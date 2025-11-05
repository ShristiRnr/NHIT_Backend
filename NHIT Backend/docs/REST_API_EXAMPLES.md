# REST API Examples

## 🎯 API Gateway Endpoint
**Base URL:** `http://localhost:8080`

All REST requests go through the API Gateway, which converts them to gRPC calls to the microservices.

---

## 👤 User Management APIs

### 1. Create User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

**Response:**
```json
{
  "user_id": "987fcdeb-51a2-43f7-9abc-123456789def",
  "name": "John Doe",
  "email": "john@example.com",
  "roles": [],
  "permissions": []
}
```

### 2. Get User
```bash
curl http://localhost:8080/api/v1/users/987fcdeb-51a2-43f7-9abc-123456789def
```

### 3. List Users
```bash
curl "http://localhost:8080/api/v1/users?tenant_id=123e4567-e89b-12d3-a456-426614174000"
```

### 4. Update User
```bash
curl -X PUT http://localhost:8080/api/v1/users/987fcdeb-51a2-43f7-9abc-123456789def \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated",
    "email": "john.updated@example.com",
    "password": "newpassword123"
  }'
```

### 5. Delete User
```bash
curl -X DELETE http://localhost:8080/api/v1/users/987fcdeb-51a2-43f7-9abc-123456789def
```

### 6. Assign Roles to User
```bash
curl -X POST http://localhost:8080/api/v1/users/987fcdeb-51a2-43f7-9abc-123456789def/roles \
  -H "Content-Type: application/json" \
  -d '{
    "roles": [
      "role-uuid-1",
      "role-uuid-2"
    ]
  }'
```

### 7. List User Roles
```bash
curl http://localhost:8080/api/v1/users/987fcdeb-51a2-43f7-9abc-123456789def/roles
```

---

## 🔐 Authentication APIs

### 1. Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Jane Smith",
    "email": "jane@example.com",
    "password": "securepass456",
    "roles": ["ADMIN"]
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login": "john@example.com",
    "password": "securepass123",
    "tenant_id": "123e4567-e89b-12d3-a456-426614174000"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_here",
  "user_id": "987fcdeb-51a2-43f7-9abc-123456789def",
  "email": "john@example.com",
  "name": "John Doe",
  "roles": ["ADMIN"],
  "permissions": ["manage-users", "view-reports"],
  "tenant_id": "123e4567-e89b-12d3-a456-426614174000",
  "token_expires_at": 1699564800,
  "refresh_expires_at": 1700169600
}
```

### 3. Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": "987fcdeb-51a2-43f7-9abc-123456789def",
    "refresh_token": "refresh_token_here"
  }'
```

### 4. Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "refresh_token_here",
    "tenant_id": "123e4567-e89b-12d3-a456-426614174000"
  }'
```

### 5. Forgot Password
```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com"
  }'
```

### 6. Reset Password
```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "reset_token_from_email",
    "new_password": "newSecurePass789"
  }'
```

### 7. Verify Email
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "987fcdeb-51a2-43f7-9abc-123456789def",
    "verification_token": "token_from_email"
  }'
```

---

## 🧪 Testing with Postman

### Import Collection
1. Open Postman
2. Click "Import"
3. Create requests with the examples above
4. Set base URL as environment variable: `{{baseUrl}} = http://localhost:8080`

### Example Postman Request
```
Method: POST
URL: {{baseUrl}}/api/v1/auth/login
Headers:
  Content-Type: application/json
Body (raw JSON):
{
  "login": "john@example.com",
  "password": "securepass123",
  "tenant_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

## 🔍 Testing with Browser

You can test GET requests directly in your browser:

```
http://localhost:8080/api/v1/users?tenant_id=123e4567-e89b-12d3-a456-426614174000
```

---

## 📊 Response Formats

### Success Response
```json
{
  "user_id": "...",
  "name": "...",
  "email": "..."
}
```

### Error Response
```json
{
  "code": 3,
  "message": "invalid user_id: invalid UUID length: 5",
  "details": []
}
```

### gRPC Error Codes
- `0` - OK
- `3` - INVALID_ARGUMENT
- `5` - NOT_FOUND
- `13` - INTERNAL
- `16` - UNAUTHENTICATED

---

## 🎨 Benefits

✅ **Standard REST APIs** - No need to learn gRPC for clients  
✅ **JSON Format** - Easy to read and debug  
✅ **Browser Compatible** - Works in any HTTP client  
✅ **Automatic Conversion** - Gateway handles gRPC translation  
✅ **CORS Enabled** - Works with web applications  
✅ **Type Safe** - Validated by protobuf definitions  

---

## 🚀 Quick Start

1. **Start Services:**
```bash
# Terminal 1 - User Service
cd services/user-service
go run cmd/server/main.go

# Terminal 2 - Auth Service  
cd services/auth-service
go run cmd/server/main.go

# Terminal 3 - API Gateway
cd services/api-gateway
go run cmd/server/main.go
```

2. **Test API:**
```bash
curl http://localhost:8080/api/v1/users
```

3. **Check Gateway Logs:**
You'll see requests being proxied to gRPC services!

---

## 📝 Notes

- All requests go through port **8080** (API Gateway)
- Gateway forwards to gRPC services (50051, 50052, 50053)
- Responses are automatically converted from Protobuf to JSON
- CORS is enabled for browser requests
- Authentication tokens should be passed in `Authorization` header

Happy API testing! 🎉
