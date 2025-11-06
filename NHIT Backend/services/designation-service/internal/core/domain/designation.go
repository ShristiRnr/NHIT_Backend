package domain

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// Designation represents a job designation/position in the organization
type Designation struct {
	ID          uuid.UUID
	Name        string
	Description string
	Slug        string
	IsActive    bool
	ParentID    *uuid.UUID // Nullable for hierarchical structure
	Level       int32      // Hierarchy level (0 = top level)
	UserCount   int32      // Cached count of users with this designation
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewDesignation creates a new designation with validation
func NewDesignation(name, description string, isActive bool, parentID *uuid.UUID) (*Designation, error) {
	// Validate name
	if err := ValidateDesignationName(name); err != nil {
		return nil, err
	}

	// Validate description
	if err := ValidateDesignationDescription(description); err != nil {
		return nil, err
	}

	// Generate slug from name
	slug := GenerateSlug(name)

	now := time.Now()
	designation := &Designation{
		ID:          uuid.New(),
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Slug:        slug,
		IsActive:    isActive,
		ParentID:    parentID,
		Level:       0, // Will be calculated based on parent
		UserCount:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return designation, nil
}

// Update updates the designation fields with validation
func (d *Designation) Update(name, description string, isActive bool, parentID *uuid.UUID) error {
	// Validate name
	if err := ValidateDesignationName(name); err != nil {
		return err
	}

	// Validate description
	if err := ValidateDesignationDescription(description); err != nil {
		return err
	}

	// Prevent circular reference (designation cannot be its own parent)
	if parentID != nil && *parentID == d.ID {
		return ErrCircularReference
	}

	// Update fields
	d.Name = strings.TrimSpace(name)
	d.Description = strings.TrimSpace(description)
	d.Slug = GenerateSlug(name)
	d.IsActive = isActive
	d.ParentID = parentID
	d.UpdatedAt = time.Now()

	return nil
}

// Activate sets the designation as active
func (d *Designation) Activate() {
	d.IsActive = true
	d.UpdatedAt = time.Now()
}

// Deactivate sets the designation as inactive
func (d *Designation) Deactivate() {
	d.IsActive = false
	d.UpdatedAt = time.Now()
}

// CanBeDeleted checks if the designation can be deleted
func (d *Designation) CanBeDeleted() error {
	if d.UserCount > 0 {
		return ErrDesignationHasUsers
	}
	return nil
}

// UpdateUserCount updates the cached user count
func (d *Designation) UpdateUserCount(count int32) {
	d.UserCount = count
	d.UpdatedAt = time.Now()
}

// SetLevel sets the hierarchy level
func (d *Designation) SetLevel(level int32) {
	d.Level = level
	d.UpdatedAt = time.Now()
}

// ValidateDesignationName validates the designation name
func ValidateDesignationName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrDesignationNameRequired
	}

	if len(name) < 2 {
		return ErrDesignationNameTooShort
	}

	if len(name) > 250 {
		return ErrDesignationNameTooLong
	}

	// Check if name contains only valid characters (letters, numbers, spaces, hyphens, underscores)
	if !isValidDesignationName(name) {
		return ErrDesignationNameInvalidChars
	}

	// Check for reserved names
	if isReservedName(name) {
		return ErrDesignationNameReserved
	}

	return nil
}

// ValidateDesignationDescription validates the designation description
func ValidateDesignationDescription(description string) error {
	description = strings.TrimSpace(description)

	if description == "" {
		return ErrDesignationDescriptionRequired
	}

	if len(description) < 5 {
		return ErrDesignationDescriptionTooShort
	}

	if len(description) > 500 {
		return ErrDesignationDescriptionTooLong
	}

	return nil
}

// isValidDesignationName checks if the name contains only valid characters
func isValidDesignationName(name string) bool {
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != ' ' && r != '-' && r != '_' && r != '/' && r != '&' && r != '.' {
			return false
		}
	}
	return true
}

// isReservedName checks if the name is reserved
func isReservedName(name string) bool {
	reserved := []string{
		"admin", "administrator", "root", "system", "superuser",
		"null", "undefined", "none", "default",
	}

	lowerName := strings.ToLower(strings.TrimSpace(name))
	for _, r := range reserved {
		if lowerName == r {
			return true
		}
	}
	return false
}

// GenerateSlug generates a URL-friendly slug from the name
func GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and special characters with hyphens
	slug = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return '-'
	}, slug)

	// Remove consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 100 {
		slug = slug[:100]
	}

	return slug
}

// NormalizeDesignationName normalizes the name for comparison
func NormalizeDesignationName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
