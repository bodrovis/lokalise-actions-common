package normalizers

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeOptionalNamePattern(t *testing.T) {
	type tc struct {
		name        string
		in          string
		want        string
		expectError string
	}

	absPath, _ := filepath.Abs("some/abs/path")

	cases := []tc{
		{
			name: "empty string returns empty without error",
			in:   "",
			want: "",
		},
		{
			name: "whitespace only returns empty without error",
			in:   "   \t   ",
			want: "",
		},
		{
			name: "simple relative pattern",
			in:   "locales/en.json",
			want: "locales/en.json",
		},
		{
			name: "glob pattern is allowed",
			in:   "locales/**/*.yaml",
			want: "locales/**/*.yaml",
		},
		{
			name: "normalizes path-like pattern",
			in:   "./locales//en/../**/*.json",
			want: "locales/**/*.json",
		},
		{
			name: "dot path is allowed",
			in:   ".",
			want: ".",
		},
		{
			name:        "absolute path is forbidden",
			in:          absPath,
			expectError: "invalid NAME_PATTERN",
		},
		{
			name:        "parent escape is forbidden",
			in:          "../outside/*.json",
			expectError: "escapes repo root",
		},
		{
			name:        "parent escape after clean is forbidden",
			in:          "a/../../b/*.json",
			expectError: "escapes repo root",
		},
		{
			name:        "tilde path is forbidden",
			in:          "~/.config/*.json",
			expectError: "no ~ expansion",
		},
		{
			name:        "tilde user path is forbidden",
			in:          "~john/file.json",
			expectError: "no ~ expansion",
		},
		{
			name:        "NUL byte is forbidden",
			in:          "abc\x00def.json",
			expectError: "contains NUL",
		},
		{
			name:        "UNC-like path is forbidden",
			in:          "//server/share/*.json",
			expectError: "path must be relative to repo",
		},
		{
			name:        "drive-prefixed relative path is forbidden",
			in:          "C:foo/*.json",
			expectError: "drive-prefixed",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeOptionalNamePattern(tt.in)
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

func TestNormalizeFileExtensions(t *testing.T) {
	type tc struct {
		name        string
		in          []string
		want        []string
		expectError string
	}

	cases := []tc{
		{
			name: "single extension",
			in:   []string{"json"},
			want: []string{"json"},
		},
		{
			name: "trims whitespace removes leading dot and lowercases",
			in:   []string{" .JSON "},
			want: []string{"json"},
		},
		{
			name: "multiple extensions preserve first-seen order",
			in:   []string{"yaml", "json", "xml"},
			want: []string{"yaml", "json", "xml"},
		},
		{
			name: "deduplicates after normalization",
			in:   []string{".JSON", " json ", "Json", ".yaml", "YAML"},
			want: []string{"json", "yaml"},
		},
		{
			name: "skips empty values after normalization",
			in:   []string{"", "   ", ".", "..json", " yml "},
			want: []string{"json", "yml"},
		},
		{
			name:        "leading dot only becomes empty and is skipped",
			in:          []string{"."},
			expectError: "no valid file extensions after normalization",
		},
		{
			name:        "errors when extensions have slashes",
			in:          []string{".jso/n"},
			expectError: "invalid file extension",
		},
		{
			name:        "all values empty after normalization",
			in:          []string{"", "   ", ".", " . "},
			expectError: "no valid file extensions after normalization",
		},
		{
			name:        "empty input slice returns error",
			in:          []string{},
			expectError: "no file extensions provided",
		},
		{
			name: "does not remove internal dots",
			in:   []string{"tar.gz", ".svg"},
			want: []string{"tar.gz", "svg"},
		},
		{
			name: "preserves first normalized occurrence",
			in:   []string{" .YML ", "json", "yml", ".JSON"},
			want: []string{"yml", "json"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeFileExtensions(tt.in)
			if tt.expectError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.expectError) {
					t.Fatalf("expected error containing %q, got %v", tt.expectError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
