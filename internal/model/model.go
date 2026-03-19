package model

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bento-diffs/bricks/diffpane"
	"github.com/cloudboy-jh/bento-diffs/bricks/fileheader"
	"github.com/cloudboy-jh/bento-diffs/internal/adapter"
	"github.com/cloudboy-jh/bento-diffs/recipes/commandpaletteflow"
	"github.com/cloudboy-jh/bento-diffs/recipes/emptystatepane"
	"github.com/cloudboy-jh/bento-diffs/recipes/filterbar"
	"github.com/cloudboy-jh/bentotui/registry/bricks/bar"
	"github.com/cloudboy-jh/bentotui/registry/bricks/dialog"
	"github.com/cloudboy-jh/bentotui/registry/bricks/surface"
	"github.com/cloudboy-jh/bentotui/registry/rooms"
	"github.com/cloudboy-jh/bentotui/theme"
)

type Layout int

const (
	LayoutSplit Layout = iota
	LayoutStacked
)

type Options struct {
	Layout          Layout
	SyntaxEnabled   bool
	ShowLineNumbers bool
	UseTTYInput     bool
}

type app struct {
	program *tea.Program
}

type toggleLayoutMsg struct{}
type nextFileMsg struct{}
type prevFileMsg struct{}
type startFilterMsg struct{}
type clearFilterMsg struct{}

func New(workspace *adapter.WorkspaceAdapter, t theme.Theme, opts Options) *app {
	m := &model{
		theme:     t,
		workspace: workspace,
		layout:    opts.Layout,
		syntax:    opts.SyntaxEnabled,
		lineNums:  opts.ShowLineNumbers,
		keys:      defaultKeyMap(),
	}
	m.initWidgets()

	programOpts := []tea.ProgramOption{}
	if opts.UseTTYInput {
		if tty, err := os.Open("CONIN$"); err == nil {
			programOpts = append(programOpts, tea.WithInput(tty))
		}
	}
	p := tea.NewProgram(m, programOpts...)
	return &app{program: p}
}

func (a *app) Run() error {
	_, err := a.program.Run()
	return err
}

type model struct {
	theme      theme.Theme
	workspace  *adapter.WorkspaceAdapter
	activeFile int
	layout     Layout
	syntax     bool
	lineNums   bool

	fileHeader *fileheader.Model
	diffPane   *diffpane.Model
	emptyPane  *emptystatepane.Model
	filterBar  *filterbar.Model
	footer     *bar.Model
	dialogs    *dialog.Manager

	width       int
	height      int
	keys        keyMap
	filterQuery string
	filterMode  bool
	showEmpty   bool
	visible     []int
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch mm := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = mm.Width
		m.height = mm.Height
		m.resize()
		m.refreshWorkspace(false)
		m.dialogs.SetSize(mm.Width, mm.Height)
	case dialog.OpenMsg, dialog.CloseMsg:
		u, cmd := m.dialogs.Update(mm)
		m.dialogs = u.(*dialog.Manager)
		return m, cmd
	case toggleLayoutMsg:
		m.toggleLayout()
	case nextFileMsg:
		m.nextFile()
	case prevFileMsg:
		m.prevFile()
	case startFilterMsg:
		m.filterMode = true
		m.filterBar.Input.SetValue(m.filterQuery)
		return m, m.filterBar.Focus()
	case clearFilterMsg:
		m.filterMode = false
		m.filterQuery = ""
		m.filterBar.Input.SetValue("")
		m.filterBar.Input.Blur()
		m.refreshWorkspace(true)
	case tea.MouseMsg:
		if m.filterMode || m.dialogs.IsOpen() {
			return m, nil
		}
		mouse := mm.Mouse()
		switch mouse.Button {
		case tea.MouseWheelDown:
			m.diffPane.ScrollDown(3)
		case tea.MouseWheelUp:
			m.diffPane.ScrollUp(3)
		}
	case tea.KeyMsg:
		if m.dialogs.IsOpen() {
			u, cmd := m.dialogs.Update(mm)
			m.dialogs = u.(*dialog.Manager)
			return m, cmd
		}

		if m.filterMode {
			switch {
			case keyMatch(mm, m.keys.Clear):
				m.filterMode = false
				m.filterQuery = ""
				m.filterBar.Input.SetValue("")
				m.filterBar.Input.Blur()
				m.refreshWorkspace(true)
				return m, nil
			case keyMatch(mm, m.keys.Apply):
				m.filterMode = false
				m.filterQuery = strings.TrimSpace(m.filterBar.Input.Value())
				m.filterBar.Input.Blur()
				m.refreshWorkspace(true)
				return m, nil
			}

			_, cmd := m.filterBar.Input.Update(mm)
			m.filterQuery = strings.TrimSpace(m.filterBar.Input.Value())
			m.refreshWorkspace(false)
			return m, cmd
		}

		switch {
		case keyMatch(mm, m.keys.Quit):
			return m, tea.Quit
		case keyMatch(mm, m.keys.Down):
			m.diffPane.ScrollDown(1)
		case keyMatch(mm, m.keys.Up):
			m.diffPane.ScrollUp(1)
		case keyMatch(mm, m.keys.PageDown):
			m.diffPane.ScrollDown(max(1, m.contentHeight()/2))
		case keyMatch(mm, m.keys.PageUp):
			m.diffPane.ScrollUp(max(1, m.contentHeight()/2))
		case keyMatch(mm, m.keys.Toggle):
			m.toggleLayout()
		case keyMatch(mm, m.keys.NextFile):
			m.nextFile()
		case keyMatch(mm, m.keys.PrevFile):
			m.prevFile()
		case keyMatch(mm, m.keys.Filter):
			m.filterMode = true
			m.filterBar.Input.SetValue(m.filterQuery)
			return m, m.filterBar.Focus()
		case keyMatch(mm, m.keys.Palette):
			return m, m.openPalette()
		}
	case theme.ThemeChangedMsg:
		m.theme = mm.Theme
		m.fileHeader.SetTheme(mm.Theme)
		m.diffPane.SetTheme(mm.Theme)
		m.emptyPane.SetTheme(mm.Theme)
		m.filterBar.SetTheme(mm.Theme)
		m.footer.SetTheme(mm.Theme)
		m.dialogs.SetTheme(mm.Theme)
		m.refreshWorkspace(false)
	}

	return m, nil
}

