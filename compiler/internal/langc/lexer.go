package langc

import (
	"fmt"
	"strings"
)

type TokenKind int

const (
	TokEOF TokenKind = iota
	TokNewline
	TokIdent
	TokInt
	TokFloat
	TokString
	TokChar
	TokTemplate
	TokKeyword
	TokPunct
)

type Pos struct {
	Line int
	Col  int
}

func (p Pos) String() string { return fmt.Sprintf("%d:%d", p.Line, p.Col) }

type Token struct {
	Kind TokenKind
	Text string
	Pos  Pos
}

var keywords = map[string]bool{
	"function": true, "export": true, "import": true, "from": true, "as": true,
	"const": true, "let": true, "if": true, "else": true, "while": true,
	"for": true, "of": true, "return": true, "break": true, "continue": true,
	"struct": true, "enum": true, "match": true, "true": true, "false": true,
	"mutable": true, "uses": true, "test": true,
}

// Longest punctuation first, so "==" is matched before "=".
var puncts = []string{
	"===", "!==", "...", "=>", "==", "!=", "<=", ">=", "&&", "||", "+%", "-%", "*%", "<<", ">>",
	"+", "-", "*", "/", "%", "<", ">", "=", "!", "&", "|", "^", "~",
	"(", ")", "{", "}", "[", "]", ",", ":", ".", ";",
}

type Lexer struct {
	src    []rune
	i      int
	line   int
	col    int
	depth  int // nesting of ( and [, where line breaks never end a statement
	last   *Token
	tokens []Token
}

func Lex(src string) ([]Token, error) {
	lx := &Lexer{src: []rune(src), line: 1, col: 1}
	if err := lx.run(); err != nil {
		return nil, err
	}
	return lx.tokens, nil
}

func (lx *Lexer) errorf(pos Pos, format string, args ...any) error {
	return fmt.Errorf("%s: %s", pos, fmt.Sprintf(format, args...))
}

func (lx *Lexer) peek(off int) rune {
	if lx.i+off < len(lx.src) {
		return lx.src[lx.i+off]
	}
	return 0
}

func (lx *Lexer) advance() rune {
	r := lx.src[lx.i]
	lx.i++
	if r == '\n' {
		lx.line++
		lx.col = 1
	} else {
		lx.col++
	}
	return r
}

func (lx *Lexer) emit(kind TokenKind, text string, pos Pos) {
	lx.tokens = append(lx.tokens, Token{Kind: kind, Text: text, Pos: pos})
	lx.last = &lx.tokens[len(lx.tokens)-1]
}

// A line break ends a statement only after a token that can end one.
func (lx *Lexer) lastEndsStatement() bool {
	if lx.last == nil {
		return false
	}
	switch lx.last.Kind {
	case TokIdent, TokInt, TokFloat, TokString, TokChar, TokTemplate:
		return true
	case TokKeyword:
		switch lx.last.Text {
		case "return", "break", "continue", "true", "false":
			return true
		}
	case TokPunct:
		switch lx.last.Text {
		case ")", "]", "}":
			return true
		}
	}
	return false
}

func (lx *Lexer) run() error {
	for lx.i < len(lx.src) {
		r := lx.peek(0)
		pos := Pos{lx.line, lx.col}
		switch {
		case r == '\n':
			lx.advance()
			if lx.depth == 0 && lx.lastEndsStatement() {
				lx.emit(TokNewline, "\n", pos)
			}
		case r == ' ' || r == '\t' || r == '\r':
			lx.advance()
		case r == '/' && lx.peek(1) == '/':
			for lx.i < len(lx.src) && lx.peek(0) != '\n' {
				lx.advance()
			}
		case r == '/' && lx.peek(1) == '*':
			lx.advance()
			lx.advance()
			for lx.i < len(lx.src) && !(lx.peek(0) == '*' && lx.peek(1) == '/') {
				lx.advance()
			}
			if lx.i >= len(lx.src) {
				return lx.errorf(pos, "unterminated block comment")
			}
			lx.advance()
			lx.advance()
		case isLetter(r):
			start := lx.i
			for lx.i < len(lx.src) && (isLetter(lx.peek(0)) || isDigit(lx.peek(0))) {
				lx.advance()
			}
			text := string(lx.src[start:lx.i])
			if keywords[text] {
				lx.emit(TokKeyword, text, pos)
			} else {
				lx.emit(TokIdent, text, pos)
			}
		case isDigit(r):
			if err := lx.number(pos); err != nil {
				return err
			}
		case r == '"':
			s, err := lx.quoted('"', pos)
			if err != nil {
				return err
			}
			lx.emit(TokString, s, pos)
		case r == '\'':
			s, err := lx.quoted('\'', pos)
			if err != nil {
				return err
			}
			if len([]rune(s)) != 1 {
				return lx.errorf(pos, "a character literal holds exactly one character; use double quotes for strings")
			}
			lx.emit(TokChar, s, pos)
		case r == '`':
			s, err := lx.template(pos)
			if err != nil {
				return err
			}
			lx.emit(TokTemplate, s, pos)
		default:
			matched := false
			for _, p := range puncts {
				if lx.hasPrefix(p) {
					for range p {
						lx.advance()
					}
					if p == ";" {
						return lx.errorf(pos, "semicolons are not used; end the statement with a line break")
					}
					switch p {
					case "(", "[":
						lx.depth++
					case ")", "]":
						if lx.depth > 0 {
							lx.depth--
						}
					}
					lx.emit(TokPunct, p, pos)
					matched = true
					break
				}
			}
			if !matched {
				return lx.errorf(pos, "unexpected character %q", r)
			}
		}
	}
	pos := Pos{lx.line, lx.col}
	if lx.lastEndsStatement() {
		lx.emit(TokNewline, "\n", pos)
	}
	lx.emit(TokEOF, "", pos)
	return nil
}

