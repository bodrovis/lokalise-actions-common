package parsers

import (
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
			name:     "Classic Mac newlines (CR only)",
			envKey:   "TEST_ENV",
			envValue: "a\rb\rc",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Mixed newline styles and trimming",
			envKey:   "TEST_ENV",
			envValue: " a \r\nb\r c \n\n",
			expected: []string{"a", "b", "c"},
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
			name:     "Empty variable",
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
			t.Setenv(tt.envKey, tt.envValue)

			result := ParseStringArrayEnv(tt.envKey)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Fatalf("ParseStringArrayEnv(%q) = %v, want %v", tt.envKey, result, tt.expected)
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
		{"Trimmed true value", "TRIMMED_TRUE_ENV", "  true \n", true, false},
		{"Trimmed false value", "TRIMMED_FALSE_ENV", "\t false  ", false, false},
		{"Whitespace around invalid value", "TRIMMED_INVALID_ENV", "  nope  ", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envKey, tt.envValue)

			result, err := ParseBoolEnv(tt.envKey)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseBoolEnv(%q) error = %v, wantErr = %v", tt.envKey, err, tt.wantErr)
			}

			if result != tt.expected {
				t.Fatalf("ParseBoolEnv(%q) = %v, want %v", tt.envKey, result, tt.expected)
			}
		})
	}
}

