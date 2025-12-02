1) Create Tenant
POST: http://localhost:8083/api/v1/users/tenants

Request:
{
    "name": "SuperAdmin789",
    "email": "superadmin789@gmail.com",
    "password": "AdminPassword789!"
}

Response:
{
    "tenantId": "9e1e8cf8-65e4-4ba4-ab04-16b642a3c9a0",
    "name": "SuperAdmin789",
    "email": "superadmin789@gmail.com",
    "password": ""
}

2) Create Orgaizations
POST: http://localhost:8083/api/v1/organizations

{
    "tenant_id": "f4eebf0e-733e-4d17-bdca-165c8e5bb90f",
    "name": "NHIT",
    "code": "NHIT123456",
    "description": "Main Organization",
    "super_admin": {
      "name": "SuperAdmin123",
      "email": "superadmin123@example.com",
      "password": "AdminPassword123!"
    },
    "initial_projects": ["Project Alpha", "Project Beta"]
}

Response:
{
    "organization": {
        "orgId": "611015ab-bb4a-4ce3-b6f0-25ac48469a0d",
        "tenantId": "f4eebf0e-733e-4d17-bdca-165c8e5bb90f",
        "name": "NHIT",
        "code": "NHIT123456",
        "databaseName": "org_nhit123456",
        "description": "Main Organization",
        "logo": "",
        "isActive": true,
        "createdBy": "00000000-0000-0000-0000-000000000001",
        "createdAt": "2025-11-12T23:31:42.740652Z",
        "updatedAt": "2025-11-12T23:31:42.740652Z"
    },
    "message": "Organization created successfully"
}

3) Login:
POST: 