func (lx *Lexer) hasPrefix(p string) bool {
	for k, c := range p {
		if lx.peek(k) != c {
			return false
		}
	}
	return true
}

func (lx *Lexer) number(pos Pos) error {
	start := lx.i
	isFloat := false
	if lx.peek(0) == '0' && (lx.peek(1) == 'x' || lx.peek(1) == 'b' || lx.peek(1) == 'o') {
		lx.advance()
		lx.advance()
		for isHexDigit(lx.peek(0)) || lx.peek(0) == '_' {
			lx.advance()
		}
	} else {
		for isDigit(lx.peek(0)) || lx.peek(0) == '_' {
			lx.advance()
		}
		if lx.peek(0) == '.' && isDigit(lx.peek(1)) {
			isFloat = true
			lx.advance()
			for isDigit(lx.peek(0)) || lx.peek(0) == '_' {
				lx.advance()
			}
		}
		if lx.peek(0) == 'e' || lx.peek(0) == 'E' {
			isFloat = true
			lx.advance()
			if lx.peek(0) == '+' || lx.peek(0) == '-' {
				lx.advance()
			}
			for isDigit(lx.peek(0)) {
				lx.advance()
			}
		}
	}
	text := strings.ReplaceAll(string(lx.src[start:lx.i]), "_", "")
	if isFloat {
		lx.emit(TokFloat, text, pos)
	} else {
		lx.emit(TokInt, text, pos)
	}
	return nil
}

func (lx *Lexer) escape(pos Pos) (string, error) {
	lx.advance() // the backslash
	if lx.i >= len(lx.src) {
		return "", lx.errorf(pos, "unterminated escape")
	}
	c := lx.advance()
	switch c {
	case 'n':
		return "\n", nil
	case 't':
		return "\t", nil
	case 'r':
		return "\r", nil
	case '0':
		return "\x00", nil
	case '\\', '"', '\'', '`', '$':
		return string(c), nil
	case 'u':
		if lx.peek(0) != '{' {
			return "", lx.errorf(pos, "expected { after \\u")
		}
		lx.advance()
		v := 0
		for lx.peek(0) != '}' {
			if lx.i >= len(lx.src) || !isHexDigit(lx.peek(0)) {
				return "", lx.errorf(pos, "invalid \\u escape")
			}
			v = v*16 + hexValue(lx.advance())
		}
		lx.advance()
		return string(rune(v)), nil
	}
	return "", lx.errorf(pos, "unknown escape \\%c", c)
}

func (lx *Lexer) quoted(q rune, pos Pos) (string, error) {
	lx.advance()
	var b strings.Builder
	for {
		if lx.i >= len(lx.src) || lx.peek(0) == '\n' {
			return "", lx.errorf(pos, "unterminated literal")
		}
		c := lx.peek(0)
		if c == q {
			lx.advance()
			return b.String(), nil
		}
		if c == '\\' {
			s, err := lx.escape(pos)
			if err != nil {
				return "", err
			}
			b.WriteString(s)
			continue
		}
		b.WriteRune(lx.advance())
	}
}

// Templates keep their raw source. The parser splits out each ${...} part
// and parses it on its own.
func (lx *Lexer) template(pos Pos) (string, error) {
	lx.advance()
	start := lx.i
	for {
		if lx.i >= len(lx.src) {
			return "", lx.errorf(pos, "unterminated template string")
		}
		c := lx.peek(0)
		if c == '\\' {
			lx.advance()
			if lx.i < len(lx.src) {
				lx.advance()
			}
			continue
		}
		if c == '`' {
			raw := string(lx.src[start:lx.i])
			lx.advance()
			return raw, nil
		}
		lx.advance()
	}
}

func isLetter(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func isHexDigit(r rune) bool {
	return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func hexValue(r rune) int {
	switch {
	case isDigit(r):
		return int(r - '0')
	case r >= 'a' && r <= 'f':
		return int(r-'a') + 10
	default:
		return int(r-'A') + 10
	}
}
