package githuboutput

import (
	"os"
	"path/filepath"
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
			name: "GITHUB_OUTPUT not set",
			setupEnv: func(t *testing.T) string {
				_ = os.Unsetenv("GITHUB_OUTPUT")
				return ""
			},
			outputName:   "key",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name: "GITHUB_OUTPUT set, write succeeds",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "key",
			outputValue:  "value",
			wantOK:       true,
			wantFileBody: "key=value\n",
		},
		{
			name: "GITHUB_OUTPUT set to invalid path",
			setupEnv: func(t *testing.T) string {
				p := filepath.Join(os.TempDir(), "definitely-not-exist-12345", "out.txt")
				t.Setenv("GITHUB_OUTPUT", p)
				return ""
			},
			outputName:   "key",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name: "Write multiple times appends",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)

				if got := WriteToGitHubOutput("key1", "value1"); !got {
					t.Fatalf("first write failed")
				}
				return path
			},
			outputName:   "key2",
			outputValue:  "value2",
			wantOK:       true,
			wantFileBody: "key1=value1\nkey2=value2\n",
		},
		{
			name: "Empty name is rejected",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name: "Whitespace-only name is rejected",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "   \t  ",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name: "Trimmed name is accepted",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "  key  ",
			outputValue:  "value",
			wantOK:       true,
			wantFileBody: "key=value\n",
		},
		{
			name: "Empty value is allowed",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "key",
			outputValue:  "",
			wantOK:       true,
			wantFileBody: "key=\n",
		},
		{
			name: "Name with special characters is allowed",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "special_key!@#$",
			outputValue:  "special_value%^&*",
			wantOK:       true,
			wantFileBody: "special_key!@#$=special_value%^&*\n",
		},
		{
			name: "Name containing equals is rejected",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "bad=key",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name: "Name containing newline is rejected",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "bad\nkey",
			outputValue:  "value",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name: "Value containing newline is rejected",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "key",
			outputValue:  "value\nwithnewline",
			wantOK:       false,
			wantFileBody: "",
		},
		{
			name: "Value containing carriage return is rejected",
			setupEnv: func(t *testing.T) string {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path
			},
			outputName:   "key",
			outputValue:  "value\rbroken",
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

			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile(%s): %v", path, err)
			}
			if string(b) != tt.wantFileBody {
				t.Fatalf("file content mismatch.\nwant:\n%q\ngot:\n%q", tt.wantFileBody, string(b))
			}
		})
	}
}
