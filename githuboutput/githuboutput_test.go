package githuboutput

import (
	"os"
	"testing"
)

func TestWriteToGitHubOutput(t *testing.T) {
	tests := []struct {
		name         string
		setupEnv     func(t *testing.T) string
		outputName   string
		outputValue  string
		wantOK       bool
		wantFileBody string
	}{
		{
			name:     "GITHUB_OUTPUT not set",
			setupEnv: setupGitHubOutput,
		},
		{
			name:         "GITHUB_OUTPUT set, write succeeds",
			setupEnv:     setupGitHubOutput,
			outputName:   "key",
			outputValue:  "value",
			wantOK:       true,
			wantFileBody: "key=value\n",
		},
		{
			name:     "GITHUB_OUTPUT set to invalid path",
			setupEnv: setupGitHubOutput,
		},
		{
			name: "Write multiple times appends",
			setupEnv: func(t *testing.T) string {
				path := setupGitHubOutput(t)

				if ok := WriteToGitHubOutput("key1", "value1"); !ok {
					t.Fatal("first write failed")
				}

				return path
			},
			outputName:   "key2",
			outputValue:  "value2",
			wantOK:       true,
			wantFileBody: "key1=value1\nkey2=value2\n",
		},
		{
			name:         "Empty name is rejected",
			setupEnv:     setupGitHubOutput,
			outputName:   "",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name:         "Whitespace-only name is rejected",
			setupEnv:     setupGitHubOutput,
			outputName:   "   \t  ",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name:         "Trimmed name is accepted",
			setupEnv:     setupGitHubOutput,
			outputName:   "  key  ",
			outputValue:  "value",
			wantOK:       true,
			wantFileBody: "key=value\n",
		},
		{
			name:         "Empty value is allowed",
			setupEnv:     setupGitHubOutput,
			outputName:   "key",
			outputValue:  "",
			wantOK:       true,
			wantFileBody: "key=\n",
		},
		{
			name:         "Name with special characters is allowed",
			setupEnv:     setupGitHubOutput,
			outputName:   "special_key!@#$",
			outputValue:  "special_value%^&*",
			wantOK:       true,
			wantFileBody: "special_key!@#$=special_value%^&*\n",
		},
		{
			name:         "Name containing equals is rejected",
			setupEnv:     setupGitHubOutput,
			outputName:   "bad=key",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name:         "Name containing newline is rejected",
			setupEnv:     setupGitHubOutput,
			outputName:   "bad\nkey",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name:         "Value containing newline is rejected",
			setupEnv:     setupGitHubOutput,
			outputName:   "key",
			outputValue:  "value\nwithnewline",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name:         "Value containing carriage return is rejected",
			setupEnv:     setupGitHubOutput,
			outputName:   "key",
			outputValue:  "value\rbroken",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name:         "Name containing carriage return is rejected",
			setupEnv:     setupGitHubOutput,
			outputName:   "bad\rkey",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupEnv(t)

			got := WriteToGitHubOutput(tt.outputName, tt.outputValue)
			if got != tt.wantOK {
				t.Fatalf("WriteToGitHubOutput(%q, %q) = %v, want %v", tt.outputName, tt.outputValue, got, tt.wantOK)
			}

			if path == "" {
				return
			}

			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile(%s): %v", path, err)
			}

			if string(body) != tt.wantFileBody {
				t.Fatalf(
					"file content mismatch.\nwant:\n%q\ngot:\n%q",
					tt.wantFileBody,
					string(body),
				)
			}
		})
	}
}

func setupGitHubOutput(t *testing.T) string {
	t.Helper()

	f, err := os.CreateTemp("", "github_output_test")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}

	path := f.Name()

	if err := f.Close(); err != nil {
		t.Fatalf("Close temp file: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(path)
	})

	t.Setenv("GITHUB_OUTPUT", path)

	return path
}
