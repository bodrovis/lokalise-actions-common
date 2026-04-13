package managedpaths

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type mockCaptureRunner struct {
	outputs map[string]string
	errors  map[string]error
}

func (m mockCaptureRunner) Capture(name string, args ...string) (string, error) {
	key := captureKey(name, args...)
	out, outOK := m.outputs[key]
	err, errOK := m.errors[key]

	if !outOK && !errOK {
		return "", fmt.Errorf("unexpected command: %s", key)
	}
	return out, err
}

func captureKey(name string, args ...string) string {
	return name + " " + strings.Join(args, " ")
}

func p(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}

func assertEqualStrings(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result:\n got: %#v\nwant: %#v", got, want)
	}
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want substring %q", err.Error(), want)
	}
}

func TestTranslationScope_ToTranslationFilesConfig(t *testing.T) {
	t.Parallel()

	scope := TranslationScope{
		Paths:          []string{"locales", "pkg/locales"},
		FileExt:        []string{"strings", "stringsdict"},
		FlatNaming:     false,
		AlwaysPullBase: true,
		BaseLang:       "en",
	}

	got := scope.ToTranslationFilesConfig()

	if !reflect.DeepEqual(got.TranslationPaths, scope.Paths) {
		t.Fatalf("TranslationPaths = %#v, want %#v", got.TranslationPaths, scope.Paths)
	}
	if !reflect.DeepEqual(got.FileExt, scope.FileExt) {
		t.Fatalf("FileExt = %#v, want %#v", got.FileExt, scope.FileExt)
	}
	if got.FlatNaming != scope.FlatNaming {
		t.Fatalf("FlatNaming = %v, want %v", got.FlatNaming, scope.FlatNaming)
	}
	if got.AlwaysPullBase != scope.AlwaysPullBase {
		t.Fatalf("AlwaysPullBase = %v, want %v", got.AlwaysPullBase, scope.AlwaysPullBase)
	}
	if got.BaseLang != scope.BaseLang {
		t.Fatalf("BaseLang = %q, want %q", got.BaseLang, scope.BaseLang)
	}
}

