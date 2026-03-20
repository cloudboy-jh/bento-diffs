package tui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	bentodiffs "github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs"
)

type repoItem struct {
	Name string
	Path string
}

type commitItem struct {
	SHA     string
	Short   string
	Date    string
	Subject string
}

func discoverRepos(roots []string) ([]repoItem, error) {
	items := make([]repoItem, 0)
	seen := map[string]struct{}{}

	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			if d.Name() == ".git" {
				repo := filepath.Dir(path)
				if _, ok := seen[repo]; !ok {
					seen[repo] = struct{}{}
					items = append(items, repoItem{Name: filepath.Base(repo), Path: repo})
				}
				return filepath.SkipDir
			}
			if strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			continue
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Name == items[j].Name {
			return items[i].Path < items[j].Path
		}
		return items[i].Name < items[j].Name
	})
	return items, nil
}

func loadCommits(repoPath string, limit int) ([]commitItem, error) {
	if limit <= 0 {
		limit = 200
	}
	cmd := exec.Command("git", "-C", repoPath, "log", fmt.Sprintf("-%d", limit), "--date=short", "--pretty=format:%H%x1f%h%x1f%ad%x1f%s")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	lines := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
	items := make([]commitItem, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(string(line), "\x1f", 4)
		if len(parts) < 4 {
			continue
		}
		items = append(items, commitItem{SHA: parts[0], Short: parts[1], Date: parts[2], Subject: parts[3]})
	}
	return items, nil
}

func loadCommitDiffs(repoPath, sha string) ([]bentodiffs.DiffResult, error) {
	cmd := exec.Command("git", "-C", repoPath, "show", sha, "--patch", "--no-color", "--pretty=format:")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show: %w", err)
	}
	patch := string(out)
	diffs, err := bentodiffs.ParseUnifiedDiffs(patch)
	if err != nil {
		return nil, fmt.Errorf("parse patch: %w", err)
	}
	if len(diffs) == 0 {
		return nil, fmt.Errorf("no file diffs found for commit %s", sha)
	}
	return diffs, nil
}
