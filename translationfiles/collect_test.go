package translationfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCollect_Flat_SingleExt_ExcludeBase(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "locales"))

	writeFile(t, filepath.Join(root, "locales", "en.strings"), "base")
	writeFile(t, filepath.Join(root, "locales", "fr.strings"), "fr")
	writeFile(t, filepath.Join(root, "locales", "de.strings"), "de")
	writeFile(t, filepath.Join(root, "locales", "README.md"), "ignore")
	mkdirAll(t, filepath.Join(root, "locales", "nested"))
	writeFile(t, filepath.Join(root, "locales", "nested", "it.strings"), "ignore nested in flat")

	got, err := Collect(Config{
		TranslationPaths: []string{filepath.Join(root, "locales")},
		FileExt:          []string{"strings"},
		FlatNaming:       true,
		AlwaysPullBase:   false,
		BaseLang:         "en",
	})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	want := []string{
		toSlash(filepath.Join(root, "locales", "de.strings")),
		toSlash(filepath.Join(root, "locales", "fr.strings")),
	}

	assertEqualStrings(t, got, want)
}

func TestCollect_Flat_MultipleExt_ExcludeBase(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "pkg"))

	writeFile(t, filepath.Join(root, "pkg", "en.strings"), "base")
	writeFile(t, filepath.Join(root, "pkg", "en.stringsdict"), "base dict")
	writeFile(t, filepath.Join(root, "pkg", "fr.strings"), "fr")
	writeFile(t, filepath.Join(root, "pkg", "fr.stringsdict"), "fr dict")
	writeFile(t, filepath.Join(root, "pkg", "es.strings"), "es")

	got, err := Collect(Config{
		TranslationPaths: []string{filepath.Join(root, "pkg")},
		FileExt:          []string{"strings", "stringsdict"},
		FlatNaming:       true,
		AlwaysPullBase:   false,
		BaseLang:         "en",
	})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	want := []string{
		toSlash(filepath.Join(root, "pkg", "es.strings")),
		toSlash(filepath.Join(root, "pkg", "fr.strings")),
		toSlash(filepath.Join(root, "pkg", "fr.stringsdict")),
	}

	assertEqualStrings(t, got, want)
}

func TestCollect_Flat_IncludeBase(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "locales"))

	writeFile(t, filepath.Join(root, "locales", "en.strings"), "base")
	writeFile(t, filepath.Join(root, "locales", "fr.strings"), "fr")

	got, err := Collect(Config{
		TranslationPaths: []string{filepath.Join(root, "locales")},
		FileExt:          []string{"strings"},
		FlatNaming:       true,
		AlwaysPullBase:   true,
		BaseLang:         "en",
	})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	want := []string{
		toSlash(filepath.Join(root, "locales", "en.strings")),
		toSlash(filepath.Join(root, "locales", "fr.strings")),
	}

	assertEqualStrings(t, got, want)
}

func TestCollect_Nested_SingleExt_ExcludeBaseDir(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "locales", "en"))
	mkdirAll(t, filepath.Join(root, "locales", "fr"))
	mkdirAll(t, filepath.Join(root, "locales", "de"))

	writeFile(t, filepath.Join(root, "locales", "en", "app.strings"), "base")
	writeFile(t, filepath.Join(root, "locales", "fr", "app.strings"), "fr")
	writeFile(t, filepath.Join(root, "locales", "de", "app.strings"), "de")
	writeFile(t, filepath.Join(root, "locales", "fr", "ignore.txt"), "ignore")

	got, err := Collect(Config{
		TranslationPaths: []string{filepath.Join(root, "locales")},
		FileExt:          []string{"strings"},
		FlatNaming:       false,
		AlwaysPullBase:   false,
		BaseLang:         "en",
	})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	want := []string{
		toSlash(filepath.Join(root, "locales", "de", "app.strings")),
		toSlash(filepath.Join(root, "locales", "fr", "app.strings")),
	}

	assertEqualStrings(t, got, want)
}

func TestCollect_Nested_MultipleExt_IncludeBaseDir(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "pkg", "en"))
	mkdirAll(t, filepath.Join(root, "pkg", "fr"))

	writeFile(t, filepath.Join(root, "pkg", "en", "Localizable.strings"), "base")
	writeFile(t, filepath.Join(root, "pkg", "en", "Plural.stringsdict"), "base dict")
	writeFile(t, filepath.Join(root, "pkg", "fr", "Localizable.strings"), "fr")
	writeFile(t, filepath.Join(root, "pkg", "fr", "Plural.stringsdict"), "fr dict")

	got, err := Collect(Config{
		TranslationPaths: []string{filepath.Join(root, "pkg")},
		FileExt:          []string{"strings", "stringsdict"},
		FlatNaming:       false,
		AlwaysPullBase:   true,
		BaseLang:         "en",
	})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	want := []string{
		toSlash(filepath.Join(root, "pkg", "en", "Localizable.strings")),
		toSlash(filepath.Join(root, "pkg", "en", "Plural.stringsdict")),
		toSlash(filepath.Join(root, "pkg", "fr", "Localizable.strings")),
		toSlash(filepath.Join(root, "pkg", "fr", "Plural.stringsdict")),
	}

	assertEqualStrings(t, got, want)
}

