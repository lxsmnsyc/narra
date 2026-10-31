package narra

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Parser struct {
	toks []Token
	i    int
}

// templateDot finds `Name.` inside a template string. The parser does not keep
// positions inside ${...}, so a fix cannot rename a name used there.
var templateDot = regexp.MustCompile(`([A-Za-z_][A-Za-z_0-9]*)\.`)

func Parse(src string) (*File, error) {
	toks, err := Lex(src)
	if err != nil {
		return nil, syntaxDiagnostic(err)
	}
	p := &Parser{toks: toks}
	f, err := p.file()
	if err != nil {
		return nil, syntaxDiagnostic(err)
	}
	f.dotUses = map[string][]dotUse{}
	f.inTemplate = map[string]bool{}
	for i := 0; i < len(toks); i++ {
		if toks[i].Kind == TokTemplate {
			for _, m := range templateDot.FindAllStringSubmatch(toks[i].Text, -1) {
				f.inTemplate[m[1]] = true
			}
		}
		if i+1 < len(toks) && toks[i].Kind == TokIdent && toks[i+1].Kind == TokPunct && toks[i+1].Text == "." &&
			(i == 0 || toks[i-1].Kind != TokPunct || toks[i-1].Text != ".") {
			member := ""
			if i+2 < len(toks) && toks[i+2].Kind == TokIdent {
				member = toks[i+2].Text
			}
			f.dotUses[toks[i].Text] = append(f.dotUses[toks[i].Text], dotUse{toks[i].Pos, member})
		}
	}
	// The header ends at the last leading comment line that a blank line
	// follows; a comment right above a declaration is its doc comment.
	f.Header = 1
	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines) && strings.HasPrefix(lines[i], "//"); i++ {
		if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "" {
			f.Header = i + 2
		}
	}
	return f, nil
}

type parseError struct{ d Diagnostic }

func (p *Parser) fail(pos Pos, format string, args ...any) {
	p.failFix(pos, nil, format, args...)
}

// failFix stops parsing with an error that carries exact fixes.
func (p *Parser) failFix(pos Pos, fixes []Fix, format string, args ...any) {
	panic(parseError{Diagnostic{Code: codeFor(format), Pos: pos, Message: fmt.Sprintf(format, args...), Fixes: fixes}})
}

func (p *Parser) tok() Token { return p.toks[p.i] }

func (p *Parser) next() Token {
	t := p.toks[p.i]
	if p.i < len(p.toks)-1 {
		p.i++
	}
	return t
}

func (p *Parser) is(text string) bool {
	t := p.tok()
	return (t.Kind == TokPunct || t.Kind == TokKeyword) && t.Text == text
}

func (p *Parser) accept(text string) bool {
	if p.is(text) {
		p.next()
		return true
	}
	return false
}

func (p *Parser) expect(text string) Token {
	if !p.is(text) {
		p.fail(p.tok().Pos, "expected %q, found %s", text, describe(p.tok()))
	}
	return p.next()
}

func (p *Parser) ident() Token {
	if p.tok().Kind != TokIdent {
		p.fail(p.tok().Pos, "expected a name, found %s", describe(p.tok()))
	}
	return p.next()
}

func (p *Parser) skipNewlines() {
	for p.tok().Kind == TokNewline {
		p.next()
	}
}

// endStatement accepts the end of a statement: a line break, or a closing
// brace or end of file that follows on the same line.
func (p *Parser) endStatement() {
	switch {
	case p.tok().Kind == TokNewline:
		p.skipNewlines()
	case p.is("}") || p.tok().Kind == TokEOF:
	default:
		p.fail(p.tok().Pos, "expected the end of the statement, found %s", describe(p.tok()))
	}
}

// endTypeDecl ends a declaration whose last token is a type. The lexer emits
// no line break after a `>`, since `>` also continues an expression, so a
// declaration ending in a generic type is ended by the next token on a later
// line.
func (p *Parser) endTypeDecl() {
	if p.i > 0 {
		prev, cur := p.toks[p.i-1], p.tok()
		if prev.Kind == TokPunct && prev.Text == ">" && cur.Pos.Line > prev.Pos.Line {
			return
		}
	}
	p.endStatement()
}

func describe(t Token) string {
	switch t.Kind {
	case TokEOF:
		return "end of file"
	case TokNewline:
		return "line break"
	}
	return strconv.Quote(t.Text)
}

func (p *Parser) file() (f *File, err error) {
	defer func() {
		if r := recover(); r != nil {
			pe, ok := r.(parseError)
			if !ok {
				panic(r)
			}
			err = pe.d
		}
	}()
	f = &File{}
	p.skipNewlines()
	for p.is("import") {
		f.Imports = append(f.Imports, p.importDecl())
		p.skipNewlines()
	}
	for p.tok().Kind != TokEOF {
		f.Decls = append(f.Decls, p.decl())
		p.skipNewlines()
	}
	return f, nil
}

