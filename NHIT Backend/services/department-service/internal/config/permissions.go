package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Department CRUD operations
		"/department.v1.DepartmentService/CreateDepartment": {"create-department"},
		"/department.v1.DepartmentService/GetDepartment":    {"view-departments"},
		"/department.v1.DepartmentService/ListDepartments":  {"view-departments"},
		"/department.v1.DepartmentService/UpdateDepartment": {"edit-department"},
		"/department.v1.DepartmentService/DeleteDepartment": {"delete-department"},
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		// Add public methods here if any
	}
}
