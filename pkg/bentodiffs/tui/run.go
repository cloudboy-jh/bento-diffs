package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bentotui/theme"
)

type Options struct {
	ThemeName       string
	Layout          string
	SyntaxEnabled   bool
	ShowLineNumbers bool
}

func Run(opts Options) error {
	if opts.Layout == "" {
		opts.Layout = "split"
	}
	if opts.ThemeName == "" {
		opts.ThemeName = "catppuccin-mocha"
	}
	if _, err := theme.SetTheme(opts.ThemeName); err != nil {
		return fmt.Errorf("unknown theme %q (available themes: %s)", opts.ThemeName, strings.Join(theme.AvailableThemes(), ", "))
	}

	m := newModel(opts, theme.CurrentTheme())
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