func (p *Parser) importDecl() *ImportDecl {
	pos := p.expect("import").Pos
	d := &ImportDecl{Pos: pos}
	if p.accept("*") {
		p.expect("as")
		d.AliasPos = p.tok().Pos
		d.Alias = p.ident().Text
	} else if p.accept("{") {
		p.skipNewlines()
		for !p.is("}") {
			n := p.ident().Text
			local := n
			if p.accept("as") {
				local = p.ident().Text
			}
			d.Names = append(d.Names, ImportName{Name: n, Local: local})
			p.skipNewlines()
			if !p.accept(",") {
				break
			}
			p.skipNewlines()
		}
		p.skipNewlines()
		p.expect("}")
	} else {
		at := p.tok().Pos
		var fixes []Fix
		if p.tok().Kind == TokIdent {
			fixes = []Fix{{Title: "import the module as a namespace", Edits: []Edit{{Start: at, End: at, Text: "* as "}}}}
		}
		p.failFix(at, fixes, "use `import * as Name from \"path\"` or `import { a, b } from \"path\"`; there are no default imports")
	}
	p.expect("from")
	if p.tok().Kind != TokString {
		p.fail(p.tok().Pos, "expected a module path string")
	}
	d.Path = p.next().Text
	p.endStatement()
	return d
}

func (p *Parser) decl() Decl {
	exported := p.accept("export")
	if d := p.opaqueOrDistinct(exported); d != nil {
		return d
	}
	switch {
	case p.is("function"):
		return p.funcDecl(exported)
	case p.is("const"):
		return p.constDecl(exported)
	case p.is("struct"):
		return p.structDecl(exported)
	case p.is("enum"):
		return p.enumDecl(exported)
	case p.is("trait"):
		return p.traitDecl(exported)
	case p.is("implement"):
		if exported {
			p.fail(p.tok().Pos, "implement blocks cannot be exported; export the trait and the type")
		}
		return p.implDecl()
	case p.is("test"):
		if exported {
			p.fail(p.tok().Pos, "tests cannot be exported")
		}
		pos := p.next().Pos
		if p.tok().Kind != TokString {
			p.fail(p.tok().Pos, "expected the test name as a string")
		}
		name := p.next().Text
		return &TestDecl{Pos: pos, Name: name, Body: p.block()}
	}
	p.fail(p.tok().Pos, "expected a declaration, found %s", describe(p.tok()))
	return nil
}

func (p *Parser) typeExpr() *TypeExpr {
	if p.is("(") {
		return p.funcTypeExpr()
	}
	if p.is("#[") {
		// A tuple type, written like the literal: #[int, string].
		te := &TypeExpr{Pos: p.next().Pos, Tuple: true}
		for !p.is("]") {
			te.Args = append(te.Args, p.typeExpr())
			if !p.accept(",") {
				break
			}
		}
		p.expect("]")
		return te
	}
	t := p.ident()
	te := &TypeExpr{Pos: t.Pos, Name: t.Text}
	if p.accept(".") { // a type from a namespace import, such as Text.Token
		te.Name += "." + p.ident().Text
	}
	if p.accept("<") {
		for {
			te.Args = append(te.Args, p.typeExpr())
			if !p.accept(",") {
				break
			}
		}
		// In `Array<Array<int>>` the lexer reads `>>` as one token. Split it.
		if p.is(">>") {
			t := p.tok()
			p.toks[p.i] = Token{Kind: TokPunct, Text: ">", Pos: Pos{t.Pos.Line, t.Pos.Col + 1}}
			return te
		}
		p.expect(">")
	}
	if p.is("[") && p.toks[p.i+1].Kind == TokPunct && p.toks[p.i+1].Text == "]" {
		open, shut := p.tok().Pos, p.toks[p.i+1].Pos
		fix := Fix{Title: "write Array<...>", Edits: []Edit{
			{Start: te.Pos, End: te.Pos, Text: "Array<"},
			{Start: open, End: Pos{shut.Line, shut.Col + 1}, Text: ">"},
		}}
		p.failFix(open, []Fix{fix}, "there is no `T[]` array type; write `Array<T>`")
	}
	return te
}

// typeParams parses `<A, B: Show & Ordered>` and returns the names and the
// trait constraints of each.
func (p *Parser) typeParams() ([]string, [][]*TypeExpr) {
	names, bounds, _ := p.typeParamsEffects()
	return names, bounds
}

