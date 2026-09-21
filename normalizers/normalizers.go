package normalizers

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bodrovis/lokalise-actions-common/v2/parsers"

	"github.com/bmatcuk/doublestar/v4"
)

// NormalizeOptionalNamePattern normalizes an optional repo-relative glob pattern.
// Empty or whitespace-only input is allowed and returns an empty string.
func NormalizeOptionalNamePattern(pattern string) (string, error) {
	if strings.TrimSpace(pattern) == "" {
		return "", nil
	}

	normalized, err := normalizeRepoRelativeGlobPattern(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid NAME_PATTERN %q: %w", pattern, err)
	}

	return normalized, nil
}

// NormalizeGlobPatterns normalizes repo-relative glob patterns.
// Empty entries are ignored.
func NormalizeGlobPatterns(patterns []string) ([]string, error) {
	normalized := make([]string, 0, len(patterns))

	for _, pattern := range patterns {
		if strings.TrimSpace(pattern) == "" {
			continue
		}

		value, err := normalizeRepoRelativeGlobPattern(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}

		normalized = append(normalized, value)
	}

	return normalized, nil
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
		return nil, errors.New("no file extensions provided")
	}

	seen := make(map[string]struct{}, len(exts))
	out := make([]string, 0, len(exts))

	for _, raw := range exts {
		ext := strings.TrimSpace(raw)
		ext = strings.TrimPrefix(ext, ".")
		ext = strings.ToLower(ext)

		if ext == "" {
			continue
		}

		if strings.ContainsAny(ext, `/\`) {
			return nil, fmt.Errorf("invalid file extension %q", raw)
		}

		if _, ok := seen[ext]; ok {
			continue
		}

		seen[ext] = struct{}{}
		out = append(out, ext)
	}

	if len(out) == 0 {
		return nil, errors.New("no valid file extensions after normalization")
	}

	return out, nil
}

func normalizeRepoRelativeGlobPattern(pattern string) (string, error) {
	normalized, err := parsers.EnsureRepoRelativePattern(pattern)
	if err != nil {
		return "", err
	}

	normalized = filepath.ToSlash(normalized)

	if !doublestar.ValidatePattern(normalized) {
		return "", errors.New("invalid glob syntax")
	}

	return normalized, nil
}
