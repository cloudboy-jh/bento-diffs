package bentodiffs

import "github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs/parser"

type LineType = parser.LineType

const (
	LineContext = parser.LineContext
	LineAdded   = parser.LineAdded
	LineRemoved = parser.LineRemoved
)

type SegmentType = parser.SegmentType

const (
	SegmentEqual   = parser.SegmentEqual
	SegmentAdded   = parser.SegmentAdded
	SegmentRemoved = parser.SegmentRemoved
)

type Segment = parser.Segment
type DiffLine = parser.DiffLine
type Hunk = parser.Hunk
type DiffResult = parser.DiffResult
