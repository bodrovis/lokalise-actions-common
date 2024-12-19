package parsers

import (
	"os"
	"strconv"
	"strings"
)

// ParseStringArrayEnv parses a string environment variable into an array of strings.
// It trims spaces, normalizes line endings, and removes empty lines.
func ParseStringArrayEnv(envVar string) []string {
	val := os.Getenv(envVar)
	if val == "" {
		return []string{}
	}

	// Normalize line endings and other potential separators
	normalized := strings.ReplaceAll(val, "\r\n", "\n")     // Normalize Windows to Unix
	normalized = strings.ReplaceAll(normalized, "\r", "\n") // Handle stray carriage returns

	// Split and clean
	var paths []string
	for _, line := range strings.Split(normalized, "\n") {
		path := strings.TrimSpace(line) // Remove extra spaces
		if path != "" {                 // Ignore empty lines
			paths = append(paths, path)
		}
	}

	if len(paths) == 0 {
		return []string{}
	}

	return paths
}

// ParseBoolEnv parses a boolean environment variable.
// Returns false if the variable is not set or empty.
// Returns an error if the value cannot be parsed as a boolean.
func ParseBoolEnv(envVar string) (bool, error) {
	val := os.Getenv(envVar)
	if val == "" {
		return false, nil
	}
	return strconv.ParseBool(val)
}

// ParseUintEnv retrieves an environment variable as a positive integer.
// Returns the default value if the variable is not set, invalid, or less than 1.
func ParseUintEnv(envVar string, defaultVal int) int {
	valStr := os.Getenv(envVar)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil || val < 1 {
		return defaultVal
	}
	return val
}
