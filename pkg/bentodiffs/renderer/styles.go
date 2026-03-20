package renderer

import (
	"image/color"

	"github.com/cloudboy-jh/bentotui/theme"
)

type styles struct {
	removedBG    color.Color
	addedBG      color.Color
	contextBG    color.Color
	lineNumberFG color.Color
	removedFG    color.Color
	addedFG      color.Color
}

func createStyles(t theme.Theme) styles {
	return styles{
		removedBG:    t.DiffRemovedBG(),
		addedBG:      t.DiffAddedBG(),
		contextBG:    t.DiffContextBG(),
		lineNumberFG: t.DiffLineNum(),
		removedFG:    t.DiffRemoved(),
		addedFG:      t.DiffAdded(),
	}
}
