package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	bentodiffs "github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs"
	"github.com/cloudboy-jh/bento-diffs/recipes/commandpaletteflow"
	"github.com/cloudboy-jh/bentotui/registry/bricks/bar"
	"github.com/cloudboy-jh/bentotui/registry/bricks/dialog"
	"github.com/cloudboy-jh/bentotui/registry/bricks/input"
	"github.com/cloudboy-jh/bentotui/registry/bricks/list"
	"github.com/cloudboy-jh/bentotui/registry/bricks/surface"
	"github.com/cloudboy-jh/bentotui/registry/rooms"
	"github.com/cloudboy-jh/bentotui/theme"
)

type route int

const (
	routeHome route = iota
	routeCommits
	routeDiff
)

const homeWordmark = "" +
	"██████╗ ███████╗███╗   ██╗████████╗ ██████╗ ██████╗ ██╗███████╗███████╗███████╗\n" +
	"██╔══██╗██╔════╝████╗  ██║╚══██╔══╝██╔═══██╗██╔══██╗██║██╔════╝██╔════╝██╔════╝\n" +
	"██████╔╝█████╗  ██╔██╗ ██║   ██║   ██║   ██║██║  ██║██║█████╗  █████╗  █████╗  \n" +
	"██╔══██╗██╔══╝  ██║╚██╗██║   ██║   ██║   ██║██║  ██║██║██╔══╝  ██╔══╝  ██╔══╝  \n" +
	"██████╔╝███████╗██║ ╚████║   ██║   ╚██████╔╝██████╔╝██║██║     ██║     ██║     \n" +
	"╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝╚═╝     ╚═╝     ╚═╝     "

type reposLoadedMsg struct {
	repos []repoItem
	err   error
}

type commitsLoadedMsg struct {
	commits []commitItem
	err     error
}

type diffLoadedMsg struct {
	diffs []bentodiffs.DiffResult
	err   error
}

type openThemePickerMsg struct{}
type refreshReposMsg struct{}
type openRootsConfigMsg struct{}
type backToHomeMsg struct{}

const mockRepoPath = "mock://demo"

type model struct {
	theme theme.Theme
	opts  Options

	route  route
	width  int
	height int

	footer *bar.Model
	dialog *dialog.Manager

	configPath string
	roots      []string

	homeFilterMode bool
	homeInput      *input.Model
	homeList       *list.Model
	repos          []repoItem
	homeVisible    []repoItem
	homeIndex      int

	commitFilterMode bool
	commitInput      *input.Model
	commitList       *list.Model
	commits          []commitItem
	commitVisible    []commitItem
	commitIndex      int

	activeRepo repoItem

	viewer bentodiffs.Viewer

	loading bool
	loadErr string
}

func newModel(opts Options, t theme.Theme) *model {
	homeInput := input.New()
	homeInput.SetPlaceholder("Filter repos...")
	homeInput.SetTheme(t)

	commitInput := input.New()
	commitInput.SetPlaceholder("Filter commits...")
	commitInput.SetTheme(t)

	homeList := list.New(1000)
	homeList.SetTheme(t)
	homeList.SetDensity(list.DensityCompact)

	commitList := list.New(1000)
	commitList.SetTheme(t)
	commitList.SetDensity(list.DensityCompact)

	footer := bar.New(bar.FooterAnchored(), bar.CompactCards(), bar.WithTheme(t))

	d := dialog.New()
	d.SetTheme(t)

	return &model{
		theme:       t,
		opts:        opts,
		route:       routeHome,
		footer:      footer,
		dialog:      d,
		homeInput:   homeInput,
		homeList:    homeList,
		commitInput: commitInput,
		commitList:  commitList,
	}
}

