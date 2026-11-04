package narra

import "strings"

// errorfFix reports a diagnostic that carries exact fixes.
func (c *Checker) errorfFix(pos Pos, fixes []Fix, format string, args ...any) {
	n := len(c.errs)
	c.errorf(pos, format, args...)
	if n < len(c.errs) {
		// In a chain `a + b + c` the inner errors start at the same text. Only
		// the outermost keeps its fix, because the others would leave an error.
		for i := 0; i < n && len(fixes) > 0; i++ {
			if old := c.errs[i].Fixes; len(old) > 0 && old[0].Title == fixes[0].Title && old[0].Edits[0].Start == fixes[0].Edits[0].Start {
				c.errs[i].Fixes = nil
			}
		}
		c.errs[n].Fixes = fixes
	}
}

// importFix adds `import * as ns from "path"` before the module's first import,
// or after the leading comment lines when it has none.
func (c *Checker) importFix(ns, path string) []Fix {
	at := Pos{1, 1}
	if c.mod != nil && c.mod.importAt.Line > 0 {
		at = c.mod.importAt
	}
	return []Fix{{Title: "import " + path, Edits: []Edit{{Start: at, End: at, Text: "import * as " + ns + " from \"" + path + "\"\n"}}}}
}

// concatFix rewrites `a + "x" + b` as the template string `${a}x${b}`. It
// offers the fix only when every operand is a string name or string literal,
// because only then are the start and the end of the text known exactly.
func concatFix(e *BinaryExpr) []Fix {
	var leaves []Expr
	var walk func(x Expr) bool
	walk = func(x Expr) bool {
		switch x := x.(type) {
		case *BinaryExpr:
			return x.Op == "+" && walk(x.Left) && walk(x.Right)
		case *Ident:
			leaves = append(leaves, x)
			return x.Type() == TStr
		case *StringLit:
			leaves = append(leaves, x)
			return true
		}
		return false
	}
	if !walk(e) {
		return nil
	}
	var b strings.Builder
	b.WriteByte('`')
	for _, leaf := range leaves {
		switch leaf := leaf.(type) {
		case *Ident:
			b.WriteString("${" + leaf.Name + "}")
		case *StringLit:
			text, ok := templateEscape(leaf.Value)
			if !ok {
				return nil
			}
			b.WriteString(text)
		}
	}
	b.WriteByte('`')
	first, last := leaves[0], leaves[len(leaves)-1]
	var end Pos
	switch last := last.(type) {
	case *Ident:
		end = Pos{last.P.Line, last.P.Col + len(last.Name)}
	case *StringLit:
		end = last.End
	}
	return []Fix{{Title: "join the strings with a template string", Edits: []Edit{{Start: first.position(), End: end, Text: b.String()}}}}
}

