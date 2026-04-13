package translationfiles

import (
	"path/filepath"
	"strings"
)

type Config struct {
	TranslationPaths []string
	FileExt          []string
	FlatNaming       bool
	AlwaysPullBase   bool
	BaseLang         string
}

func Matches(cfg Config, path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}

	allowedExts := buildAllowedExts(cfg.FileExt)
	if len(allowedExts) == 0 {
		return false
	}

	cleanPath, ext, ok := matchesAllowedExtension(path, allowedExts)
	if !ok {
		return false
	}

	return matchesAnyTranslationRoot(cfg, cleanPath, ext)
}

func matchesAllowedExtension(path string, allowedExts map[string]struct{}) (string, string, bool) {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	ext := normalizeExt(filepath.Ext(cleanPath))
	if ext == "" {
		return "", "", false
	}

	if _, ok := allowedExts[ext]; !ok {
		return "", "", false
	}

	return cleanPath, ext, true
}

func matchesAnyTranslationRoot(cfg Config, cleanPath, ext string) bool {
	baseName := filepath.Base(cleanPath)

	for _, rawRoot := range cfg.TranslationPaths {
		if matchesTranslationRoot(cfg, rawRoot, cleanPath, baseName, ext) {
			return true
		}
	}

	return false
}

func matchesTranslationRoot(cfg Config, rawRoot, cleanPath, baseName, ext string) bool {
	root := strings.TrimSpace(rawRoot)
	if root == "" {
		return false
	}

	rel, ok := relativePathWithinRoot(root, cleanPath)
	if !ok {
		return false
	}

	relSlash := filepath.ToSlash(rel)

	// Skip the translation root itself; only files under the root are eligible.
	if relSlash == "." {
		return false
	}

	if cfg.FlatNaming {
		return matchesFlatNaming(relSlash, baseName, cfg.BaseLang, ext, cfg.AlwaysPullBase)
	}

	return matchesNestedNaming(relSlash, cfg.BaseLang, cfg.AlwaysPullBase)
}

func matchesFlatNaming(relSlash, baseName, baseLang, ext string, alwaysPullBase bool) bool {
	// In flat mode, only direct children of the translation root are allowed.
	if strings.Contains(relSlash, "/") {
		return false
	}

	if !alwaysPullBase && baseName == baseLang+"."+ext {
		return false
	}

	return true
}

func matchesNestedNaming(relSlash, baseLang string, alwaysPullBase bool) bool {
	if alwaysPullBase {
		return true
	}

	parts := strings.Split(relSlash, "/")
	if len(parts) > 0 && parts[0] == baseLang {
		return false
	}

	return true
}

func buildAllowedExts(fileExt []string) map[string]struct{} {
	allowedExts := make(map[string]struct{}, len(fileExt))

	for _, ext := range fileExt {
		e := normalizeExt(ext)
		if e == "" {
			continue
		}
		allowedExts[e] = struct{}{}
	}

	return allowedExts
}

func relativePathWithinRoot(root, path string) (string, bool) {
	root = filepath.Clean(strings.TrimSpace(root))
	path = filepath.Clean(strings.TrimSpace(path))

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", false
	}

	rel = filepath.Clean(rel)
	relSlash := filepath.ToSlash(rel)

	if relSlash == ".." || strings.HasPrefix(relSlash, "../") {
		return "", false
	}

	return rel, true
}

func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	ext = strings.TrimPrefix(ext, ".")
	return ext
}
