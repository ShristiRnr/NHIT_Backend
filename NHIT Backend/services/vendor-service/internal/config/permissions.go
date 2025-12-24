package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Vendor CRUD operations
		"/vendor.VendorService/CreateVendor":       {"create-vendors"}, // Plural
		"/vendor.VendorService/GetVendor":          {"view-vendors"},
		"/vendor.VendorService/GetVendorByCode":    {"view-vendors"},
		"/vendor.VendorService/ListVendors":        {"view-vendors"},
		"/vendor.VendorService/UpdateVendor":       {"edit-vendors"}, // Plural
		"/vendor.VendorService/DeleteVendor":       {"delete-vendors"}, // Plural
		
		// Vendor Code operations
		"/vendor.VendorService/GenerateVendorCode":   {"create-vendors"}, // Plural
		"/vendor.VendorService/UpdateVendorCode":     {"edit-vendors"}, // Plural
		"/vendor.VendorService/RegenerateVendorCode": {"edit-vendors"}, // Plural
		
		// Vendor Account operations
		"/vendor.VendorService/CreateVendorAccount":     {"edit-vendors"}, // Aligned to edit-vendors or create-vendors? User list has create-vendors. Accounts are sub-resource. I'll use edit-vendors for updating, create-vendors for creating? User list: "edit-vendors", "create-vendors". Vendor accounts are part of vendor mgmt. I'll map to edit-vendors generally or verify if "manage-vendor-accounts" is in list. It is NOT. I will use `edit-vendors` for account management to be safe and consistent.
		"/vendor.VendorService/GetVendorAccounts":       {"view-vendors"},
		"/vendor.VendorService/GetVendorBankingDetails": {"view-vendors"},
		"/vendor.VendorService/UpdateVendorAccount":     {"edit-vendors"},
		"/vendor.VendorService/DeleteVendorAccount":     {"edit-vendors"},
		"/vendor.VendorService/ToggleAccountStatus":     {"edit-vendors"},
		
		// Dropdown operations
		"/vendor.VendorService/GetProjectsDropdown":     {"view-vendors"},
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		// Vendor code generation can be public for UI
		"/vendor.VendorService/GenerateVendorCode",
	}
}
