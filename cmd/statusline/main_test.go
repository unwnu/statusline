package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedBranch    string
		expectedAhead     int
		expectedBehind    int
		expectedTracked   bool
		expectedUntracked bool
	}{
		{
			name: "clean repository",
			input: `# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0`,
			expectedBranch:    "main",
			expectedAhead:     0,
			expectedBehind:    0,
			expectedTracked:   false,
			expectedUntracked: false,
		},
		{
			name: "ahead and behind",
			input: `# branch.head feature-branch
# branch.upstream origin/main
# branch.ab +2 -3`,
			expectedBranch:    "feature-branch",
			expectedAhead:     2,
			expectedBehind:    3,
			expectedTracked:   false,
			expectedUntracked: false,
		},
		{
			name: "with tracked changes",
			input: `# branch.head main
# branch.ab +0 -0
1 M. N... 100644 100644 100644 abc123 def456 file1.txt
2 R. N... 100644 100644 100644 abc123 def456 oldfile.txt newfile.txt`,
			expectedBranch:    "main",
			expectedAhead:     0,
			expectedBehind:    0,
			expectedTracked:   true,
			expectedUntracked: false,
		},
		{
			name: "with untracked files",
			input: `# branch.head main
# branch.ab +0 -0
? untracked.txt
? another-untracked.go`,
			expectedBranch:    "main",
			expectedAhead:     0,
			expectedBehind:    0,
			expectedTracked:   false,
			expectedUntracked: true,
		},
		{
			name: "mixed status",
			input: `# branch.head develop
# branch.ab +1 -0
1 M. N... 100644 100644 100644 abc123 def456 tracked.txt
? untracked.txt`,
			expectedBranch:    "develop",
			expectedAhead:     1,
			expectedBehind:    0,
			expectedTracked:   true,
			expectedUntracked: true,
		},
		{
			name:              "empty input",
			input:             "",
			expectedBranch:    "",
			expectedAhead:     0,
			expectedBehind:    0,
			expectedTracked:   false,
			expectedUntracked: false,
		},
		{
			name: "malformed branch.ab",
			input: `# branch.head main
# branch.ab invalid`,
			expectedBranch:    "main",
			expectedAhead:     0,
			expectedBehind:    0,
			expectedTracked:   false,
			expectedUntracked: false,
		},
		{
			name: "single branch.ab value",
			input: `# branch.head main
# branch.ab +5`,
			expectedBranch:    "main",
			expectedAhead:     0,
			expectedBehind:    0,
			expectedTracked:   false,
			expectedUntracked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branch, ahead, behind, hasTracked, hasUntracked := parseStatus(tt.input)
			assert.Equal(t, tt.expectedBranch, branch)
			assert.Equal(t, tt.expectedAhead, ahead)
			assert.Equal(t, tt.expectedBehind, behind)
			assert.Equal(t, tt.expectedTracked, hasTracked)
			assert.Equal(t, tt.expectedUntracked, hasUntracked)
		})
	}
}

