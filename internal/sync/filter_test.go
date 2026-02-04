package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter_ShouldExclude_BasicPatterns(t *testing.T) {
	f := NewFilter([]string{"debug", "cache"})

	tests := []struct {
		name    string
		path    string
		exclude bool
	}{
		// "debug" should match debug/ and debug/foo.log
		{"debug dir", "debug", true},
		{"debug subfile", "debug/foo.log", true},
		// "debug" should NOT match debugger/foo
		{"debugger not matched", "debugger/foo", false},
		// "cache" should match cache/ and cache/data
		{"cache dir", "cache", true},
		{"cache subfile", "cache/data", true},
		// Other paths should not be excluded
		{"projects", "projects/session.jsonl", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.exclude, f.ShouldExclude(tt.path))
		})
	}
}

func TestFilter_WildcardPatterns(t *testing.T) {
	f := NewFilter([]string{"*.json"})

	tests := []struct {
		name    string
		path    string
		exclude bool
	}{
		// "*.json" should match .json files
		{"settings.json", "settings.json", true},
		{"nested json", "foo/bar.json", true},
		// "*.json" should NOT match .jsonl
		{"history.jsonl not matched", "history.jsonl", false},
		{"nested jsonl", "projects/session.jsonl", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.exclude, f.ShouldExclude(tt.path))
		})
	}
}

func TestFilter_NegationPatterns(t *testing.T) {
	f := NewFilter([]string{"*.json", "!history.jsonl"})

	tests := []struct {
		name    string
		path    string
		exclude bool
	}{
		{"settings.json excluded", "settings.json", true},
		{"history.jsonl NOT excluded (negation)", "history.jsonl", false},
		{"other json excluded", "config/app.json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.exclude, f.ShouldExclude(tt.path))
		})
	}
}

func TestFilter_EmptyPatterns(t *testing.T) {
	f := NewFilter([]string{})
	assert.False(t, f.ShouldExclude("anything"))
}

func TestFilter_ComplexPatterns(t *testing.T) {
	// Real-world config from design doc
	f := NewFilter([]string{
		"debug",
		"cache",
		"statsig",
		"telemetry",
		"plugins",
		"ide",
		"paste-cache",
		"shell-snapshots",
		"*.json",
		"!history.jsonl",
	})

	tests := []struct {
		name    string
		path    string
		exclude bool
	}{
		// Directories to exclude
		{"debug dir", "debug/logs/error.log", true},
		{"cache dir", "cache/temp", true},
		{"statsig dir", "statsig/data", true},
		// Files to exclude
		{"json files", "settings.json", true},
		{"nested json", "config/app.json", true},
		// Files to include (negation)
		{"history.jsonl", "history.jsonl", false},
		// Files to include (not matched)
		{"projects", "projects/session.jsonl", false},
		{"file-history", "file-history/file.txt", false},
		{"plans", "plans/plan.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.exclude, f.ShouldExclude(tt.path))
		})
	}
}
