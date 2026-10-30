package narra

import (
	"fmt"
	"strings"
)

// LexDump lexes src and returns one line per token, `<line>:<col> <kind>
// <text>`, with the text escaped so it stays on one line. When lexing fails
// the result is the single line `error <message>` and ok is false.
// `narra lex` prints it, and the self-hosted lexer in selfhost/ prints the
// same format, so the two can be diffed.
func LexDump(src string) (dump string, ok bool) {
	tokens, err := Lex(src)
	if err != nil {
		return "error " + err.Error() + "\n", false
	}
	var b strings.Builder
	for _, t := range tokens {
		fmt.Fprintf(&b, "%d:%d %s %s\n", t.Pos.Line, t.Pos.Col, tokenKindName(t.Kind), escapeTokenText(t.Text))
	}
	return b.String(), true
}

func tokenKindName(k TokenKind) string {
	switch k {
	case TokEOF:
		return "eof"
	case TokNewline:
		return "newline"
	case TokIdent:
		return "ident"
	case TokInt:
		return "int"
	case TokFloat:
		return "float"
	case TokString:
		return "string"
	case TokChar:
		return "char"
	case TokTemplate:
		return "template"
	case TokKeyword:
		return "keyword"
	default:
		return "punct"
	}
}

// escapeTokenText writes `\` as `\\`, newline, tab and carriage return as
// `\n`, `\t` and `\r`, and other control characters as `\u{<hex>}`.
func escapeTokenText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\r':
			b.WriteString(`\r`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\u{%x}`, r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
