package sync

import (
	"path/filepath"
	"strings"
)

// Filter handles file inclusion based on patterns.
type Filter struct {
	includePatterns []string
}

// NewFilter creates a Filter from a list of include patterns.
func NewFilter(patterns []string) *Filter {
	return &Filter{includePatterns: patterns}
}

// ShouldInclude returns true if the path should be included.
func (f *Filter) ShouldInclude(path string) bool {
	for _, pattern := range f.includePatterns {
		if matchPattern(pattern, path) {
			return true
		}
	}
	return false
}

// matchPattern checks if a path matches a pattern.
// Supports:
// - Simple directory names: "projects" matches "projects" and "projects/foo"
// - Exact file names: "history.jsonl" matches "history.jsonl"
// - Wildcards: "*.jsonl" matches "foo.jsonl" and "bar/baz.jsonl"
func matchPattern(pattern, path string) bool {
	// Wildcard pattern (e.g., "*.jsonl")
	if strings.Contains(pattern, "*") {
		// Match against the base name of the path
		base := filepath.Base(path)
		matched, _ := filepath.Match(pattern, base)
		return matched
	}

	// Directory pattern (e.g., "projects")
	// Match if path equals pattern or starts with pattern/
	if path == pattern {
		return true
	}
	if strings.HasPrefix(path, pattern+"/") {
		return true
	}

	return false
}
