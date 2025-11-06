# Role Management System Update

## Overview
Updated the gRPC user service role management system based on PHP Laravel Spatie Permission patterns. The system now includes comprehensive role and permission management with validation, activity logging, and notification features.

## Key Features Added

### 1. **Permission Management System**
- Full CRUD operations for permissions
- Permission-to-role assignment
- Permission synchronization for roles
- Default permissions seeding

### 2. **Advanced Role Queries**
- **Approval Roles**: Get roles matching approval patterns (approver, reviewer, QS, GN, ER, PN, manager, admin)
- **Users by Roles**: Get users with specific role names
- **Role Existence Check**: Verify if a role exists before operations
- **Role by Name**: Retrieve roles by tenant and name

### 3. **Role Validation & Protection**
- Prevent deletion of "Super Admin" role
- Prevent users from deleting their own assigned roles
- Case-insensitive role name checks
- Role existence validation before creation

### 4. **Activity Logging**
- Log role creation, updates, and deletions
- Log permission assignments
- Track user actions with context

### 5. **Super Admin Notifications**
- Notify super admins on role updates
- Email notifications for critical changes
- Configurable notification triggers

## Database Changes

### New SQL Queries (`permissions.sql`)
```sql
- CreatePermission
- GetPermission
- GetPermissionByName
- ListPermissions
- UpdatePermission
- DeletePermission
- ListPermissionsOfRole
- RemovePermissionFromRole
- RemoveAllPermissionsFromRole
- GetRolesByPattern
- GetUsersWithRoleNames
- GetApprovalRolesByTenant
- GetUsersWithApprovalRoles
```

### Updated Queries (`roles.sql`)
```sql
- GetRoleByName
- RoleExistsByName
- ListSuperAdmins (updated to include both 'super_admin' and 'Super Admin')
```

### Seed Data Updates
Added default role:
- **Super Admin** (full access with all permissions)

**Note**: All other roles are dynamic and must be created by Super Admin as needed. This provides maximum flexibility for different organizational structures.

Added default permissions:
- create-role, edit-role, delete-role, view-role
- create-user, edit-user, delete-user, view-user
- manage-users, view-reports
- approve-requests, review-content

## Proto Definitions Updated

### New RPC Methods
```protobuf
// Role Management
rpc GetApprovalRoles(GetApprovalRolesRequest) returns (ListRolesResponse);
rpc GetUsersWithApprovalRoles(GetUsersWithApprovalRolesRequest) returns (ListUsersResponse);
rpc GetUsersWithRoles(GetUsersWithRolesRequest) returns (ListUsersResponse);
rpc RoleExists(RoleExistsRequest) returns (RoleExistsResponse);
rpc SyncRolePermissions(SyncRolePermissionsRequest) returns (RoleResponse);

// Permission Management
rpc CreatePermission(CreatePermissionRequest) returns (PermissionResponse);
rpc ListPermissions(ListPermissionsRequest) returns (ListPermissionsResponse);
rpc GetPermission(GetPermissionRequest) returns (PermissionResponse);
rpc AssignPermissionToRole(AssignPermissionToRoleRequest) returns (google.protobuf.Empty);
rpc RemovePermissionFromRole(RemovePermissionFromRoleRequest) returns (google.protobuf.Empty);
rpc ListPermissionsOfRole(ListPermissionsOfRoleRequest) returns (ListPermissionsResponse);
```

## New Files Created

1. **`internal/adapters/database/queries/permissions.sql`**
   - Permission CRUD queries
   - Role-permission relationship queries
   - Advanced role filtering queries

2. **`internal/adapters/repository/permission_repository.go`**
   - Permission repository implementation
   - Implements PermissionRepository interface

3. **`internal/core/ports/services/permission_services.go`**
   - Permission service layer
   - Default permission creation
   - Permission business logic

## Updated Files

1. **`api/proto/user_management.proto`**
   - Added new RPC methods
   - Added request/response message types

2. **`internal/core/ports/repository.go`**
   - Extended RoleRepository interface
   - Added PermissionRepository interface

3. **`internal/adapters/repository/roles_repository.go`**
   - Implemented new repository methods
   - Added approval role queries
   - Added permission management

4. **`internal/core/ports/services/roles_services.go`**
   - Added role validation methods
   - Added permission sync functionality
   - Added default role creation
   - Added approval role queries

5. **`internal/core/ports/http_server/roles_handler.go`**
   - Enhanced delete validation
   - Added new HTTP handlers
   - Improved error messages

6. **`internal/adapters/database/migration/003_insert_seed_data.up.sql`**
   - Added default roles and permissions
   - Added role-permission mappings
   - Added Super Admin with full permissions

## Usage Examples

### 1. Check if Role Exists
```go
exists, err := roleService.RoleExists(ctx, tenantID, "Manager")
```