func (m *model) Init() tea.Cmd {
	return m.loadReposCmd()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dialog.IsOpen() {
		u, cmd := m.dialog.Update(msg)
		m.dialog = u.(*dialog.Manager)
		if tc, ok := msg.(theme.ThemeChangedMsg); ok {
			m.applyTheme(tc.Theme)
		}
		return m, cmd
	}

	switch mm := msg.(type) {
	case dialog.OpenMsg, dialog.CloseMsg:
		u, cmd := m.dialog.Update(mm)
		m.dialog = u.(*dialog.Manager)
		return m, cmd
	case tea.WindowSizeMsg:
		m.width = mm.Width
		m.height = mm.Height
		m.footer.SetSize(mm.Width, 1)
		m.dialog.SetSize(mm.Width, mm.Height)
		m.resizeRouteWidgets()
		if m.viewer != nil {
			m.viewer.SetSize(mm.Width, mm.Height)
		}
		m.updateFooter()
		return m, nil
	case theme.ThemeChangedMsg:
		m.applyTheme(mm.Theme)
		return m, nil
	case reposLoadedMsg:
		m.loading = false
		if mm.err != nil {
			m.loadErr = mm.err.Error()
		} else {
			m.loadErr = ""
			m.repos = mm.repos
			m.filterRepos()
		}
		m.updateFooter()
		return m, nil
	case commitsLoadedMsg:
		m.loading = false
		if mm.err != nil {
			m.loadErr = mm.err.Error()
		} else {
			m.loadErr = ""
			m.commits = mm.commits
			m.filterCommits()
		}
		m.updateFooter()
		return m, nil
	case diffLoadedMsg:
		m.loading = false
		if mm.err != nil {
			m.loadErr = mm.err.Error()
			m.updateFooter()
			return m, nil
		}
		m.loadErr = ""
		m.viewer = bentodiffs.NewViewer(bentodiffs.ViewerOptions{
			Diffs:           mm.diffs,
			Layout:          m.opts.Layout,
			SyntaxEnabled:   m.opts.SyntaxEnabled,
			ShowLineNumbers: m.opts.ShowLineNumbers,
			Theme:           m.theme,
		})
		if m.width > 0 && m.height > 0 {
			m.viewer.SetSize(m.width, m.height)
		}
		m.route = routeDiff
		m.updateFooter()
		return m, nil
	case tea.KeyMsg:
		if mm.String() == "q" || mm.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if keyIs(mm, "ctrl+k", "p") && m.route != routeDiff {
			return m, m.openCommandPalette()
		}
		switch m.route {
		case routeHome:
			return m.updateHome(mm)
		case routeCommits:
			return m.updateCommits(mm)
		case routeDiff:
			if mm.String() == "esc" {
				m.route = routeCommits
				m.updateFooter()
				return m, nil
			}
			if m.viewer != nil {
				v, cmd := m.viewer.Update(mm)
				m.viewer = v
				return m, cmd
			}
		}
	case openThemePickerMsg:
		return m, m.openThemePicker()
	case refreshReposMsg:
		m.loadErr = ""
		m.loading = true
		m.updateFooter()
		return m, m.loadReposCmd()
	case openRootsConfigMsg:
		return m, m.openConfigDialog()
	case backToHomeMsg:
		m.route = routeHome
		m.updateFooter()
		return m, nil
	}

	if m.route == routeDiff && m.viewer != nil {
		v, cmd := m.viewer.Update(msg)
		m.viewer = v
		return m, cmd
	}

	return m, nil
}

func (m *model) View() tea.View {
	if m.route == routeDiff && m.viewer != nil {
		v := tea.NewView(m.viewer.View())
		v.AltScreen = true
		v.BackgroundColor = m.theme.Background()
		return v
	}

	if m.width <= 0 || m.height <= 0 {
		v := tea.NewView("")
		v.AltScreen = true
		v.BackgroundColor = m.theme.Background()
		return v
	}

	content := rooms.RenderFunc(func(width, height int) string {
		switch m.route {
		case routeCommits:
			return m.commitsBody(width, height)
		default:
			return m.homeBody(width, height)
		}
	})

	screen := rooms.Focus(m.width, m.height, content, m.footer)
	surf := surface.New(m.width, m.height)
	surf.Fill(m.theme.Background())
	surf.Draw(0, 0, screen)
	if m.dialog.IsOpen() {
		surf.DrawCenter(viewString(m.dialog.View()))
	}

	v := tea.NewView(surf.Render())
	v.AltScreen = true
	v.BackgroundColor = m.theme.Background()
	return v
}

