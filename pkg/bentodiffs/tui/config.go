package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type config struct {
	RepoRoots []string `json:"repo_roots"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(dir, "bentodiffs", "config.json"), nil
}

func loadConfig() (config, string, error) {
	path, err := configPath()
	if err != nil {
		return config{}, "", err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config{}, path, nil
		}
		return config{}, path, fmt.Errorf("read config: %w", err)
	}
	var c config
	if err := json.Unmarshal(b, &c); err != nil {
		return config{}, path, fmt.Errorf("parse config: %w", err)
	}
	c.RepoRoots = normalizeRoots(c.RepoRoots)
	return c, path, nil
}

func normalizeRoots(roots []string) []string {
	out := make([]string, 0, len(roots))
	seen := map[string]struct{}{}
	for _, r := range roots {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		abs, err := filepath.Abs(r)
		if err != nil {
			continue
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		out = append(out, abs)
	}
	return out
}