// typeParamsEffects also reads effect variables, `<T, effect E>`. They are
// returned apart from the type parameters (check_effects.go).
func (p *Parser) typeParamsEffects() ([]string, [][]*TypeExpr, []string) {
	var names []string
	var bounds [][]*TypeExpr
	var effects []string
	if !p.accept("<") {
		return nil, nil, nil
	}
	for !p.is(">") {
		if t := p.tok(); t.Kind == TokIdent && t.Text == "effect" && p.i+1 < len(p.toks) && p.toks[p.i+1].Kind == TokIdent {
			p.next()
			effects = append(effects, p.ident().Text)
			if !p.accept(",") {
				break
			}
			continue
		}
		names = append(names, p.ident().Text)
		var bs []*TypeExpr
		if p.accept(":") {
			for {
				t := p.ident()
				bs = append(bs, &TypeExpr{Pos: t.Pos, Name: t.Text})
				if !p.accept("&") {
					break
				}
			}
		}
		bounds = append(bounds, bs)
		if !p.accept(",") {
			break
		}
	}
	p.expect(">")
	return names, bounds, effects
}

// typeParamsNoBounds parses type parameters of a struct or enum, which cannot
// have constraints in this compiler.
func (p *Parser) typeParamsNoBounds() []string {
	pos := p.tok().Pos
	names, bounds := p.typeParams()
	for _, bs := range bounds {
		if len(bs) > 0 {
			p.fail(pos, "trait constraints on struct and enum type parameters are not supported by this compiler yet")
		}
	}
	return names
}

// constDecl parses a top-level `const NAME: T = value`.
func (p *Parser) constDecl(exported bool) *ConstDecl {
	d := &ConstDecl{Pos: p.expect("const").Pos, Exported: exported}
	d.Name = p.ident().Text
	if p.accept(":") {
		d.Type = p.typeExpr()
	}
	p.expect("=")
	d.Value = p.expr()
	p.endStatement()
	return d
}

func (p *Parser) funcDecl(exported bool) *FuncDecl {
	fd := p.funcHeader(exported)
	fd.Body = p.block()
	return fd
}

// funcHeader parses a function declaration up to its body.
func (p *Parser) funcHeader(exported bool) *FuncDecl {
	pos := p.expect("function").Pos
	if p.is("*") {
		p.fail(p.tok().Pos, "generators are not supported by this compiler yet")
	}
	fd := &FuncDecl{Pos: pos, Exported: exported, Name: p.ident().Text}
	fd.TypeParams, fd.Bounds, fd.EffectParams = p.typeParamsEffects()
	p.expect("(")
	for !p.is(")") {
		param := &Param{Pos: p.tok().Pos}
		param.Mutable = p.accept("mutable")
		param.Escaping = !param.Mutable && p.acceptEscaping()
		param.Name = p.ident().Text
		p.expect(":")
		param.Type = p.typeExpr()
		fd.Params = append(fd.Params, param)
		if !p.accept(",") {
			break
		}
	}
	p.expect(")")
	if p.accept(":") {
		fd.Result = p.typeExpr()
	}
	fd.Effects = p.usesList()
	return fd
}

func (p *Parser) fieldList() []*FieldDecl {
	var fields []*FieldDecl
	p.skipNewlines()
	for !p.is("}") {
		fdecl := &FieldDecl{Pos: p.tok().Pos, Name: p.ident().Text}
		p.expect(":")
		fdecl.Type = p.typeExpr()
		if p.accept("=") {
			fdecl.Default = p.expr()
		}
		fields = append(fields, fdecl)
		p.accept(",")
		p.skipNewlines()
	}
	return fields
}

func (p *Parser) structDecl(exported bool) *StructDecl {
	pos := p.expect("struct").Pos
	sd := &StructDecl{Pos: pos, Exported: exported, Name: p.ident().Text}
	sd.TypeParams = p.typeParamsNoBounds()
	sd.Implements, sd.Derives = p.traitLists()
	p.expect("{")
	sd.Fields = p.fieldList()
	p.expect("}")
	return sd
}

func (p *Parser) enumDecl(exported bool) *EnumDecl {
	pos := p.expect("enum").Pos
	ed := &EnumDecl{Pos: pos, Exported: exported, Name: p.ident().Text}
	ed.TypeParams = p.typeParamsNoBounds()
	ed.Implements, ed.Derives = p.traitLists()
	p.expect("{")
	p.skipNewlines()
	for !p.is("}") {
		v := &VariantDecl{Pos: p.tok().Pos, Name: p.ident().Text}
		if p.accept("(") {
			for !p.is(")") {
				f := &FieldDecl{Pos: p.tok().Pos, Name: p.ident().Text}
				p.expect(":")
				f.Type = p.typeExpr()
				v.Fields = append(v.Fields, f)
				if !p.accept(",") {
					break
				}
			}
			p.expect(")")
		}
		ed.Variants = append(ed.Variants, v)
		p.accept(",")
		p.skipNewlines()
	}
	p.expect("}")
	return ed
}

