# Dynamic Roles Implementation

## Overview
The role management system has been updated to support **fully dynamic roles**. Only the **Super Admin** role is predefined in the system. All other roles must be created by Super Admin as needed.

## Key Changes

### 1. **Single Default Role**
- **Only Super Admin** is created by default
- All other roles (Admin, Manager, User, Approver, Reviewer, etc.) are removed from seed data
- Super Admin has full access to all permissions

### 2. **Dynamic Role Creation**
Super Admin can create any role dynamically based on organizational needs:
- Custom role names
- Custom permission assignments
- Flexible role structures

## Benefits

### ✅ **Maximum Flexibility**
- Organizations can define their own role structures
- No predefined roles that may not fit the organization
- Easy to adapt to different business models

### ✅ **Simplified Onboarding**
- Start with just Super Admin
- Create roles as the organization grows
- No unused default roles cluttering the system

### ✅ **Better Security**
- Only necessary roles exist in the system
- Reduced attack surface
- Clear role ownership and management

## Implementation Details

### Database Seed Data
```sql
-- Only Super Admin role is created
INSERT INTO roles (role_id, name, tenant_id, created_at, updated_at)
VALUES 
('11111111-1111-1111-1111-111111111111', 'Super Admin', '11111111-1111-1111-1111-111111111111', NOW(), NOW())
ON CONFLICT (role_id) DO NOTHING;
```

### Service Method
```go
// CreateSuperAdminRole creates Super Admin role if it doesn't exist
func (s *RoleService) CreateSuperAdminRole(ctx context.Context, tenantID uuid.UUID) error {
    exists, err := s.repo.RoleExists(ctx, tenantID, "Super Admin")
    if err != nil {
        return err
    }

    if !exists {
        params := db.CreateRoleParams{
            TenantID: tenantID,
            Name:     "Super Admin",
        }
        _, err := s.repo.Create(ctx, params)
        if err != nil {
            log.Printf("[RoleService] Error creating Super Admin role: %v\n", err)
            return err
        }
        log.Printf("[RoleService] Created Super Admin role for tenant: %s\n", tenantID)
    }

    return nil
}
```

## How Super Admin Creates Roles

### Via gRPC API
```protobuf
rpc CreateRole(CreateRoleRequest) returns (RoleResponse);

message CreateRoleRequest {
  string tenant_id = 1;
  string name = 2;
  repeated string permissions = 3;
}
```

### Via HTTP API
```http
POST /roles
Content-Type: application/json

{
  "tenant_id": "11111111-1111-1111-1111-111111111111",
  "name": "Project Manager",
  "permissions": [
    "55555555-5555-5555-5555-555555555555",
    "66666666-6666-6666-6666-666666666666"
  ]
}
```

### Example: Creating Common Roles

#### 1. Create "Manager" Role
```go
role, err := roleService.CreateRole(ctx, tenantID, "Manager")
if err != nil {
    return err
}

// Assign permissions
permissionIDs := []uuid.UUID{
    uuid.MustParse("55555555-5555-5555-5555-555555555555"), // create-user
    uuid.MustParse("66666666-6666-6666-6666-666666666666"), // edit-user
    uuid.MustParse("88888888-8888-8888-8888-888888888888"), // view-user
    uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), // approve-requests
}

err = roleService.SyncRolePermissions(ctx, role.RoleID, permissionIDs)
```

#### 2. Create "Viewer" Role
```go
role, err := roleService.CreateRole(ctx, tenantID, "Viewer")
if err != nil {
    return err
}

// Assign read-only permissions
permissionIDs := []uuid.UUID{
    uuid.MustParse("44444444-4444-4444-4444-444444444444"), // view-role
    uuid.MustParse("88888888-8888-8888-8888-888888888888"), // view-user
    uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), // view-reports
}

err = roleService.SyncRolePermissions(ctx, role.RoleID, permissionIDs)
```

#### 3. Create "HR Manager" Role
```go
role, err := roleService.CreateRole(ctx, tenantID, "HR Manager")
if err != nil {
    return err
}

// Assign HR-specific permissions
permissionIDs := []uuid.UUID{
    uuid.MustParse("55555555-5555-5555-5555-555555555555"), // create-user
    uuid.MustParse("66666666-6666-6666-6666-666666666666"), // edit-user
    uuid.MustParse("77777777-7777-7777-7777-777777777777"), // delete-user
    uuid.MustParse("88888888-8888-8888-8888-888888888888"), // view-user
    uuid.MustParse("99999999-9999-9999-9999-999999999999"), // manage-users
}

err = roleService.SyncRolePermissions(ctx, role.RoleID, permissionIDs)
```

