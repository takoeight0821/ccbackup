package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter_ShouldInclude_BasicPatterns(t *testing.T) {
	f := NewFilter([]string{"projects", "history.jsonl", "plans"})

	tests := []struct {
		name    string
		path    string
		include bool
	}{
		// "projects" should match projects/ and projects/session.jsonl
		{"projects dir", "projects", true},
		{"projects subfile", "projects/session.jsonl", true},
		// "projects" should NOT match projectsExtra/foo
		{"projectsExtra not matched", "projectsExtra/foo", false},
		// "history.jsonl" should match exactly
		{"history.jsonl", "history.jsonl", true},
		// "history.jsonl" should NOT match other files
		{"other jsonl not matched", "other.jsonl", false},
		// "plans" should match plans/ and plans/plan.md
		{"plans dir", "plans", true},
		{"plans subfile", "plans/plan.md", true},
		// Files not in include list should not be included
		{"debug not included", "debug/log.txt", false},
		{"settings.json not included", "settings.json", false},
		{"tasks not included", "tasks/task.json", false},
		{"file-history not included", "file-history/abc/file@v1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.include, f.ShouldInclude(tt.path))
		})
	}
}

func TestFilter_WildcardPatterns(t *testing.T) {
	f := NewFilter([]string{"*.jsonl"})

	tests := []struct {
		name    string
		path    string
		include bool
	}{
		// "*.jsonl" should match .jsonl files
		{"history.jsonl", "history.jsonl", true},
		{"nested jsonl", "projects/session.jsonl", true},
		// "*.jsonl" should NOT match .json
		{"settings.json not matched", "settings.json", false},
		{"nested json not matched", "foo/bar.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.include, f.ShouldInclude(tt.path))
		})
	}
}

func TestFilter_EmptyPatterns(t *testing.T) {
	f := NewFilter([]string{})
	// With empty patterns, nothing should be included
	assert.False(t, f.ShouldInclude("anything"))
	assert.False(t, f.ShouldInclude("projects/session.jsonl"))
}

func TestFilter_RealWorldConfig(t *testing.T) {
	// Real-world config from design doc
	f := NewFilter([]string{
		"projects",
		"history.jsonl",
		"plans",
		"todos",
	})

	tests := []struct {
		name    string
		path    string
		include bool
	}{
		// Directories to include
		{"projects dir", "projects/session.jsonl", true},
		{"plans dir", "plans/plan.md", true},
		{"todos dir", "todos/task.json", true},
		// Files to include
		{"history.jsonl", "history.jsonl", true},
		// Files/directories NOT to include
		{"file-history dir", "file-history/abc/file@v1", false},
		{"debug dir", "debug/logs/error.log", false},
		{"cache dir", "cache/temp", false},
		{"statsig dir", "statsig/data", false},
		{"settings.json", "settings.json", false},
		{"nested json", "config/app.json", false},
		{"tasks dir", "tasks/task.json", false},
		{"usage-data", "usage-data/stats.json", false},
		{"settings.json.backup", "settings.json.backup", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.include, f.ShouldInclude(tt.path))
		})
	}
}
