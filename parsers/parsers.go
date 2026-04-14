package parsers

import (
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

	val = strings.ReplaceAll(val, "\r\n", "\n")
	val = strings.ReplaceAll(val, "\r", "\n")

	lines := strings.Split(val, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	return result
}

// EnsureRepoRelativePattern validates a single repo-relative path or pattern.
// Allowed:
//   - "." => repo root
//   - relative subdirs/files like "locales", "packages/app/locales", "./locales"
//   - repo-relative glob patterns like "**/*.yaml", "locales/**/en.json"
//
// Forbidden:
//   - empty/whitespace
//   - absolute (POSIX, Windows drive, UNC)
//   - parent escapes ("..", "../")
//   - drive-relative like "C:foo"
//   - tilde-expansion "~", "~user"
//   - NUL byte
//
// Returns a cleaned path/pattern (OS-native separators). Caller may ToSlash it.
func EnsureRepoRelativePattern(p string) (string, error) {
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

	if strings.HasPrefix(s, "/") {
		return "", fmt.Errorf("path must be relative to repo: %q", p)
	}

	if s == ".." || strings.HasPrefix(s, "../") {
		return "", fmt.Errorf("path escapes repo root: %q", p)
	}

	// Windows drive-relative "C:foo"
	if len(s) >= 2 && s[1] == ':' && ((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z')) {
		return "", fmt.Errorf("path must be relative (drive-prefixed): %q", p)
	}

	return clean, nil
}

// EnsureRepoRelativePath validates a single path is repo-relative and safe.
// Same rules as EnsureRepoRelativePattern, but glob metacharacters are forbidden.
func EnsureRepoRelativePath(p string) (string, error) {
	clean, err := EnsureRepoRelativePattern(p)
	if err != nil {
		return "", err
	}

	s := filepath.ToSlash(clean)
	if strings.ContainsAny(s, `*?[]`) {
		return "", fmt.Errorf("invalid path %q: glob characters are not allowed", p)
	}

	return clean, nil
}

// ParseRepoRelativePathsEnv reads an env var as multiline list (using ParseStringArrayEnv),
// validates each item with EnsureRepoRelativePath, normalizes to forward slashes,
// deduplicates (order-preserving), and returns the set.
// Returns an error if the env var is empty or any entry is invalid.
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
	val := strings.TrimSpace(os.Getenv(envVar))
	if val == "" {
		return false, nil
	}

	return strconv.ParseBool(val)
}

// ParseUintEnv retrieves an environment variable as a positive integer.
// Returns the default value if the variable is not set, invalid, or less than 1.
func ParseUintEnv(envVar string, defaultVal int) int {
	valStr := strings.TrimSpace(os.Getenv(envVar))
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
	if dst == nil {
		return fmt.Errorf("destination map must not be nil")
	}

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

// NormalizeFileExtensions normalizes a list of file extensions.
//
// Behavior:
//   - trims surrounding whitespace
//   - removes a leading dot (".json" -> "json")
//   - lowercases all values
//   - removes duplicates while preserving order
//   - skips empty values after normalization
//
// Errors:
//   - returns an error if input slice is empty
//   - returns an error if no valid extensions remain after normalization
//
// Example:
//
//	input:  []string{".JSON", " yaml ", "json", ""}
//	output: []string{"json", "yaml"}
func NormalizeFileExtensions(exts []string) ([]string, error) {
	if len(exts) == 0 {
		return nil, fmt.Errorf("cannot infer file extension: FILE_EXT is empty")
	}

	seen := make(map[string]struct{}, len(exts))
	out := make([]string, 0, len(exts))

	for _, ext := range exts {
		ext = strings.TrimSpace(ext)
		ext = strings.TrimPrefix(ext, ".")
		ext = strings.ToLower(ext)

		if ext == "" {
			continue
		}
		if _, ok := seen[ext]; ok {
			continue
		}
		seen[ext] = struct{}{}
		out = append(out, ext)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no valid file extensions after normalization")
	}

	return out, nil
}