func (m *model) View() tea.View {
	if m.width <= 0 || m.height <= 0 {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	header := m.headerView()
	mainPane := m.mainView()
	footer := m.footerView()

	content := &focusContent{header: header, main: mainPane}
	screen := rooms.Focus(m.width, m.height, content, footer)

	surf := surface.New(m.width, m.height)
	surf.Fill(m.theme.Background())
	surf.Draw(0, 0, screen)
	if m.dialogs.IsOpen() {
		surf.DrawCenter(viewString(m.dialogs.View()))
	}

	v := tea.NewView(surf.Render())
	v.AltScreen = true
	v.BackgroundColor = m.theme.Background()
	return v
}

func (m *model) initWidgets() {
	m.fileHeader = fileheader.New(m.theme)
	m.diffPane = diffpane.New(m.theme)
	m.emptyPane = emptystatepane.New("No diff content", "Nothing to show yet.", m.theme)
	m.filterBar = filterbar.New("bento-diffs", m.theme)
	m.dialogs = dialog.New()
	m.dialogs.SetTheme(m.theme)
	m.footer = bar.New(
		bar.FooterAnchored(),
		bar.WithTheme(m.theme),
	)
	m.refreshWorkspace(true)
}

func (m *model) resize() {
	if m.fileHeader == nil {
		return
	}
	m.fileHeader.SetSize(m.width, 1)
	m.footer.SetSize(m.width, 1)
	m.filterBar.SetSize(m.width)
	diffWidth := max(1, m.width)
	m.diffPane.SetSize(diffWidth, m.contentHeight())
	m.emptyPane.SetSize(diffWidth, m.contentHeight())
}

func (m *model) toggleLayout() {
	if m.layout == LayoutSplit {
		m.layout = LayoutStacked
	} else {
		m.layout = LayoutSplit
	}
	m.resize()
	m.refreshWorkspace(true)
}

func (m *model) setActiveVisible(index int) {
	if len(m.visible) == 0 {
		m.activeFile = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(m.visible) {
		index = len(m.visible) - 1
	}
	m.activeFile = m.visible[index]
	m.refreshWorkspace(true)
}

func (m *model) nextFile() {
	if len(m.visible) == 0 {
		return
	}
	idx := m.indexOfActiveVisible()
	if idx < 0 {
		idx = 0
	} else {
		idx = (idx + 1) % len(m.visible)
	}
	m.activeFile = m.visible[idx]
	m.refreshWorkspace(true)
}

func (m *model) prevFile() {
	if len(m.visible) == 0 {
		return
	}
	idx := m.indexOfActiveVisible()
	if idx < 0 {
		idx = 0
	} else {
		idx--
		if idx < 0 {
			idx = len(m.visible) - 1
		}
	}
	m.activeFile = m.visible[idx]
	m.refreshWorkspace(true)
}

func (m *model) indexOfActiveVisible() int {
	for i, v := range m.visible {
		if v == m.activeFile {
			return i
		}
	}
	return -1
}

func (m *model) contentHeight() int {
	h := m.height - 2
	if h < 1 {
		return 1
	}
	return h
}

func toPaneLayout(l Layout) diffpane.Layout {
	if l == LayoutStacked {
		return diffpane.LayoutStacked
	}
	return diffpane.LayoutSplit
}

func toWorkspaceLayout(l Layout) adapter.Layout {
	if l == LayoutStacked {
		return adapter.LayoutStacked
	}
	return adapter.LayoutSplit
}

func (m *model) refreshWorkspace(resetScroll bool) {
	if m.workspace == nil {
		return
	}

	diffWidth := m.width
	if diffWidth < 1 {
		diffWidth = 1
	}

	ws := m.workspace.Build(m.activeFile, adapter.RenderOptions{
		Width:           diffWidth,
		Layout:          toWorkspaceLayout(m.layout),
		Theme:           m.theme,
		SyntaxEnabled:   m.syntax,
		ShowLineNumbers: m.lineNums,
		FilterQuery:     m.filterQuery,
	})

	m.fileHeader.SetLayout(toPaneLayout(m.layout))
	m.fileHeader.SetFile(ws.Header.File, ws.Header.Additions, ws.Header.Removals)

	m.visible = m.visible[:0]
	for _, item := range ws.Rail.Items {
		m.visible = append(m.visible, item.Index)
	}
	if ws.Rail.ActiveFile >= 0 && ws.Rail.ActiveFile < len(m.visible) {
		m.activeFile = m.visible[ws.Rail.ActiveFile]
	}

	m.diffPane.SetLines(ws.Main.Lines, resetScroll)
	m.showEmpty = len(ws.Main.Lines) == 0
	if m.showEmpty {
		switch {
		case m.workspace.FileCount() == 0:
			m.emptyPane.SetTitle("No diffs loaded")
			m.emptyPane.SetMessage("Pipe a diff into bento-diffs, pass --patch, or use two files.")
		case strings.TrimSpace(m.filterQuery) != "" && len(ws.Rail.Items) == 0:
			m.emptyPane.SetTitle("No files match")
			m.emptyPane.SetMessage("Update the filter query or press Esc to clear it.")
		default:
			m.emptyPane.SetTitle("No diff content")
			m.emptyPane.SetMessage("The selected file has no renderable hunks.")
		}
	}

	cards := make([]bar.Card, 0, len(ws.Footer.Cards))
	for _, c := range ws.Footer.Cards {
		cards = append(cards, bar.Card{Command: c.Command, Label: c.Label, Enabled: c.Enabled})
	}
	m.footer.SetCards(cards)
}

func keyMatch(msg tea.KeyMsg, binding interface{ Keys() []string }) bool {
	for _, k := range binding.Keys() {
		if msg.String() == k {
			return true
		}
	}
	return false
}

func (m *model) openPalette() tea.Cmd {
	commands := []dialog.Command{
		{Label: "Toggle layout", Group: "View", Keybind: "tab", Action: func() tea.Msg { return toggleLayoutMsg{} }},
		{Label: "Next file", Group: "Navigate", Keybind: "]", Action: func() tea.Msg { return nextFileMsg{} }},
		{Label: "Previous file", Group: "Navigate", Keybind: "[", Action: func() tea.Msg { return prevFileMsg{} }},
		{Label: "Filter files", Group: "Search", Keybind: "/", Action: func() tea.Msg { return startFilterMsg{} }},
		{Label: "Clear filter", Group: "Search", Keybind: "esc", Action: func() tea.Msg { return clearFilterMsg{} }},
		{Label: "Quit", Group: "App", Keybind: "q", Action: func() tea.Msg { return tea.Quit() }},
	}
	return commandpaletteflow.Open(commands)
}

func (m *model) headerView() rooms.Sizable {
	if m.filterMode || m.filterQuery != "" {
		return m.filterBar.Input
	}
	return m.fileHeader
}

type focusContent struct {
	header rooms.Sizable
	main   rooms.Sizable
}

func (c *focusContent) SetSize(width, height int) {
	if c.header != nil {
		c.header.SetSize(width, 1)
	}
	if c.main != nil {
		h := height - 1
		if h < 1 {
			h = 1
		}
		c.main.SetSize(width, h)
	}
}

func (c *focusContent) View() tea.View {
	head := ""
	if c.header != nil {
		head = viewString(c.header.View())
	}
	body := ""
	if c.main != nil {
		body = viewString(c.main.View())
	}
	if head == "" {
		return tea.NewView(body)
	}
	if body == "" {
		return tea.NewView(head)
	}
	return tea.NewView(head + "\n" + body)
}

func (m *model) mainView() rooms.Sizable {
	if m.showEmpty {
		return m.emptyPane
	}
	return m.diffPane
}

func (m *model) footerView() rooms.Sizable {
	if m.filterMode {
		return m.filterBar.Footer
	}
	return m.footer
}

func viewString(v tea.View) string {
	if v.Content == nil {
		return ""
	}
	if r, ok := v.Content.(interface{ Render() string }); ok {
		return r.Render()
	}
	if s, ok := v.Content.(fmt.Stringer); ok {
		return s.String()
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
