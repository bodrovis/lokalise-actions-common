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

	cleanPath := filepath.Clean(path)
	ext := normalizeExt(filepath.Ext(cleanPath))
	if ext == "" {
		return false
	}
	if _, ok := allowedExts[ext]; !ok {
		return false
	}

	baseName := filepath.Base(cleanPath)

	for _, rawRoot := range cfg.TranslationPaths {
		root := strings.TrimSpace(rawRoot)
		if root == "" {
			continue
		}

		rel, ok := relativePathWithinRoot(root, cleanPath)
		if !ok {
			continue
		}

		relSlash := filepath.ToSlash(rel)
		if relSlash == "." || relSlash == "" {
			continue
		}

		if cfg.FlatNaming {
			// only direct children of the translation root are allowed
			if strings.Contains(relSlash, "/") {
				continue
			}

			if !cfg.AlwaysPullBase && baseName == cfg.BaseLang+"."+ext {
				continue
			}
		} else {
			if !cfg.AlwaysPullBase {
				parts := strings.Split(relSlash, "/")
				if len(parts) > 0 && parts[0] == cfg.BaseLang {
					continue
				}
			}
		}

		return true
	}

	return false
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