func TestParseBoolEnv_UnsetVariable(t *testing.T) {
	key := "SOME_UNSET_ENV_FOR_TEST"

	result, err := ParseBoolEnv(key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != false {
		t.Fatalf("got %v, want false", result)
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
		{"Trimmed valid integer", "TRIMMED_VALID_ENV", "  42 \n", 10, 42},
		{"Trimmed zero value", "TRIMMED_ZERO_ENV", "  0  ", 10, 10},
		{"Trimmed negative value", "TRIMMED_NEGATIVE_ENV", "  -7 ", 15, 15},
		{"Trimmed invalid value", "TRIMMED_INVALID_ENV", "\t abc \n", 20, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envKey, tt.envValue)

			result := ParseUintEnv(tt.envKey, tt.defaultVal)

			if result != tt.expected {
				t.Fatalf("ParseUintEnv(%q, %d) = %d, want %d", tt.envKey, tt.defaultVal, result, tt.expected)
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
			name: "dot path is allowed as repo root",
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
			name:        "empty path forbidden",
			in:          "",
			expectError: "empty path",
		},
		{
			name:        "whitespace-only path forbidden",
			in:          "   \t  ",
			expectError: "empty path",
		},
		{
			name:        "tilde path forbidden",
			in:          "~/.config",
			expectError: "no ~ expansion",
		},
		{
			name:        "tilde-user path forbidden",
			in:          "~john/file",
			expectError: "no ~ expansion",
		},
		{
			name:        "NUL byte forbidden",
			in:          "abc\x00def",
			expectError: "contains NUL",
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

	t.Run("repo root dot path is allowed", func(t *testing.T) {
		t.Setenv("TEST_PATHS", ".")
		got, err := ParseRepoRelativePathsEnv("TEST_PATHS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"."}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

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

func TestParseObject_JSON(t *testing.T) {
	raw := `
{
  "indentation": "2sp",
  "replace_breaks": false,
  "include_tags": ["custom-1","custom-2"],
  "nested": {"a": 1}
}
`

	got, err := ParseObject(raw)
	if err != nil {
		t.Fatalf("ParseObject(JSON) returned error: %v", err)
	}

	want := map[string]any{
		"indentation":    "2sp",
		"replace_breaks": false,
		"include_tags":   []any{"custom-1", "custom-2"},
		"nested":         map[string]any{"a": float64(1)}, // JSON numbers -> float64
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseObject(JSON) mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestParseObject_YAML(t *testing.T) {
	raw := `
indentation: 2sp
replace_breaks: false
include_tags:
  - custom-1
  - custom-2
nested:
  a: 1
`

	got, err := ParseObject(raw)
	if err != nil {
		t.Fatalf("ParseObject(YAML) returned error: %v", err)
	}

	// YAML numbers usually decode into int, not float64 (depends on lib),
	// so keep expected as int to match go.yaml.in/yaml/v4 typical behavior.
	want := map[string]any{
		"indentation":    "2sp",
		"replace_breaks": false,
		"include_tags":   []any{"custom-1", "custom-2"},
		"nested":         map[string]any{"a": 1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseObject(YAML) mismatch.\n got: %#v\nwant: %#v", got, want)
	}
}

func TestParseObject_Empty_ReturnsEmptyMap(t *testing.T) {
	got, err := ParseObject("   \n\t  ")
	if err != nil {
		t.Fatalf("ParseObject(empty) returned error: %v", err)
	}
	if got == nil {
		t.Fatalf("ParseObject(empty) returned nil map, want empty map")
	}
	if len(got) != 0 {
		t.Fatalf("ParseObject(empty) len=%d, want 0. map=%#v", len(got), got)
	}
}

func TestParseObject_InvalidCases(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"Invalid JSON", `{"indentation":"2sp",`},
		{"Invalid YAML", "indentation: \"2sp\n"},
		{"YAML Not a Mapping", "- a\n- b\n"},
		// JSON "null" is valid JSON but invalid for our object-only rule.
		{"JSON Null", "null"},
		{"JSON Array", `["a","b"]`},
		{"JSON Number", `123`},
		{"YAML Null", `null`},
		{"YAML Scalar", `hello`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseObject(tt.raw)
			if err == nil {
				t.Fatalf("ParseObject(%s) expected error, got nil", tt.name)
			}
		})
	}
}

func TestParseAdditionalParamsAndMerge_JSON_Overrides(t *testing.T) {
	type Params map[string]any

	dst := Params{
		"format":             "json",
		"original_filenames": true,
		"include_tags":       []any{"default-tag"},
	}

	raw := `
{
  "include_tags": ["custom-1","custom-2"],
  "indentation": "2sp",
  "replace_breaks": false
}
`

	err := ParseAdditionalParamsAndMerge(dst, raw)
	if err != nil {
		t.Fatalf("ParseAdditionalParamsAndMerge(JSON) error: %v", err)
	}

	want := Params{
		"format":             "json",
		"original_filenames": true,
		"include_tags":       []any{"custom-1", "custom-2"}, // overridden
		"indentation":        "2sp",
		"replace_breaks":     false,
	}

	if !reflect.DeepEqual(dst, want) {
		t.Fatalf("merge mismatch.\n got: %#v\nwant: %#v", dst, want)
	}
}

func TestParseAdditionalParamsAndMerge_YAML_Overrides(t *testing.T) {
	type Params map[string]any

	dst := Params{
		"format":       "yaml",
		"include_tags": []any{"default-tag"},
	}

	raw := `
include_tags:
  - custom-1
  - custom-2
indentation: 2sp
replace_breaks: false
`

	err := ParseAdditionalParamsAndMerge(dst, raw)
	if err != nil {
		t.Fatalf("ParseAdditionalParamsAndMerge(YAML) error: %v", err)
	}

	want := Params{
		"format":         "yaml",
		"include_tags":   []any{"custom-1", "custom-2"},
		"indentation":    "2sp",
		"replace_breaks": false,
	}

	if !reflect.DeepEqual(dst, want) {
		t.Fatalf("merge mismatch.\n got: %#v\nwant: %#v", dst, want)
	}
}

func TestParseAdditionalParamsAndMerge_NilDst(t *testing.T) {
	type Params map[string]any
	var dst Params

	err := ParseAdditionalParamsAndMerge(dst, `{"a":1}`)
	if err == nil || !strings.Contains(err.Error(), "must not be nil") {
		t.Fatalf("expected nil-dst error, got %v", err)
	}
}

func TestParseAdditionalParamsAndMerge_Empty_NoChanges(t *testing.T) {
	type Params map[string]any

	dst := Params{
		"format": "json",
	}

	err := ParseAdditionalParamsAndMerge(dst, " \n\t ")
	if err != nil {
		t.Fatalf("ParseAdditionalParamsAndMerge(empty) error: %v", err)
	}

	want := Params{"format": "json"}
	if !reflect.DeepEqual(dst, want) {
		t.Fatalf("expected no changes.\n got: %#v\nwant: %#v", dst, want)
	}
}

func TestParseAdditionalParamsAndMerge_Invalid_ReturnsError(t *testing.T) {
	type Params map[string]any

	dst := Params{
		"format": "json",
	}

	err := ParseAdditionalParamsAndMerge(dst, `{"indentation":"2sp",`)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
