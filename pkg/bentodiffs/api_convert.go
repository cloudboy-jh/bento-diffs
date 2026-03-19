package bentodiffs

import "github.com/cloudboy-jh/bento-diffs/internal/parser"

func toPublicDiffResult(in parser.DiffResult) DiffResult {
	out := DiffResult{
		OldFile:     in.OldFile,
		NewFile:     in.NewFile,
		DisplayFile: in.DisplayFile,
		Additions:   in.Additions,
		Removals:    in.Removals,
		Hunks:       make([]Hunk, 0, len(in.Hunks)),
	}
	for _, h := range in.Hunks {
		out.Hunks = append(out.Hunks, toPublicHunk(h))
	}
	return out
}

func toPublicHunk(in parser.Hunk) Hunk {
	out := Hunk{Header: in.Header, Lines: make([]DiffLine, 0, len(in.Lines))}
	for _, l := range in.Lines {
		out.Lines = append(out.Lines, toPublicDiffLine(l))
	}
	return out
}

func toPublicDiffLine(in parser.DiffLine) DiffLine {
	out := DiffLine{
		Kind:      LineType(in.Kind),
		OldLineNo: in.OldLineNo,
		NewLineNo: in.NewLineNo,
		Content:   in.Content,
		Segments:  make([]Segment, 0, len(in.Segments)),
	}
	for _, s := range in.Segments {
		out.Segments = append(out.Segments, Segment{
			Start: s.Start,
			End:   s.End,
			Type:  SegmentType(s.Type),
			Text:  s.Text,
		})
	}
	return out
}

func toInternalDiffResult(in DiffResult) parser.DiffResult {
	out := parser.DiffResult{
		OldFile:     in.OldFile,
		NewFile:     in.NewFile,
		DisplayFile: in.DisplayFile,
		Additions:   in.Additions,
		Removals:    in.Removals,
		Hunks:       make([]parser.Hunk, 0, len(in.Hunks)),
	}
	for _, h := range in.Hunks {
		out.Hunks = append(out.Hunks, toInternalHunk(h))
	}
	return out
}

func toInternalHunk(in Hunk) parser.Hunk {
	out := parser.Hunk{Header: in.Header, Lines: make([]parser.DiffLine, 0, len(in.Lines))}
	for _, l := range in.Lines {
		out.Lines = append(out.Lines, toInternalDiffLine(l))
	}
	return out
}

func toInternalDiffLine(in DiffLine) parser.DiffLine {
	out := parser.DiffLine{
		Kind:      parser.LineType(in.Kind),
		OldLineNo: in.OldLineNo,
		NewLineNo: in.NewLineNo,
		Content:   in.Content,
		Segments:  make([]parser.Segment, 0, len(in.Segments)),
	}
	for _, s := range in.Segments {
		out.Segments = append(out.Segments, parser.Segment{
			Start: s.Start,
			End:   s.End,
			Type:  parser.SegmentType(s.Type),
			Text:  s.Text,
		})
	}
	return out
}
