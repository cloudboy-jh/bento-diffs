package fileheader

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/cloudboy-jh/bento-diffs/bricks/diffpane"
	"github.com/cloudboy-jh/bentotui/theme"
	"github.com/cloudboy-jh/bentotui/theme/styles"
)

type Model struct {
	width     int
	height    int
	theme     theme.Theme
	file      string
	additions int
	removals  int
	layout    diffpane.Layout
}

func New(t theme.Theme) *Model {
	return &Model{theme: t, height: 1}
}

func (m *Model) SetTheme(t theme.Theme) { m.theme = t }

func (m *Model) SetSize(width, height int) {
	m.width = width
	if height > 0 {
		m.height = height
	}
}

func (m *Model) SetFile(name string, additions, removals int) {
	m.file = name
	m.additions = additions
	m.removals = removals
}

func (m *Model) SetLayout(layout diffpane.Layout) {
	m.layout = layout
}

func (m *Model) View() tea.View {
	if m.width <= 0 {
		return tea.NewView("")
	}

	layout := "split"
	if m.layout == diffpane.LayoutStacked {
		layout = "stacked"
	}
	left := lipgloss.NewStyle().Foreground(m.theme.Text()).Bold(true).Render(m.file)
	stats := lipgloss.NewStyle().Foreground(m.theme.DiffAdded()).Render(fmt.Sprintf("+%d", m.additions)) +
		" " +
		lipgloss.NewStyle().Foreground(m.theme.DiffRemoved()).Render(fmt.Sprintf("-%d", m.removals))
	badge := lipgloss.NewStyle().Foreground(m.theme.TextMuted()).Render("[" + layout + "]")

	line := left + "  " + stats + "  " + badge
	line = ansi.Truncate(line, m.width, "...")
	row := styles.Row(m.theme.DiffContextBG(), m.theme.Text(), m.width, line)
	return tea.NewView(row)
}
