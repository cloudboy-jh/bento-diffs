package bentodiffs

import (
	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs/viewer"
	"github.com/cloudboy-jh/bentotui/theme"
)

// Viewer exposes Bubble Tea lifecycle plus imperative navigation controls.
type Viewer interface {
	Init() tea.Cmd
	Update(tea.Msg) (Viewer, tea.Cmd)
	View() string
	SetSize(width, height int)
	SetTheme(theme.Theme)
	SetDiffs([]DiffResult)
	SetFileIndex(int)
	NextFile()
	PrevFile()
	NextHunk()
	PrevHunk()
	ScrollUp(n int)
	ScrollDown(n int)
	State() ViewerState
	TeaModel() tea.Model
}

// ViewerOptions configures NewViewer.
type ViewerOptions struct {
	Diffs           []DiffResult
	Layout          string
	SyntaxEnabled   bool
	ShowLineNumbers bool
	Theme           theme.Theme
}

// ViewerState is a readonly snapshot of viewer state.
type ViewerState struct {
	Width      int
	Height     int
	Layout     string
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

// NewViewer constructs an embeddable interactive diff viewer.
func NewViewer(opts ViewerOptions) Viewer {
	layout := viewer.LayoutSplit
	if opts.Layout == "stacked" {
		layout = viewer.LayoutStacked
	}
	t := opts.Theme
	if t == nil {
		t = theme.CurrentTheme()
	}
	m := viewer.New(opts.Diffs, t, viewer.Options{
		Layout:          layout,
		SyntaxEnabled:   opts.SyntaxEnabled,
		ShowLineNumbers: opts.ShowLineNumbers,
	})
	return &viewerHandle{model: m}
}

type viewerHandle struct {
	model *viewer.Model
}

func (v *viewerHandle) Init() tea.Cmd {
	return v.model.Init()
}

func (v *viewerHandle) Update(msg tea.Msg) (Viewer, tea.Cmd) {
	_, cmd := v.model.Update(msg)
	return v, cmd
}

func (v *viewerHandle) View() string {
	return v.model.Render()
}

func (v *viewerHandle) SetSize(width, height int) {
	v.model.SetSize(width, height)
}

func (v *viewerHandle) SetTheme(t theme.Theme) {
	v.model.SetTheme(t)
}

func (v *viewerHandle) SetDiffs(diffs []DiffResult) {
	v.model.SetDiffs(diffs)
}

func (v *viewerHandle) SetFileIndex(index int) {
	v.model.SetFileIndex(index)
}

func (v *viewerHandle) NextFile() {
	v.model.NextFile()
}

func (v *viewerHandle) PrevFile() {
	v.model.PrevFile()
}

func (v *viewerHandle) NextHunk() {
	v.model.NextHunk()
}

func (v *viewerHandle) PrevHunk() {
	v.model.PrevHunk()
}

func (v *viewerHandle) ScrollUp(n int) {
	v.model.ScrollUp(n)
}

func (v *viewerHandle) ScrollDown(n int) {
	v.model.ScrollDown(n)
}

func (v *viewerHandle) State() ViewerState {
	s := v.model.State()
	layout := "split"
	if s.Layout == viewer.LayoutStacked {
		layout = "stacked"
	}
	return ViewerState{
		Width:      s.Width,
		Height:     s.Height,
		Layout:     layout,
		ActiveFile: s.ActiveFile,
		Visible:    s.Visible,
		FileCount:  s.FileCount,
		Scroll:     s.Scroll,
		MaxScroll:  s.MaxScroll,
		HunkStarts: s.HunkStarts,
		ActiveHunk: s.ActiveHunk,
		FilterMode: s.FilterMode,
		Filter:     s.Filter,
		ShowEmpty:  s.ShowEmpty,
	}
}

func (v *viewerHandle) TeaModel() tea.Model {
	return v.model
}
