package bentodiffs

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	"github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs/parser"
	"github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs/renderer"
	"github.com/cloudboy-jh/bentotui/theme"
)

func TestGenerateDiffAndParseUnifiedDiffs(t *testing.T) {
	patch, adds, rems := GenerateDiff("a\nb\nc\n", "a\nB\nc\n", "sample.txt", 3)
	if adds != 1 || rems != 1 {
		t.Fatalf("counts = +%d -%d, want +1 -1", adds, rems)
	}
	diffs, err := ParseUnifiedDiffs(patch)
	if err != nil {
		t.Fatalf("ParseUnifiedDiffs error: %v", err)
	}
	if len(diffs) != 1 {
		t.Fatalf("parsed files = %d, want 1", len(diffs))
	}
	if diffs[0].Additions != 1 || diffs[0].Removals != 1 {
		t.Fatalf("parsed stats = +%d -%d, want +1 -1", diffs[0].Additions, diffs[0].Removals)
	}
}

func TestViewerConstructionAndImperativeNavigation(t *testing.T) {
	before := strings.Join([]string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"",
	}, "\n")
	after := strings.Join([]string{
		"line1",
		"LINE2",
		"line3",
		"line4",
		"LINE5",
		"line6",
		"",
	}, "\n")

	p1, _, _ := GenerateDiff(before, after, "one.txt", 0)
	p2, _, _ := GenerateDiff("hello\n", "hello there\n", "two.txt", 0)
	diffs, err := ParseUnifiedDiffs(p1 + "\n" + p2)
	if err != nil {
		t.Fatalf("ParseUnifiedDiffs error: %v", err)
	}

	v := NewViewer(ViewerOptions{
		Diffs:           diffs,
		Layout:          "split",
		SyntaxEnabled:   true,
		ShowLineNumbers: true,
		Theme:           theme.CurrentTheme(),
	})
	v.SetSize(100, 3)
	if view := v.View(); view == "" {
		t.Fatal("viewer view should not be empty")
	}

	st := v.State()
	if len(st.HunkStarts) < 2 {
		t.Fatalf("hunk starts = %v, want at least 2 hunks", st.HunkStarts)
	}

	v.NextHunk()
	st = v.State()
	wantNext := st.HunkStarts[1]
	if wantNext > st.MaxScroll {
		wantNext = st.MaxScroll
	}
	if st.Scroll != wantNext {
		t.Fatalf("scroll after NextHunk = %d, want %d", st.Scroll, wantNext)
	}

	v.PrevHunk()
	st = v.State()
	if st.Scroll != st.HunkStarts[0] {
		t.Fatalf("scroll after PrevHunk = %d, want %d", st.Scroll, st.HunkStarts[0])
	}

	active := st.ActiveFile
	v.NextFile()
	if v.State().ActiveFile == active && len(v.State().Visible) > 1 {
		t.Fatal("NextFile should move to next file when multiple files are visible")
	}

	v.SetFileIndex(999)
	if got, want := v.State().ActiveFile, v.State().FileCount-1; got != want {
		t.Fatalf("SetFileIndex clamp = %d, want %d", got, want)
	}

	v.ScrollDown(999)
	st = v.State()
	if st.Scroll != st.MaxScroll {
		t.Fatalf("ScrollDown clamp = %d, want %d", st.Scroll, st.MaxScroll)
	}
	v.ScrollUp(999)
	if v.State().Scroll != 0 {
		t.Fatalf("ScrollUp clamp = %d, want 0", v.State().Scroll)
	}
}

func TestIntralineSegmentsRenderedDistinctly(t *testing.T) {
	h := parser.Hunk{Lines: []parser.DiffLine{
		{Kind: parser.LineRemoved, OldLineNo: 1, Content: "hello cat"},
		{Kind: parser.LineAdded, NewLineNo: 1, Content: "hello dog"},
	}}
	parser.HighlightIntralineChanges(&h)

	if len(h.Lines[0].Segments) == 0 || len(h.Lines[1].Segments) == 0 {
		t.Fatal("expected intraline segments on removed and added lines")
	}

	rendered := strings.Join(renderer.RenderUnifiedHunk(h, 120, "demo.txt", theme.CurrentTheme(), false, true), "\n")
	addedSeq := bgSeq(theme.CurrentTheme().DiffHighlightAdded())
	removedSeq := bgSeq(theme.CurrentTheme().DiffHighlightRemoved())
	if !strings.Contains(rendered, addedSeq) {
		t.Fatalf("rendered output missing added intraline highlight sequence %q", addedSeq)
	}
	if !strings.Contains(rendered, removedSeq) {
		t.Fatalf("rendered output missing removed intraline highlight sequence %q", removedSeq)
	}
}

func bgSeq(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("48;2;%d;%d;%d", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}
