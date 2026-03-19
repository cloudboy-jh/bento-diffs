package bentodiffs

import "github.com/cloudboy-jh/bento-diffs/internal/parser"

func ParseUnifiedDiff(patch string) (DiffResult, error) {
	res, err := parser.ParseUnifiedDiff(patch)
	if err != nil {
		return DiffResult{}, err
	}
	return toPublicDiffResult(res), nil
}

func ParseUnifiedDiffs(patch string) ([]DiffResult, error) {
	all, err := parser.ParseUnifiedDiffs(patch)
	if err != nil {
		return nil, err
	}
	res := make([]DiffResult, 0, len(all))
	for _, d := range all {
		res = append(res, toPublicDiffResult(d))
	}
	return res, nil
}

func GenerateDiff(before, after, filename string, context int) (patch string, additions, removals int) {
	return parser.GenerateDiff(before, after, filename, context)
}

func MockDiffs(context int) ([]DiffResult, error) {
	beforeGo := "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"hello\")\n}\n"
	afterGo := "package main\n\nimport (\n    \"fmt\"\n    \"time\"\n)\n\nfunc main() {\n    fmt.Println(\"hello from bento-diffs\")\n    fmt.Println(time.Now().Format(time.RFC3339))\n}\n"

	beforeReadme := "# Demo\n\nA tiny demo file.\n\n- first item\n"
	afterReadme := "# Demo\n\nA tiny demo file for bento-diffs.\n\n- first item\n- second item\n"

	p1, _, _ := parser.GenerateDiff(beforeGo, afterGo, "cmd/demo/main.go", context)
	p2, _, _ := parser.GenerateDiff(beforeReadme, afterReadme, "README.md", context)
	all, err := parser.ParseUnifiedDiffs(p1 + "\n" + p2)
	if err != nil {
		return nil, err
	}
	out := make([]DiffResult, 0, len(all))
	for _, d := range all {
		out = append(out, toPublicDiffResult(d))
	}
	return out, nil
}
