package translationfiles

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Config struct {
	TranslationPaths []string
	FileExt          []string
	FlatNaming       bool
	AlwaysPullBase   bool
	BaseLang         string
}

// Collect returns normalized, slash-separated, sorted, deduplicated paths
// under the configured translation roots.
func Collect(cfg Config) ([]string, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	allowedExts := make(map[string]struct{}, len(cfg.FileExt))
	for _, ext := range cfg.FileExt {
		e := normalizeExt(ext)
		if e != "" {
			allowedExts[e] = struct{}{}
		}
	}

	seen := make(map[string]struct{})
	files := make([]string, 0)

	for _, rawPath := range cfg.TranslationPaths {
		root := filepath.Clean(strings.TrimSpace(rawPath))

		info, err := os.Stat(root)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("translation path does not exist: %s", root)
			}
			return nil, fmt.Errorf("failed to stat translation path %q: %w", root, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("translation path is not a directory: %s", root)
		}

		var collected []string
		if cfg.FlatNaming {
			collected, err = collectFlat(root, cfg, allowedExts)
		} else {
			collected, err = collectRecursive(root, cfg, allowedExts)
		}
		if err != nil {
			return nil, err
		}

		for _, file := range collected {
			file = filepath.ToSlash(filepath.Clean(file))
			if _, ok := seen[file]; ok {
				continue
			}
			seen[file] = struct{}{}
			files = append(files, file)
		}
	}

	slices.Sort(files)
	return files, nil
}

func validateConfig(cfg Config) error {
	if len(cfg.TranslationPaths) == 0 {
		return fmt.Errorf("translation paths are required")
	}
	if len(cfg.FileExt) == 0 {
		return fmt.Errorf("file extensions are required")
	}
	if strings.TrimSpace(cfg.BaseLang) == "" {
		return fmt.Errorf("base language is required")
	}
	if strings.ContainsAny(cfg.BaseLang, `/\`) {
		return fmt.Errorf("base language must not contain path separators")
	}

	for _, p := range cfg.TranslationPaths {
		if strings.TrimSpace(p) == "" {
			return fmt.Errorf("translation path cannot be empty")
		}
	}

	validExtFound := false
	for _, ext := range cfg.FileExt {
		e := normalizeExt(ext)
		if e == "" {
			continue
		}
		if strings.ContainsAny(e, `/\`) {
			return fmt.Errorf("invalid file extension %q", ext)
		}
		validExtFound = true
	}
	if !validExtFound {
		return fmt.Errorf("no valid file extensions after normalization")
	}

	return nil
}

func collectFlat(root string, cfg Config, allowedExts map[string]struct{}) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read translation path %q: %w", root, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		fullPath := filepath.Join(root, name)

		if !hasAllowedExt(name, allowedExts) {
			continue
		}
		if shouldExcludeFlatBaseFile(name, cfg) {
			continue
		}

		files = append(files, fullPath)
	}

	return files, nil
}

func collectRecursive(root string, cfg Config, allowedExts map[string]struct{}) ([]string, error) {
	files := make([]string, 0)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("failed while walking %q: %w", path, walkErr)
		}

		if d.IsDir() {
			if !cfg.AlwaysPullBase && isBaseLangDir(root, path, cfg.BaseLang) {
				return filepath.SkipDir
			}
			return nil
		}

		if !hasAllowedExt(d.Name(), allowedExts) {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func hasAllowedExt(name string, allowedExts map[string]struct{}) bool {
	ext := normalizeExt(filepath.Ext(name))
	if ext == "" {
		return false
	}
	_, ok := allowedExts[ext]
	return ok
}

func shouldExcludeFlatBaseFile(fileName string, cfg Config) bool {
	if cfg.AlwaysPullBase {
		return false
	}

	ext := normalizeExt(filepath.Ext(fileName))
	if ext == "" {
		return false
	}

	expected := cfg.BaseLang + "." + ext
	return fileName == expected
}

func isBaseLangDir(root, path, baseLang string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return false
	}

	parts := strings.Split(filepath.ToSlash(rel), "/")
	return len(parts) > 0 && parts[0] == baseLang
}

func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	ext = strings.TrimPrefix(ext, ".")
	return ext
}
