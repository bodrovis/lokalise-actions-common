package parsers

import (
	"bufio"
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

	// Normalize line endings to Unix-style
	val = strings.ReplaceAll(val, "\r\n", "\n")
	val = strings.ReplaceAll(val, "\r", "\n")

	scanner := bufio.NewScanner(strings.NewReader(val))
	var result []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			result = append(result, line)
		}
	}

	return result
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
