package sync

import (
	"path/filepath"
	"strings"
)

// Filter handles file exclusion based on patterns.
type Filter struct {
	excludePatterns []string
	includePatterns []string // negation patterns (start with !)
}

// NewFilter creates a Filter from a list of patterns.
// Patterns starting with ! are negation patterns (include).
func NewFilter(patterns []string) *Filter {
	f := &Filter{}
	for _, p := range patterns {
		if strings.HasPrefix(p, "!") {
			f.includePatterns = append(f.includePatterns, p[1:])
		} else {
			f.excludePatterns = append(f.excludePatterns, p)
		}
	}
	return f
}

// ShouldExclude returns true if the path should be excluded.
func (f *Filter) ShouldExclude(path string) bool {
	// First check if it matches any include (negation) pattern
	for _, pattern := range f.includePatterns {
		if matchPattern(pattern, path) {
			return false
		}
	}

	// Then check if it matches any exclude pattern
	for _, pattern := range f.excludePatterns {
		if matchPattern(pattern, path) {
			return true
		}
	}

	return false
}

// matchPattern checks if a path matches a pattern.
// Supports:
// - Simple directory names: "debug" matches "debug" and "debug/foo"
// - Wildcards: "*.json" matches "foo.json" and "bar/baz.json"
func matchPattern(pattern, path string) bool {
	// Wildcard pattern (e.g., "*.json")
	if strings.Contains(pattern, "*") {
		// Match against the base name of the path
		base := filepath.Base(path)
		matched, _ := filepath.Match(pattern, base)
		return matched
	}

	// Directory pattern (e.g., "debug")
	// Match if path equals pattern or starts with pattern/
	if path == pattern {
		return true
	}
	if strings.HasPrefix(path, pattern+"/") {
		return true
	}

	return false
}
