package managedpaths

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bodrovis/lokalise-actions-common/v2/translationfiles"
)

// CaptureRunner runs a command and returns its captured stdout.
type CaptureRunner interface {
	Capture(name string, args ...string) (string, error)
}

// TranslationScope describes which translation files are managed by the action.
type TranslationScope struct {
	Paths          []string
	FileExt        []string
	FlatNaming     bool
	AlwaysPullBase bool
	BaseLang       string
}

// ToTranslationFilesConfig converts the managed-path scope to translationfiles config.
func (s TranslationScope) ToTranslationFilesConfig() translationfiles.Config {
	return translationfiles.Config{
		TranslationPaths: s.Paths,
		FileExt:          s.FileExt,
		FlatNaming:       s.FlatNaming,
		AlwaysPullBase:   s.AlwaysPullBase,
		BaseLang:         s.BaseLang,
	}
}

// CollectManagedGitPaths returns changed or untracked Git paths that belong to the
// managed translation scope. The result is normalized, deduplicated, and sorted.
func CollectManagedGitPaths(r CaptureRunner, scope TranslationScope) ([]string, error) {
	paths, err := collectCandidateGitPaths(r)
	if err != nil {
		return nil, err
	}

	return FilterManaged(scope, paths), nil
}

// HasManagedGitPaths reports whether Git currently contains any managed translation paths.
func HasManagedGitPaths(r CaptureRunner, scope TranslationScope) (bool, error) {
	paths, err := CollectManagedGitPaths(r, scope)
	if err != nil {
		return false, err
	}
	return len(paths) > 0, nil
}

// FilterManaged keeps only paths that match the translation scope.
// The result is normalized, deduplicated, and sorted.
func FilterManaged(scope TranslationScope, paths []string) []string {
	return filterManagedPaths(scope.ToTranslationFilesConfig(), paths)
}

func filterManagedPaths(cfg translationfiles.Config, paths []string) []string {
	filtered := make([]string, 0, len(paths))

	for _, p := range paths {
		p, ok := normalizePath(p)
		if !ok {
			continue
		}
		if translationfiles.Matches(cfg, p) {
			filtered = append(filtered, p)
		}
	}

	return mergeAndNormalize(filtered)
}

func collectCandidateGitPaths(r CaptureRunner) ([]string, error) {
	changed, err := collectChangedTrackedPaths(r)
	if err != nil {
		return nil, err
	}

	untracked, err := collectUntrackedPaths(r)
	if err != nil {
		return nil, err
	}

	return mergeAndNormalize(changed, untracked), nil
}

func collectChangedTrackedPaths(r CaptureRunner) ([]string, error) {
	if hasHEAD(r) {
		return collectChangedTrackedPathsFromHEAD(r)
	}

	return collectChangedTrackedPathsWithoutHEAD(r)
}

// hasHEAD reports whether the repository already has at least one commit.
func hasHEAD(r CaptureRunner) bool {
	_, err := r.Capture("git", "rev-parse", "--verify", "HEAD")
	return err == nil
}

// When HEAD exists, `git diff --name-only HEAD` covers both staged and unstaged
// tracked changes relative to the current commit.
func collectChangedTrackedPathsFromHEAD(r CaptureRunner) ([]string, error) {
	return captureGitPathList(
		r,
		"git diff HEAD failed",
		"diff", "--name-only", "HEAD",
	)
}

// In repositories without HEAD yet, staged and unstaged tracked changes must be
// collected separately.
func collectChangedTrackedPathsWithoutHEAD(r CaptureRunner) ([]string, error) {
	cached, err := captureGitPathList(
		r,
		"git diff --cached (no HEAD) failed",
		"diff", "--name-only", "--cached",
	)
	if err != nil {
		return nil, err
	}

	worktree, err := captureGitPathList(
		r,
		"git diff (worktree) failed",
		"diff", "--name-only",
	)
	if err != nil {
		return nil, err
	}

	return mergeAndNormalize(cached, worktree), nil
}

func collectUntrackedPaths(r CaptureRunner) ([]string, error) {
	return captureGitPathList(
		r,
		"git ls-files failed",
		"ls-files", "--others", "--exclude-standard",
	)
}

func captureGitPathList(r CaptureRunner, errPrefix string, gitArgs ...string) ([]string, error) {
	args := append([]string{"-c", "core.quotepath=false"}, gitArgs...)

	out, err := r.Capture("git", args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w\nOutput: %s", errPrefix, err, out)
	}

	return normalize(parseNonEmptyLines(out)), nil
}

func parseNonEmptyLines(s string) []string {
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
		p, ok := normalizePath(p)
		if !ok {
			continue
		}
		out = append(out, p)
	}

	return out
}

func normalizePath(p string) (string, bool) {
	p = filepath.ToSlash(strings.TrimSpace(p))
	if p == "" {
		return "", false
	}
	return p, true
}

// mergeAndNormalize flattens path groups, drops empty entries, removes duplicates,
// normalizes separators, and sorts the result for deterministic output.
func mergeAndNormalize(groups ...[]string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)

	for _, group := range groups {
		for _, p := range group {
			p, ok := normalizePath(p)
			if !ok {
				continue
			}
			if _, exists := seen[p]; exists {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}

	slices.Sort(out)
	return out
}
