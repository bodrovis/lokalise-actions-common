package githuboutput

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteToGitHubOutput(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (outputPath string, verify bool)
		write        func(t *testing.T)
		wantReturn   bool
		wantFile     bool
		wantFileBody string
	}{
		{
			name: "GITHUB_OUTPUT not set",
			setup: func(t *testing.T) (string, bool) {
				// Ensure truly unset even on GitHub runner
				_ = os.Unsetenv("GITHUB_OUTPUT")
				return "", false
			},
			write: func(t *testing.T) {
				got := WriteToGitHubOutput("key", "value")
				if got != false {
					t.Fatalf("expected false, got %v", got)
				}
			},
			wantReturn: false,
			wantFile:   false,
		},
		{
			name: "GITHUB_OUTPUT set, write succeeds",
			setup: func(t *testing.T) (string, bool) {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path, true
			},
			write: func(t *testing.T) {
				got := WriteToGitHubOutput("key", "value")
				if got != true {
					t.Fatalf("expected true, got %v", got)
				}
			},
			wantReturn:   true,
			wantFile:     true,
			wantFileBody: "key=value\n",
		},
		{
			name: "GITHUB_OUTPUT set to invalid path",
			setup: func(t *testing.T) (string, bool) {
				// Use a path that should be invalid on all OSes:
				// a file inside a non-existent directory.
				p := filepath.Join(os.TempDir(), "definitely-not-exist-12345", "out.txt")
				t.Setenv("GITHUB_OUTPUT", p)
				return p, false
			},
			write: func(t *testing.T) {
				got := WriteToGitHubOutput("key", "value")
				if got != false {
					t.Fatalf("expected false, got %v", got)
				}
			},
			wantReturn: false,
			wantFile:   false,
		},
		{
			name: "Write multiple times, data should append",
			setup: func(t *testing.T) (string, bool) {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path, true
			},
			write: func(t *testing.T) {
				if got := WriteToGitHubOutput("key1", "value1"); got != true {
					t.Fatalf("first write: expected true, got %v", got)
				}
				if got := WriteToGitHubOutput("key2", "value2"); got != true {
					t.Fatalf("second write: expected true, got %v", got)
				}
			},
			wantReturn:   true,
			wantFile:     true,
			wantFileBody: "key1=value1\nkey2=value2\n",
		},
		{
			name: "Empty name and value",
			setup: func(t *testing.T) (string, bool) {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path, true
			},
			write: func(t *testing.T) {
				got := WriteToGitHubOutput("", "")
				if got != true {
					t.Fatalf("expected true, got %v", got)
				}
			},
			wantReturn:   true,
			wantFile:     true,
			wantFileBody: "=\n",
		},
		{
			name: "Name and value with special characters",
			setup: func(t *testing.T) (string, bool) {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path, true
			},
			write: func(t *testing.T) {
				got := WriteToGitHubOutput("special_key!@#$", "special_value%^&*")
				if got != true {
					t.Fatalf("expected true, got %v", got)
				}
			},
			wantReturn:   true,
			wantFile:     true,
			wantFileBody: "special_key!@#$=special_value%^&*\n",
		},
		{
			name: "Value contains newline (unsupported scenario)",
			setup: func(t *testing.T) (string, bool) {
				f, err := os.CreateTemp("", "github_output_test")
				if err != nil {
					t.Fatalf("CreateTemp: %v", err)
				}
				path := f.Name()
				_ = f.Close()

				t.Cleanup(func() { _ = os.Remove(path) })
				t.Setenv("GITHUB_OUTPUT", path)
				return path, true
			},
			write: func(t *testing.T) {
				got := WriteToGitHubOutput("key", "value\nwithnewline")
				if got != true {
					t.Fatalf("expected true, got %v", got)
				}
			},
			wantReturn:   true,
			wantFile:     true,
			wantFileBody: "key=value\nwithnewline\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, verify := tt.setup(t)

			tt.write(t)

			if !verify {
				// nothing to verify on disk for "unset" or invalid path cases
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
