package githuboutput

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// WriteToGitHubOutput appends a single-line output in "name=value" format
// to the file pointed to by the GITHUB_OUTPUT environment variable.
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

func githubOutputPath() (string, bool) {
	path := os.Getenv("GITHUB_OUTPUT")
	return path, path != ""
}

func normalizeOutputName(name string) (string, bool) {
	name = strings.TrimSpace(name)

	if name == "" || strings.ContainsAny(name, "\r\n=") {
		return "", false
	}

	return name, true
}

func isSingleLineValue(value string) bool {
	return !strings.ContainsAny(value, "\r\n")
}

func appendOutputLine(path, name, value string) (ok bool) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		log.Printf("Failed to open GITHUB_OUTPUT file (%s): %v", path, err)
		return false
	}

	ok = true

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close GITHUB_OUTPUT file (%s): %v", path, err)
			ok = false
		}
	}()

	if _, err := fmt.Fprintf(file, "%s=%s\n", name, value); err != nil {
		log.Printf("Failed to write GitHub output %q: %v", name, err)
		return false
	}

	return true
}
