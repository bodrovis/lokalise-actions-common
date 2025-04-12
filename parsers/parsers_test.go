package parsers

import (
	"log"
	"os"
	"reflect"
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
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable for the test
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

func normalizeSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
