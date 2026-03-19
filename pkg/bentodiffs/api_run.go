package bentodiffs

import (
	"fmt"
	"strings"

	"github.com/cloudboy-jh/bento-diffs/internal/adapter"
	"github.com/cloudboy-jh/bento-diffs/internal/model"
	"github.com/cloudboy-jh/bento-diffs/internal/parser"
	"github.com/cloudboy-jh/bentotui/theme"
)

type Options struct {
	Layout          string
	ThemeName       string
	SyntaxEnabled   bool
	ShowLineNumbers bool
	UseTTYInput     bool
}

func DefaultOptions() Options {
	return Options{
		Layout:          "split",
		ThemeName:       "catppuccin-mocha",
		SyntaxEnabled:   true,
		ShowLineNumbers: true,
	}
}

func AvailableThemes() []string {
	return theme.AvailableThemes()
}

func RunDiffs(diffs []DiffResult, opts Options) error {
	internal := make([]parser.DiffResult, 0, len(diffs))
	for _, d := range diffs {
		internal = append(internal, toInternalDiffResult(d))
	}
	return runInternalDiffs(internal, opts)
}

func RunPatch(patch string, fileName string, opts Options) error {
	diffs, err := parser.ParseUnifiedDiffs(patch)
	if err != nil {
		return fmt.Errorf("parse diff: %w", err)
	}
	if len(diffs) == 0 {
		return fmt.Errorf("no file diffs found")
	}
	if len(diffs) == 1 && diffs[0].DisplayFile == "" && fileName != "" {
		diffs[0].DisplayFile = fileName
	}
	return runInternalDiffs(diffs, opts)
}

func RunFiles(before, after, filename string, context int, opts Options) error {
	patch, _, _ := parser.GenerateDiff(before, after, filename, context)
	return RunPatch(patch, filename, opts)
}

func runInternalDiffs(diffs []parser.DiffResult, opts Options) error {
	layout := model.LayoutSplit
	if opts.Layout == "stacked" {
		layout = model.LayoutStacked
	} else if opts.Layout != "split" {
		return fmt.Errorf("--layout must be split or stacked")
	}

	if _, err := theme.SetTheme(opts.ThemeName); err != nil {
		return fmt.Errorf("unknown theme %q (available themes: %s)", opts.ThemeName, strings.Join(theme.AvailableThemes(), ", "))
	}

	workspace := adapter.NewWorkspaceAdapter(diffs)
	app := model.New(workspace, theme.CurrentTheme(), model.Options{
		Layout:          layout,
		SyntaxEnabled:   opts.SyntaxEnabled,
		ShowLineNumbers: opts.ShowLineNumbers,
		UseTTYInput:     opts.UseTTYInput,
	})
	return app.Run()
}