func (m *model) updateHome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.homeFilterMode {
		switch msg.String() {
		case "esc":
			m.homeFilterMode = false
			m.homeInput.SetValue("")
			m.filterRepos()
			m.updateFooter()
			return m, nil
		case "enter":
			m.homeFilterMode = false
			m.updateFooter()
			return m, nil
		}
		u, cmd := m.homeInput.Update(msg)
		m.homeInput = u.(*input.Model)
		m.filterRepos()
		m.updateFooter()
		return m, cmd
	}

	switch msg.String() {
	case "/":
		m.homeFilterMode = true
		return m, m.homeInput.Focus()
	case "ctrl+r", "r":
		m.loadErr = ""
		m.loading = true
		m.updateFooter()
		return m, m.loadReposCmd()
	case "ctrl+comma", "ctrl+,", ",":
		return m, m.openConfigDialog()
	case "down", "j":
		m.homeIndex++
		if m.homeIndex >= len(m.homeVisible) {
			m.homeIndex = len(m.homeVisible) - 1
		}
		if m.homeIndex < 0 {
			m.homeIndex = 0
		}
		m.homeList.SetCursor(m.homeIndex)
	case "up", "k":
		m.homeIndex--
		if m.homeIndex < 0 {
			m.homeIndex = 0
		}
		m.homeList.SetCursor(m.homeIndex)
	case "enter":
		if len(m.homeVisible) == 0 {
			return m, nil
		}
		m.activeRepo = m.homeVisible[m.homeIndex]
		m.route = routeCommits
		m.loading = true
		m.commitIndex = 0
		m.commitInput.SetValue("")
		m.filterCommits()
		m.updateFooter()
		return m, m.loadCommitsCmd(m.activeRepo.Path)
	}
	m.updateFooter()
	return m, nil
}

func (m *model) updateCommits(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.commitFilterMode {
		switch msg.String() {
		case "esc":
			m.commitFilterMode = false
			m.commitInput.SetValue("")
			m.filterCommits()
			m.updateFooter()
			return m, nil
		case "enter":
			m.commitFilterMode = false
			m.updateFooter()
			return m, nil
		}
		u, cmd := m.commitInput.Update(msg)
		m.commitInput = u.(*input.Model)
		m.filterCommits()
		m.updateFooter()
		return m, cmd
	}

	switch msg.String() {
	case "esc":
		m.route = routeHome
		m.updateFooter()
		return m, nil
	case "/":
		m.commitFilterMode = true
		return m, m.commitInput.Focus()
	case "down", "j":
		m.commitIndex++
		if m.commitIndex >= len(m.commitVisible) {
			m.commitIndex = len(m.commitVisible) - 1
		}
		if m.commitIndex < 0 {
			m.commitIndex = 0
		}
		m.commitList.SetCursor(m.commitIndex)
	case "up", "k":
		m.commitIndex--
		if m.commitIndex < 0 {
			m.commitIndex = 0
		}
		m.commitList.SetCursor(m.commitIndex)
	case "enter":
		if len(m.commitVisible) == 0 {
			return m, nil
		}
		m.loading = true
		m.loadErr = ""
		m.updateFooter()
		return m, m.loadDiffCmd(m.activeRepo.Path, m.commitVisible[m.commitIndex].SHA)
	}
	m.updateFooter()
	return m, nil
}