// traitLists parses `implements A, B` and `derives C, D` after a type name.
func (p *Parser) traitLists() (impls, derives []*TypeExpr) {
	names := func() []*TypeExpr {
		var out []*TypeExpr
		for {
			t := p.ident()
			out = append(out, &TypeExpr{Pos: t.Pos, Name: t.Text})
			if !p.accept(",") {
				return out
			}
		}
	}
	if p.accept("implements") {
		impls = names()
	}
	if p.accept("derives") {
		derives = names()
	}
	return impls, derives
}

func (p *Parser) traitDecl(exported bool) *TraitDecl {
	pos := p.expect("trait").Pos
	td := &TraitDecl{Pos: pos, Exported: exported, Name: p.ident().Text}
	if p.is("<") {
		p.fail(p.tok().Pos, "a trait has no type parameters; use Self for the implementing type")
	}
	p.expect("{")
	p.skipNewlines()
	for !p.is("}") {
		if !p.is("function") {
			p.fail(p.tok().Pos, "a trait has functions only, found %s", describe(p.tok()))
		}
		fd := p.funcHeader(false)
		if p.is("{") {
			p.fail(p.tok().Pos, "trait functions have no default implementation; remove the body")
		}
		td.Funcs = append(td.Funcs, fd)
		p.endTypeDecl()
	}
	p.expect("}")
	return td
}

func (p *Parser) implDecl() *ImplDecl {
	pos := p.expect("implement").Pos
	d := &ImplDecl{Pos: pos}
	t := p.ident()
	d.Trait = &TypeExpr{Pos: t.Pos, Name: t.Text}
	p.expect("for")
	d.Type = p.typeExpr()
	p.expect("{")
	p.skipNewlines()
	for !p.is("}") {
		if p.is("export") {
			p.fail(p.tok().Pos, "functions in an implement block cannot be exported")
		}
		if !p.is("function") {
			p.fail(p.tok().Pos, "an implement block has functions only, found %s", describe(p.tok()))
		}
		d.Funcs = append(d.Funcs, p.funcDecl(false))
		p.skipNewlines()
	}
	p.expect("}")
	return d
}

// funcTypeExpr parses a function type such as `(item: T) => U`. Parameter
// names are optional and ignored.
func (p *Parser) funcTypeExpr() *TypeExpr {
	te := &TypeExpr{Pos: p.expect("(").Pos, Func: true}
	for !p.is(")") {
		if p.tok().Kind == TokIdent && p.i+1 < len(p.toks) && p.toks[p.i+1].Kind == TokPunct && p.toks[p.i+1].Text == ":" {
			p.next()
			p.next()
		}
		te.Args = append(te.Args, p.typeExpr())
		if !p.accept(",") {
			break
		}
	}
	p.expect(")")
	p.expect("=>")
	te.Result = p.typeExpr()
	te.Uses = p.usesList()
	return te
}

// usesList reads an optional `uses a, b` list of effect names.
func (p *Parser) usesList() []string {
	var names []string
	if p.accept("uses") {
		for {
			names = append(names, p.ident().Text)
			if p.is(",") && p.paramAfterComma() {
				break
			}
			if !p.accept(",") {
				break
			}
		}
	}
	return names
}

// acceptEscaping accepts the `escaping` marker before a parameter name.
// It is not a keyword, so a parameter can still be named escaping.
func (p *Parser) acceptEscaping() bool {
	t := p.tok()
	if t.Kind == TokIdent && t.Text == "escaping" && p.i+1 < len(p.toks) && p.toks[p.i+1].Kind == TokIdent {
		p.next()
		return true
	}
	return false
}

// arrowAhead reports whether the `(` at the current token starts an arrow
// function: its matching `)` is followed by `=>` or by `:` and a result type.
func (p *Parser) arrowAhead() bool {
	depth := 0
	for j := p.i; j < len(p.toks); j++ {
		t := p.toks[j]
		if t.Kind == TokEOF {
			return false
		}
		if t.Kind != TokPunct {
			continue
		}
		switch t.Text {
		case "(":
			depth++
		case ")":
			depth--
			if depth == 0 {
				if j+1 >= len(p.toks) || p.toks[j+1].Kind != TokPunct {
					return false
				}
				next := p.toks[j+1].Text
				return next == "=>" || next == ":"
			}
		}
	}
	return false
}

