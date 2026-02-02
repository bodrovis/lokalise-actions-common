package parsers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	yaml "go.yaml.in/yaml/v4"
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

// EnsureRepoRelativePath validates a single path is repo-relative and safe.
// Allowed:
//   - "." => repo root
//   - relative subdirs/files like "locales", "packages/app/locales", "./locales"
//
// Forbidden:
//   - empty/whitespace
//   - absolute (POSIX, Windows drive, UNC)
//   - parent escapes ("..", "../")
//   - drive-relative like "C:foo"
//   - glob metachars: * ? [ ]
//   - tilde-expansion "~", "~user"
//
// Returns a cleaned path (OS-native separators). Caller may ToSlash it.
func EnsureRepoRelativePath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("empty path")
	}

	if strings.ContainsRune(p, '\x00') {
		return "", fmt.Errorf("invalid path: contains NUL")
	}
	if strings.HasPrefix(p, "~") {
		return "", fmt.Errorf("path must be relative to repo (no ~ expansion): %q", p)
	}

	clean := filepath.Clean(p)

	if clean == "." {
		return ".", nil
	}

	if filepath.IsAbs(clean) {
		return "", fmt.Errorf("path must be relative to repo: %q", p)
	}

	s := filepath.ToSlash(clean)

	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "//") {
		return "", fmt.Errorf("path must be relative to repo: %q", p)
	}

	if s == ".." || strings.HasPrefix(s, "../") {
		return "", fmt.Errorf("path escapes repo root: %q", p)
	}

	// Windows drive-relative "C:foo"
	if len(s) >= 2 && s[1] == ':' && ((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z')) {
		return "", fmt.Errorf("path must be relative (drive-prefixed): %q", p)
	}

	if strings.ContainsAny(s, `*?[]`) {
		return "", fmt.Errorf("invalid path %q in TRANSLATIONS_PATH: glob characters are not allowed", p)
	}

	return clean, nil
}

// ParseRepoRelativePathsEnv reads an env var as multiline list (using ParseStringArrayEnv),
// validates each item with EnsureRepoRelativePath, normalizes to forward slashes,
// deduplicates (order-preserving), and returns the set.
// Returns an error if the env var is empty or all entries are invalid.
func ParseRepoRelativePathsEnv(envVar string) ([]string, error) {
	raw := ParseStringArrayEnv(envVar)
	if len(raw) == 0 {
		return nil, fmt.Errorf("environment variable %s is required", envVar)
	}

	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))

	for _, p := range raw {
		clean, err := EnsureRepoRelativePath(p)
		if err != nil {
			return nil, fmt.Errorf("invalid path %q in %s: %w", p, envVar, err)
		}
		norm := filepath.ToSlash(clean)
		if _, dup := seen[norm]; dup {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, norm)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no valid paths found in %s", envVar)
	}
	return out, nil
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

// ParseAdditionalParamsAndMerge parses raw (JSON object or YAML mapping) and copies keys into dst.
// Caller-specified values override existing keys in dst.
func ParseAdditionalParamsAndMerge[M ~map[string]any](dst M, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	add, err := ParseObject(raw)
	if err != nil {
		return err
	}
	maps.Copy(dst, add)
	return nil
}

// ParseObject parses JSON object or YAML mapping into map[string]any.
// Detection: leading "{" => JSON, otherwise YAML.
func ParseObject(raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}, nil
	}
	if strings.HasPrefix(raw, "{") {
		return parseJSONMap(raw)
	}
	return parseYAMLMap(raw)
}

// parseYAMLMap parses a YAML mapping into map[string]any.
func parseYAMLMap(s string) (map[string]any, error) {
	var m map[string]any
	err := yaml.Unmarshal([]byte(s), &m)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("YAML must be a mapping (key: value)")
	}
	return m, nil
}

// parseJSONMap parses a JSON object string into map[string]any.
// Validation: we only accept objects; arrays/primitives are rejected by unmarshal error.
func parseJSONMap(s string) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("JSON must be an object (not null)")
	}
	return m, nil
}
