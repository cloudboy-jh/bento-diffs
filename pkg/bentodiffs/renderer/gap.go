package renderer

import (
	"fmt"

	"github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs/parser"
)

func hunkGapLines(prev, next parser.Hunk) int {
	prevOldStart, prevOldLines, prevNewStart, prevNewLines, okPrev := parseHunkBounds(prev.Header)
	nextOldStart, _, nextNewStart, _, okNext := parseHunkBounds(next.Header)
	if !okPrev || !okNext {
		return 0
	}

	prevOldEnd := prevOldStart + prevOldLines
	prevNewEnd := prevNewStart + prevNewLines
	gapOld := nextOldStart - prevOldEnd
	gapNew := nextNewStart - prevNewEnd
	if gapOld < 0 {
		gapOld = 0
	}
	if gapNew < 0 {
		gapNew = 0
	}
	if gapOld > gapNew {
		return gapOld
	}
	return gapNew
}

func parseHunkBounds(header string) (oldStart, oldLines, newStart, newLines int, ok bool) {
	if _, err := fmt.Sscanf(header, "-%d,%d +%d,%d", &oldStart, &oldLines, &newStart, &newLines); err != nil {
		return 0, 0, 0, 0, false
	}
	return oldStart, oldLines, newStart, newLines, true
}
