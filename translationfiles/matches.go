package translationfiles

import (
	"path/filepath"
	"strings"
)

type Config struct {
	TranslationPaths []string
	FileExts         []string
	FlatNaming       bool
	AlwaysPullBase   bool
	BaseLang         string
}

func Matches(cfg Config, path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}

	allowedExts := buildAllowedExts(cfg.FileExts)
	if len(allowedExts) == 0 {
		return false
	}

	cleanPath, ext, ok := matchesAllowedExtension(path, allowedExts)
	if !ok {
		return false
	}

	return matchesAnyTranslationRoot(cfg, cleanPath, ext)
}

func matchesAllowedExtension(
	filePath string,
	allowedExts map[string]struct{},
) (string, string, bool) {
	cleanPath := filepath.Clean(normalizePathSeparators(filePath))
	lowerPath := strings.ToLower(cleanPath)

	var matchedExt string

	for ext := range allowedExts {
		if strings.HasSuffix(lowerPath, "."+ext) && len(ext) > len(matchedExt) {
			matchedExt = ext
		}
	}

	if matchedExt == "" {
		return "", "", false
	}

	return cleanPath, matchedExt, true
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

func matchesFlatNaming(
	relSlash,
	baseName,
	baseLang,
	ext string,
	alwaysPullBase bool,
) bool {
	if strings.Contains(relSlash, "/") {
		return false
	}

	if !alwaysPullBase &&
		strings.EqualFold(baseName, baseLang+"."+ext) {
		return false
	}

	return true
}

func matchesNestedNaming(
	relSlash,
	baseLang string,
	alwaysPullBase bool,
) bool {
	if alwaysPullBase {
		return true
	}

	first, _, _ := strings.Cut(relSlash, "/")
	return first != baseLang
}

func buildAllowedExts(fileExts []string) map[string]struct{} {
	allowedExts := make(map[string]struct{}, len(fileExts))

	for _, ext := range fileExts {
		e := normalizeExt(ext)
		if e == "" {
			continue
		}
		allowedExts[e] = struct{}{}
	}

	return allowedExts
}

func relativePathWithinRoot(root, filePath string) (string, bool) {
	root = filepath.Clean(normalizePathSeparators(root))
	filePath = filepath.Clean(normalizePathSeparators(filePath))

	rel, err := filepath.Rel(root, filePath)
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

func normalizePathSeparators(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), `\`, "/")
}
