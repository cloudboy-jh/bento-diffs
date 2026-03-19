package emptystatepane

import (
	tea "charm.land/bubbletea/v2"
	"github.com/cloudboy-jh/bentotui/registry/bricks/card"
	"github.com/cloudboy-jh/bentotui/registry/bricks/text"
	"github.com/cloudboy-jh/bentotui/theme"
)

type Model struct {
	Card *card.Model
	Body *text.Model
}

func New(title, message string, t theme.Theme) *Model {
	if t == nil {
		t = theme.CurrentTheme()
	}
	body := text.New(message)
	panel := card.New(
		card.Title(title),
		card.Flat(),
		card.Content(body),
		card.WithTheme(t),
	)
	return &Model{Card: panel, Body: body}
}

func (m *Model) SetSize(width, height int) {
	m.Card.SetSize(width, height)
}

func (m *Model) View() tea.View { return m.Card.View() }

func (m *Model) SetTheme(t theme.Theme) {
	m.Card.SetTheme(t)
}

func (m *Model) SetTitle(title string) {
	m.Card.SetTitle(title)
}

func (m *Model) SetMessage(message string) {
	m.Body.SetText(message)
}
