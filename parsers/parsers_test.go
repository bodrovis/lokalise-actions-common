package parsers

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseStringArrayEnv(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		expected []string
	}{
		{
			name:     "Single path, no newline",
			envKey:   "TEST_ENV",
			envValue: "/path/to/dir",
			expected: []string{"/path/to/dir"},
		},
		{
			name:     "Multiple paths, Unix newlines",
			envKey:   "TEST_ENV",
			envValue: "/path/to/dir\n/another/path\n/yet/another/path",
			expected: []string{"/path/to/dir", "/another/path", "/yet/another/path"},
		},
		{
			name:     "Multiple paths, Windows newlines",
			envKey:   "TEST_ENV",
			envValue: "/path/to/dir\r\n/another/path\r\n/yet/another/path",
			expected: []string{"/path/to/dir", "/another/path", "/yet/another/path"},
		},
		{
			name:     "Empty lines and whitespace",
			envKey:   "TEST_ENV",
			envValue: "/path/to/dir\n\n \n/another/path\n   \n/yet/another/path\n",
			expected: []string{"/path/to/dir", "/another/path", "/yet/another/path"},
		},
		{
			name:     "Only empty lines and whitespace",
			envKey:   "TEST_ENV",
			envValue: "\n\n \n  \n",
			expected: []string{},
		},
		{
			name:     "Empty string input",
			envKey:   "TEST_ENV",
			envValue: "",
			expected: []string{},
		},
		{
			name:     "Paths with special characters",
			envKey:   "TEST_ENV",
			envValue: "/path/with/special!@#$%^&*()_+\n/path/with/[]{}|;:',.<>?",
			expected: []string{"/path/with/special!@#$%^&*()_+", "/path/with/[]{}|;:',.<>?"},
		},
		{
			name:     "Environment key not set",
			envKey:   "UNSET_ENV",
			envValue: "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(tt.envKey, tt.envValue)
			if err != nil {
				log.Printf("Failed to set ENV %s -> %s: %v", tt.envKey, tt.envValue, err)
			}

			defer func() {
				// cleanup
				if err := os.Unsetenv(tt.envKey); err != nil {
					log.Printf("Failed to unset %s: %v", tt.envKey, err)
				}
			}()

			result := ParseStringArrayEnv(tt.envKey)
			if !reflect.DeepEqual(normalizeSlice(result), normalizeSlice(tt.expected)) {
				t.Errorf("ParseStringArrayEnv(%q) = %v, want %v", tt.envKey, result, tt.expected)
			}
		})
	}
}

func TestParseBoolEnv(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		expected bool
		wantErr  bool
	}{
		{"Unset variable", "UNSET_ENV", "", false, false},
		{"Empty value", "EMPTY_ENV", "", false, false},
		{"True value (true)", "TRUE_ENV", "true", true, false},
		{"True value (TRUE)", "TRUE_ENV_UPPER", "TRUE", true, false},
		{"False value (false)", "FALSE_ENV", "false", false, false},
		{"Invalid value", "INVALID_ENV", "invalid", false, true},
		{"Numeric true (1)", "NUMERIC_TRUE_ENV", "1", true, false},
		{"Numeric false (0)", "NUMERIC_FALSE_ENV", "0", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable for the test
			if tt.envValue != "" {
				err := os.Setenv(tt.envKey, tt.envValue)
				if err != nil {
					log.Printf("Failed to set %s to %s: %v", tt.envKey, tt.envValue, err)
				}

				defer func() {
					// cleanup
					if err := os.Unsetenv(tt.envKey); err != nil {
						log.Printf("Failed to unset %s: %v", tt.envKey, err)
					}
				}()
			} else {
				defer func() {
					// cleanup
					if err := os.Unsetenv(tt.envKey); err != nil {
						log.Printf("Failed to unset %s: %v", tt.envKey, err)
					}
				}()
			}

			// Call the function
			result, err := ParseBoolEnv(tt.envKey)

			// Check for errors
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBoolEnv(%q) error = %v, wantErr = %v", tt.envKey, err, tt.wantErr)
				return
			}

			// Check the result
			if result != tt.expected {
				t.Errorf("ParseBoolEnv(%q) = %v, want %v", tt.envKey, result, tt.expected)
			}
		})
	}
}

func TestParseUintEnv(t *testing.T) {
	tests := []struct {
		name       string
		envKey     string
		envValue   string
		defaultVal int
		expected   int
	}{
		{"Unset variable", "UNSET_ENV", "", 10, 10},
		{"Empty value", "EMPTY_ENV", "", 5, 5},
		{"Valid positive integer", "VALID_ENV", "42", 10, 42},
		{"Zero value", "ZERO_ENV", "0", 10, 10},
		{"Negative value", "NEGATIVE_ENV", "-5", 15, 15},
		{"Non-numeric value", "INVALID_ENV", "abc", 20, 20},
		{"Whitespace input", "WHITESPACE_ENV", "   ", 25, 25},
		{"Very large value", "LARGE_ENV", "99999", 1, 99999},
		{"Boundary value (1)", "BOUNDARY_ENV", "1", 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the environment variable
			if tt.envValue != "" {
				err := os.Setenv(tt.envKey, tt.envValue)
				if err != nil {
					log.Printf("Failed to set %s to %s: %v", tt.envKey, tt.envValue, err)
				}

				defer func() {
					// cleanup
					if err := os.Unsetenv(tt.envKey); err != nil {
						log.Printf("Failed to unset %s: %v", tt.envKey, err)
					}
				}()
			} else {
				defer func() {
					// cleanup
					if err := os.Unsetenv(tt.envKey); err != nil {
						log.Printf("Failed to unset %s: %v", tt.envKey, err)
					}
				}()
			}

			// Call the function
			result := ParseUintEnv(tt.envKey, tt.defaultVal)

			// Validate the result
			if result != tt.expected {
				t.Errorf("ParseUintEnv(%q, %d) = %d, want %d", tt.envKey, tt.defaultVal, result, tt.expected)
			}
		})
	}
}

