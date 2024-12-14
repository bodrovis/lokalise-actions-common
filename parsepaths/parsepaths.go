package parsepaths

import (
	"strings"
)

// ParsePaths takes a newline-separated string from an environment variable
// and returns a slice of cleaned paths. It normalizes line endings and
// filters out empty or whitespace-only lines.
func ParsePaths(envVar string) []string {
	if envVar == "" {
		return []string{} // Explicitly return an empty slice
	}

	// Normalize line endings (Windows-style to Unix-style)
	normalized := strings.ReplaceAll(envVar, "\r\n", "\n")

	// Split the input string into lines and process each line
	var paths []string
	for _, line := range strings.Split(normalized, "\n") {
		path := strings.TrimSpace(line) // Trim whitespace
		if path != "" {                 // Ignore empty lines
			paths = append(paths, path)
		}
	}

	if len(paths) == 0 {
		return []string{} // Explicitly return an empty slice
	}

	return paths
}
