package parser

import (
	"bytes"
	"fmt"
	"strings"

	sgd "github.com/sourcegraph/go-diff/diff"
)

func ParseUnifiedDiff(patch string) (DiffResult, error) {
	all, err := ParseUnifiedDiffs(patch)
	if err != nil {
		return DiffResult{}, err
	}
	if len(all) == 0 {
		return DiffResult{}, nil
	}
	return all[0], nil
}

func ParseUnifiedDiffs(patch string) ([]DiffResult, error) {
	files, err := sgd.ParseMultiFileDiff([]byte(patch))
	if err != nil {
		return nil, err
	}

	results := make([]DiffResult, 0, len(files))
	for _, f := range files {
		res := DiffResult{
			OldFile:     trimFilePrefix(f.OrigName),
			NewFile:     trimFilePrefix(f.NewName),
			DisplayFile: chooseDisplayName(trimFilePrefix(f.OrigName), trimFilePrefix(f.NewName)),
		}

		for _, h := range f.Hunks {
			hunk := Hunk{Header: fmt.Sprintf("-%d,%d +%d,%d", h.OrigStartLine, h.OrigLines, h.NewStartLine, h.NewLines)}

			oldNo := int(h.OrigStartLine)
			newNo := int(h.NewStartLine)

			lines := bytes.Split(bytes.TrimSuffix(h.Body, []byte("\n")), []byte("\n"))
			for _, raw := range lines {
				if len(raw) == 0 {
					continue
				}
				s := string(raw)
				prefix := s[0]
				content := ""
				if len(s) > 1 {
					content = s[1:]
				}

				switch prefix {
				case ' ':
					hunk.Lines = append(hunk.Lines, DiffLine{Kind: LineContext, OldLineNo: oldNo, NewLineNo: newNo, Content: content})
					oldNo++
					newNo++
				case '-':
					hunk.Lines = append(hunk.Lines, DiffLine{Kind: LineRemoved, OldLineNo: oldNo, NewLineNo: 0, Content: content})
					res.Removals++
					oldNo++
				case '+':
					hunk.Lines = append(hunk.Lines, DiffLine{Kind: LineAdded, OldLineNo: 0, NewLineNo: newNo, Content: content})
					res.Additions++
					newNo++
				case '\\':
					continue
				}
			}

			res.Hunks = append(res.Hunks, hunk)
		}

		results = append(results, res)
	}

	return results, nil
}

func trimFilePrefix(name string) string {
	name = strings.TrimPrefix(name, "a/")
	name = strings.TrimPrefix(name, "b/")
	return name
}

func chooseDisplayName(oldName, newName string) string {
	if newName != "" && newName != "/dev/null" {
		return newName
	}
	return oldName
}
