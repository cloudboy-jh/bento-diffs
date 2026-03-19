# bento-diffs — Spec

> The @pierre/diffs of the terminal. A standalone TUI diff viewer built on BentoTUI.

**Repo:** `github.com/cloudboy-jh/bento-diffs`
**Platform:** `github.com/cloudboy-jh/bentotui` v0.5.3+
**Status:** Pre-build spec

---

## What it is

A standalone Go TUI application and module. Not a library. Not a bento in the
bentotui registry. A complete, shippable terminal diff viewer — the first
proof-of-platform tool built entirely on BentoTUI, the same way Charm builds
tools with their own stack.

It is to bentotui what `delta` is to `git` — a beautiful terminal renderer for
a format that git already produces.

## Architecture split with BentoTUI

- BentoTUI is Bento-first platform surface:
  - Bricks = official components
  - Recipes = copy/own composed flows
  - Rooms = named layout contracts
  - Bentos = template apps
- `bento-diffs` stays a separate standalone tool and module.
- Parser/hunks/intraline/syntax logic remain in this repository.
- UI composition consumes workspace DTOs from an adapter layer; composition does not read parser internals directly.

## What stays in bento-diffs vs what belongs in bentotui

`bento-diffs` owns:
- unified diff parsing
- hunk modeling and pairing
- intraline segment detection
- syntax-highlighted diff line rendering
- adapter DTOs for the Bento workspace pattern

`bentotui` owns:
- reusable Bricks
- layout Rooms contracts
- theme contracts and preset runtime
- Bento app/template concerns

`bento-diffs` does not move parser/syntax logic into BentoTUI.

---

## Feature parity with @pierre/diffs (v1 scope)

| Feature | Pierre/diffs | bento-diffs v1 |
|---|---|---|
| Stacked (unified) layout | ✓ | ✓ |
| Split (side-by-side) layout | ✓ | ✓ |
| Toggle between layouts | ✓ | ✓ (tab key) |
| Full-line background colors (+/-/context) | ✓ | ✓ |
| Intraline word/char diff highlighting | ✓ | ✓ |
| Line numbers | ✓ | ✓ |
| +/- indicators (classic style) | ✓ | ✓ |
| Hunk separators | ✓ | ✓ |
| File header (name + stats) | ✓ | ✓ |
| Syntax highlighting (chroma) | ✓ | ✓ |
| Multi-file navigation (palette/filter-first) | ✓ | ✓ |
| Synchronized scroll (split view) | ✓ | ✓ |
| Adapts to theme | ✓ (Shiki) | ✓ (bentotui Theme interface) |
| Annotations/comments | ✓ | ✗ v2 |
| Accept/reject UI | ✓ | ✗ v2 |
| Line selection | ✓ | ✗ v2 |
| Merge conflict UI | ✓ | ✗ v2 |

---

## Input model

```bash
# Pipe from git diff — primary usage, like delta
git diff | bento-diffs
git diff HEAD~1 | bento-diffs
git show abc123 | bento-diffs

# Two file arguments — generates diff internally via go-udiff
bento-diffs before.go after.go

# Explicit patch file
bento-diffs --patch changes.patch

# With options
git diff | bento-diffs --layout split
git diff | bento-diffs --theme dracula
git diff | bento-diffs --no-syntax
```

Flags:
- `--layout split|stacked` — initial layout (default: split)
- `--theme <name>` — bentotui preset name (default: catppuccin-mocha)
- `--no-syntax` — disable chroma syntax highlighting
- `--context <n>` — context lines around hunks (default: 3)
- `--no-line-numbers` — hide line number column

---

## Layout

### Full layout (multi-file diff)

```
rooms.Focus(w, h,
    fileHeader,   // top  — filename, +N -N stats, layout toggle indicator
    contentPane,  // header + split/stacked diff content
    footerBar,    // foot — keybinds bar (bentotui bar brick, FooterAnchored)
)
```

```
+----------------------------+-------------------------------------+
| fileHeader: path/to/file.go            +12 -4   [split|stacked] |
+-------------------+--------+-------------------------------------+
| ● path/to/file.go |   113  "-"  ...       |  113  "+"  ...      |
|   other/file.go   |   114  "  " ...       |  114  "  " ...      |
|   README.md       |   115  "-"  ...       |  115  "+" ...       |
|                   |        ...            |        ...           |
|                   | hunk separator        | hunk separator       |
|                   |        ...            |        ...           |
+-------------------+--------+-------------------------------------+
| j/k scroll   tab layout   [ ] file   q quit                      |
+------------------------------------------------------------------+
```

