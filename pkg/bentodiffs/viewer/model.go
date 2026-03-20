package viewer

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bento-diffs/bricks/diffpane"
	"github.com/cloudboy-jh/bento-diffs/bricks/fileheader"
	"github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs/parser"
	"github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs/workspace"
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
	// LayoutSplit renders two side-by-side columns.
	LayoutSplit Layout = iota
	// LayoutStacked renders a single unified column.
	LayoutStacked
)

// Options configures viewer model behavior.
type Options struct {
	Layout          Layout
	SyntaxEnabled   bool
	ShowLineNumbers bool
}

// State is a readonly snapshot of model state.
type State struct {
	Width      int
	Height     int
	Layout     Layout
	ActiveFile int
	Visible    []int
	FileCount  int
	Scroll     int
	MaxScroll  int
	HunkStarts []int
	ActiveHunk int
	FilterMode bool
	Filter     string
	ShowEmpty  bool
}

type toggleLayoutMsg struct{}
type nextFileMsg struct{}
type prevFileMsg struct{}
type nextHunkMsg struct{}
type prevHunkMsg struct{}
type startFilterMsg struct{}
type clearFilterMsg struct{}
type openThemePickerMsg struct{}

type Model struct {
	theme      theme.Theme
	workspace  *workspace.Adapter
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
	hunkStarts  []int
}

// New builds a viewer model from parsed diffs.
func New(diffs []parser.DiffResult, t theme.Theme, opts Options) *Model {
	m := &Model{
		theme:     t,
		workspace: workspace.New(diffs),
		layout:    opts.Layout,
		syntax:    opts.SyntaxEnabled,
		lineNums:  opts.ShowLineNumbers,
		keys:      defaultKeyMap(),
	}
	m.initWidgets()
	return m
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch mm := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(mm.Width, mm.Height)
		m.dialogs.SetSize(mm.Width, mm.Height)
	case dialog.OpenMsg, dialog.CloseMsg:
		u, cmd := m.dialogs.Update(mm)
		m.dialogs = u.(*dialog.Manager)
		return m, cmd
	case toggleLayoutMsg:
		m.toggleLayout()
	case nextFileMsg:
		m.NextFile()
	case prevFileMsg:
		m.PrevFile()
	case nextHunkMsg:
		m.NextHunk()
	case prevHunkMsg:
		m.PrevHunk()
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
	case openThemePickerMsg:
		return m, openThemePicker()
	case tea.MouseMsg:
		if m.filterMode || m.dialogs.IsOpen() {
			return m, nil
		}
		mouse := mm.Mouse()
		switch mouse.Button {
		case tea.MouseWheelDown:
			m.ScrollDown(3)
		case tea.MouseWheelUp:
			m.ScrollUp(3)
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
			m.ScrollDown(1)
		case keyMatch(mm, m.keys.Up):
			m.ScrollUp(1)
		case keyMatch(mm, m.keys.PageDown):
			m.ScrollDown(max(1, m.contentHeight()/2))
		case keyMatch(mm, m.keys.PageUp):
			m.ScrollUp(max(1, m.contentHeight()/2))
		case keyMatch(mm, m.keys.Toggle):
			m.toggleLayout()
		case keyMatch(mm, m.keys.NextFile):
			m.NextFile()
		case keyMatch(mm, m.keys.PrevFile):
			m.PrevFile()
		case keyMatch(mm, m.keys.NextHunk):
			m.NextHunk()
		case keyMatch(mm, m.keys.PrevHunk):
			m.PrevHunk()
		case keyMatch(mm, m.keys.Filter):
			m.filterMode = true
			m.filterBar.Input.SetValue(m.filterQuery)
			return m, m.filterBar.Focus()
		case keyMatch(mm, m.keys.Palette):
			return m, m.openPalette()
		}
	case theme.ThemeChangedMsg:
		m.SetTheme(mm.Theme)
	}

	return m, nil
}

// View implements tea.Model.
func (m *Model) View() tea.View {
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

// Render returns the current view as a string.
func (m *Model) Render() string {
	return viewString(m.View())
}

// SetSize updates viewport dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.resize()
	m.refreshWorkspace(false)
}

// SetTheme applies a new theme and rerenders.
func (m *Model) SetTheme(t theme.Theme) {
	m.theme = t
	m.fileHeader.SetTheme(t)
	m.diffPane.SetTheme(t)
	m.emptyPane.SetTheme(t)
	m.filterBar.SetTheme(t)
	m.footer.SetTheme(t)
	m.dialogs.SetTheme(t)
	m.refreshWorkspace(false)
}

// SetDiffs replaces the loaded diff dataset.
func (m *Model) SetDiffs(diffs []parser.DiffResult) {
	m.workspace.SetDiffs(diffs)
	if m.activeFile >= m.workspace.FileCount() {
		m.activeFile = max(0, m.workspace.FileCount()-1)
	}
	m.refreshWorkspace(true)
}

