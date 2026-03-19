package filterbar

import (
	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bentotui/registry/bricks/bar"
	"github.com/cloudboy-jh/bentotui/registry/bricks/input"
	"github.com/cloudboy-jh/bentotui/theme"
)

type Model struct {
	Input  *input.Model
	Footer *bar.Model
}

func New(appName string, t theme.Theme) *Model {
	if t == nil {
		t = theme.CurrentTheme()
	}

	inp := input.New()
	inp.SetPlaceholder("Filter files...")
	inp.SetTheme(t)

	footer := bar.New(
		bar.FooterAnchored(),
		bar.Left("~ "+appName),
		bar.Cards(
			bar.Card{Command: "enter", Label: "apply", Variant: bar.CardPrimary, Enabled: true, Priority: 3},
			bar.Card{Command: "esc", Label: "clear", Variant: bar.CardMuted, Enabled: true, Priority: 2},
		),
		bar.CompactCards(),
		bar.WithTheme(t),
	)

	return &Model{Input: inp, Footer: footer}
}

func (m *Model) SetSize(width int) {
	m.Input.SetSize(max(1, width-2), 1)
	m.Footer.SetSize(width, 1)
}

func (m *Model) Focus() tea.Cmd { return m.Input.Focus() }

func (m *Model) SetTheme(t theme.Theme) {
	m.Input.SetTheme(t)
	m.Footer.SetTheme(t)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