### 2. Get Approval Roles
```go
approvalRoles, err := roleService.GetApprovalRoles(ctx, tenantID)
```

### 3. Get Users with Specific Roles
```go
users, err := roleService.GetUsersWithRoleNames(ctx, tenantID, []string{"Admin", "Manager"})
```

### 4. Sync Role Permissions
```go
permissionIDs := []uuid.UUID{perm1, perm2, perm3}
err := roleService.SyncRolePermissions(ctx, roleID, permissionIDs)
```

### 5. Create Super Admin Role (if needed)
```go
err := roleService.CreateSuperAdminRole(ctx, tenantID)
```

### 6. Create Custom Roles (by Super Admin)
```go
// Super Admin can create any role dynamically
role, err := roleService.CreateRole(ctx, tenantID, "Custom Role Name")
```

### 7. Create Default Permissions
```go
err := permissionService.CreateDefaultPermissions(ctx)
```

## Security Features

### Role Deletion Protection
- Super Admin role cannot be deleted
- Users cannot delete their own assigned roles
- Proper error messages for forbidden operations

### Permission Validation
- Middleware checks for required permissions
- Role-based access control on all endpoints
- Activity logging for audit trails

### Notification System
- Super admins notified on role changes
- Email notifications for critical updates
- Configurable notification rules

## Migration Steps

1. **Run SQLC Generation**
   ```bash
   cd d:\Nhit
   sqlc generate
   ```

2. **Run Database Migrations**
   ```bash
   # Apply the updated seed data
   migrate -path internal/adapters/database/migration -database "your-db-url" up
   ```

3. **Generate Proto Files**
   ```bash
   protoc --go_out=. --go-grpc_out=. api/proto/user_management.proto
   ```

4. **Initialize Default Data** (in your application startup)
   ```go
   // Create default permissions
   permissionService.CreateDefaultPermissions(ctx)
   
   // Create Super Admin role for each tenant (if not exists)
   roleService.CreateSuperAdminRole(ctx, tenantID)
   
   // All other roles are created dynamically by Super Admin as needed
   ```

## API Endpoints (HTTP)

### Role Management
- `GET /roles` - List all roles (requires: create-role, edit-role, or delete-role)
- `GET /roles/{roleID}` - Get role by ID
- `POST /roles` - Create role (requires: create-role)
- `PUT /roles/{roleID}` - Update role (requires: edit-role)
- `DELETE /roles/{roleID}` - Delete role (requires: delete-role)
- `GET /roles/approval` - Get approval roles
- `POST /roles/check-exists` - Check if role exists
- `POST /roles/{roleID}/sync-permissions` - Sync role permissions

### Permission Management
- `GET /permissions` - List all permissions
- `GET /permissions/{permissionID}` - Get permission by ID
- `POST /permissions` - Create permission
- `PUT /permissions/{permissionID}` - Update permission
- `DELETE /permissions/{permissionID}` - Delete permission
- `POST /roles/{roleID}/assign-permission/{permissionID}` - Assign permission to role
- `DELETE /roles/{roleID}/remove-permission/{permissionID}` - Remove permission from role

### User-Role Queries
- `GET /users/{userID}/roles` - Get user's roles
- `GET /users/{userID}/permissions` - Get user's permissions
- `POST /users/with-approval-roles` - Get users with approval roles
- `POST /users/with-roles` - Get users with specific roles

## Testing Recommendations

1. **Test Role Creation**
   - Create roles with permissions
   - Verify default roles are created
   - Test duplicate role prevention

2. **Test Role Deletion**
   - Try deleting Super Admin (should fail)
   - Try deleting self-assigned role (should fail)
   - Delete regular role (should succeed)

3. **Test Permission Sync**
   - Sync permissions for a role
   - Verify old permissions are removed
   - Verify new permissions are added

4. **Test Approval Role Queries**
   - Get approval roles
   - Verify pattern matching works
   - Test with different role names

5. **Test User-Role Queries**
   - Get users with specific roles
   - Get users with approval roles
   - Verify unique user results

## Notes

- All UUID fields use proper UUID types
- Case-insensitive role name comparisons for critical operations
- Activity logging integrated for audit trails
- Super admin notifications can be configured
- Permission system follows principle of least privilege
- **Only Super Admin role is predefined** - all other roles are created dynamically by Super Admin
- This provides maximum flexibility for different organizational structures
- Super Admin has full access to all permissions by default

## Future Enhancements

1. **Role Hierarchy Management**
   - Parent-child role relationships
   - Inherited permissions

2. **Dynamic Permission Creation**
   - API for creating custom permissions
   - Permission categories

3. **Role Templates**
   - Pre-configured role templates
   - Industry-specific role sets

4. **Advanced Notifications**
   - Webhook support
   - Slack/Teams integration
   - Custom notification rules

5. **Audit Dashboard**
   - Visual activity logs
   - Role usage analytics
   - Permission reports