package dashboard

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// LoadTemplateContent loads the HTML template from the template.html file
// It tries multiple path configurations to handle different execution contexts
// (running from project root, running from built binary, etc.)
func LoadTemplateContent() (string, error) {
	// Define canonical paths in priority order
	possiblePaths := []string{
		// When running from project root (most common during development)
		"cmd/internal/dashboard/template.html",
		// When binary is in cmd/dashboard directory
		"internal/dashboard/template.html",
		// Fallback: explicit relative path construction
		filepath.Join(".", "cmd", "internal", "dashboard", "template.html"),
	}

	var cwd string
	if wd, err := os.Getwd(); err == nil {
		cwd = wd
	}

	// Try each path
	for _, path := range possiblePaths {
		if content, err := os.ReadFile(path); err == nil {
			log.Printf("✅ Loaded template from: %s\n", path)
			return string(content), nil
		}
	}

	// Enhanced error message with debugging info
	return "", fmt.Errorf(
		"template.html not found. Current working directory: %s. Tried paths: %v",
		cwd, possiblePaths,
	)
}