// arrow parses `(a: A, b) : R => body`, where the body is an expression or
// a block. Parameter and result types may be left out.
func (p *Parser) arrow() Expr {
	a := &ArrowExpr{exprBase: exprBase{P: p.expect("(").Pos}}
	for !p.is(")") {
		if p.is("mutable") {
			p.fail(p.tok().Pos, "arrow function parameters cannot be mutable")
		}
		param := &Param{Pos: p.tok().Pos, Name: p.ident().Text}
		if p.accept(":") {
			param.Type = p.typeExpr()
		}
		a.Params = append(a.Params, param)
		if !p.accept(",") {
			break
		}
	}
	p.expect(")")
	if p.accept(":") {
		a.Result = p.typeExpr()
	}
	p.expect("=>")
	if p.is("{") {
		a.Block = p.block()
	} else {
		a.Body = p.expr()
	}
	return a
}

func (p *Parser) block() *Block {
	b := &Block{Pos: p.expect("{").Pos}
	p.skipNewlines()
	for !p.is("}") {
		if p.tok().Kind == TokEOF {
			p.fail(p.tok().Pos, "unexpected end of file inside a block")
		}
		b.Stmts = append(b.Stmts, p.stmt())
	}
	p.expect("}")
	return b
}

func (p *Parser) stmt() Stmt {
	t := p.tok()
	switch {
	case p.is("const") || p.is("let"):
		mutable := p.next().Text == "let"
		s := &VarStmt{Pos: t.Pos, Mutable: mutable}
		next := p.toks[p.i+1] // the token stream always ends with EOF
		ctor := p.tok().Kind == TokIdent && next.Kind == TokPunct && (next.Text == "(" || next.Text == ".")
		if p.is("{") || p.is("#[") || ctor {
			s.Pattern = p.pattern()
		} else {
			s.Name = p.ident().Text
		}
		if p.accept(":") {
			s.Type = p.typeExpr()
		}
		p.expect("=")
		s.Value = p.expr()
		p.endStatement()
		return s
	case p.is("return"):
		p.next()
		s := &ReturnStmt{Pos: t.Pos}
		if p.tok().Kind != TokNewline && !p.is("}") {
			s.Value = p.expr()
		}
		p.endStatement()
		return s
	case p.is("if"):
		return p.ifStmt()
	case p.is("while"):
		p.next()
		p.expect("(")
		cond := p.expr()
		p.expect(")")
		s := &WhileStmt{Pos: t.Pos, Cond: cond, Body: p.block()}
		p.skipNewlines()
		return s
	case p.is("for"):
		p.next()
		p.expect("(")
		if !p.accept("const") {
			p.fail(p.tok().Pos, "write `for (const item of items)`")
		}
		name := p.ident().Text
		if p.is("in") || (p.tok().Kind == TokIdent && p.tok().Text == "in") {
			p.fail(p.tok().Pos, "use `of`; `for ... in` is not part of the language")
		}
		p.expect("of")
		iter := p.expr()
		p.expect(")")
		s := &ForStmt{Pos: t.Pos, Name: name, Iter: iter, Body: p.block()}
		p.skipNewlines()
		return s
	case p.is("break"):
		p.next()
		p.endStatement()
		return &BreakStmt{Pos: t.Pos}
	case p.is("continue"):
		p.next()
		p.endStatement()
		return &ContinueStmt{Pos: t.Pos}
	case p.is("{"):
		p.fail(t.Pos, "a block must belong to a statement such as if or while")
	}
	e := p.expr()
	if p.accept("=") {
		v := p.expr()
		p.endStatement()
		return &AssignStmt{Pos: t.Pos, Target: e, Value: v}
	}
	if m, ok := e.(*MatchExpr); ok {
		p.skipNewlines()
		return &ExprStmt{Pos: t.Pos, Expr: m}
	}
	p.endStatement()
	return &ExprStmt{Pos: t.Pos, Expr: e}
}

func (p *Parser) ifStmt() *IfStmt {
	pos := p.expect("if").Pos
	p.expect("(")
	cond := p.expr()
	p.expect(")")
	s := &IfStmt{Pos: pos, Cond: cond, Then: p.block()}
	save := p.i
	p.skipNewlines()
	if p.accept("else") {
		if p.is("if") {
			s.Else = p.ifStmt()
			return s
		}
		s.Else = p.block()
		p.skipNewlines()
		return s
	}
	p.i = save
	p.skipNewlines()
	return s
}

// Operator precedence, lowest first.
var binaryLevels = [][]string{
	{"||"},
	{"&&"},
	{"==", "!="},
	{"<", "<=", ">", ">="},
	{"|"},
	{"^"},
	{"&"},
	{"<<", ">>"},
	{"+", "-", "+%", "-%"},
	{"*", "/", "%", "*%"},
}

// `x is Pattern` binds tighter than == and looser than < and +.
const isLevel = 2

func (p *Parser) expr() Expr { return p.binary(0) }