func TestFilterManaged(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		scope TranslationScope
		paths []string
		want  []string
	}{
		{
			name: "flat direct child matches and sorts deduplicated",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			paths: []string{
				"  " + p("locales", "fr.strings") + "  ",
				p("locales", "de.strings"),
				p("locales", "fr.strings"),
				p("other", "skip.strings"),
				"",
			},
			want: []string{
				p("locales", "de.strings"),
				p("locales", "fr.strings"),
			},
		},
		{
			name: "flat excludes nested child",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			paths: []string{
				p("locales", "nested", "fr.strings"),
				p("locales", "fr.strings"),
			},
			want: []string{
				p("locales", "fr.strings"),
			},
		},
		{
			name: "flat excludes base language when always pull base false",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings", "stringsdict"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			paths: []string{
				p("locales", "en.strings"),
				p("locales", "en.stringsdict"),
				p("locales", "fr.strings"),
				p("locales", "fr.stringsdict"),
			},
			want: []string{
				p("locales", "fr.strings"),
				p("locales", "fr.stringsdict"),
			},
		},
		{
			name: "flat includes base language when always pull base true",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: true,
				BaseLang:       "en",
			},
			paths: []string{
				p("locales", "en.strings"),
				p("locales", "fr.strings"),
			},
			want: []string{
				p("locales", "en.strings"),
				p("locales", "fr.strings"),
			},
		},
		{
			name: "nested excludes base language subtree",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings", "stringsdict"},
				FlatNaming:     false,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			paths: []string{
				p("locales", "en", "Localizable.strings"),
				p("locales", "en", "Plural.stringsdict"),
				p("locales", "fr", "Localizable.strings"),
				p("locales", "fr", "module", "Plural.stringsdict"),
			},
			want: []string{
				p("locales", "fr", "Localizable.strings"),
				p("locales", "fr", "module", "Plural.stringsdict"),
			},
		},
		{
			name: "nested includes base language subtree when always pull base true",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     false,
				AlwaysPullBase: true,
				BaseLang:       "en",
			},
			paths: []string{
				p("locales", "en", "Localizable.strings"),
				p("locales", "fr", "Localizable.strings"),
			},
			want: []string{
				p("locales", "en", "Localizable.strings"),
				p("locales", "fr", "Localizable.strings"),
			},
		},
		{
			name: "multiple roots and mixed extension coverage",
			scope: TranslationScope{
				Paths: []string{
					p("path", "to", "package", "localizables"),
					p("path", "to", "app", "supporting-files"),
				},
				FileExt:        []string{"strings", "stringsdict"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			paths: []string{
				p("path", "to", "package", "localizables", "fr.stringsdict"),
				p("path", "to", "app", "supporting-files", "fr.strings"),
				p("path", "to", "app", "supporting-files", "en.strings"),
				p("path", "to", "app", "supporting-files", "fr.json"),
			},
			want: []string{
				p("path", "to", "app", "supporting-files", "fr.strings"),
				p("path", "to", "package", "localizables", "fr.stringsdict"),
			},
		},
		{
			name: "outside similar prefix root is ignored",
			scope: TranslationScope{
				Paths:          []string{p("path", "to", "app")},
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			paths: []string{
				p("path", "to", "application", "fr.strings"),
				p("path", "to", "app", "fr.strings"),
			},
			want: []string{
				p("path", "to", "app", "fr.strings"),
			},
		},
		{
			name: "deleted style path still matches because filesystem is not consulted",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     false,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			paths: []string{
				p("locales", "fr", "DeletedButTracked.strings"),
			},
			want: []string{
				p("locales", "fr", "DeletedButTracked.strings"),
			},
		},
		{
			name: "unsupported extension is ignored",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     false,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			paths: []string{
				p("locales", "fr", "Localizable.json"),
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FilterManaged(tt.scope, tt.paths)
			assertEqualStrings(t, got, tt.want)
		})
	}
}

func TestCollectChangedTrackedPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		runner         mockCaptureRunner
		want           []string
		wantErrContain string
	}{
		{
			name: "uses HEAD diff when HEAD exists",
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"): "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): strings.Join([]string{
						"  " + p("locales", "fr.strings") + "  ",
						p("locales", "de.strings"),
						"",
					}, "\n"),
				},
			},
			want: []string{
				p("locales", "fr.strings"),
				p("locales", "de.strings"),
			},
		},
		{
			name: "falls back to cached and worktree when HEAD missing",
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"): "",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "--cached"): strings.Join([]string{
						p("locales", "fr.strings"),
						p("locales", "de.strings"),
					}, "\n"),
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only"): strings.Join([]string{
						p("locales", "de.strings"),
						p("locales", "it.strings"),
					}, "\n"),
				},
				errors: map[string]error{
					captureKey("git", "rev-parse", "--verify", "HEAD"): errors.New("no HEAD"),
				},
			},
			want: []string{
				p("locales", "de.strings"),
				p("locales", "fr.strings"),
				p("locales", "it.strings"),
			},
		},
		{
			name: "returns error when HEAD diff fails",
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                             "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): "boom",
				},
				errors: map[string]error{
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): errors.New("diff failed"),
				},
			},
			wantErrContain: "git diff HEAD failed",
		},
		{
			name: "returns error when cached diff fails in no HEAD mode",
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                                 "",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "--cached"): "cached fail",
				},
				errors: map[string]error{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                                 errors.New("no HEAD"),
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "--cached"): errors.New("cached failed"),
				},
			},
			wantErrContain: "git diff --cached (no HEAD) failed",
		},
		{
			name: "returns error when worktree diff fails in no HEAD mode",
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                                 "",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "--cached"): p("locales", "fr.strings"),
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only"):             "worktree fail",
				},
				errors: map[string]error{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                     errors.New("no HEAD"),
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only"): errors.New("worktree failed"),
				},
			},
			wantErrContain: "git diff (worktree) failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := collectChangedTrackedPaths(tt.runner)
			if tt.wantErrContain != "" {
				assertErrorContains(t, err, tt.wantErrContain)
				return
			}
			if err != nil {
				t.Fatalf("collectChangedTrackedPaths() error = %v", err)
			}
			assertEqualStrings(t, got, tt.want)
		})
	}
}

func TestCollectUntrackedPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		runner         mockCaptureRunner
		want           []string
		wantErrContain string
	}{
		{
			name: "success",
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): strings.Join([]string{
						"  " + p("locales", "fr.strings") + "  ",
						p("locales", "de.strings"),
					}, "\n"),
				},
			},
			want: []string{
				p("locales", "fr.strings"),
				p("locales", "de.strings"),
			},
		},
		{
			name: "error",
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): "ls-files fail",
				},
				errors: map[string]error{
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): errors.New("ls-files failed"),
				},
			},
			wantErrContain: "git ls-files failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := collectUntrackedPaths(tt.runner)
			if tt.wantErrContain != "" {
				assertErrorContains(t, err, tt.wantErrContain)
				return
			}
			if err != nil {
				t.Fatalf("collectUntrackedPaths() error = %v", err)
			}
			assertEqualStrings(t, got, tt.want)
		})
	}
}

func TestCollectManagedGitPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		scope          TranslationScope
		runner         mockCaptureRunner
		want           []string
		wantErrContain string
	}{
		{
			name: "mixed roots extensions base exclusion and deleted path",
			scope: TranslationScope{
				Paths: []string{
					p("path", "to", "package", "localizables"),
					p("path", "to", "app", "supporting-files"),
				},
				FileExt:        []string{"strings", "stringsdict"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"): "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): strings.Join([]string{
						p("path", "to", "package", "localizables", "fr.stringsdict"),
						p("path", "to", "app", "supporting-files", "fr.strings"),
						p("path", "to", "app", "supporting-files", "en.strings"),
						p("path", "to", "package", "localizables", "deleted.strings"),
						p("other", "place", "nope.json"),
					}, "\n"),
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): "",
				},
			},
			want: []string{
				p("path", "to", "app", "supporting-files", "fr.strings"),
				p("path", "to", "package", "localizables", "deleted.strings"),
				p("path", "to", "package", "localizables", "fr.stringsdict"),
			},
		},
		{
			name: "includes untracked managed files",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     false,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                             "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): "",
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): strings.Join([]string{
						p("locales", "fr", "Localizable.strings"),
						p("other", "ignore.strings"),
					}, "\n"),
				},
			},
			want: []string{
				p("locales", "fr", "Localizable.strings"),
			},
		},
		{
			name: "returns empty when no managed paths",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"): "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): strings.Join([]string{
						p("other", "file.strings"),
						p("locales", "nested", "fr.strings"),
						p("locales", "en.strings"),
					}, "\n"),
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): "",
				},
			},
			want: []string{},
		},
		{
			name: "propagates changed paths error",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                             "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): "boom",
				},
				errors: map[string]error{
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): errors.New("diff failed"),
				},
			},
			wantErrContain: "git diff HEAD failed",
		},
		{
			name: "propagates untracked paths error",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                                            "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"):                "",
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): "boom",
				},
				errors: map[string]error{
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): errors.New("ls-files failed"),
				},
			},
			wantErrContain: "git ls-files failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := CollectManagedGitPaths(tt.runner, tt.scope)
			if tt.wantErrContain != "" {
				assertErrorContains(t, err, tt.wantErrContain)
				return
			}
			if err != nil {
				t.Fatalf("CollectManagedGitPaths() error = %v", err)
			}
			assertEqualStrings(t, got, tt.want)
		})
	}
}

func TestHasManagedGitPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		scope          TranslationScope
		runner         mockCaptureRunner
		want           bool
		wantErrContain string
	}{
		{
			name: "true when managed paths exist",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     false,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                                            "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"):                p("locales", "fr", "Localizable.strings"),
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): "",
				},
			},
			want: true,
		},
		{
			name: "false when no managed paths exist",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                                            "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"):                p("other", "file.strings"),
					captureKey("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard"): "",
				},
			},
			want: false,
		},
		{
			name: "propagates errors",
			scope: TranslationScope{
				Paths:          []string{"locales"},
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			runner: mockCaptureRunner{
				outputs: map[string]string{
					captureKey("git", "rev-parse", "--verify", "HEAD"):                             "ok",
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): "boom",
				},
				errors: map[string]error{
					captureKey("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD"): errors.New("diff failed"),
				},
			},
			wantErrContain: "git diff HEAD failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := HasManagedGitPaths(tt.runner, tt.scope)
			if tt.wantErrContain != "" {
				assertErrorContains(t, err, tt.wantErrContain)
				return
			}
			if err != nil {
				t.Fatalf("HasManagedGitPaths() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("HasManagedGitPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	t.Parallel()

	got := normalize([]string{
		"  " + p("locales", "fr.strings") + "  ",
		"",
		"   ",
		p("locales", "de.strings"),
	})

	want := []string{
		p("locales", "fr.strings"),
		p("locales", "de.strings"),
	}

	assertEqualStrings(t, got, want)
}

func TestMergeAndNormalize(t *testing.T) {
	t.Parallel()

	got := mergeAndNormalize(
		[]string{
			"  " + p("locales", "fr.strings") + "  ",
			p("locales", "de.strings"),
		},
		[]string{
			p("locales", "fr.strings"),
			"",
			"   ",
			p("locales", "it.strings"),
		},
	)

	want := []string{
		p("locales", "de.strings"),
		p("locales", "fr.strings"),
		p("locales", "it.strings"),
	}

	assertEqualStrings(t, got, want)
}
