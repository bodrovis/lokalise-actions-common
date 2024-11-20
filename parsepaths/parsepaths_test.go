package parsepaths

import (
	"reflect"
	"strings"
	"testing"
)

func TestParsePaths(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		expected []string
	}{
		{
			name:     "Single path, no newline",
			envVar:   "/path/to/dir",
			expected: []string{"/path/to/dir"},
		},
		{
			name:     "Multiple paths, Unix newlines",
			envVar:   "/path/to/dir\n/another/path\n/yet/another/path",
			expected: []string{"/path/to/dir", "/another/path", "/yet/another/path"},
		},
		{
			name:     "Multiple paths, Windows newlines",
			envVar:   "/path/to/dir\r\n/another/path\r\n/yet/another/path",
			expected: []string{"/path/to/dir", "/another/path", "/yet/another/path"},
		},
		{
			name:     "Empty lines and whitespace",
			envVar:   "/path/to/dir\n\n \n/another/path\n   \n/yet/another/path\n",
			expected: []string{"/path/to/dir", "/another/path", "/yet/another/path"},
		},
		{
			name:     "Empty lines and whitespace plus whitespaces inside",
			envVar:   "/path/to/dir\n\n \n/another/path\n   \n/yet/another/path\n/path/to my/dir",
			expected: []string{"/path/to/dir", "/another/path", "/yet/another/path", "/path/to my/dir"},
		},
		{
			name:     "Only empty lines and whitespace",
			envVar:   "\n\n \n  \n",
			expected: []string{},
		},
		{
			name:     "Empty string input",
			envVar:   "",
			expected: []string{},
		},
		{
			name:     "Mixed newline types in input",
			envVar:   "/path/one\r\n/path/two\n/path/three\r/path/four",
			expected: []string{"/path/one", "/path/two", "/path/three\r/path/four"},
		},
		{
			name:   "Very long paths",
			envVar: "/very/long/path/" + strings.Repeat("subdir/", 100) + "\n/another/very/long/path/" + strings.Repeat("subdir/", 100),
			expected: []string{
				"/very/long/path/" + strings.Repeat("subdir/", 100),
				"/another/very/long/path/" + strings.Repeat("subdir/", 100),
			},
		},
		{
			name:   "Very long paths",
			envVar: "/very/long/path/" + strings.Repeat("subdir/", 100) + "\n/another/very/long/path/" + strings.Repeat("subdir/", 100),
			expected: []string{
				"/very/long/path/" + strings.Repeat("subdir/", 100),
				"/another/very/long/path/" + strings.Repeat("subdir/", 100),
			},
		},
		{
			name:     "Paths with internal whitespace and tabs",
			envVar:   "/path/with space\n/path/with\ttab\n/path/with multiple   spaces",
			expected: []string{"/path/with space", "/path/with\ttab", "/path/with multiple   spaces"},
		},
		{
			name:     "Input with only whitespace",
			envVar:   "   \t  ",
			expected: []string{},
		},
		{
			name:     "Input with null characters",
			envVar:   "/path/one\000/path/two",
			expected: []string{"/path/one\000/path/two"},
		},
		{
			name:     "Paths with special characters",
			envVar:   "/path/with/special!@#$%^&*()_+\n/path/with/[]{}|;:',.<>?",
			expected: []string{"/path/with/special!@#$%^&*()_+", "/path/with/[]{}|;:',.<>?"},
		},
		{
			name:     "Input ending with newline",
			envVar:   "/path/one\n/path/two\n",
			expected: []string{"/path/one", "/path/two"},
		},
		{
			name:     "Input starting with newline",
			envVar:   "\n/path/one\n/path/two",
			expected: []string{"/path/one", "/path/two"},
		},
		{
			name:     "Combination of empty lines and paths",
			envVar:   "\n\n/path/one\n\n\n/path/two\n\n",
			expected: []string{"/path/one", "/path/two"},
		},
		{
			name:     "Paths with leading/trailing tabs and spaces",
			envVar:   "\t /path/with/leading/tab \n/path/with/trailing/space \t\n\t /path/with/both \t",
			expected: []string{"/path/with/leading/tab", "/path/with/trailing/space", "/path/with/both"},
		},
		{
			name:     "Paths with Unicode characters",
			envVar:   "/path/with/üñîçødê\n/路径/含/非ASCII字符\n/путь/с/юникодом",
			expected: []string{"/path/with/üñîçødê", "/路径/含/非ASCII字符", "/путь/с/юникодом"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ParsePaths(tt.envVar)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParsePaths() = %v, want %v", result, tt.expected)
			}
		})
	}
}
