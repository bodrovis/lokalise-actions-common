package githuboutput

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// WriteToGitHubOutput appends a single-line output in "name=value" format
// to the file pointed to by the GITHUB_OUTPUT environment variable.
//
// This helper is intentionally strict:
//   - GITHUB_OUTPUT must be set
//   - name must be non-empty after trimming
//   - name must not contain '\r', '\n', or '='
//   - value must be single-line (no '\r' or '\n')
//
// Returns true on success, false on validation or I/O failure.
func WriteToGitHubOutput(name, value string) bool {
	githubOutput, ok := githubOutputPath()
	if !ok {
		return false
	}

	name, ok = normalizeOutputName(name)
	if !ok {
		return false
	}

	if !isSingleLineValue(value) {
		return false
	}

	return appendOutputLine(githubOutput, name, value)
}

// githubOutputPath returns the GITHUB_OUTPUT file path if it is available.
func githubOutputPath() (string, bool) {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return "", false
	}
	return path, true
}

// normalizeOutputName trims the name and validates that it is safe
// for the "name=value" GitHub Actions output format.
func normalizeOutputName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}

	if strings.ContainsAny(name, "\r\n=") {
		return "", false
	}

	return name, true
}

// isSingleLineValue reports whether value can be safely written
// using the simple single-line GitHub Actions output format.
func isSingleLineValue(value string) bool {
	return !strings.ContainsAny(value, "\r\n")
}

// appendOutputLine opens the GitHub output file in append mode
// and writes one "name=value" line to it.
func appendOutputLine(path, name, value string) bool {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		log.Printf("Failed to open GITHUB_OUTPUT file (%s): %v", path, err)
		return false
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Failed to close GITHUB_OUTPUT file (%s): %v", path, cerr)
		}
	}()

	if _, err := fmt.Fprintf(file, "%s=%s\n", name, value); err != nil {
		log.Printf("Failed to write GitHub output %q: %v", name, err)
		return false
	}

	return true
}