func (p *Parser) binary(level int) Expr {
	if level == len(binaryLevels) {
		return p.unary()
	}
	left := p.binary(level + 1)
	for {
		t := p.tok()
		if level == isLevel && p.is("is") {
			p.next()
			left = &IsExpr{exprBase: exprBase{P: t.Pos}, Subject: left, Pattern: p.pattern()}
			continue
		}
		if t.Kind != TokPunct {
			return left
		}
		matched := false
		for _, op := range binaryLevels[level] {
			if t.Text == op {
				matched = true
			}
		}
		if t.Text == "===" || t.Text == "!==" {
			fix := Fix{Title: "use " + t.Text[:2], Edits: []Edit{{Start: t.Pos, End: Pos{t.Pos.Line, t.Pos.Col + 3}, Text: t.Text[:2]}}}
			p.failFix(t.Pos, []Fix{fix}, "use `%s`; there is no strict equality operator", t.Text[:2])
		}
		if !matched {
			return left
		}
		p.next()
		if p.is("=") && p.tok().Pos.Col == t.Pos.Col+len(t.Text) {
			p.fail(t.Pos, "compound assignment is not supported; write `x = x %s y`", t.Text)
		}
		right := p.binary(level + 1)
		left = &BinaryExpr{exprBase: exprBase{P: t.Pos}, Op: t.Text, Left: left, Right: right}
	}
}

func (p *Parser) unary() Expr {
	t := p.tok()
	if p.is("-") || p.is("!") || p.is("~") {
		p.next()
		return &UnaryExpr{exprBase: exprBase{P: t.Pos}, Op: t.Text, Operand: p.unary()}
	}
	return p.postfix(p.primary())
}

func (p *Parser) postfix(e Expr) Expr {
	for {
		t := p.tok()
		switch {
		case p.is("."):
			p.next()
			name := p.memberName()
			e = &FieldExpr{exprBase: exprBase{P: name.Pos}, Target: e, Name: name.Text}
		case p.is("("):
			p.next()
			call := &CallExpr{exprBase: exprBase{P: t.Pos}, Callee: e}
			for !p.is(")") {
				arg := &Arg{Mutable: p.accept("mutable")}
				arg.Value = p.expr()
				call.Args = append(call.Args, arg)
				if !p.accept(",") {
					break
				}
			}
			p.expect(")")
			e = call
		case p.is("["):
			p.next()
			idx := p.expr()
			p.expect("]")
			e = &IndexExpr{exprBase: exprBase{P: t.Pos}, Target: e, Index: idx}
		case p.is("<") || p.is("{"):
			// `Name { field: value }` is a struct literal, and
			// `Name<A, B> { ... }` is one with type arguments.
			id, ok := e.(*Ident)
			if !ok || !startsUpper(id.Name) {
				return e
			}
			var targs []*TypeExpr
			if p.is("<") {
				if !p.typeArgsLitAhead() {
					return e
				}
				targs = p.litTypeArgs()
			}
			if id.Name == "Map" {
				m := p.mapLit(id.P)
				m.TypeArgs = targs
				e = m
				continue
			}
			lit := p.structLit()
			lit.TypeName = id.Name
			lit.TypeArgs = targs
			lit.P = id.P
			e = lit
		default:
			return e
		}
	}
}

// typeArgsLitAhead reports whether the `<` at the current token starts the
// type arguments of a literal: a balanced `<` ... `>` holding only names,
// `.`, `,`, `#[`, `]` and nested `<>`, followed by `{`. Anything else is a
// comparison, so `a < B` never changes meaning. `>>` closes two levels.
func (p *Parser) typeArgsLitAhead() bool {
	depth := 0
	for i := p.i; i < len(p.toks); i++ {
		t := p.toks[i]
		switch {
		case t.Kind == TokIdent:
			if depth == 0 {
				return false
			}
		case t.Kind != TokPunct:
			return false
		case t.Text == "<":
			depth++
		case t.Text == ">" || t.Text == ">>":
			depth -= len(t.Text)
			if depth < 0 {
				return false
			}
			if depth == 0 {
				n := p.toks[min(i+1, len(p.toks)-1)]
				return n.Kind == TokPunct && n.Text == "{"
			}
		case t.Text == "," || t.Text == "." || t.Text == "#[" || t.Text == "]":
		default:
			return false
		}
	}
	return false
}

// litTypeArgs parses the `<A, B>` before a literal's `{`. A trailing `>>`
// is split by typeExpr, as in a type.
func (p *Parser) litTypeArgs() []*TypeExpr {
	p.expect("<")
	var args []*TypeExpr
	for {
		args = append(args, p.typeExpr())
		if !p.accept(",") {
			break
		}
	}
	p.expect(">")
	return args
}

func startsUpper(s string) bool { return s != "" && s[0] >= 'A' && s[0] <= 'Z' }

