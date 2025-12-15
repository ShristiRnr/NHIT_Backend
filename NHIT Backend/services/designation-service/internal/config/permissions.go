package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Designation CRUD operations
		"/designation.v1.DesignationService/CreateDesignation": {"create-designation"},
		"/designation.v1.DesignationService/GetDesignation":    {"view-designations"},
		"/designation.v1.DesignationService/ListDesignations":  {"view-designations"},
		"/designation.v1.DesignationService/UpdateDesignation": {"edit-designation"},
		"/designation.v1.DesignationService/DeleteDesignation": {"delete-designation"},
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		// Add public methods here if any
	}
}
