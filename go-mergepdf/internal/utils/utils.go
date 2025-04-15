// Package utils provides utility functions for filename sanitization and UUID generation.
//
// Functions:
//   - SanitizeFilename: Returns a safe filename for storage.
//     Input: string (filename)
//     Output: string (sanitized filename)
//   - GenerateUUID: Returns a new UUID string.
//     Output: string (UUID)
//
// Used throughout the backend for safe file handling and unique IDs.
package utils

import (
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
)

func SanitizeFilename(name string) string {
	base := filepath.Base(name)
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	safe := re.ReplaceAllString(base, "_")
	if len(safe) > 100 {
		safe = safe[:100]
	}
	return safe
}

func GenerateUUID() string {
	return uuid.New().String()
}
