package config

import (
	"fmt"
	"regexp"
)

// nameRegex validates names (alphanumeric, hyphens, underscores, must start with alphanumeric).
var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidateName checks if a name is valid for the given entity type.
// Returns a descriptive error if validation fails.
func ValidateName(name, entityType string) error {
	if name == "" {
		return fmt.Errorf("%s name cannot be empty", entityType)
	}
	if len(name) > 64 {
		return fmt.Errorf("%s name too long (max 64 characters)", entityType)
	}
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("%s name contains invalid characters (use alphanumeric, hyphens, underscores)", entityType)
	}
	return nil
}
