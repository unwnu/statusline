// Package main provides a statusline for Claude Code.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	version = "dev"
	build   = "local"
)

const (
	colGreen     = "38;5;82"
	colYellow    = "38;5;220"
	colRed       = "38;5;196"
	esc          = "\x1b"
	maxBranchLen = 48
)

type input struct {
	Cwd string `json:"cwd"`
}

type repoInfo struct {
	Project                         string
	Branch                          string
	Ahead, Behind                   int
	HasTracked, HasUntracked, IsGit bool
}

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "show version and exit")
	flag.BoolVar(&showVersion, "version", false, "show version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("statusline %s (built: %s)\n", version, build)
		os.Exit(0)
	}

	cwd := readCwd(os.Stdin)
	if cwd == "" {
		if d, err := os.Getwd(); err == nil {
			cwd = d
		}
	}
	fmt.Println(render(collect(cwd)))
}

func collect(cwd string) repoInfo {
	var ri repoInfo
	ri.Project = filepath.Base(cwd)

	root := git(cwd, "rev-parse", "--show-toplevel")
	if root == "" {
		return ri
	}
	ri.IsGit = true
	ri.Project = filepath.Base(root)

	if os.Getenv("STATUSLINE_FETCH") == "1" && shouldFetch(root) {
		up := git(root, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
		if up != "" {
			if parts := strings.SplitN(up, "/", 2); len(parts) == 2 {
				_ = git(root, "fetch", "--quiet", "--no-progress", "--prune", parts[0], parts[1])
			}
		}
	}

	status := git(root, "status", "--porcelain=2", "--branch", "--ignore-submodules=dirty")
	ri.Branch, ri.Ahead, ri.Behind, ri.HasTracked, ri.HasUntracked = parseStatus(status)

	if ri.Branch == "" {
		ri.Branch = "no-branch"
	}
	if ri.Branch == "(detached)" {
		if sha := git(root, "rev-parse", "--short", "HEAD"); sha != "" {
			ri.Branch = "detached@" + sha
		}
	}
	return ri
}

func render(ri repoInfo) string {
	if !ri.IsGit {
		return ri.Project
	}
	iconCol := colGreen
	switch {
	case ri.HasUntracked:
		iconCol = colRed
	case ri.HasTracked:
		iconCol = colYellow
	}
	icon := colorizeBold("⎇", iconCol)

	arrows := ""
	if ri.Ahead > 0 {
		arrows += " " + colorize(fmt.Sprintf("↑%d", ri.Ahead), colGreen)
	}
	if ri.Behind > 0 {
		arrows += " " + colorize(fmt.Sprintf("↓%d", ri.Behind), colRed)
	}

	return fmt.Sprintf("%s on %s %s%s", ri.Project, icon, shorten(ri.Branch, maxBranchLen), arrows)
}

func readCwd(r io.Reader) string {
	var in input
	b, _ := io.ReadAll(r)
	if len(b) == 0 {
		return ""
	}
	_ = json.Unmarshal(b, &in)
	return strings.TrimSpace(in.Cwd)
}

func git(dir string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	return strings.TrimSpace(out.String())
}

func colorize(s, col string) string {
	if os.Getenv("STATUSLINE_NO_COLOR") == "1" {
		return s
	}
	return esc + "[" + col + "m" + s + esc + "[0m"
}

func colorizeBold(s, col string) string {
	if os.Getenv("STATUSLINE_NO_COLOR") == "1" {
		return s
	}
	return esc + "[1;" + col + "m" + s + esc + "[0m"
}

func parseStatus(s string) (branch string, ahead, behind int, hasTracked, hasUntracked bool) {
	for ln := range strings.SplitSeq(s, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		if strings.HasPrefix(ln, "#") {
			switch {
			case strings.HasPrefix(ln, "# branch.head "):
				branch = strings.TrimSpace(strings.TrimPrefix(ln, "# branch.head"))
			case strings.HasPrefix(ln, "# branch.ab "):
				f := strings.Fields(strings.TrimPrefix(ln, "# branch.ab"))
				if len(f) == 2 {
					ahead = parseSigned(f[0])
					behind = parseSigned(f[1])
				}
			}
			continue
		}
		if strings.HasPrefix(ln, "? ") {
			hasUntracked = true
			continue
		}
		if strings.HasPrefix(ln, "1 ") || strings.HasPrefix(ln, "2 ") {
			hasTracked = true
		}
	}
	return
}

func parseSigned(s string) (n int) {
	if s == "" {
		return 0
	}
	if s[0] == '+' || s[0] == '-' {
		s = s[1:]
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return
}

func shorten(s string, maxLen int) string {
	if maxLen <= 4 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getFetchInterval() time.Duration {
	if s := os.Getenv("STATUSLINE_FETCH_INTERVAL"); s != "" {
		if minutes, err := strconv.Atoi(s); err == nil && minutes > 0 {
			return time.Duration(minutes) * time.Minute
		}
	}
	return 30 * time.Minute
}

func shouldFetch(root string) bool {
	interval := getFetchInterval()

	up := git(root, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if up == "" {
		return true
	}

	reflog := git(root, "reflog", "show", "--date=unix", up, "-1")
	if reflog == "" {
		return true
	}

	fields := strings.Fields(reflog)
	for i, field := range fields {
		if strings.Contains(field, "fetch") && i > 0 {
			if timestamp, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
				lastFetch := time.Unix(timestamp, 0)
				return time.Since(lastFetch) >= interval
			}
		}
	}

	return true
}