## Workflow

### Initial Setup
1. System creates tenant
2. System creates Super Admin role for tenant
3. System creates default permissions
4. First user is assigned Super Admin role

### Adding New Roles
1. Super Admin logs in
2. Super Admin creates new role (e.g., "Department Head")
3. Super Admin assigns permissions to the role
4. Super Admin assigns the role to users

### Managing Roles
1. Super Admin can update role names
2. Super Admin can sync/modify role permissions
3. Super Admin can delete roles (except Super Admin itself)
4. Super Admin can view all roles and their permissions

## Protected Operations

### Super Admin Role Protection
```go
// Prevent deleting "Super Admin" role
if role.Name == "Super Admin" || role.Name == "super_admin" {
    return errors.New("SUPER ADMIN ROLE CAN NOT BE DELETED")
}
```

### Self-Role Deletion Protection
```go
// Prevent deleting self-assigned role
if auth.IsCurrentUserInRole(role.Name, ctx) {
    return errors.New("CAN NOT DELETE SELF ASSIGNED ROLE")
}
```

## Use Cases

### Startup Company
- Start with Super Admin
- Add "Developer" and "Designer" roles as team grows
- Create "Team Lead" role when needed

### Enterprise Organization
- Super Admin for IT department
- Create department-specific roles:
  - "Finance Manager"
  - "HR Manager"
  - "Sales Manager"
  - "Operations Manager"
- Create project-specific roles as needed

### Educational Institution
- Super Admin for system administrators
- Create roles like:
  - "Faculty"
  - "Student"
  - "Department Head"
  - "Registrar"

### Healthcare Organization
- Super Admin for IT staff
- Create roles like:
  - "Doctor"
  - "Nurse"
  - "Administrator"
  - "Lab Technician"

## Migration from Old System

If you had predefined roles (Admin, Manager, User, etc.):

1. **Backup existing roles and assignments**
2. **Run new migration** (only Super Admin will be created)
3. **Recreate needed roles** via Super Admin
4. **Reassign users** to new roles
5. **Verify permissions** are correctly assigned

## Best Practices

### 1. **Role Naming Convention**
- Use clear, descriptive names
- Follow organizational terminology
- Be consistent across tenants

### 2. **Permission Assignment**
- Follow principle of least privilege
- Group related permissions together
- Document role purposes

### 3. **Role Management**
- Regularly audit roles and permissions
- Remove unused roles
- Update permissions as requirements change

### 4. **User Assignment**
- Assign minimum necessary roles
- Review user roles periodically
- Log role changes for audit

## API Reference

### Create Role
```http
POST /roles
Authorization: Bearer <super_admin_token>
Content-Type: application/json

{
  "tenant_id": "uuid",
  "name": "Role Name",
  "permissions": ["perm_id_1", "perm_id_2"]
}
```

### List Roles
```http
GET /roles?tenant_id=uuid
Authorization: Bearer <token>
```

### Update Role
```http
PUT /roles/{role_id}
Authorization: Bearer <super_admin_token>
Content-Type: application/json

{
  "name": "Updated Role Name",
  "permissions": ["perm_id_1", "perm_id_2", "perm_id_3"]
}
```

### Delete Role
```http
DELETE /roles/{role_id}
Authorization: Bearer <super_admin_token>
```

### Sync Role Permissions
```http
POST /roles/{role_id}/sync-permissions
Authorization: Bearer <super_admin_token>
Content-Type: application/json

{
  "permission_ids": ["perm_id_1", "perm_id_2"]
}
```

## Summary

The system now provides **complete flexibility** in role management:
- ✅ Only Super Admin is predefined
- ✅ All other roles are created dynamically
- ✅ Super Admin has full control over role creation
- ✅ Roles can be customized per organizational needs
- ✅ No unused default roles
- ✅ Better security and cleaner system

This approach allows organizations to build their role structure exactly as they need it, without being constrained by predefined roles.
