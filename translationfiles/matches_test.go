package translationfiles

import (
	"path/filepath"
	"testing"
)

func TestMatches(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		path string
		want bool
	}{
		{
			name: "empty path",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: "",
			want: false,
		},
		{
			name: "path equal to translation root with valid extension does not match",
			cfg: Config{
				TranslationPaths: []string{"locales.strings"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: "locales.strings",
			want: false,
		},
		{
			name: "no allowed extensions",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"", ".", "   "},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "fr.strings"),
			want: false,
		},
		{
			name: "unsupported extension",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "fr.json"),
			want: false,
		},
		{
			name: "path outside translation root",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("other", "fr.strings"),
			want: false,
		},
		{
			name: "flat direct child matches",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "fr.strings"),
			want: true,
		},
		{
			name: "flat nested child does not match",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "nested", "fr.strings"),
			want: false,
		},
		{
			name: "flat excludes base language file",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings", "stringsdict"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "en.strings"),
			want: false,
		},
		{
			name: "flat includes base language file when always pull base",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   true,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "en.strings"),
			want: true,
		},
		{
			name: "flat second extension matches",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings", "stringsdict"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "fr.stringsdict"),
			want: true,
		},
		{
			name: "flat excludes base language for second extension",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings", "stringsdict"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "en.stringsdict"),
			want: false,
		},
		{
			name: "nested file matches",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       false,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "fr", "Localizable.strings"),
			want: true,
		},
		{
			name: "nested deeper file matches",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       false,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "fr", "module", "Localizable.strings"),
			want: true,
		},
		{
			name: "nested excludes base language directory",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings", "stringsdict"},
				FlatNaming:       false,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "en", "Localizable.strings"),
			want: false,
		},
		{
			name: "nested excludes deeper file under base language directory",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       false,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "en", "module", "Localizable.strings"),
			want: false,
		},
		{
			name: "nested includes base language directory when always pull base",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       false,
				AlwaysPullBase:   true,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "en", "Localizable.strings"),
			want: true,
		},
		{
			name: "multiple roots matches first root",
			cfg: Config{
				TranslationPaths: []string{
					filepath.Join("path", "to", "package", "localizables"),
					filepath.Join("path", "to", "app", "supporting-files"),
				},
				FileExt:        []string{"strings", "stringsdict"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			path: filepath.Join("path", "to", "package", "localizables", "fr.stringsdict"),
			want: true,
		},
		{
			name: "multiple roots matches second root even if first root would not",
			cfg: Config{
				TranslationPaths: []string{
					filepath.Join("path", "to", "package", "localizables"),
					filepath.Join("path", "to", "app", "supporting-files"),
				},
				FileExt:        []string{"strings", "stringsdict"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			path: filepath.Join("path", "to", "app", "supporting-files", "fr.strings"),
			want: true,
		},
		{
			name: "mixed coverage case no stringsdict in app still matches strings file",
			cfg: Config{
				TranslationPaths: []string{
					filepath.Join("path", "to", "package", "localizables"),
					filepath.Join("path", "to", "app", "supporting-files"),
				},
				FileExt:        []string{"strings", "stringsdict"},
				FlatNaming:     true,
				AlwaysPullBase: false,
				BaseLang:       "en",
			},
			path: filepath.Join("path", "to", "app", "supporting-files", "fr.strings"),
			want: true,
		},
		{
			name: "deleted file style path still matches because filesystem is not consulted",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       false,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "fr", "DeletedButTracked.strings"),
			want: true,
		},
		{
			name: "extension matching is case insensitive",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"STRINGS"},
				FlatNaming:       false,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "fr", "Localizable.STRINGS"),
			want: true,
		},
		{
			name: "empty root entries are ignored",
			cfg: Config{
				TranslationPaths: []string{"", "locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: filepath.Join("locales", "fr.strings"),
			want: true,
		},
		{
			name: "path equal to root does not match",
			cfg: Config{
				TranslationPaths: []string{"locales"},
				FileExt:          []string{"strings"},
				FlatNaming:       true,
				AlwaysPullBase:   false,
				BaseLang:         "en",
			},
			path: "locales",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Matches(tt.cfg, tt.path)
			if got != tt.want {
				t.Fatalf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelativePathWithinRoot(t *testing.T) {
	tests := []struct {
		name    string
		root    string
		path    string
		wantRel string
		wantOK  bool
	}{
		{
			name:    "path inside root",
			root:    filepath.Join("locales"),
			path:    filepath.Join("locales", "fr", "Localizable.strings"),
			wantRel: filepath.Join("fr", "Localizable.strings"),
			wantOK:  true,
		},
		{
			name:    "path equal to root",
			root:    filepath.Join("locales"),
			path:    filepath.Join("locales"),
			wantRel: ".",
			wantOK:  true,
		},
		{
			name:   "path outside root",
			root:   filepath.Join("locales"),
			path:   filepath.Join("other", "fr.strings"),
			wantOK: false,
		},
		{
			name:    "root and path are cleaned",
			root:    filepath.Join("locales", "."),
			path:    filepath.Join("locales", "fr", "..", "fr", "Localizable.strings"),
			wantRel: filepath.Join("fr", "Localizable.strings"),
			wantOK:  true,
		},
		{
			name:    "whitespace is trimmed",
			root:    "  locales  ",
			path:    "  " + filepath.Join("locales", "fr.strings") + "  ",
			wantRel: "fr.strings",
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRel, gotOK := relativePathWithinRoot(tt.root, tt.path)
			if gotOK != tt.wantOK {
				t.Fatalf("relativePathWithinRoot() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotRel != tt.wantRel {
				t.Fatalf("relativePathWithinRoot() rel = %q, want %q", gotRel, tt.wantRel)
			}
		})
	}
}

func TestBuildAllowedExts(t *testing.T) {
	got := buildAllowedExts([]string{"strings", ".stringsdict", " STRINGS ", "", ".", "   "})

	if _, ok := got["strings"]; !ok {
		t.Fatal(`buildAllowedExts() missing "strings"`)
	}
	if _, ok := got["stringsdict"]; !ok {
		t.Fatal(`buildAllowedExts() missing "stringsdict"`)
	}
	if len(got) != 2 {
		t.Fatalf("buildAllowedExts() len = %d, want 2", len(got))
	}
}
