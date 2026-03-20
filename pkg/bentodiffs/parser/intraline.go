package parser

import "github.com/sergi/go-diff/diffmatchpatch"

func HighlightIntralineChanges(h *Hunk) {
	if h == nil {
		return
	}

	dmp := diffmatchpatch.New()
	for i := 0; i < len(h.Lines)-1; i++ {
		left := &h.Lines[i]
		right := &h.Lines[i+1]
		if left.Kind != LineRemoved || right.Kind != LineAdded {
			continue
		}

		diffs := dmp.DiffMain(left.Content, right.Content, false)
		left.Segments = buildSegments(diffs, false)
		right.Segments = buildSegments(diffs, true)
	}
}

func buildSegments(diffs []diffmatchpatch.Diff, forAdded bool) []Segment {
	segments := make([]Segment, 0)
	pos := 0
	for _, d := range diffs {
		length := len([]rune(d.Text))
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			segments = append(segments, Segment{Start: pos, End: pos + length, Type: SegmentEqual, Text: d.Text})
			pos += length
		case diffmatchpatch.DiffInsert:
			if forAdded {
				segments = append(segments, Segment{Start: pos, End: pos + length, Type: SegmentAdded, Text: d.Text})
			}
			if forAdded {
				pos += length
			}
		case diffmatchpatch.DiffDelete:
			if !forAdded {
				segments = append(segments, Segment{Start: pos, End: pos + length, Type: SegmentRemoved, Text: d.Text})
				pos += length
			}
		}
	}
	return segments
}