func TestCollect_MultiplePaths_DifferentExtensionCoverage(t *testing.T) {
	root := t.TempDir()

	pkg := filepath.Join(root, "path", "to", "package", "localizables")
	app := filepath.Join(root, "path", "to", "app", "supporting-files")
	mkdirAll(t, pkg)
	mkdirAll(t, app)

	writeFile(t, filepath.Join(pkg, "en.strings"), "base")
	writeFile(t, filepath.Join(pkg, "en.stringsdict"), "base dict")
	writeFile(t, filepath.Join(pkg, "fr.strings"), "fr")
	writeFile(t, filepath.Join(pkg, "fr.stringsdict"), "fr dict")

	writeFile(t, filepath.Join(app, "en.strings"), "base")
	writeFile(t, filepath.Join(app, "fr.strings"), "fr")
	// app has no .stringsdict on purpose

	got, err := Collect(Config{
		TranslationPaths: []string{pkg, app},
		FileExt:          []string{"strings", "stringsdict"},
		FlatNaming:       true,
		AlwaysPullBase:   false,
		BaseLang:         "en",
	})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	want := []string{
		toSlash(filepath.Join(app, "fr.strings")),
		toSlash(filepath.Join(pkg, "fr.strings")),
		toSlash(filepath.Join(pkg, "fr.stringsdict")),
	}

	assertEqualStrings(t, got, want)
}

func TestCollect_PathExists_ButSomeExtensionsMissing_IsOK(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "app"))

	writeFile(t, filepath.Join(root, "app", "fr.strings"), "fr")

	got, err := Collect(Config{
		TranslationPaths: []string{filepath.Join(root, "app")},
		FileExt:          []string{"strings", "stringsdict"},
		FlatNaming:       true,
		AlwaysPullBase:   false,
		BaseLang:         "en",
	})
	if err != nil {
		t.Fatalf("Collect() unexpected error = %v", err)
	}

	want := []string{
		toSlash(filepath.Join(root, "app", "fr.strings")),
	}

	assertEqualStrings(t, got, want)
}

func TestCollect_EmptyExistingDirectory_IsOK(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "empty"))

	got, err := Collect(Config{
		TranslationPaths: []string{filepath.Join(root, "empty")},
		FileExt:          []string{"strings"},
		FlatNaming:       true,
		AlwaysPullBase:   false,
		BaseLang:         "en",
	})
	if err != nil {
		t.Fatalf("Collect() unexpected error = %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("Collect() = %v, want empty result", got)
	}
}

func TestCollect_MissingPath_ReturnsError(t *testing.T) {
	root := t.TempDir()

	_, err := Collect(Config{
		TranslationPaths: []string{filepath.Join(root, "missing")},
		FileExt:          []string{"strings"},
		FlatNaming:       true,
		AlwaysPullBase:   false,
		BaseLang:         "en",
	})
	if err == nil {
		t.Fatal("Collect() error = nil, want error")
	}
}

func TestCollect_PathIsFile_ReturnsError(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "not-a-dir")
	writeFile(t, filePath, "hello")

	_, err := Collect(Config{
		TranslationPaths: []string{filePath},
		FileExt:          []string{"strings"},
		FlatNaming:       true,
		AlwaysPullBase:   false,
		BaseLang:         "en",
	})
	if err == nil {
		t.Fatal("Collect() error = nil, want error")
	}
}

func TestCollect_DeduplicatesAndSortsPaths(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "locales")
	mkdirAll(t, dir)

	writeFile(t, filepath.Join(dir, "z.strings"), "z")
	writeFile(t, filepath.Join(dir, "a.strings"), "a")

	got, err := Collect(Config{
		TranslationPaths: []string{dir, dir},
		FileExt:          []string{"strings", "strings"},
		FlatNaming:       true,
		AlwaysPullBase:   true,
		BaseLang:         "en",
	})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	want := []string{
		toSlash(filepath.Join(dir, "a.strings")),
		toSlash(filepath.Join(dir, "z.strings")),
	}

	assertEqualStrings(t, got, want)
}

func TestCollect_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "empty paths",
			cfg: Config{
				FileExt:        []string{"strings"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			wantErr: "translation paths are required",
		},
		{
			name: "empty file ext",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			wantErr: "file extensions are required",
		},
		{
			name: "empty base lang",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
			},
			wantErr: "base language is required",
		},
		{
			name: "base lang contains slash",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en/us",
			},
			wantErr: "base language must not contain path separators",
		},
		{
			name: "base lang contains backslash",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         `en\us`,
			},
			wantErr: "base language must not contain path separators",
		},
		{
			name: "only invalid file ext values",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"", ".", "   "},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			wantErr: "no valid file extensions after normalization",
		},
		{
			name: "invalid ext with slash",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"str/ings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			wantErr: fmt.Sprintf("invalid file extension %q", "str/ings"),
		},
		{
			name: "invalid ext with backslash",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{`str\ings`},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			wantErr: fmt.Sprintf("invalid file extension %q", `str\ings`),
		},
		{
			name: "empty translation path item",
			cfg: Config{
				TranslationPaths: []string{"", "locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			wantErr: "translation path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Collect(tt.cfg)
			if err == nil {
				t.Fatal("Collect() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Collect() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %q: %v", path, err)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create dir %q: %v", path, err)
	}
}

func assertEqualStrings(t *testing.T, got, want []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected result:\n got: %#v\nwant: %#v", got, want)
	}
}

func toSlash(path string) string {
	return filepath.ToSlash(path)
}
