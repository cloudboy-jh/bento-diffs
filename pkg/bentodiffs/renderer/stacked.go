package renderer

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
	"github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs/parser"
	"github.com/cloudboy-jh/bentotui/theme"
	tstyles "github.com/cloudboy-jh/bentotui/theme/styles"
)

func RenderUnifiedHunk(h parser.Hunk, width int, fileName string, t theme.Theme, syntax bool, showLineNumbers bool) []string {
	copyHunk := h
	parser.HighlightIntralineChanges(&copyHunk)

	out := make([]string, 0, len(copyHunk.Lines))
	for _, dl := range copyHunk.Lines {
		out = append(out, renderStackedLine(dl, width, fileName, t, syntax, showLineNumbers))
	}
	return out
}

func RenderUnifiedDiff(result parser.DiffResult, width int, fileName string, t theme.Theme, syntax bool, showLineNumbers bool) []string {
	r := RenderUnifiedDiffWithMeta(result, width, fileName, t, syntax, showLineNumbers)
	return r.Lines
}

func RenderUnifiedDiffWithMeta(result parser.DiffResult, width int, fileName string, t theme.Theme, syntax bool, showLineNumbers bool) RenderedDiff {
	out := make([]string, 0)
	starts := make([]int, 0, len(result.Hunks))
	for i, h := range result.Hunks {
		if i > 0 {
			if gap := hunkGapLines(result.Hunks[i-1], h); gap > 0 {
				out = append(out, renderStackedCollapsedContextRow(width, gap, t))
			}
		}
		starts = append(starts, len(out))
		out = append(out, RenderUnifiedHunk(h, width, fileName, t, syntax, showLineNumbers)...)
	}
	return RenderedDiff{Lines: out, HunkStarts: starts}
}

func renderStackedCollapsedContextRow(width, lines int, t theme.Theme) string {
	text := fmt.Sprintf("  ^ %d unmodified lines", lines)
	text = ansi.Truncate(text, width, "...")
	return tstyles.RowClip(t.BackgroundPanel(), t.TextMuted(), width, text)
}