func (m *model) homeBody(width, height int) string {
	word := lipgloss.NewStyle().Bold(true).Foreground(m.theme.TextAccent()).Render(homeWordmark)
	pair := lipgloss.NewStyle().Bold(true).Foreground(m.theme.DiffAdded()).Render("+") +
		" " +
		lipgloss.NewStyle().Bold(true).Foreground(m.theme.DiffRemoved()).Render("-")
	hero := lipgloss.JoinHorizontal(lipgloss.Top, word, lipgloss.NewStyle().PaddingLeft(2).Render(pair))
	dim := lipgloss.NewStyle().Foreground(m.theme.TextMuted())

	blockW := clamp(width*3/4, 48, 96)
	listH := clamp(height/2, 8, 18)
	m.homeList.SetSize(blockW, listH)
	listView := viewString(m.homeList.View())

	inputView := m.homeInput.Value()
	if m.homeFilterMode {
		inputView = viewString(m.homeInput.View())
		if strings.TrimSpace(inputView) == "" {
			inputView = m.homeInput.Value()
		}
	} else if m.homeInput.Value() == "" {
		inputView = dim.Render("press / to filter repos")
	}

	inputLine := lipgloss.NewStyle().
		Background(m.theme.InputBG()).
		Foreground(m.theme.InputFG()).
		Padding(0, 2).
		Width(max(1, blockW-4)).
		Render(inputView)

	stack := strings.Join([]string{
		lipgloss.NewStyle().Width(blockW).Align(lipgloss.Center).Render(hero),
		"",
		inputLine,
		"",
		listView,
	}, "\n")

	if m.loading {
		stack += "\n\n" + dim.Render("loading repositories...")
	}
	if strings.TrimSpace(m.loadErr) != "" {
		stack += "\n\n" + lipgloss.NewStyle().Foreground(m.theme.Error()).Render(m.loadErr)
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, stack)
}

func (m *model) commitsBody(width, height int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(m.theme.TextAccent()).Render(m.activeRepo.Name)
	sub := lipgloss.NewStyle().Foreground(m.theme.TextMuted()).Render(m.activeRepo.Path)

	blockW := clamp(width*4/5, 56, 110)
	listH := clamp(height/2, 8, 18)
	m.commitList.SetSize(blockW, listH)
	listView := viewString(m.commitList.View())

	inputView := m.commitInput.Value()
	if m.commitFilterMode {
		inputView = viewString(m.commitInput.View())
		if strings.TrimSpace(inputView) == "" {
			inputView = m.commitInput.Value()
		}
	} else if m.commitInput.Value() == "" {
		inputView = lipgloss.NewStyle().Foreground(m.theme.TextMuted()).Render("press / to filter commits")
	}

	inputLine := lipgloss.NewStyle().
		Background(m.theme.InputBG()).
		Foreground(m.theme.InputFG()).
		Padding(0, 2).
		Width(max(1, blockW-4)).
		Render(inputView)

	stack := strings.Join([]string{
		title,
		sub,
		"",
		inputLine,
		"",
		listView,
	}, "\n")

	if m.loading {
		stack += "\n\n" + lipgloss.NewStyle().Foreground(m.theme.TextMuted()).Render("loading commits...")
	}
	if strings.TrimSpace(m.loadErr) != "" {
		stack += "\n\n" + lipgloss.NewStyle().Foreground(m.theme.Error()).Render(m.loadErr)
	}

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, stack)
}

func (m *model) filterRepos() {
	q := strings.ToLower(strings.TrimSpace(m.homeInput.Value()))
	m.homeVisible = m.homeVisible[:0]
	for _, r := range m.repos {
		if q == "" || strings.Contains(strings.ToLower(r.Name), q) || strings.Contains(strings.ToLower(r.Path), q) {
			m.homeVisible = append(m.homeVisible, r)
		}
	}
	if m.homeIndex >= len(m.homeVisible) {
		m.homeIndex = len(m.homeVisible) - 1
	}
	if m.homeIndex < 0 {
		m.homeIndex = 0
	}
	m.homeList.Clear()
	for _, r := range m.homeVisible {
		m.homeList.AppendRow(list.Row{Kind: list.RowItem, Primary: r.Name, Secondary: r.Path, Label: r.Path})
	}
	m.homeList.SetCursor(m.homeIndex)
}

