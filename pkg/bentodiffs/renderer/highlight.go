package renderer

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	chromastyles "github.com/alecthomas/chroma/v2/styles"
	"github.com/cloudboy-jh/bentotui/theme"
)

func SyntaxHighlight(fileName, content string, t theme.Theme, lineBG color.Color) string {
	if content == "" {
		return ""
	}

	lexer := lexers.Match(fileName)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	iter, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}

	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		return content
	}

	bg := hexColor(lineBG)
	textFG := hexColor(t.Text())
	keywordFG := hexColor(t.SyntaxKeyword())
	typeFG := hexColor(t.SyntaxType())
	funcFG := hexColor(t.SyntaxFunction())
	varFG := hexColor(t.SyntaxVariable())
	stringFG := hexColor(t.SyntaxString())
	numberFG := hexColor(t.SyntaxNumber())
	commentFG := hexColor(t.SyntaxComment())
	opFG := hexColor(t.SyntaxOperator())
	punctFG := hexColor(t.SyntaxPunctuation())
	warnFG := hexColor(t.Warning())
	errorFG := hexColor(t.Error())
	mutedFG := hexColor(t.TextMuted())
	addedFG := hexColor(t.DiffAdded())
	removedFG := hexColor(t.DiffRemoved())

	style, err := chroma.NewStyle("bento-diff", chroma.StyleEntries{
		chroma.Background:             withBG("", bg),
		chroma.Text:                   withBG(textFG, bg),
		chroma.Keyword:                withBG(keywordFG, bg, "bold"),
		chroma.KeywordType:            withBG(typeFG, bg),
		chroma.NameFunction:           withBG(funcFG, bg),
		chroma.NameVariable:           withBG(varFG, bg),
		chroma.LiteralString:          withBG(stringFG, bg),
		chroma.LiteralNumber:          withBG(numberFG, bg),
		chroma.Comment:                withBG(commentFG, bg, "italic"),
		chroma.Operator:               withBG(opFG, bg),
		chroma.Punctuation:            withBG(punctFG, bg),
		chroma.GenericDeleted:         withBG(removedFG, bg),
		chroma.GenericInserted:        withBG(addedFG, bg),
		chroma.GenericSubheading:      withBG(mutedFG, bg),
		chroma.GenericEmph:            withBG(textFG, bg),
		chroma.GenericStrong:          withBG(textFG, bg, "bold"),
		chroma.GenericOutput:          withBG(textFG, bg),
		chroma.GenericTraceback:       withBG(errorFG, bg),
		chroma.LiteralStringAffix:     withBG(stringFG, bg),
		chroma.LiteralStringBacktick:  withBG(stringFG, bg),
		chroma.LiteralStringChar:      withBG(stringFG, bg),
		chroma.LiteralStringDelimiter: withBG(stringFG, bg),
		chroma.LiteralStringDoc:       withBG(commentFG, bg),
		chroma.LiteralStringDouble:    withBG(stringFG, bg),
		chroma.LiteralStringEscape:    withBG(warnFG, bg),
		chroma.LiteralStringHeredoc:   withBG(stringFG, bg),
		chroma.LiteralStringInterpol:  withBG(varFG, bg),
		chroma.LiteralStringOther:     withBG(stringFG, bg),
		chroma.LiteralStringRegex:     withBG(stringFG, bg),
		chroma.LiteralStringSingle:    withBG(stringFG, bg),
		chroma.LiteralStringSymbol:    withBG(stringFG, bg),
	})
	if err != nil || style == nil {
		style = chromastyles.Fallback
	}

	buf := &bytes.Buffer{}
	if err := formatter.Format(buf, style, iter); err != nil {
		return content
	}
	return buf.String()
}

func hexColor(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func withBG(fgHex, bgHex string, attrs ...string) string {
	out := ""
	if fgHex != "" {
		out = "#" + fgHex
	}
	if bgHex != "" {
		if out != "" {
			out += " "
		}
		out += "bg:#" + bgHex
	}
	for _, attr := range attrs {
		if attr == "" {
			continue
		}
		if out != "" {
			out += " "
		}
		out += attr
	}
	return out
}
