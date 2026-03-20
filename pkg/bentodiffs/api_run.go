package bentodiffs

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs/parser"
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
	return runProgram(diffs, opts)
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
	return runProgram(diffs, opts)
}

func RunFiles(before, after, filename string, context int, opts Options) error {
	patch, _, _ := parser.GenerateDiff(before, after, filename, context)
	return RunPatch(patch, filename, opts)
}

func runProgram(diffs []DiffResult, opts Options) error {
	if opts.Layout != "split" && opts.Layout != "stacked" {
		return fmt.Errorf("--layout must be split or stacked")
	}

	if _, err := theme.SetTheme(opts.ThemeName); err != nil {
		return fmt.Errorf("unknown theme %q (available themes: %s)", opts.ThemeName, strings.Join(theme.AvailableThemes(), ", "))
	}

	v := NewViewer(ViewerOptions{
		Diffs:           diffs,
		Layout:          opts.Layout,
		SyntaxEnabled:   opts.SyntaxEnabled,
		ShowLineNumbers: opts.ShowLineNumbers,
		Theme:           theme.CurrentTheme(),
	})

	programOpts := []tea.ProgramOption{}
	if opts.UseTTYInput {
		if tty, err := os.Open("CONIN$"); err == nil {
			programOpts = append(programOpts, tea.WithInput(tty))
		}
	}
	p := tea.NewProgram(v.TeaModel(), programOpts...)
	_, err := p.Run()
	return err
}