func (m *model) filterCommits() {
	q := strings.ToLower(strings.TrimSpace(m.commitInput.Value()))
	m.commitVisible = m.commitVisible[:0]
	for _, c := range m.commits {
		hay := strings.ToLower(c.Short + " " + c.Subject + " " + c.Date)
		if q == "" || strings.Contains(hay, q) {
			m.commitVisible = append(m.commitVisible, c)
		}
	}
	if m.commitIndex >= len(m.commitVisible) {
		m.commitIndex = len(m.commitVisible) - 1
	}
	if m.commitIndex < 0 {
		m.commitIndex = 0
	}
	m.commitList.Clear()
	for _, c := range m.commitVisible {
		m.commitList.AppendRow(list.Row{Kind: list.RowItem, Primary: c.Short + "  " + c.Subject, Secondary: c.Date, Label: c.Short, RightStat: c.Date})
	}
	m.commitList.SetCursor(m.commitIndex)
}

func (m *model) updateFooter() {
	switch m.route {
	case routeCommits:
		m.footer.SetLeft("~ bentodiffs/commits")
		m.footer.SetCards([]bar.Card{
			{Command: "enter", Label: "open", Variant: bar.CardPrimary, Enabled: len(m.commitVisible) > 0, Priority: 3},
			{Command: "j/k", Label: "move", Enabled: true, Priority: 2},
			{Command: "/", Label: "filter", Enabled: true, Priority: 2},
			{Command: "p", Label: "palette", Enabled: true, Priority: 2},
			{Command: "esc", Label: "back", Enabled: true, Priority: 2},
			{Command: "q", Label: "quit", Enabled: true, Priority: 1},
		})
		m.footer.SetStatusPill(fmt.Sprintf("%d/%d commits", len(m.commitVisible), len(m.commits)))
	case routeDiff:
		m.footer.SetLeft("~ bentodiffs/diff")
		if m.viewer != nil {
			st := m.viewer.State()
			files := max(1, st.FileCount)
			hunks := max(1, len(st.HunkStarts))
			m.footer.SetStatusPill(fmt.Sprintf("file %d/%d  hunk %d/%d", st.ActiveFile+1, files, st.ActiveHunk+1, hunks))
		} else {
			m.footer.SetStatusPill("")
		}
		m.footer.SetCards([]bar.Card{
			{Command: "esc", Label: "back", Enabled: true, Priority: 3},
			{Command: "q", Label: "quit", Enabled: true, Priority: 2},
		})
	default:
		m.footer.SetLeft("~ bentodiffs/home")
		m.footer.SetCards([]bar.Card{
			{Command: "enter", Label: "select", Variant: bar.CardPrimary, Enabled: len(m.homeVisible) > 0, Priority: 3},
			{Command: "j/k", Label: "move", Enabled: true, Priority: 2},
			{Command: "/", Label: "filter", Enabled: true, Priority: 2},
			{Command: "p", Label: "palette", Enabled: true, Priority: 2},
			{Command: "r", Label: "refresh", Enabled: true, Priority: 2},
			{Command: ",", Label: "roots", Enabled: true, Priority: 2},
			{Command: "q", Label: "quit", Enabled: true, Priority: 1},
		})
		m.footer.SetStatusPill(fmt.Sprintf("%d/%d repos", len(m.homeVisible), len(m.repos)))
	}
}

func (m *model) loadReposCmd() tea.Cmd {
	m.loading = true
	m.loadErr = ""
	c, path, err := loadConfig()
	m.configPath = path
	if err != nil {
		return func() tea.Msg { return reposLoadedMsg{err: err} }
	}
	m.roots = c.RepoRoots
	if len(m.roots) == 0 {
		m.loading = false
		m.repos = []repoItem{{Name: "Mock repo", Path: mockRepoPath}}
		m.filterRepos()
		m.loadErr = "no repo roots configured (press ,) - using mock repo"
		m.updateFooter()
		return nil
	}
	roots := append([]string{}, m.roots...)
	return func() tea.Msg {
		repos, err := discoverRepos(roots)
		return reposLoadedMsg{repos: repos, err: err}
	}
}

