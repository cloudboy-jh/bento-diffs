package parser

import (
	"strings"

	udiff "github.com/aymanbagabas/go-udiff"
)

func GenerateDiff(before, after, filename string, context int) (patch string, additions, removals int) {
	edits := udiff.Strings(before, after)
	if context < 0 {
		context = 0
	}
	if p, err := udiff.ToUnified("a/"+filename, "b/"+filename, before, edits, context); err == nil {
		patch = p
	} else {
		patch = udiff.Unified("a/"+filename, "b/"+filename, before, after)
	}

	for _, line := range strings.Split(strings.TrimSuffix(patch, "\n"), "\n") {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}
		if strings.HasPrefix(line, "+") {
			additions++
		}
		if strings.HasPrefix(line, "-") {
			removals++
		}
	}

	return patch, additions, removals
}
