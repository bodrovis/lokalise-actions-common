package managedpaths

import (
	"fmt"
	"maps"
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
	FileExts       []string
	FlatNaming     bool
	AlwaysPullBase bool
	BaseLang       string
}

// ToTranslationFilesConfig converts the managed-path scope to translationfiles config.
func (s TranslationScope) ToTranslationFilesConfig() translationfiles.Config {
	return translationfiles.Config{
		TranslationPaths: s.Paths,
		FileExts:         s.FileExts,
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

func filterManagedPaths(
	cfg translationfiles.Config,
	paths []string,
) []string {
	matched := make(map[string]struct{})

	for _, path := range paths {
		path, ok := normalizePath(path)
		if !ok || !translationfiles.Matches(cfg, path) {
			continue
		}

		matched[path] = struct{}{}
	}

	return slices.Sorted(maps.Keys(matched))
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

func captureGitPathList(
	r CaptureRunner,
	errPrefix string,
	gitArgs ...string,
) ([]string, error) {
	args := append([]string{"-c", "core.quotepath=false"}, gitArgs...)

	out, err := r.Capture("git", args...)
	if err != nil {
		return nil, fmt.Errorf(
			"%s: %w\nOutput: %s",
			errPrefix,
			err,
			strings.TrimSpace(out),
		)
	}

	return parseNonEmptyLines(out), nil
}

func parseNonEmptyLines(s string) []string {
	var result []string

	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	return result
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

	for _, group := range groups {
		for _, path := range group {
			path, ok := normalizePath(path)
			if ok {
				seen[path] = struct{}{}
			}
		}
	}

	return slices.Sorted(maps.Keys(seen))
}
