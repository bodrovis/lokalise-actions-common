package managedpaths

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bodrovis/lokalise-actions-common/v2/translationfiles"
)

type CaptureRunner interface {
	Capture(name string, args ...string) (string, error)
}

type TranslationScope struct {
	Paths          []string
	FileExt        []string
	FlatNaming     bool
	AlwaysPullBase bool
	BaseLang       string
}

func (s TranslationScope) ToTranslationFilesConfig() translationfiles.Config {
	return translationfiles.Config{
		TranslationPaths: s.Paths,
		FileExt:          s.FileExt,
		FlatNaming:       s.FlatNaming,
		AlwaysPullBase:   s.AlwaysPullBase,
		BaseLang:         s.BaseLang,
	}
}

func CollectManagedGitPaths(r CaptureRunner, scope TranslationScope) ([]string, error) {
	changed, err := collectChangedTrackedPaths(r)
	if err != nil {
		return nil, err
	}

	untracked, err := collectUntrackedPaths(r)
	if err != nil {
		return nil, err
	}

	all := mergeAndNormalize(changed, untracked)
	return FilterManaged(scope, all), nil
}

func HasManagedGitPaths(r CaptureRunner, scope TranslationScope) (bool, error) {
	paths, err := CollectManagedGitPaths(r, scope)
	if err != nil {
		return false, err
	}
	return len(paths) > 0, nil
}

func FilterManaged(scope TranslationScope, paths []string) []string {
	tfCfg := scope.ToTranslationFilesConfig()

	filtered := make([]string, 0, len(paths))
	for _, p := range paths {
		p = filepath.ToSlash(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		if translationfiles.Matches(tfCfg, p) {
			filtered = append(filtered, p)
		}
	}

	return mergeAndNormalize(filtered)
}

func collectChangedTrackedPaths(r CaptureRunner) ([]string, error) {
	if _, err := r.Capture("git", "rev-parse", "--verify", "HEAD"); err == nil {
		out, err := r.Capture("git", "-c", "core.quotepath=false", "diff", "--name-only", "HEAD")
		if err != nil {
			return nil, fmt.Errorf("git diff HEAD failed: %w\nOutput: %s", err, out)
		}
		return normalize(splitNonEmptyLines(out)), nil
	}

	var all []string

	outCached, err := r.Capture("git", "-c", "core.quotepath=false", "diff", "--name-only", "--cached")
	if err != nil {
		return nil, fmt.Errorf("git diff --cached (no HEAD) failed: %w\nOutput: %s", err, outCached)
	}
	all = append(all, splitNonEmptyLines(outCached)...)

	outWT, err := r.Capture("git", "-c", "core.quotepath=false", "diff", "--name-only")
	if err != nil {
		return nil, fmt.Errorf("git diff (worktree) failed: %w\nOutput: %s", err, outWT)
	}
	all = append(all, splitNonEmptyLines(outWT)...)

	return mergeAndNormalize(all), nil
}

func collectUntrackedPaths(r CaptureRunner) ([]string, error) {
	out, err := r.Capture("git", "-c", "core.quotepath=false", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w\nOutput: %s", err, out)
	}
	return normalize(splitNonEmptyLines(out)), nil
}

func splitNonEmptyLines(s string) []string {
	var res []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		res = append(res, line)
	}
	return res
}

func normalize(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		p = filepath.ToSlash(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func mergeAndNormalize(groups ...[]string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)

	for _, group := range groups {
		for _, p := range group {
			p = filepath.ToSlash(strings.TrimSpace(p))
			if p == "" {
				continue
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}

	slices.Sort(out)
	return out
}