// mapLit parses `Map { "a": 1, [key]: 2 }`. A key that is not a literal is
// written in brackets.
func (p *Parser) mapLit(pos Pos) *MapLit {
	lit := &MapLit{exprBase: exprBase{P: pos}}
	p.expect("{")
	p.skipNewlines()
	for !p.is("}") {
		var key Expr
		if p.accept("[") {
			key = p.expr()
			p.expect("]")
		} else {
			key = p.expr()
			if !isLiteralKey(key) {
				p.fail(key.position(), "a map key that is not a literal is written [key]: value")
			}
		}
		p.expect(":")
		lit.Keys = append(lit.Keys, key)
		lit.Vals = append(lit.Vals, p.expr())
		p.skipNewlines()
		if !p.accept(",") {
			break
		}
		p.skipNewlines()
	}
	p.skipNewlines()
	p.expect("}")
	return lit
}

func isLiteralKey(e Expr) bool {
	switch e := e.(type) {
	case *IntLit, *FloatLit, *StringLit, *CharLit, *BoolLit:
		return true
	case *UnaryExpr:
		return e.Op == "-" && isLiteralKey(e.Operand)
	}
	return false
}

func (p *Parser) structLit() *StructLit {
	lit := &StructLit{exprBase: exprBase{P: p.expect("{").Pos}}
	p.skipNewlines()
	for !p.is("}") {
		if p.is("...") {
			if lit.Spread != nil || len(lit.Fields) > 0 {
				p.fail(p.tok().Pos, "a struct literal has at most one spread, and it comes first")
			}
			p.next()
			lit.Spread = p.expr()
			p.skipNewlines()
			if !p.accept(",") {
				break
			}
			p.skipNewlines()
			continue
		}
		name := p.ident()
		if !p.is(":") {
			p.fail(name.Pos, "write `%s: %s`; field shorthand is not allowed", name.Text, name.Text)
		}
		p.expect(":")
		lit.Fields = append(lit.Fields, &FieldInit{Pos: name.Pos, Name: name.Text, Value: p.expr()})
		p.skipNewlines()
		if !p.accept(",") {
			break
		}
		p.skipNewlines()
	}
	p.skipNewlines()
	p.expect("}")
	return lit
}

func (p *Parser) primary() Expr {
	t := p.tok()
	base := exprBase{P: t.Pos}
	switch t.Kind {
	case TokInt:
		p.next()
		return &IntLit{exprBase: base, Text: t.Text}
	case TokFloat:
		p.next()
		return &FloatLit{exprBase: base, Text: t.Text}
	case TokString:
		p.next()
		return &StringLit{exprBase: base, Value: t.Text, End: t.End}
	case TokChar:
		p.next()
		return &CharLit{exprBase: base, Value: []rune(t.Text)[0]}
	case TokTemplate:
		p.next()
		return p.template(t)
	case TokIdent:
		p.next()
		if t.Text == "null" || t.Text == "undefined" {
			fix := Fix{Title: "use None", Edits: []Edit{{Start: t.Pos, End: Pos{t.Pos.Line, t.Pos.Col + len(t.Text)}, Text: "None"}}}
			p.failFix(t.Pos, []Fix{fix}, "there is no %s; use `None` with an `Option` type", t.Text)
		}
		return &Ident{exprBase: base, Name: t.Text}
	case TokKeyword:
		switch t.Text {
		case "true", "false":
			p.next()
			return &BoolLit{exprBase: base, Value: t.Text == "true"}
		case "match":
			return p.matchExpr()
		}
	case TokPunct:
		switch t.Text {
		case "(":
			if p.arrowAhead() {
				return p.arrow()
			}
			p.next()
			e := p.expr()
			p.expect(")")
			return e
		case "[", "#[":
			p.next()
			lit := &ArrayLit{exprBase: base, IsTuple: t.Text == "#["}
			for !p.is("]") {
				lit.Elems = append(lit.Elems, p.expr())
				if !p.accept(",") {
					break
				}
			}
			p.expect("]")
			return lit
		case "{":
			return p.structLit()
		}
	}
	p.fail(t.Pos, "expected an expression, found %s", describe(t))
	return nil
}

