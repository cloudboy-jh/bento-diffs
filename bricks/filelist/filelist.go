package filelist

import (
	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bentotui/registry/bricks/list"
	"github.com/cloudboy-jh/bentotui/theme"
)

type Model struct {
	width  int
	height int
	theme  theme.Theme
	title  string
	files  []Entry
	active int
	inner  *list.Model
}

type Entry struct {
	Name  string
	Stats string
}

func New(files []Entry, t theme.Theme) *Model {
	inner := list.New(200)
	inner.SetTheme(t)
	m := &Model{
		theme: t,
		title: "Changed files",
		files: append([]Entry{}, files...),
		inner: inner,
	}
	m.rebuild()
	return m
}

func (m *Model) SetTheme(t theme.Theme) {
	m.theme = t
	m.inner.SetTheme(t)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.inner.SetSize(width, height)
}

func (m *Model) SetActive(index int) {
	if len(m.files) == 0 {
		m.active = 0
		m.inner.SetCursor(0)
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(m.files) {
		index = len(m.files) - 1
	}
	m.active = index
	m.inner.SetCursor(index + 1)
}

func (m *Model) SetTitle(title string) {
	if title == "" {
		title = "Changed files"
	}
	m.title = title
	m.rebuild()
}

func (m *Model) SetFiles(files []Entry) {
	m.files = append([]Entry{}, files...)
	if m.active >= len(m.files) {
		m.active = len(m.files) - 1
	}
	if m.active < 0 {
		m.active = 0
	}
	m.rebuild()
}

func (m *Model) Next() int {
	if len(m.files) == 0 {
		return 0
	}
	m.active = (m.active + 1) % len(m.files)
	m.inner.SetCursor(m.active + 1)
	return m.active
}

func (m *Model) Prev() int {
	if len(m.files) == 0 {
		return 0
	}
	m.active--
	if m.active < 0 {
		m.active = len(m.files) - 1
	}
	m.inner.SetCursor(m.active + 1)
	return m.active
}

func (m *Model) View() tea.View {
	if m.width <= 0 || m.height <= 0 {
		return tea.NewView("")
	}
	m.inner.SetCursor(m.active + 1)
	return m.inner.View()
}

func (m *Model) rebuild() {
	m.inner.Clear()
	m.inner.AppendSection(m.title)
	for _, f := range m.files {
		m.inner.AppendRow(list.Row{Kind: list.RowItem, Primary: f.Name, Label: f.Name, RightStat: f.Stats})
	}
	m.inner.SetCursor(m.active + 1)
}
