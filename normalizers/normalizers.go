package normalizers

import (
	"fmt"
	"strings"

	"github.com/bodrovis/lokalise-actions-common/v2/parsers"
)

// NormalizeOptionalNamePattern normalizes an optional repo-relative name pattern.
// Empty or whitespace-only input is allowed and returns an empty string.
//
// Non-empty input must be a valid repo-relative pattern.
func NormalizeOptionalNamePattern(pattern string) (string, error) {
	if strings.TrimSpace(pattern) == "" {
		return "", nil
	}

	normalized, err := parsers.EnsureRepoRelativePattern(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid NAME_PATTERN %q: %w", pattern, err)
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
