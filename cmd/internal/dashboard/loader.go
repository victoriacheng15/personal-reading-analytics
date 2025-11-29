package dashboard

import (
	"fmt"
	"os"
	"path/filepath"
)

// LoadTemplateContent loads the HTML template from the template.html file
// It tries multiple path configurations to handle different execution contexts
func LoadTemplateContent() (string, error) {
	// Possible paths where template.html might be located
	possiblePaths := []string{
		// Relative to current working directory (common when running from project root)
		"cmd/internal/dashboard/template.html",
		// Relative to binary location (when built and run from cmd/dashboard)
		"dashboard/template.html",
		"internal/dashboard/template.html",
		// Absolute path patterns based on common project structures
		filepath.Join(".", "cmd", "internal", "dashboard", "template.html"),
	}

	// Try to find the template file
	for _, path := range possiblePaths {
		if content, err := os.ReadFile(path); err == nil {
			return string(content), nil
		}
	}

	// If file not found in any location, return error with debugging info
	return "", fmt.Errorf("template.html not found. Tried paths: %v", possiblePaths)
}
