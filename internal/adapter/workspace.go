package adapter

import (
	"fmt"
	"strings"

	"github.com/cloudboy-jh/bento-diffs/internal/parser"
	"github.com/cloudboy-jh/bento-diffs/internal/renderer"
	"github.com/cloudboy-jh/bentotui/theme"
)

type Layout int

const (
	LayoutSplit Layout = iota
	LayoutStacked
)

type HeaderDTO struct {
	File      string
	Additions int
	Removals  int
	Layout    Layout
}

type FileRailItemDTO struct {
	Name     string
	Stats    string
	Selected bool
	Index    int
}

type FileRailDTO struct {
	Title      string
	Items      []FileRailItemDTO
	ActiveFile int
}

type MainDiffPaneDTO struct {
	FileName string
	Lines    []string
}

type FooterCardDTO struct {
	Command string
	Label   string
	Enabled bool
}

type FooterStatusDTO struct {
	Cards []FooterCardDTO
}

type WorkspaceDTO struct {
	Header HeaderDTO
	Rail   FileRailDTO
	Main   MainDiffPaneDTO
	Footer FooterStatusDTO
}

type RenderOptions struct {
	Width           int
	Layout          Layout
	Theme           theme.Theme
	SyntaxEnabled   bool
	ShowLineNumbers bool
	FilterQuery     string
}

type WorkspaceAdapter struct {
	diffs []parser.DiffResult
}

func NewWorkspaceAdapter(diffs []parser.DiffResult) *WorkspaceAdapter {
	return &WorkspaceAdapter{diffs: append([]parser.DiffResult{}, diffs...)}
}

func (a *WorkspaceAdapter) FileCount() int {
	return len(a.diffs)
}

func (a *WorkspaceAdapter) Build(activeFile int, opts RenderOptions) WorkspaceDTO {
	count := len(a.diffs)
	if count == 0 {
		return WorkspaceDTO{
			Header: HeaderDTO{Layout: opts.Layout},
			Rail:   FileRailDTO{Title: "Changed files"},
			Footer: FooterStatusDTO{Cards: []FooterCardDTO{{Command: "q", Label: "quit", Enabled: true}}},
		}
	}

	visible := a.FilteredIndices(opts.FilterQuery)
	if len(visible) == 0 {
		return WorkspaceDTO{
			Header: HeaderDTO{Layout: opts.Layout},
			Rail:   FileRailDTO{Title: "Changed files", ActiveFile: 0},
			Main:   MainDiffPaneDTO{Lines: nil},
			Footer: FooterStatusDTO{Cards: []FooterCardDTO{
				{Command: "/", Label: "filter", Enabled: true},
				{Command: "esc", Label: "clear", Enabled: strings.TrimSpace(opts.FilterQuery) != ""},
				{Command: "q", Label: "quit", Enabled: true},
			}},
		}
	}

	if activeFile < 0 || activeFile >= count || !containsIndex(visible, activeFile) {
		activeFile = visible[0]
	}

	active := a.diffs[activeFile]
	name := fileName(active)

	railItems := make([]FileRailItemDTO, 0, len(visible))
	visibleActive := 0
	for i, src := range visible {
		d := a.diffs[src]
		railItems = append(railItems, FileRailItemDTO{
			Name:     fileName(d),
			Stats:    fmt.Sprintf("+%d -%d", d.Additions, d.Removals),
			Selected: src == activeFile,
			Index:    src,
		})
		if src == activeFile {
			visibleActive = i
		}
	}

	lines := make([]string, 0)
	if opts.Layout == LayoutStacked {
		lines = append(lines, renderer.RenderUnifiedDiff(active, opts.Width, name, opts.Theme, opts.SyntaxEnabled, opts.ShowLineNumbers)...)
	} else {
		lines = append(lines, renderer.RenderSideBySideDiff(active, opts.Width, name, opts.Theme, opts.SyntaxEnabled, opts.ShowLineNumbers)...)
	}

	return WorkspaceDTO{
		Header: HeaderDTO{
			File:      name,
			Additions: active.Additions,
			Removals:  active.Removals,
			Layout:    opts.Layout,
		},
		Rail: FileRailDTO{
			Title:      "Changed files",
			Items:      railItems,
			ActiveFile: visibleActive,
		},
		Main: MainDiffPaneDTO{
			FileName: name,
			Lines:    lines,
		},
		Footer: FooterStatusDTO{Cards: []FooterCardDTO{
			{Command: "j/k", Label: "scroll", Enabled: true},
			{Command: "tab", Label: "layout", Enabled: true},
			{Command: "[/]", Label: "next/prev", Enabled: len(visible) > 1},
			{Command: "/", Label: "filter", Enabled: true},
			{Command: "ctrl+k", Label: "palette", Enabled: true},
			{Command: "q", Label: "quit", Enabled: true},
		}},
	}
}

func (a *WorkspaceAdapter) FilteredIndices(query string) []int {
	query = strings.TrimSpace(strings.ToLower(query))
	indices := make([]int, 0, len(a.diffs))
	for i, d := range a.diffs {
		name := strings.ToLower(fileName(d))
		if query == "" || strings.Contains(name, query) {
			indices = append(indices, i)
		}
	}
	return indices
}

func containsIndex(items []int, target int) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func fileName(d parser.DiffResult) string {
	if d.DisplayFile != "" {
		return d.DisplayFile
	}
	if d.NewFile != "" {
		return d.NewFile
	}
	return d.OldFile
}