### Single-pane stacked layout

```
rooms.Focus(w, h, diffPane, footerBar)
```

```
+------------------------------------------------------------------+
| fileHeader                                                        |
+------------------------------------------------------------------+
|  113  "-"  old content...                                        |
|  113  "+"  new content...                                        |
|  114  "  " context...                                            |
+------------------------------------------------------------------+
| j/k scroll   tab layout   [ ] file   q quit                      |
+------------------------------------------------------------------+
```

---

## Repository structure

```
bento-diffs/
  main.go                        ← entry: reads input, runs tea.NewProgram
  go.mod
  go.sum
  README.md
  CHANGELOG.md

  internal/
    adapter/
      workspace.go               ← maps parsed diffs to Bento workspace DTOs

    parser/
      unified.go                 ← ParseUnifiedDiff — parses unified patch format
      generate.go                ← GenerateDiff — two-file diff via go-udiff
      intraline.go               ← HighlightIntralineChanges via sergi/go-diff
      types.go                   ← LineType, DiffLine, Hunk, DiffResult, linePair

    renderer/
      split.go                   ← RenderSideBySideHunk — paired left/right columns
      stacked.go                 ← RenderUnifiedHunk — classic unified view
      highlight.go               ← SyntaxHighlight via chroma
      line.go                    ← renderLeftColumn, renderRightColumn
      styles.go                  ← createStyles(t theme.Theme) — lipgloss styles from theme tokens

    model/
      model.go                   ← root Bubble Tea model (Init/Update/View) using adapter DTOs
      keymap.go                  ← keybindings
      scroll.go                  ← synchronized scroll state (offset, max)

  bricks/
    diffpane/
      diffpane.go                ← scrollable render-ready diff line pane (Sizable, SetSize, View)
    fileheader/
      fileheader.go              ← top bar: filename, +N -N, layout mode badge
    filelist/
      filelist.go                ← optional file list wrapper (state/nav support)
```

---

## Dependencies

```go
// go.mod
require (
    github.com/cloudboy-jh/bentotui       v0.5.3
    charm.land/bubbletea/v2               // via bentotui
    charm.land/lipgloss/v2                // via bentotui
    github.com/aymanbagabas/go-udiff      // unified diff generation (two-file input)
    github.com/sergi/go-diff              // intraline char-level diff (diffmatchpatch)
    github.com/alecthomas/chroma/v2       // syntax highlighting
    github.com/charmbracelet/x/ansi       // ANSI-safe truncation
    github.com/sourcegraph/go-diff/diff   // unified diff parser (alternative to hand-rolling)
)
```

---

## Parser — `internal/parser`

### `ParseUnifiedDiff(patch string) (DiffResult, error)`

Parses unified diff format (`git diff` output). Extracts:
- `OldFile`, `NewFile` from `--- a/` / `+++ b/` headers
- `[]Hunk` — each with `Header string` (`@@ -N,N +N,N @@`) and `[]DiffLine`
- `DiffLine.Kind` — `LineAdded`, `LineRemoved`, `LineContext`
- `DiffLine.OldLineNo`, `DiffLine.NewLineNo`
- `DiffLine.Content` — line text without the `+`/`-`/` ` prefix

### `HighlightIntralineChanges(h *Hunk)`

Walks adjacent `-`/`+` pairs in a hunk. Uses `diffmatchpatch.DiffMain` to find
character-level segments. Populates `DiffLine.Segments []Segment` with `Start`,
`End`, `Type`, `Text` for each changed span.

### `GenerateDiff(before, after, filename string) (patch string, additions, removals int)`

Takes two file content strings, runs `go-udiff` to produce a unified diff string,
counts `+`/`-` lines. Used for the two-file argument input mode.

---

## Renderer — `internal/renderer`

### `styles.go` — `createStyles(t theme.Theme)`

Returns four lipgloss styles from the theme's Diff tokens:

```go
removedLine  = lipgloss.NewStyle().Background(t.DiffRemovedBG())
addedLine    = lipgloss.NewStyle().Background(t.DiffAddedBG())
contextLine  = lipgloss.NewStyle().Background(t.DiffContextBG())
lineNumber   = lipgloss.NewStyle().Foreground(t.DiffLineNum())
```

### `highlight.go` — `SyntaxHighlight`