func TestEnsureRepoRelativePath(t *testing.T) {
	type tc struct {
		name        string
		in          string
		want        string
		expectError string
	}
	absPath, _ := filepath.Abs("some/abs/path")

	cases := []tc{
		{
			name: "simple relative",
			in:   "a/b",
			want: "a/b",
		},
		{
			name: "normalizes ./ and // and ..",
			in:   "./a//b/../c",
			want: "a/c",
		},
		{
			name: "trims trailing slash",
			in:   "x/",
			want: "x",
		},
		{
			name: "dot path is forbidden",
			in:   ".",
			want: ".",
		},
		{
			name:        "absolute path is forbidden",
			in:          absPath,
			expectError: "path must be relative to repo",
		},
		{
			name:        "parent escape is forbidden (../)",
			in:          "../outside",
			expectError: "escapes repo root",
		},
		{
			name:        "parent escape after clean a/../../b",
			in:          "a/../../b",
			expectError: "escapes repo root",
		},
		{
			name:        "UNC-like path forbidden",
			in:          "//server/share",
			expectError: "path must be relative to repo",
		},
		{
			name:        "drive-prefixed relative forbidden",
			in:          "C:foo",
			expectError: "drive-prefixed",
		},
		{
			name:        "glob meta * forbidden",
			in:          "locales/*",
			expectError: "glob characters are not allowed",
		},
		{
			name:        "glob meta ? forbidden",
			in:          "foo?.yml",
			expectError: "glob characters are not allowed",
		},
		{
			name:        "glob meta [] forbidden",
			in:          "bar[0]",
			expectError: "glob characters are not allowed",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnsureRepoRelativePath(tt.in)
			if tt.expectError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.expectError) {
					t.Fatalf("expected error containing %q, got %v", tt.expectError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if filepath.ToSlash(got) != tt.want {
				t.Fatalf("got %q, want %q", filepath.ToSlash(got), tt.want)
			}
		})
	}
}

func TestParseRepoRelativePathsEnv(t *testing.T) {
	allKeys := []string{
		"TEST_PATHS",
	}
	for _, k := range allKeys {
		t.Setenv(k, "")
	}

	t.Run("valid multiple, normalize and dedupe order-preserving", func(t *testing.T) {
		// ./x, x/, x  -> "x"; a//b/../c -> a/c
		val := strings.Join([]string{
			"./x",
			"x/",
			"./y",
			"a//b/../c",
			"x",
		}, "\n")
		t.Setenv("TEST_PATHS", val)

		got, err := ParseRepoRelativePathsEnv("TEST_PATHS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"x", "y", "a/c"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("single valid", func(t *testing.T) {
		t.Setenv("TEST_PATHS", "locales")
		got, err := ParseRepoRelativePathsEnv("TEST_PATHS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"locales"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("required env missing -> error", func(t *testing.T) {
		t.Setenv("TEST_PATHS", "")
		_, err := ParseRepoRelativePathsEnv("TEST_PATHS")
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Fatalf("expected required error, got %v", err)
		}
	})

	t.Run("first invalid path -> error (glob)", func(t *testing.T) {
		t.Setenv("TEST_PATHS", "locales/*\nvalid")
		_, err := ParseRepoRelativePathsEnv("TEST_PATHS")
		if err == nil || !strings.Contains(err.Error(), "glob characters") {
			t.Fatalf("expected glob error, got %v", err)
		}
	})

	t.Run("mixed valid and invalid -> error at invalid", func(t *testing.T) {
		t.Setenv("TEST_PATHS", "a\n../up\nb")
		_, err := ParseRepoRelativePathsEnv("TEST_PATHS")
		if err == nil || !strings.Contains(err.Error(), "escapes repo root") {
			t.Fatalf("expected escape error, got %v", err)
		}
	})

	t.Run("UNC-like path -> error", func(t *testing.T) {
		t.Setenv("TEST_PATHS", "//server/share")
		_, err := ParseRepoRelativePathsEnv("TEST_PATHS")
		if err == nil || !strings.Contains(err.Error(), "relative to repo") {
			t.Fatalf("expected relative-to-repo error, got %v", err)
		}
	})

	t.Run("drive-prefixed relative -> error", func(t *testing.T) {
		t.Setenv("TEST_PATHS", "C:foo")
		_, err := ParseRepoRelativePathsEnv("TEST_PATHS")
		if err == nil || !strings.Contains(err.Error(), "drive-prefixed") {
			t.Fatalf("expected drive-prefixed error, got %v", err)
		}
	})
}

func normalizeSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
