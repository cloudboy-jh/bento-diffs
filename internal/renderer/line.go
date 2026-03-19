package renderer

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/cloudboy-jh/bento-diffs/internal/parser"
	"github.com/cloudboy-jh/bentotui/theme"
	tstyles "github.com/cloudboy-jh/bentotui/theme/styles"
)

const lineNumberWidth = 6

func renderSplitColumn(dl *parser.DiffLine, colWidth int, fileName string, t theme.Theme, syntax bool, showLineNumbers bool, side string) string {
	if dl == nil {
		return tstyles.Row(t.DiffContextBG(), t.Text(), colWidth, "")
	}

	st := createStyles(t)
	bg := st.contextBG
	lineNoFG := st.lineNumberFG
	markerBG := bg
	lineNo := 0

	switch dl.Kind {
	case parser.LineAdded:
		bg = st.addedBG
		lineNoFG = st.addedFG
		markerBG = st.addedFG
		lineNo = dl.NewLineNo
	case parser.LineRemoved:
		bg = st.removedBG
		lineNoFG = st.removedFG
		markerBG = st.removedFG
		lineNo = dl.OldLineNo
	default:
		markerBG = bg
		if side == "left" {
			lineNo = dl.OldLineNo
		} else {
			lineNo = dl.NewLineNo
		}
	}

	marker := lipgloss.NewStyle().Background(markerBG).Render(" ")
	prefix := ""
	if showLineNumbers {
		num := lipgloss.NewStyle().Foreground(lineNoFG).Background(bg).Render(fmt.Sprintf("%*d", lineNumberWidth, lineNo))
		prefix = num
	}

	body := dl.Content
	if syntax {
		body = SyntaxHighlight(fileName, body, t, bg)
	}
	body = applySegments(body, dl.Content, dl.Segments, dl.Kind, t)

	avail := colWidth - ansi.StringWidth(marker) - ansi.StringWidth(prefix)
	if avail < 0 {
		avail = 0
	}
	body = ansi.Truncate(body, avail, "...")
	line := marker + prefix + body
	return tstyles.RowClip(bg, t.Text(), colWidth, line)
}

func renderStackedLine(dl parser.DiffLine, width int, fileName string, t theme.Theme, syntax bool, showLineNumbers bool) string {
	copyLine := dl
	return renderSplitColumn(&copyLine, width, fileName, t, syntax, showLineNumbers, "right")
}

func applySegments(renderedText, rawText string, segments []parser.Segment, kind parser.LineType, t theme.Theme) string {
	if len(segments) == 0 {
		return renderedText
	}
	if strings.Contains(renderedText, "\x1b[") {
		return renderedText
	}

	runes := []rune(rawText)
	var b strings.Builder
	for _, seg := range segments {
		start := max(seg.Start, 0)
		end := min(seg.End, len(runes))
		if start >= end {
			continue
		}
		chunk := string(runes[start:end])
		switch {
		case kind == parser.LineAdded && seg.Type == parser.SegmentAdded:
			b.WriteString(lipgloss.NewStyle().Background(t.DiffHighlightAdded()).Render(chunk))
		case kind == parser.LineRemoved && seg.Type == parser.SegmentRemoved:
			b.WriteString(lipgloss.NewStyle().Background(t.DiffHighlightRemoved()).Render(chunk))
		default:
			b.WriteString(chunk)
		}
	}
	if b.Len() == 0 {
		return renderedText
	}
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