// templateEscape writes text so a template string reads it back unchanged.
func templateEscape(s string) (string, bool) {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\\' || r == '`' || r == '$':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7f:
			return "", false
		default:
			b.WriteRune(r)
		}
	}
	return b.String(), true
}

// pushFix rewrites `xs.push(v)` as `Arrays.push(mutable xs, v)`, with the alias the module uses. It checks that
// the text between the parts has no spaces, since the AST keeps no spacing.
func pushFix(call *CallExpr, callee *FieldExpr, t Type, alias string) []Fix {
	if alias == "" {
		alias = "Arrays"
	}
	if _, ok := t.(*ArrayType); !ok || callee.Name != "push" || len(call.Args) == 0 {
		return nil
	}
	id, ok := callee.Target.(*Ident)
	if !ok {
		return nil
	}
	dot := Pos{callee.P.Line, callee.P.Col - 1}
	if id.P.Line != dot.Line || id.P.Col+len(id.Name) != dot.Col || call.P.Line != dot.Line || call.P.Col != callee.P.Col+len(callee.Name) {
		return nil
	}
	return []Fix{{Title: "call " + alias + ".push", Edits: []Edit{
		{Start: id.P, End: id.P, Text: alias + ".push(mutable "},
		{Start: dot, End: Pos{call.P.Line, call.P.Col + 1}, Text: ", "},
	}}}
}

// quoteFix rewrites 'abc' as "abc". open and close are the positions of the
// two quote marks. It offers the fix only when the text has no escape and no
// double quote, because then swapping the marks keeps the meaning.
func quoteFix(open, close Pos, raw string) []Fix {
	if strings.ContainsAny(raw, "\\\"") {
		return nil
	}
	return []Fix{{Title: "use double quotes", Edits: []Edit{
		{Start: open, End: Pos{open.Line, open.Col + 1}, Text: "\""},
		{Start: close, End: Pos{close.Line, close.Col + 1}, Text: "\""},
	}}}
}

// stdAliasPaths maps the aliases the spec's import table names to their std
// module, for the add-import fix of an unknown module name.
var stdAliasPaths = map[string]string{
	"Arrays": "std/array", "Maps": "std/map", "Boxes": "std/box",
	"Ranges": "std/range", "Options": "std/option", "Results": "std/result",
	"Str": "std/string", "Int": "std/int", "Float": "std/float",
	"Char": "std/char", "Env": "std/env", "Fs": "std/fs",
	"Process": "std/process", "Test": "std/test",
}

// lengthFix rewrites `x.length` as `Arrays.length(x)`, `Str.length(x)` or
// `Maps.size(x)`, and adds the import when the module lacks it. The target
// must be a name that touches the dot on one line.
func (c *Checker) lengthFix(e *FieldExpr, t Type) []Fix {
	id, ok := e.Target.(*Ident)
	if !ok {
		return nil
	}
	dot := Pos{e.P.Line, e.P.Col - 1}
	if id.P.Line != dot.Line || id.P.Col+len(id.Name) != dot.Col {
		return nil
	}
	var alias, path, fn string
	switch tt := t.(type) {
	case *ArrayType:
		alias, path, fn = "Arrays", "std/array", "length"
	case *MapType:
		alias, path, fn = "Maps", "std/map", "size"
	case *PrimType:
		if tt != TStr {
			return nil
		}
		alias, path, fn = "Str", "std/string", "length"
	default:
		return nil
	}
	edit := Edit{Start: id.P, End: Pos{e.P.Line, e.P.Col + len("length")}}
	if have := c.aliasFor(path); have != "" {
		edit.Text = have + "." + fn + "(" + id.Name + ")"
		return []Fix{{Title: "call " + have + "." + fn, Edits: []Edit{edit}}}
	}
	edit.Text = alias + "." + fn + "(" + id.Name + ")"
	fix := c.importFix(alias, path)
	fix[0].Title = "call " + alias + "." + fn
	fix[0].Edits = append(fix[0].Edits, edit)
	return fix
}

// tokenAt returns the token n places after the current one, or the final EOF.
func (p *Parser) tokenAt(n int) Token {
	return p.toks[min(p.i+n, len(p.toks)-1)]
}

// habitDecl reports `class` and `interface` where a declaration should start.
// The fix renames the keyword to `struct` only when the body holds fields
// and nothing else, because only then the result is a valid struct.
func (p *Parser) habitDecl() {
	t := p.tok()
	if t.Kind != TokIdent || (t.Text != "class" && t.Text != "interface") {
		return
	}
	var fixes []Fix
	if p.fieldsOnlyBody() {
		fixes = []Fix{{Title: "use struct", Edits: []Edit{{Start: t.Pos, End: t.End, Text: "struct"}}}}
	}
	if t.Text == "class" {
		p.failFix(t.Pos, fixes, "there is no `class`; use a `struct` and functions in a module")
	}
	p.failFix(t.Pos, fixes, "there is no `interface`; use a `struct` for data or a `trait` for behavior")
}

// fieldsOnlyBody reports whether the tokens after the current `class` or
// `interface` are `Name {` and then lines of `field: type`, with no method,
// modifier, separator or generic parameter.
func (p *Parser) fieldsOnlyBody() bool {
	if p.tokenAt(1).Kind != TokIdent || p.tokenAt(2).Kind != TokPunct || p.tokenAt(2).Text != "{" {
		return false
	}
	n := 3
	for {
		for p.tokenAt(n).Kind == TokNewline {
			n++
		}
		if p.tokenAt(n).Kind == TokPunct && p.tokenAt(n).Text == "}" {
			return true
		}
		if p.tokenAt(n).Kind != TokIdent || p.tokenAt(n+1).Kind != TokPunct || p.tokenAt(n+1).Text != ":" {
			return false
		}
		n += 2
		// The type runs to the end of the line.
		for p.tokenAt(n).Kind != TokNewline {
			t := p.tokenAt(n)
			if t.Kind == TokEOF || (t.Kind == TokPunct && (t.Text == ";" || t.Text == "," || t.Text == "{" || t.Text == "}" || t.Text == "?" || t.Text == "(" || t.Text == "=")) {
				return false
			}
			n++
		}
	}
}

// habitExpr reports `this`, `new`, `try` and `throw` at the start of an
// expression. t is the identifier just consumed; p.tok() is the token after it.
func (p *Parser) habitExpr(t Token) {
	next := p.tok()
	switch t.Text {
	case "this":
		p.fail(t.Pos, "there is no `this`; pass the value as a parameter")
	case "new":
		if next.Kind == TokIdent {
			p.fail(t.Pos, "there is no `new`; write a struct literal such as `Name { field: value }`, or call a function")
		}
	case "try":
		if next.Kind == TokPunct && next.Text == "{" {
			p.fail(t.Pos, "there is no `try` or `catch`; return a `Result` and check it with `is Err` or `match`")
		}
	case "throw":
		switch next.Kind {
		case TokIdent, TokString, TokTemplate, TokInt, TokFloat, TokChar:
			var fixes []Fix
			if p.resultFunc && p.startsStatement(t) {
				if end, ok := p.speculateExpr(); ok {
					fixes = []Fix{{Title: "return an Err", Edits: []Edit{
						{Start: t.Pos, End: next.Pos, Text: "return Err("},
						{Start: end, End: end, Text: ")"},
					}}}
				}
			}
			p.failFix(t.Pos, fixes, "there is no `throw`; return `Err(value)`, or call `panic` for a bug")
		}
	}
}

// startsStatement reports whether t, the token just consumed, begins a
// statement.
func (p *Parser) startsStatement(t Token) bool {
	if p.i < 2 {
		return false
	}
	before := p.toks[p.i-2]
	return before.Kind == TokNewline || (before.Kind == TokPunct && (before.Text == "{" || before.Text == "}"))
}

// speculateExpr parses an expression at the current token and reports where
// it ends, which must be the end of the line. The parser position is
// restored; a syntax error gives ok false.
func (p *Parser) speculateExpr() (end Pos, ok bool) {
	saved := p.i
	defer func() {
		if r := recover(); r != nil {
			if _, isParse := r.(parseError); !isParse {
				panic(r)
			}
			ok = false
		}
		p.i = saved
	}()
	p.expr()
	last := p.toks[p.i-1]
	if after := p.tok(); after.Kind != TokNewline && !(after.Kind == TokPunct && after.Text == "}") {
		return last.End, false
	}
	return last.End, true
}