func (p *Parser) template(t Token) Expr {
	lit := &TemplateLit{exprBase: exprBase{P: t.Pos}}
	raw := t.Text
	var text strings.Builder
	flush := func() {
		if text.Len() > 0 {
			lit.Parts = append(lit.Parts, &StringLit{exprBase: exprBase{P: t.Pos}, Value: text.String(), End: t.End})
			text.Reset()
		}
	}
	for i := 0; i < len(raw); {
		c := raw[i]
		if c == '\\' && i+1 < len(raw) {
			switch raw[i+1] {
			case 'n':
				text.WriteByte('\n')
			case 't':
				text.WriteByte('\t')
			default:
				text.WriteByte(raw[i+1])
			}
			i += 2
			continue
		}
		if c == '$' && i+1 < len(raw) && raw[i+1] == '{' {
			depth := 1
			j := i + 2
			for j < len(raw) && depth > 0 {
				switch raw[j] {
				case '{':
					depth++
				case '}':
					depth--
				}
				j++
			}
			if depth != 0 {
				p.fail(t.Pos, "unterminated ${ in template string")
			}
			flush()
			inner, err := Lex(raw[i+2 : j-1])
			if err != nil {
				p.fail(t.Pos, "in template string: %v", err)
			}
			// Report errors inside ${...} at the template's position.
			for k := range inner {
				inner[k].Pos = t.Pos
			}
			sub := &Parser{toks: inner}
			e := sub.expr()
			lit.Parts = append(lit.Parts, e)
			i = j
			continue
		}
		text.WriteByte(c)
		i++
	}
	flush()
	return lit
}

func (p *Parser) matchExpr() Expr {
	pos := p.expect("match").Pos
	p.expect("(")
	subject := p.expr()
	p.expect(")")
	p.expect("{")
	m := &MatchExpr{exprBase: exprBase{P: pos}, Subject: subject}
	p.skipNewlines()
	for !p.is("}") {
		arm := &MatchArm{Pos: p.tok().Pos, Pattern: p.pattern()}
		if p.accept("if") {
			p.expect("(")
			arm.Guard = p.expr()
			p.expect(")")
		}
		p.expect("=>")
		if p.is("{") {
			arm.Block = p.block()
		} else {
			arm.Body = p.expr()
		}
		m.Arms = append(m.Arms, arm)
		p.accept(",")
		p.skipNewlines()
	}
	p.expect("}")
	return m
}

func (p *Parser) pattern() *Pattern {
	t := p.tok()
	pat := &Pattern{Pos: t.Pos}
	switch {
	case p.is("{"):
		p.next()
		pat.Kind = PatStruct
		p.skipNewlines()
		for !p.is("}") {
			name := p.ident()
			fp := &FieldPattern{Pos: name.Pos, Name: name.Text}
			if p.accept(":") {
				fp.Pattern = p.pattern()
			} else {
				fp.Pattern = &Pattern{Pos: name.Pos, Kind: PatName, Name: name.Text}
			}
			pat.Fields = append(pat.Fields, fp)
			p.skipNewlines()
			if !p.accept(",") {
				break
			}
			p.skipNewlines()
		}
		p.skipNewlines()
		p.expect("}")
		return pat
	case p.is("#["):
		p.next()
		pat.Kind = PatTuple
		pat.Args = p.patternList("]")
		return pat
	case p.is("-"):
		p.next()
		n := p.tok()
		if n.Kind != TokInt && n.Kind != TokFloat {
			p.fail(n.Pos, "expected a number after - in a pattern, found %s", describe(n))
		}
		pat.Kind = PatLit
		pat.Lit = &UnaryExpr{exprBase: exprBase{P: t.Pos}, Op: "-", Operand: p.primary()}
		return pat
	case t.Kind == TokInt || t.Kind == TokFloat || t.Kind == TokString || t.Kind == TokChar || p.is("true") || p.is("false"):
		pat.Kind = PatLit
		pat.Lit = p.primary()
		return pat
	}
	t = p.ident()
	if t.Text == "_" {
		pat.Kind = PatWildcard
		return pat
	}
	pat.Kind = PatName
	pat.Name = t.Text
	if p.accept(".") {
		pat.Kind = PatCtor
		pat.Enum = t.Text
		pat.Name = p.ident().Text
	}
	if p.accept("(") {
		pat.Kind = PatCtor
		pat.Args = p.patternList(")")
	}
	return pat
}

// patternList parses patterns up to close, which it consumes. The result is
// never nil, so Name() keeps zero arguments apart from a bare Name.
func (p *Parser) patternList(close string) []*Pattern {
	list := []*Pattern{}
	p.skipNewlines()
	for !p.is(close) {
		list = append(list, p.pattern())
		p.skipNewlines()
		if !p.accept(",") {
			break
		}
		p.skipNewlines()
	}
	p.expect(close)
	return list
}

// paramAfterComma reports whether the `,` at the current token is followed
// by `name:`, `mutable name` or `escaping name`, the next parameter of a
// function. It ends a `uses` list written inside a parameter list.
func (p *Parser) paramAfterComma() bool {
	if p.i+2 >= len(p.toks) {
		return false
	}
	if k := p.toks[p.i+1]; k.Kind != TokIdent && !(k.Kind == TokKeyword && k.Text == "mutable") {
		return false
	}
	n := p.toks[p.i+2]
	return n.Kind == TokIdent || (n.Kind == TokPunct && n.Text == ":")
}
