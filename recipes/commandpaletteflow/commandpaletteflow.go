package commandpaletteflow

import (
	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bentotui/registry/bricks/dialog"
)

func Open(commands []dialog.Command) tea.Cmd {
	return func() tea.Msg {
		palette := dialog.NewCommandPalette(commands)
		return dialog.Open(dialog.Custom{
			DialogTitle: "Commands",
			Content:     palette,
			Width:       56,
			Height:      18,
		})
	}
}
