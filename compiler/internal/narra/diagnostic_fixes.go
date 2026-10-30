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

// pushFix rewrites `xs.push(v)` as `Array.push(mutable xs, v)`. It checks that
// the text between the parts has no spaces, since the AST keeps no spacing.
func pushFix(call *CallExpr, callee *FieldExpr, t Type) []Fix {
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
	return []Fix{{Title: "call Array.push", Edits: []Edit{
		{Start: id.P, End: id.P, Text: "Array.push(mutable "},
		{Start: dot, End: Pos{call.P.Line, call.P.Col + 1}, Text: ", "},
	}}}
}