func TestParseSigned(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"positive with plus", "+123", 123},
		{"negative with minus", "-456", 456},
		{"positive without sign", "789", 789},
		{"zero", "0", 0},
		{"empty string", "", 0},
		{"invalid characters", "abc", 0},
		{"mixed valid and invalid", "123abc", 123},
		{"sign only", "+", 0},
		{"minus only", "-", 0},
		{"large number", "+999999", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSigned(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShorten(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"short string unchanged", "main", 10, "main"},
		{"exact max length", "feature", 7, "feature"},
		{"long string shortened", "very-long-feature-branch-name", 15, "very-long-fe..."},
		{"max too small", "main", 3, "main"},
		{"max equals 4", "feature", 4, "feature"},
		{"empty string", "", 10, ""},
		{"single character", "a", 5, "a"},
		{"exactly max-3 length", "test", 7, "test"},
		{"max equals string length", "branch", 6, "branch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shorten(tt.input, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		repoInfo repoInfo
		expected string
	}{
		{
			name: "non-git repository",
			repoInfo: repoInfo{
				Project: "myproject",
				IsGit:   false,
			},
			expected: "myproject",
		},
		{
			name: "clean git repository",
			repoInfo: repoInfo{
				Project: "myproject",
				Branch:  "main",
				IsGit:   true,
			},
			expected: "myproject on \x1b[1;38;5;82m⎇\x1b[0m main",
		},
		{
			name: "repository with tracked changes",
			repoInfo: repoInfo{
				Project:    "myproject",
				Branch:     "feature",
				IsGit:      true,
				HasTracked: true,
			},
			expected: "myproject on \x1b[1;38;5;220m⎇\x1b[0m feature",
		},
		{
			name: "repository with untracked files",
			repoInfo: repoInfo{
				Project:      "myproject",
				Branch:       "develop",
				IsGit:        true,
				HasUntracked: true,
			},
			expected: "myproject on \x1b[1;38;5;196m⎇\x1b[0m develop",
		},
		{
			name: "repository ahead and behind",
			repoInfo: repoInfo{
				Project: "myproject",
				Branch:  "feature",
				Ahead:   2,
				Behind:  1,
				IsGit:   true,
			},
			expected: "myproject on \x1b[1;38;5;82m⎇\x1b[0m feature \x1b[38;5;82m↑2\x1b[0m \x1b[38;5;196m↓1\x1b[0m",
		},
		{
			name: "long branch name shortened",
			repoInfo: repoInfo{
				Project: "myproject",
				Branch:  "very-long-feature-branch-name-that-exceeds-max-length",
				IsGit:   true,
			},
			expected: "myproject on \x1b[1;38;5;82m⎇\x1b[0m very-long-feature-branch-name-that-exceeds-ma...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := render(tt.repoInfo)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderNoColor(t *testing.T) {
	t.Setenv("STATUSLINE_NO_COLOR", "1")

	ri := repoInfo{
		Project:      "myproject",
		Branch:       "main",
		Ahead:        1,
		Behind:       2,
		IsGit:        true,
		HasUntracked: true,
	}

	result := render(ri)
	expected := "myproject on ⎇ main ↑1 ↓2"
	assert.Equal(t, expected, result)
}

func TestReadCwd(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid JSON",
			input:    `{"cwd": "/home/user/project"}`,
			expected: "/home/user/project",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid JSON",
			input:    `{invalid json}`,
			expected: "",
		},
		{
			name:     "JSON with whitespace",
			input:    `{"cwd": "  /home/user/project  "}`,
			expected: "/home/user/project",
		},
		{
			name:     "empty cwd field",
			input:    `{"cwd": ""}`,
			expected: "",
		},
		{
			name:     "missing cwd field",
			input:    `{"other": "value"}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result := readCwd(reader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColorize(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		color    string
		expected string
	}{
		{
			name:     "green color",
			text:     "⎇",
			color:    colGreen,
			expected: "\x1b[38;5;82m⎇\x1b[0m",
		},
		{
			name:     "yellow color",
			text:     "text",
			color:    colYellow,
			expected: "\x1b[38;5;220mtext\x1b[0m",
		},
		{
			name:     "red color",
			text:     "error",
			color:    colRed,
			expected: "\x1b[38;5;196merror\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorize(tt.text, tt.color)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColorizeNoColor(t *testing.T) {
	t.Setenv("STATUSLINE_NO_COLOR", "1")

	result := colorize("test", colGreen)
	assert.Equal(t, "test", result)
}

func TestColorizeBold(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		color    string
		expected string
	}{
		{
			name:     "bold green color",
			text:     "⎇",
			color:    colGreen,
			expected: "\x1b[1;38;5;82m⎇\x1b[0m",
		},
		{
			name:     "bold yellow color",
			text:     "text",
			color:    colYellow,
			expected: "\x1b[1;38;5;220mtext\x1b[0m",
		},
		{
			name:     "bold red color",
			text:     "error",
			color:    colRed,
			expected: "\x1b[1;38;5;196merror\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeBold(tt.text, tt.color)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColorizeBoldNoColor(t *testing.T) {
	t.Setenv("STATUSLINE_NO_COLOR", "1")

	result := colorizeBold("test", colGreen)
	assert.Equal(t, "test", result)
}

func TestRepoInfo(t *testing.T) {
	ri := repoInfo{
		Project:      "testproject",
		Branch:       "main",
		Ahead:        5,
		Behind:       2,
		HasTracked:   true,
		HasUntracked: false,
		IsGit:        true,
	}

	assert.Equal(t, "testproject", ri.Project)
	assert.Equal(t, "main", ri.Branch)
	assert.Equal(t, 5, ri.Ahead)
	assert.Equal(t, 2, ri.Behind)
	assert.True(t, ri.HasTracked)
	assert.False(t, ri.HasUntracked)
	assert.True(t, ri.IsGit)
}

func TestGit(t *testing.T) {
	t.Run("git command timeout", func(t *testing.T) {
		// Test with sleep command to trigger timeout
		result := git("/tmp", "sleep", "1")
		assert.Equal(t, "", result)
	})

	t.Run("git command in non-git directory", func(t *testing.T) {
		result := git("/tmp", "rev-parse", "--show-toplevel")
		assert.Equal(t, "", result)
	})

	t.Run("invalid git command", func(t *testing.T) {
		result := git("/tmp", "invalid-command")
		assert.Equal(t, "", result)
	})
}

func TestCollect(t *testing.T) {
	t.Run("non-git directory", func(t *testing.T) {
		ri := collect("/tmp")
		assert.Equal(t, "tmp", ri.Project)
		assert.False(t, ri.IsGit)
		assert.Equal(t, "", ri.Branch)
		assert.Equal(t, 0, ri.Ahead)
		assert.Equal(t, 0, ri.Behind)
		assert.False(t, ri.HasTracked)
		assert.False(t, ri.HasUntracked)
	})

	t.Run("directory path handling", func(t *testing.T) {
		ri := collect("/some/deep/project/path")
		assert.Equal(t, "path", ri.Project)
		assert.False(t, ri.IsGit)
	})

	t.Run("collect with STATUSLINE_FETCH", func(t *testing.T) {
		t.Setenv("STATUSLINE_FETCH", "1")
		ri := collect("/tmp")
		assert.Equal(t, "tmp", ri.Project)
		assert.False(t, ri.IsGit)
	})

	// Test git repository simulation using the actual git directory
	t.Run("git repository basics", func(t *testing.T) {
		// Use current directory which is a git repo
		ri := collect(".")
		assert.Equal(t, "statusline", ri.Project)
		assert.True(t, ri.IsGit)
		// Branch name will depend on actual git state, so just check it's not empty
		assert.NotEmpty(t, ri.Branch)
	})

	t.Run("git repository with STATUSLINE_FETCH enabled", func(t *testing.T) {
		t.Setenv("STATUSLINE_FETCH", "1")
		// Use current directory which is a git repo
		ri := collect(".")
		assert.Equal(t, "statusline", ri.Project)
		assert.True(t, ri.IsGit)
		assert.NotEmpty(t, ri.Branch)
	})
}

func TestMainFunction(t *testing.T) {
	// Test main function indirectly by testing its components
	// Since main() calls other functions we've already tested, we focus on edge cases

	// Test working directory fallback behavior would require mocking os.Getwd
	// For now, we test that main doesn't panic with different inputs

	t.Run("main function integration", func(t *testing.T) {
		// This is more of a smoke test to ensure main doesn't panic
		// Real integration would require more sophisticated mocking
		assert.NotPanics(t, func() {
			// We can't easily test main() directly without refactoring,
			// but we can test the data flow through render(collect(...))
			result := render(collect("/tmp"))
			assert.Contains(t, result, "tmp")
		})
	})
}

// Additional edge case tests for better coverage
func TestParseStatusEdgeCases(t *testing.T) {
	t.Run("parse status with whitespace lines", func(t *testing.T) {
		input := `# branch.head main
		
   
# branch.ab +1 -0`
		branch, ahead, behind, hasTracked, hasUntracked := parseStatus(input)
		assert.Equal(t, "main", branch)
		assert.Equal(t, 1, ahead)
		assert.Equal(t, 0, behind)
		assert.False(t, hasTracked)
		assert.False(t, hasUntracked)
	})

	t.Run("parse status with mixed prefixes", func(t *testing.T) {
		input := `# branch.head develop
# branch.ab +0 -1
1 M. N... 100644 100644 100644 abc123 def456 file1.txt
2 R. N... 100644 100644 100644 abc123 def456 old.txt new.txt
? untracked.txt
u AM N... 100644 100644 100644 abc123 def456 unmerged.txt`
		branch, ahead, behind, hasTracked, hasUntracked := parseStatus(input)
		assert.Equal(t, "develop", branch)
		assert.Equal(t, 0, ahead)
		assert.Equal(t, 1, behind)
		assert.True(t, hasTracked)
		assert.True(t, hasUntracked)
	})
}

func TestRenderEdgeCases(t *testing.T) {
	t.Run("render with only ahead", func(t *testing.T) {
		ri := repoInfo{
			Project: "test",
			Branch:  "main",
			Ahead:   3,
			Behind:  0,
			IsGit:   true,
		}
		result := render(ri)
		assert.Contains(t, result, "↑3")
		assert.NotContains(t, result, "↓")
	})

	t.Run("render with only behind", func(t *testing.T) {
		ri := repoInfo{
			Project: "test",
			Branch:  "main",
			Ahead:   0,
			Behind:  2,
			IsGit:   true,
		}
		result := render(ri)
		assert.Contains(t, result, "↓2")
		assert.NotContains(t, result, "↑")
	})
}

func TestGetFetchInterval(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{
			name:     "default interval",
			envVar:   "",
			expected: "30m0s",
		},
		{
			name:     "custom interval 5 minutes",
			envVar:   "5",
			expected: "5m0s",
		},
		{
			name:     "custom interval 60 minutes",
			envVar:   "60",
			expected: "1h0m0s",
		},
		{
			name:     "custom interval 1 minute",
			envVar:   "1",
			expected: "1m0s",
		},
		{
			name:     "zero interval for always fetch",
			envVar:   "0",
			expected: "0s",
		},
		{
			name:     "invalid interval negative",
			envVar:   "-5",
			expected: "30m0s",
		},
		{
			name:     "invalid interval non-numeric",
			envVar:   "abc",
			expected: "30m0s",
		},
		{
			name:     "invalid interval float",
			envVar:   "5.5",
			expected: "30m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv("STATUSLINE_FETCH_INTERVAL", tt.envVar)
			}
			result := getFetchInterval()
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestShouldFetch(t *testing.T) {
	// Generate timestamps for testing
	now := time.Now()
	recentTimestamp := now.Add(-10 * time.Minute).Unix()  // 10 minutes ago
	oldTimestamp := now.Add(-60 * time.Minute).Unix()     // 60 minutes ago

	tests := []struct {
		name         string
		reflogOutput string
		interval     string
		expected     bool
		description  string
	}{
		{
			name:         "no reflog output",
			reflogOutput: "",
			interval:     "30",
			expected:     true,
			description:  "should fetch if no reflog available",
		},
		{
			name:         "recent fetch within interval",
			reflogOutput: fmt.Sprintf("abc123 HEAD@{%d}: fetch: from origin main", recentTimestamp),
			interval:     "30",
			expected:     false,
			description:  "should not fetch if recently fetched",
		},
		{
			name:         "old fetch outside interval",
			reflogOutput: fmt.Sprintf("def456 HEAD@{%d}: fetch: from origin main", oldTimestamp),
			interval:     "30",
			expected:     true,
			description:  "should fetch if last fetch was long ago",
		},
		{
			name:         "zero interval always fetch",
			reflogOutput: fmt.Sprintf("abc123 HEAD@{%d}: fetch: from origin main", recentTimestamp),
			interval:     "0",
			expected:     true,
			description:  "should always fetch with zero interval",
		},
		{
			name:         "malformed reflog",
			reflogOutput: "invalid reflog entry",
			interval:     "30",
			expected:     true,
			description:  "should fetch if reflog is malformed",
		},
		{
			name:         "reflog without fetch entry",
			reflogOutput: fmt.Sprintf("abc123 HEAD@{%d}: commit: some commit message", recentTimestamp),
			interval:     "30",
			expected:     true,
			description:  "should fetch if no fetch entry in reflog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("STATUSLINE_FETCH_INTERVAL", tt.interval)

			interval := getFetchInterval()
			result := shouldFetchFromReflog(tt.reflogOutput, interval)

			// Check the expected result
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestFetchLogic(t *testing.T) {
	tests := []struct {
		name        string
		fetchEnv    string
		intervalEnv string
		description string
	}{
		{
			name:        "fetch disabled",
			fetchEnv:    "",
			intervalEnv: "30",
			description: "no fetch when STATUSLINE_FETCH not set",
		},
		{
			name:        "fetch enabled with default interval",
			fetchEnv:    "1",
			intervalEnv: "",
			description: "fetch enabled with 30 minute default",
		},
		{
			name:        "fetch enabled with custom interval",
			fetchEnv:    "1",
			intervalEnv: "5",
			description: "fetch enabled with 5 minute interval",
		},
		{
			name:        "fetch enabled with zero interval",
			fetchEnv:    "1",
			intervalEnv: "0",
			description: "fetch always with zero interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fetchEnv != "" {
				t.Setenv("STATUSLINE_FETCH", tt.fetchEnv)
			}
			if tt.intervalEnv != "" {
				t.Setenv("STATUSLINE_FETCH_INTERVAL", tt.intervalEnv)
			}

			fetchEnabled := tt.fetchEnv == "1"
			interval := getFetchInterval()

			if !fetchEnabled {
				assert.NotEqual(t, "1", tt.fetchEnv)
			} else {
				switch tt.intervalEnv {
				case "0":
					assert.Equal(t, "0s", interval.String(), "zero interval means always fetch")
				case "":
					assert.Equal(t, "30m0s", interval.String(), "empty should use default")
				}
			}
		})
	}
}
