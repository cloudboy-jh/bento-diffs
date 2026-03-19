package diffpane

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bentotui/theme"
	"github.com/cloudboy-jh/bentotui/theme/styles"
)

type Layout int

const (
	LayoutSplit Layout = iota
	LayoutStacked
)

type Model struct {
	scroll    int
	maxScroll int
	width     int
	height    int
	theme     theme.Theme
	lines     []string
}

func New(t theme.Theme) *Model {
	return &Model{
		theme: t,
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) SetTheme(t theme.Theme) {
	m.theme = t
}

func (m *Model) SetLines(lines []string, resetScroll bool) {
	m.lines = append([]string{}, lines...)
	m.maxScroll = len(m.lines) - m.height
	if m.maxScroll < 0 {
		m.maxScroll = 0
	}
	if resetScroll {
		m.scroll = 0
	}
	if m.scroll > m.maxScroll {
		m.scroll = m.maxScroll
	}
	if m.scroll < 0 {
		m.scroll = 0
	}
}

func (m *Model) ScrollDown(n int) {
	m.scroll += n
	if m.scroll > m.maxScroll {
		m.scroll = m.maxScroll
	}
}

func (m *Model) ScrollUp(n int) {
	m.scroll -= n
	if m.scroll < 0 {
		m.scroll = 0
	}
}

func (m *Model) View() tea.View {
	if m.height <= 0 || m.width <= 0 {
		return tea.NewView("")
	}
	if len(m.lines) == 0 {
		blank := make([]string, m.height)
		for i := range blank {
			blank[i] = styles.Row(m.theme.DiffContextBG(), m.theme.Text(), m.width, "")
		}
		return tea.NewView(strings.Join(blank, "\n"))
	}

	start := m.scroll
	end := start + m.height
	if end > len(m.lines) {
		end = len(m.lines)
	}
	view := append([]string{}, m.lines[start:end]...)
	for len(view) < m.height {
		view = append(view, styles.Row(m.theme.DiffContextBG(), m.theme.Text(), m.width, ""))
	}
	return tea.NewView(strings.Join(view, "\n"))
}