func (m *model) loadCommitsCmd(repoPath string) tea.Cmd {
	if repoPath == mockRepoPath {
		return func() tea.Msg {
			return commitsLoadedMsg{commits: []commitItem{
				{SHA: "mock-1", Short: "mock-1", Date: "local", Subject: "demo: syntax + intraline"},
				{SHA: "mock-2", Short: "mock-2", Date: "local", Subject: "demo: multi-file changes"},
			}}
		}
	}
	return func() tea.Msg {
		commits, err := loadCommits(repoPath, 300)
		return commitsLoadedMsg{commits: commits, err: err}
	}
}

func (m *model) loadDiffCmd(repoPath, sha string) tea.Cmd {
	if repoPath == mockRepoPath {
		return func() tea.Msg {
			diffs, err := bentodiffs.MockDiffs(3)
			return diffLoadedMsg{diffs: diffs, err: err}
		}
	}
	return func() tea.Msg {
		diffs, err := loadCommitDiffs(repoPath, sha)
		return diffLoadedMsg{diffs: diffs, err: err}
	}
}

func (m *model) openConfigDialog() tea.Cmd {
	path := m.configPath
	if path == "" {
		if p, err := configPath(); err == nil {
			path = p
		}
	}
	msg := "Edit this file and add repo roots JSON:\n\n" + path + "\n\n{\n  \"repo_roots\": [\n    \"/Users/you/code\"\n  ]\n}"
	return func() tea.Msg {
		return dialog.Open(dialog.Confirm{DialogTitle: "Repo roots config", Message: msg})
	}
}

func (m *model) openThemePicker() tea.Cmd {
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

func (m *model) openCommandPalette() tea.Cmd {
	commands := []dialog.Command{
		{Label: "Theme picker", Group: "App", Keybind: "", Action: func() tea.Msg { return openThemePickerMsg{} }},
	}
	if m.route == routeHome {
		commands = append(commands,
			dialog.Command{Label: "Refresh repos", Group: "Home", Keybind: "r", Action: func() tea.Msg { return refreshReposMsg{} }},
			dialog.Command{Label: "Show roots config", Group: "Home", Keybind: ",", Action: func() tea.Msg { return openRootsConfigMsg{} }},
		)
	}
	if m.route == routeCommits {
		commands = append(commands,
			dialog.Command{Label: "Back to repos", Group: "Navigate", Keybind: "esc", Action: func() tea.Msg { return backToHomeMsg{} }},
		)
	}
	return commandpaletteflow.Open(commands)
}

func keyIs(msg tea.KeyMsg, keys ...string) bool {
	for _, k := range keys {
		if msg.String() == k {
			return true
		}
	}
	return false
}

func (m *model) resizeRouteWidgets() {
	w := clamp(m.width*3/4, 40, 110)
	m.homeInput.SetSize(max(1, w-6), 1)
	m.commitInput.SetSize(max(1, w-6), 1)
	h := clamp(m.height/2, 8, 18)
	m.homeList.SetSize(w, h)
	m.commitList.SetSize(w, h)
}

func (m *model) applyTheme(t theme.Theme) {
	if t == nil {
		return
	}
	m.theme = t
	m.footer.SetTheme(t)
	m.dialog.SetTheme(t)
	m.homeInput.SetTheme(t)
	m.commitInput.SetTheme(t)
	m.homeList.SetTheme(t)
	m.commitList.SetTheme(t)
	if m.viewer != nil {
		m.viewer.SetTheme(t)
	}
	m.updateFooter()
}

func viewString(v tea.View) string {
	if v.Content == nil {
		return ""
	}
	if r, ok := v.Content.(interface{ Render() string }); ok {
		return r.Render()
	}
	if s, ok := v.Content.(interface{ String() string }); ok {
		return s.String()
	}
	return fmt.Sprint(v.Content)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
