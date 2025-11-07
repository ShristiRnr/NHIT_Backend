package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidOrganizationName is returned when organization name is invalid
	ErrInvalidOrganizationName = errors.New("organization name must be between 3 and 255 characters")
	
	// ErrInvalidOrganizationCode is returned when organization code is invalid
	ErrInvalidOrganizationCode = errors.New("organization code must be between 2 and 10 uppercase alphanumeric characters")
	
	// ErrInvalidDatabaseName is returned when database name is invalid
	ErrInvalidDatabaseName = errors.New("invalid database name format")
	
	// ErrOrganizationNotActive is returned when organization is not active
	ErrOrganizationNotActive = errors.New("organization is not active")
	
	// ErrInvalidTenantID is returned when tenant ID is invalid
	ErrInvalidTenantID = errors.New("tenant ID cannot be empty")
	
	// ErrInvalidCreatorID is returned when creator ID is invalid
	ErrInvalidCreatorID = errors.New("creator ID cannot be empty")
	
	// ErrDuplicateOrganizationCode is returned when organization code already exists
	ErrDuplicateOrganizationCode = errors.New("organization code already exists")
)

// Organization represents an organization in the system with complete business logic
type Organization struct {
	OrgID        uuid.UUID
	TenantID     uuid.UUID
	Name         string
	Code         string     // Unique organization code (e.g., "NHIT", "ABC")
	DatabaseName string     // Database name for multi-tenancy
	Description  string     // Organization description
	Logo         string     // Logo file path/URL
	IsActive     bool       // Active status
	CreatedBy    uuid.UUID  // User ID who created this organization
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewOrganization creates a new organization with validation
func NewOrganization(tenantID uuid.UUID, name, code, description, logo string, createdBy uuid.UUID) (*Organization, error) {
	if tenantID == uuid.Nil {
		return nil, ErrInvalidTenantID
	}
	
	if createdBy == uuid.Nil {
		return nil, ErrInvalidCreatorID
	}
	
	org := &Organization{
		OrgID:       uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Code:        strings.ToUpper(strings.TrimSpace(code)),
		Description: description,
		Logo:        logo,
		IsActive:    true,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Generate database name from code
	org.DatabaseName = org.GenerateDatabaseName()
	
	// Validate the organization
	if err := org.Validate(); err != nil {
		return nil, err
	}
	
	return org, nil
}

// Validate validates the organization fields
func (o *Organization) Validate() error {
	// Validate name
	if len(o.Name) < 3 || len(o.Name) > 255 {
		return ErrInvalidOrganizationName
	}
	
	// Validate code
	if !o.IsValidCode() {
		return ErrInvalidOrganizationCode
	}
	
	// Validate database name
	if !o.IsValidDatabaseName() {
		return ErrInvalidDatabaseName
	}
	
	return nil
}

// IsValidCode checks if the organization code is valid
func (o *Organization) IsValidCode() bool {
	if len(o.Code) < 2 || len(o.Code) > 10 {
		return false
	}
	
	// Code must be uppercase alphanumeric
	matched, _ := regexp.MatchString(`^[A-Z0-9_]+$`, o.Code)
	return matched
}

// IsValidDatabaseName checks if the database name is valid
func (o *Organization) IsValidDatabaseName() bool {
	if o.DatabaseName == "" {
		return false
	}
	
	// Database name must be lowercase alphanumeric with underscores
	matched, _ := regexp.MatchString(`^[a-z0-9_]+$`, o.DatabaseName)
	return matched && len(o.DatabaseName) <= 64
}

// GenerateDatabaseName generates a database name from the organization code
func (o *Organization) GenerateDatabaseName() string {
	// Convert code to lowercase and replace spaces/hyphens with underscores
	dbName := strings.ToLower(o.Code)
	dbName = strings.ReplaceAll(dbName, " ", "_")
	dbName = strings.ReplaceAll(dbName, "-", "_")
	
	// Add a prefix to avoid conflicts
	return fmt.Sprintf("org_%s", dbName)
}

// Activate activates the organization
func (o *Organization) Activate() {
	o.IsActive = true
	o.UpdatedAt = time.Now()
}

// Deactivate deactivates the organization
func (o *Organization) Deactivate() {
	o.IsActive = false
	o.UpdatedAt = time.Now()
}

// ToggleStatus toggles the organization active status
func (o *Organization) ToggleStatus() {
	o.IsActive = !o.IsActive
	o.UpdatedAt = time.Now()
}

// Update updates the organization fields
func (o *Organization) Update(name, code, description, logo string, isActive bool) error {
	o.Name = name
	o.Code = strings.ToUpper(strings.TrimSpace(code))
	o.Description = description
	o.Logo = logo
	o.IsActive = isActive
	o.UpdatedAt = time.Now()
	
	// Regenerate database name if code changed
	o.DatabaseName = o.GenerateDatabaseName()
	
	// Validate the updated organization
	if err := o.Validate(); err != nil {
		return err
	}
	
	return nil
}

// CanBeAccessed checks if the organization can be accessed
func (o *Organization) CanBeAccessed() bool {
	return o.IsActive
}

// IsCreatedBy checks if the organization was created by the specified user
func (o *Organization) IsCreatedBy(userID uuid.UUID) bool {
	return o.CreatedBy == userID
}

// UserOrganization represents the association between users and organizations
// This is kept separate to maintain clear separation of concerns
type UserOrganization struct {
	UserID           uuid.UUID
	OrgID            uuid.UUID
	RoleID           uuid.UUID  // Role within this organization
	IsCurrentContext bool       // Whether this is the user's current organization context
	JoinedAt         time.Time
	UpdatedAt        time.Time
}

// NewUserOrganization creates a new user-organization association
func NewUserOrganization(userID, orgID, roleID uuid.UUID) (*UserOrganization, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}
	
	if orgID == uuid.Nil {
		return nil, errors.New("organization ID cannot be empty")
	}
	
	if roleID == uuid.Nil {
		return nil, errors.New("role ID cannot be empty")
	}
	
	return &UserOrganization{
		UserID:           userID,
		OrgID:            orgID,
		RoleID:           roleID,
		IsCurrentContext: false,
		JoinedAt:         time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

// SetAsCurrentContext sets this organization as the user's current context
func (uo *UserOrganization) SetAsCurrentContext() {
	uo.IsCurrentContext = true
	uo.UpdatedAt = time.Now()
}

// RemoveAsCurrentContext removes this organization as the user's current context
func (uo *UserOrganization) RemoveAsCurrentContext() {
	uo.IsCurrentContext = false
	uo.UpdatedAt = time.Now()
}

// UpdateRole updates the user's role within the organization
func (uo *UserOrganization) UpdateRole(roleID uuid.UUID) error {
	if roleID == uuid.Nil {
		return errors.New("role ID cannot be empty")
	}
	
	uo.RoleID = roleID
	uo.UpdatedAt = time.Now()
	return nil
}
