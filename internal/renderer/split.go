package renderer

import (
	"fmt"

	"github.com/cloudboy-jh/bento-diffs/internal/parser"
	"github.com/cloudboy-jh/bentotui/theme"
	tstyles "github.com/cloudboy-jh/bentotui/theme/styles"
)

type linePair struct {
	left  *parser.DiffLine
	right *parser.DiffLine
}

func RenderSideBySideHunk(h parser.Hunk, width int, fileName string, t theme.Theme, syntax bool, showLineNumbers bool) []string {
	copyHunk := h
	parser.HighlightIntralineChanges(&copyHunk)

	pairs := pairLines(copyHunk.Lines)
	colWidthLeft, colWidthRight := splitWidths(width)

	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		left := renderSplitColumn(p.left, colWidthLeft, fileName, t, syntax, showLineNumbers, "left")
		right := renderSplitColumn(p.right, colWidthRight, fileName, t, syntax, showLineNumbers, "right")
		out = append(out, joinSplitRow(left, right, t))
	}
	return out
}

func RenderSideBySideDiff(result parser.DiffResult, width int, fileName string, t theme.Theme, syntax bool, showLineNumbers bool) []string {
	out := make([]string, 0)
	for i, h := range result.Hunks {
		if i > 0 {
			if gap := hunkGapLines(result.Hunks[i-1], h); gap > 0 {
				out = append(out, renderSplitCollapsedContextRow(width, gap, t))
			}
		}
		out = append(out, RenderSideBySideHunk(h, width, fileName, t, syntax, showLineNumbers)...)
	}
	return out
}

func splitWidths(total int) (left int, right int) {
	if total < 3 {
		return 1, 1
	}
	body := total - 1
	left = body / 2
	right = body - left
	return left, right
}

func renderSplitCollapsedContextRow(width, lines int, t theme.Theme) string {
	leftWidth, rightWidth := splitWidths(width)
	text := fmt.Sprintf("^ %d unmodified lines", lines)
	left := tstyles.RowClip(t.BackgroundPanel(), t.TextMuted(), leftWidth, "  "+text)
	right := tstyles.Row(t.BackgroundPanel(), t.TextMuted(), rightWidth, "")
	return joinSplitRow(left, right, t)
}

func joinSplitRow(left, right string, t theme.Theme) string {
	divider := tstyles.Row(t.BorderSubtle(), t.BorderSubtle(), 1, "")
	return left + divider + right
}

func pairLines(lines []parser.DiffLine) []linePair {
	pairs := make([]linePair, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		cur := lines[i]
		if cur.Kind == parser.LineRemoved && i+1 < len(lines) && lines[i+1].Kind == parser.LineAdded {
			left := cur
			right := lines[i+1]
			pairs = append(pairs, linePair{left: &left, right: &right})
			i++
			continue
		}
		if cur.Kind == parser.LineRemoved {
			left := cur
			pairs = append(pairs, linePair{left: &left})
			continue
		}
		if cur.Kind == parser.LineAdded {
			right := cur
			pairs = append(pairs, linePair{right: &right})
			continue
		}
		left := cur
		right := cur
		pairs = append(pairs, linePair{left: &left, right: &right})
	}
	return pairs
}