Builds a dynamic chroma XML style from the theme's `Syntax*()` token methods.
Matches opencode's approach exactly — constructs an XML style string, parses it
with `chroma.MustNewXMLStyle`, then applies a background transform to match the
current line's background (added/removed/context).

The background transform ensures syntax highlighting adapts cleanly to diff line
backgrounds — keywords on a red-tinted line still look like keywords, not just
pale text.

### `line.go` — `renderLeftColumn` / `renderRightColumn`

Each takes a `*DiffLine` and a `colWidth int`. Returns an exact-width ANSI string:

```
[linenum] [marker] [syntax-highlighted content with intraline segments applied]
```

Line number column: 6 chars right-aligned + 1 space + marker (1 char).
Content: syntax-highlighted, then intraline segments applied on top, then truncated
to `colWidth` with a muted `...` ellipsis if it overflows.

If `dl == nil` (no matching line on this side in split view): returns a full-width
context-colored empty row.

### `split.go` — `RenderSideBySideHunk`

1. Copies the hunk, calls `HighlightIntralineChanges`
2. Calls `pairLines` to match `-`/`+` adjacent lines
3. For each `linePair`: renders left + right columns, joins them
4. Returns the full hunk string

### `stacked.go` — `RenderUnifiedHunk`

Renders each line sequentially:
- Context: `contextLine` style, line numbers for both sides
- Added: `addedLine` style, new line number, `+` marker
- Removed: `removedLine` style, old line number, `-` marker
- Hunk header: muted row with `@@ ... @@` text

---

## Bricks — `bricks/`

### `diffpane/diffpane.go`

The scrollable diff content area. Implements `Sizable` for bentotui rooms.

```go
type Model struct {
    scroll    int           // current scroll offset (line index)
    maxScroll int           // computed from line count - visible height
    width     int
    height    int
    theme     theme.Theme
    lines     []string      // render-ready ANSI lines from adapter DTO
}

func New(t theme.Theme) *Model
func (m *Model) SetSize(width, height int)
func (m *Model) SetTheme(t theme.Theme)
func (m *Model) SetLines(lines []string, resetScroll bool)
func (m *Model) ScrollDown(n int)
func (m *Model) ScrollUp(n int)
func (m *Model) View() tea.View
func (m *Model) Init() tea.Cmd
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
```

The adapter pre-renders lines; `diffpane` receives them via `SetLines` and only
handles viewport math + painting.
`View()` slices `m.lines[m.scroll : m.scroll+m.height]` and joins with `\n`.
This means scrolling is O(1) — no re-rendering per frame, just a slice.

### `fileheader/fileheader.go`

Top bar: filename (left), `+N -N` stats (center-right), layout mode badge (far right).
Uses `bentotui/theme/styles.Row` for the full-width painted row.
`SetTheme`, `SetSize`, `SetFile(name, additions, removals int)`, `SetLayout(Layout)`.

### `filelist/filelist.go`

File list remains available as an internal/optional brick, but the default app composition is rail-less.
Multi-file navigation is recipes-first through command palette (`ctrl+k`) and filter (`/`).

---

## Root model — `internal/model/model.go`

```go
type model struct {
    theme      theme.Theme
    workspace  *adapter.WorkspaceAdapter
    activeFile int                   // selected file index in adapter
    layout     Layout

    fileHeader *fileheader.Model
    fileList   *filelist.Model
    diffPane   *diffpane.Model
    footer     *bar.Model

    width  int
    height int
}
```

### Init

1. Read input (stdin or file args)
2. Parse with `parser.ParseUnifiedDiff` or `parser.GenerateDiff`
3. Build `adapter.WorkspaceAdapter` and initialize UI bricks

### Update

```
tea.WindowSizeMsg  → SetSize on all models + rebuild WorkspaceDTO
tea.KeyMsg:
  j / down        → diffPane.ScrollDown(1)
  k / up          → diffPane.ScrollUp(1)
  ctrl+d          → diffPane.ScrollDown(height/2)
  ctrl+u          → diffPane.ScrollUp(height/2)
  tab             → toggle layout (Split ↔ Stacked), rebuild WorkspaceDTO
  [               → previous file
  ]               → next file
  q / ctrl+c      → tea.Quit
  ?               → toggle help overlay (v2)

theme.ThemeChangedMsg → update theme on all models + rebuild WorkspaceDTO
```

### View

