package bentodiffs

type LineType int

const (
	LineContext LineType = iota
	LineAdded
	LineRemoved
)

type SegmentType int

const (
	SegmentEqual SegmentType = iota
	SegmentAdded
	SegmentRemoved
)

type Segment struct {
	Start int
	End   int
	Type  SegmentType
	Text  string
}

type DiffLine struct {
	Kind      LineType
	OldLineNo int
	NewLineNo int
	Content   string
	Segments  []Segment
}

type Hunk struct {
	Header string
	Lines  []DiffLine
}

type DiffResult struct {
	OldFile     string
	NewFile     string
	DisplayFile string
	Hunks       []Hunk
	Additions   int
	Removals    int
}