// SetFileIndex selects the active file by absolute dataset index.
func (m *Model) SetFileIndex(index int) {
	if m.workspace.FileCount() == 0 {
		m.activeFile = 0
		m.refreshWorkspace(true)
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= m.workspace.FileCount() {
		index = m.workspace.FileCount() - 1
	}
	m.activeFile = index
	m.refreshWorkspace(true)
}

// NextFile selects the next visible file.
func (m *Model) NextFile() {
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

// PrevFile selects the previous visible file.
func (m *Model) PrevFile() {
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

// NextHunk jumps the scroll viewport to the next hunk.
func (m *Model) NextHunk() {
	if len(m.hunkStarts) == 0 {
		return
	}
	cur := m.activeHunkIndex()
	if cur+1 < len(m.hunkStarts) {
		m.diffPane.SetScroll(m.hunkStarts[cur+1])
	}
}

// PrevHunk jumps the scroll viewport to the previous hunk.
func (m *Model) PrevHunk() {
	if len(m.hunkStarts) == 0 {
		return
	}
	cur := m.activeHunkIndex()
	if cur > 0 {
		m.diffPane.SetScroll(m.hunkStarts[cur-1])
	}
}

// ScrollUp moves the viewport up by n lines.
func (m *Model) ScrollUp(n int) {
	m.diffPane.ScrollUp(n)
}

// ScrollDown moves the viewport down by n lines.
func (m *Model) ScrollDown(n int) {
	m.diffPane.ScrollDown(n)
}

// State returns a readonly snapshot of viewer state.
func (m *Model) State() State {
	visible := append([]int{}, m.visible...)
	hunkStarts := append([]int{}, m.hunkStarts...)
	return State{
		Width:      m.width,
		Height:     m.height,
		Layout:     m.layout,
		ActiveFile: m.activeFile,
		Visible:    visible,
		FileCount:  m.workspace.FileCount(),
		Scroll:     m.diffPane.Scroll(),
		MaxScroll:  m.diffPane.MaxScroll(),
		HunkStarts: hunkStarts,
		ActiveHunk: m.activeHunkIndex(),
		FilterMode: m.filterMode,
		Filter:     m.filterQuery,
		ShowEmpty:  m.showEmpty,
	}
}

func (m *Model) initWidgets() {
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

func (m *Model) resize() {
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

func (m *Model) toggleLayout() {
	if m.layout == LayoutSplit {
		m.layout = LayoutStacked
	} else {
		m.layout = LayoutSplit
	}
	m.resize()
	m.refreshWorkspace(true)
}

func (m *Model) indexOfActiveVisible() int {
	for i, v := range m.visible {
		if v == m.activeFile {
			return i
		}
	}
	return -1
}

func (m *Model) activeHunkIndex() int {
	if len(m.hunkStarts) == 0 {
		return 0
	}
	scroll := m.diffPane.Scroll()
	idx := 0
	for i := 0; i < len(m.hunkStarts); i++ {
		if m.hunkStarts[i] <= scroll {
			idx = i
			continue
		}
		break
	}
	return idx
}

func (m *Model) contentHeight() int {
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

func toWorkspaceLayout(l Layout) workspace.Layout {
	if l == LayoutStacked {
		return workspace.LayoutStacked
	}
	return workspace.LayoutSplit
}

func (m *Model) refreshWorkspace(resetScroll bool) {
	if m.workspace == nil {
		return
	}

	diffWidth := m.width
	if diffWidth < 1 {
		diffWidth = 1
	}

	ws := m.workspace.Build(m.activeFile, workspace.RenderOptions{
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

	m.hunkStarts = append(m.hunkStarts[:0], ws.Main.HunkStarts...)
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
	fileTotal := max(1, m.workspace.FileCount())
	hunkTotal := max(1, len(m.hunkStarts))
	m.footer.SetStatusPill(fmt.Sprintf("file %d/%d  hunk %d/%d", m.activeFile+1, fileTotal, m.activeHunkIndex()+1, hunkTotal))
}

func keyMatch(msg tea.KeyMsg, binding interface{ Keys() []string }) bool {
	for _, k := range binding.Keys() {
		if msg.String() == k {
			return true
		}
	}
	return false
}

func (m *Model) openPalette() tea.Cmd {
	commands := []dialog.Command{
		{Label: "Theme picker", Group: "View", Keybind: "", Action: func() tea.Msg { return openThemePickerMsg{} }},
		{Label: "Toggle layout", Group: "View", Keybind: "tab", Action: func() tea.Msg { return toggleLayoutMsg{} }},
		{Label: "Next file", Group: "Navigate", Keybind: "]", Action: func() tea.Msg { return nextFileMsg{} }},
		{Label: "Previous file", Group: "Navigate", Keybind: "[", Action: func() tea.Msg { return prevFileMsg{} }},
		{Label: "Next hunk", Group: "Navigate", Keybind: "n", Action: func() tea.Msg { return nextHunkMsg{} }},
		{Label: "Previous hunk", Group: "Navigate", Keybind: "N", Action: func() tea.Msg { return prevHunkMsg{} }},
		{Label: "Filter files", Group: "Search", Keybind: "/", Action: func() tea.Msg { return startFilterMsg{} }},
		{Label: "Clear filter", Group: "Search", Keybind: "esc", Action: func() tea.Msg { return clearFilterMsg{} }},
		{Label: "Quit", Group: "App", Keybind: "q", Action: func() tea.Msg { return tea.Quit() }},
	}
	return commandpaletteflow.Open(commands)
}

func openThemePicker() tea.Cmd {
	return func() tea.Msg {
		h := len(theme.AvailableThemes()) + 8
		return dialog.Open(dialog.Custom{
			DialogTitle: "Themes",
			Content:     dialog.NewThemePicker(),
			Width:       44,
			Height:      h,
		})
	}
}

func (m *Model) headerView() rooms.Sizable {
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

func (m *Model) mainView() rooms.Sizable {
	if m.showEmpty {
		return m.emptyPane
	}
	return m.diffPane
}

func (m *Model) footerView() rooms.Sizable {
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