```go
func (m *model) View() tea.View {
    t := m.theme

    content := composeHeaderAndMain(m.fileHeader, m.diffPane)
    screen := rooms.Focus(m.width, m.height, content, m.footer)

    surf := surface.New(m.width, m.height)
    surf.Fill(t.Background())
    surf.Draw(0, 0, screen)

    v := tea.NewView(surf.Render())
    v.AltScreen = true
    v.BackgroundColor = t.Background()
    return v
}
```

---

## Scroll model

Split view and stacked view both use a single `scroll int` offset — number of
rendered lines from the top. Left and right columns in split view are always at
the same offset (synchronized scroll). No independent left/right scrolling.

`maxScroll = len(m.lines) - m.height`. Clamped to `[0, maxScroll]` on every
scroll operation.

Mouse wheel events (`tea.MouseMsg`) map to `ScrollDown`/`ScrollUp` — Bubble Tea
v2 exposes these without extra setup.

---

## Color model

The diff view explicitly uses only `DiffAddedBG`, `DiffRemovedBG`, `DiffContextBG`
and the other diff-specific tokens. It does NOT use `CardChrome`, `CardBody`,
`BackgroundPanel`, or `SelectionBG`.

This is intentional — the diff surface lives on `Background()` (canvas level) and
paints its own color hierarchy on top. The bentotui v0.5.3 Bento-first architecture
makes this clean: the diffpane brick receives render-ready lines + a theme and only
calls the tokens it needs.

```
Background()              ← canvas, fills the whole surface
DiffContextBG()           ← unchanged lines (same or slightly elevated)
DiffRemovedBG()           ← removed line slab
DiffAddedBG()             ← added line slab
DiffRemovedLineNumBG()    ← line number column on removed lines
DiffAddedLineNumBG()      ← line number column on added lines
DiffHighlightRemoved()    ← intraline changed char highlight on removed lines
DiffHighlightAdded()      ← intraline changed char highlight on added lines
DiffRemoved()             ← "-" marker + line number foreground
DiffAdded()               ← "+" marker + line number foreground
DiffLineNum()             ← context line number foreground
```

---

## Syntax highlighting approach

Matches opencode's approach:

1. For each diff line, determine the background color (`DiffAddedBG`, `DiffRemovedBG`, or `DiffContextBG`)
2. Build a chroma XML style from the theme's `Syntax*()` token methods
3. Apply a background transform that sets every chroma entry's `Background` to the line's actual background — this ensures ANSI background codes inside the highlighted text match the line's diff color
4. Tokenize the line content with the chroma lexer matched to the filename extension
5. Format with `terminal16m` formatter
6. Apply intraline segment highlighting on top of the syntax-colored output

If `--no-syntax` flag is set, skip steps 2-5 and just paint plain colored text.

---

## v1 scope boundary

**In v1:**
- Parse and render `git diff` unified patch format
- Two-file argument mode
- Split and stacked layouts with toggle
- Synchronized scroll
- Intraline char diff
- Syntax highlighting via chroma
- Multi-file navigation via command palette + filter flow
- 16 bentotui preset themes via `--theme`
- Mouse wheel scroll
- Pipe-friendly (stdin first-class)

**Not in v1:**
- Theme picker TUI (use `--theme` flag)
- Annotations/comments
- Accept/reject UI
- Line selection
- Merge conflict resolution
- Word-wrap toggle
- Search/filter within diff
- Horizontal scroll for long lines (truncate with `...` ellipsis)

---

## Build + release

```bash
# Run locally
git diff | go run . 
go run . before.go after.go

# Build binary
go build -o bento-diffs .

# Install
go install github.com/cloudboy-jh/bento-diffs@latest
```

GoReleaser config mirrors bentotui — 6 binaries (linux/darwin/windows × amd64/arm64).

---

## Positioning

```
charm builds tools with bubbletea + lipgloss     → delta, glow, soft-serve
bento builds tools with bentotui                 → bento-diffs, ...
```

`bento-diffs` is the first entry in the bento tools ecosystem. Every future
bento tool follows the same pattern: standalone module, built on bentotui,
uses the Theme interface for visual consistency, ships its own binary.

## How to consume with Bento v0.5.3

Use this three-stage flow:

1. Parse input into `[]parser.DiffResult` with `internal/parser`.
2. Build render-ready workspace DTOs with `internal/adapter.WorkspaceAdapter`.
3. Compose Bento UI with rooms + bricks:
   - header
   - optional file index state
   - main diff pane
   - footer/status line

The adapter boundary guarantees UI code receives ready-to-render payloads and keeps parser/render logic local to this repo.
