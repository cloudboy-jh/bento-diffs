package bentodiffs

import core "github.com/cloudboy-jh/bento-diffs/pkg/bentodiffs"

type LineType = core.LineType

const (
	LineContext = core.LineContext
	LineAdded   = core.LineAdded
	LineRemoved = core.LineRemoved
)

type SegmentType = core.SegmentType

const (
	SegmentEqual   = core.SegmentEqual
	SegmentAdded   = core.SegmentAdded
	SegmentRemoved = core.SegmentRemoved
)

type Segment = core.Segment
type DiffLine = core.DiffLine
type Hunk = core.Hunk
type DiffResult = core.DiffResult
type Options = core.Options
type Viewer = core.Viewer
type ViewerOptions = core.ViewerOptions
type ViewerState = core.ViewerState

func DefaultOptions() Options {
	return core.DefaultOptions()
}

func AvailableThemes() []string {
	return core.AvailableThemes()
}

func ParseUnifiedDiff(patch string) (DiffResult, error) {
	return core.ParseUnifiedDiff(patch)
}

func ParseUnifiedDiffs(patch string) ([]DiffResult, error) {
	return core.ParseUnifiedDiffs(patch)
}

func GenerateDiff(before, after, filename string, context int) (patch string, additions, removals int) {
	return core.GenerateDiff(before, after, filename, context)
}

func MockDiffs(context int) ([]DiffResult, error) {
	return core.MockDiffs(context)
}

func RunDiffs(diffs []DiffResult, opts Options) error {
	return core.RunDiffs(diffs, opts)
}

func RunPatch(patch string, fileName string, opts Options) error {
	return core.RunPatch(patch, fileName, opts)
}

func RunFiles(before, after, filename string, context int, opts Options) error {
	return core.RunFiles(before, after, filename, context, opts)
}

func NewViewer(opts ViewerOptions) Viewer {
	return core.NewViewer(opts)
}
