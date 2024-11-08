package githuboutput

import (
	"os"
	"testing"
)

func TestWriteToGitHubOutput(t *testing.T) {
	// Save the original GITHUB_OUTPUT value to restore it after tests
	originalGithubOutput := os.Getenv("GITHUB_OUTPUT")
	defer os.Setenv("GITHUB_OUTPUT", originalGithubOutput)

	tests := []struct {
		name                string
		envVarValue         string // Value to set GITHUB_OUTPUT to
		nameInput           string
		valueInput          string
		expectedReturn      bool
		expectedFileContent string
	}{
		{
			name:           "GITHUB_OUTPUT not set",
			envVarValue:    "",
			nameInput:      "key",
			valueInput:     "value",
			expectedReturn: false,
		},
		{
			name:                "GITHUB_OUTPUT set, write succeeds",
			envVarValue:         "tempfile",
			nameInput:           "key",
			valueInput:          "value",
			expectedReturn:      true,
			expectedFileContent: "key=value\n",
		},
		{
			name:           "GITHUB_OUTPUT set to invalid path",
			envVarValue:    "/invalid/path/to/file",
			nameInput:      "key",
			valueInput:     "value",
			expectedReturn: false,
		},
		{
			name:                "Write multiple times, data should append",
			envVarValue:         "tempfile",
			nameInput:           "key1",
			valueInput:          "value1",
			expectedReturn:      true,
			expectedFileContent: "key1=value1\nkey2=value2\n",
		},
		{
			name:                "Empty name and value",
			envVarValue:         "tempfile",
			nameInput:           "",
			valueInput:          "",
			expectedReturn:      true,
			expectedFileContent: "=\n",
		},
		{
			name:                "Name and value with special characters",
			envVarValue:         "tempfile",
			nameInput:           "special_key!@#$",
			valueInput:          "special_value%^&*",
			expectedReturn:      true,
			expectedFileContent: "special_key!@#$=special_value%^&*\n",
		},
		{
			name:                "Value contains newline (unsupported scenario)",
			envVarValue:         "tempfile",
			nameInput:           "key",
			valueInput:          "value\nwithnewline",
			expectedReturn:      true,
			expectedFileContent: "key=value\nwithnewline\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare the environment
			if tt.envVarValue != "" {
				if tt.envVarValue == "tempfile" {
					// Create a temporary file to act as GITHUB_OUTPUT
					tempFile, err := os.CreateTemp("", "github_output_test")
					if err != nil {
						t.Fatalf("Failed to create temporary file: %v", err)
					}
					defer os.Remove(tempFile.Name()) // Clean up after test
					os.Setenv("GITHUB_OUTPUT", tempFile.Name())
					defer os.Unsetenv("GITHUB_OUTPUT")

					if tt.name == "Write multiple times, data should append" {
						// First write
						result := WriteToGitHubOutput("key1", "value1")
						if result != true {
							t.Errorf("Expected WriteToGitHubOutput to return true, got %v", result)
						}
						// Second write
						result = WriteToGitHubOutput("key2", "value2")
						if result != true {
							t.Errorf("Expected WriteToGitHubOutput to return true, got %v", result)
						}
					} else {
						// Single write
						result := WriteToGitHubOutput(tt.nameInput, tt.valueInput)
						if result != tt.expectedReturn {
							t.Errorf("Expected WriteToGitHubOutput to return %v, got %v", tt.expectedReturn, result)
						}
					}

					// Verify file content
					contentBytes, err := os.ReadFile(tempFile.Name())
					if err != nil {
						t.Fatalf("Failed to read temporary file: %v", err)
					}
					content := string(contentBytes)
					if content != tt.expectedFileContent {
						t.Errorf("File content mismatch.\nExpected:\n%q\nGot:\n%q", tt.expectedFileContent, content)
					}
				} else {
					// GITHUB_OUTPUT is set to an invalid path
					os.Setenv("GITHUB_OUTPUT", tt.envVarValue)
					defer os.Unsetenv("GITHUB_OUTPUT")
					result := WriteToGitHubOutput(tt.nameInput, tt.valueInput)
					if result != tt.expectedReturn {
						t.Errorf("Expected WriteToGitHubOutput to return %v, got %v", tt.expectedReturn, result)
					}

					// Ensure that the file does not exist
					_, err := os.Stat(tt.envVarValue)
					if !os.IsNotExist(err) {
						t.Errorf("Expected file %v to not exist, but it does", tt.envVarValue)
					}
				}
			} else {
				// GITHUB_OUTPUT is not set
				os.Unsetenv("GITHUB_OUTPUT")
				result := WriteToGitHubOutput(tt.nameInput, tt.valueInput)
				if result != tt.expectedReturn {
					t.Errorf("Expected WriteToGitHubOutput to return %v, got %v", tt.expectedReturn, result)
				}
			}
		})
	}
}
