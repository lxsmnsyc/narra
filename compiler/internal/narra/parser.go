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
	// resultFunc is true inside a function declared to return Result, where
	// the throw fix is right. Closures clear it.
	resultFunc bool
	// ternaries holds the block expressions that `c ? a : b` produced.
	ternaries map[*BlockExpr]bool
	// typeNames are the type parameters of the declaration being parsed. A
	// lowercase one still counts as a type in `id<t>(x)`.
	typeNames []string
}

// templateDot finds `Name.` inside a template string. The parser does not keep
// positions inside ${...}, so a fix cannot rename a name used there.
var templateDot = regexp.MustCompile(`([A-Za-z_][A-Za-z_0-9]*)\.`)

func Parse(src string) (*File, error) {
	toks, err := Lex(src)
	if err != nil {
		return nil, syntaxDiagnostic(err)
	}
	p := &Parser{toks: toks, ternaries: map[*BlockExpr]bool{}}
	f, err := p.file()
	if err != nil {
		return nil, syntaxDiagnostic(err)
	}
	f.Src = src
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
	f.Header = headerLine(src)
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
	if p.tok().Text == "this" {
		p.fail(p.tok().Pos, "there is no `this`; pass the value as a parameter")
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
		p.typeNames = nil
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
			np := p.tok().Pos
			n := p.ident().Text
			local := n
			end := Pos{np.Line, np.Col + len(n)}
			if p.accept("as") {
				lp := p.tok().Pos
				local = p.ident().Text
				end = Pos{lp.Line, lp.Col + len(local)}
			}
			d.Names = append(d.Names, ImportName{Name: n, Local: local, Pos: np, End: end})
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
	p.habitDecl()
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
	if te.Name == "Generator" || te.Name == "Coroutine" {
		// A generator type carries the effects its body has when it runs:
		// `Generator<string> uses io` (D78).
		te.Uses = p.usesList()
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
		p.typeNames = append(p.typeNames, names[len(names)-1])
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
	saved := p.resultFunc
	p.resultFunc = fd.Result != nil && fd.Result.Name == "Result"
	fd.Body = p.block()
	p.resultFunc = saved
	return fd
}

// funcHeader parses a function declaration up to its body.
func (p *Parser) funcHeader(exported bool) *FuncDecl {
	pos := p.expect("function").Pos
	generator := p.accept("*")
	fd := &FuncDecl{Pos: pos, Exported: exported, Generator: generator, Name: p.ident().Text}
	fd.TypeParams, fd.Bounds, fd.EffectParams = p.typeParamsEffects()
	p.expect("(")
	for !p.is(")") {
		param := &Param{Pos: p.tok().Pos}
		param.Mutable = p.accept("mutable")
		param.Escaping = !param.Mutable && p.acceptEscaping()
		param.NamePos = p.tok().Pos
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
		p.nestedGenerator()
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
		p.nestedGenerator()
		d.Funcs = append(d.Funcs, p.funcDecl(false))
		p.skipNewlines()
	}
	p.expect("}")
	return d
}

// nestedGenerator rejects `function*` in a trait or an implement block: a
// generator is a top-level function (D74).
func (p *Parser) nestedGenerator() {
	if next := p.toks[p.i+1]; next.Kind == TokPunct && next.Text == "*" {
		p.fail(next.Pos, "a generator must be a top-level function; declare `function*` at module level")
	}
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
				if next == ":" {
					return p.resultTypeThenArrow(j + 2)
				}
				return next == "=>"
			}
		}
	}
	return false
}

// resultTypeThenArrow tells an arrow's `: R =>` from the `:` of a conditional
// expression such as `c ? (a) : b`. From the token after the `:`, it looks for
// a `=>` outside brackets before anything that ends an expression.
func (p *Parser) resultTypeThenArrow(start int) bool {
	depth := 0
	for k := start; k < len(p.toks); k++ {
		t := p.toks[k]
		if t.Kind == TokNewline || t.Kind == TokEOF {
			return false
		}
		if t.Kind != TokPunct {
			continue
		}
		switch t.Text {
		case "(", "[", "<":
			depth++
		case ")", "]", ">":
			depth--
		case ">>":
			depth -= 2
		case "=>":
			if depth == 0 {
				return true
			}
		case "?", ":", ",", "{", "}", ";":
			if depth == 0 {
				return false
			}
		}
		if depth < 0 {
			return false
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
		param := &Param{Pos: p.tok().Pos, NamePos: p.tok().Pos, Name: p.ident().Text}
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
	saved := p.resultFunc
	p.resultFunc = false
	if p.is("{") {
		a.Block = p.block()
	} else {
		a.Body = p.expr()
	}
	p.resultFunc = saved
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

// loopExitLabel reads the optional label after break or continue.
func (p *Parser) loopExitLabel() string {
	if p.tok().Kind == TokIdent {
		return p.next().Text
	}
	return ""
}

// labeledStmt parses `name: for ...` and `name: while ...`.
func (p *Parser) labeledStmt() Stmt {
	t := p.next()
	p.next() // the colon
	if !p.is("for") && !p.is("while") {
		p.fail(t.Pos, "label %q must be on a for or while loop", t.Text)
	}
	switch s := p.stmt().(type) {
	case *WhileStmt:
		s.Label, s.Pos = t.Text, t.Pos
		return s
	case *ForStmt:
		s.Label, s.Pos = t.Text, t.Pos
		return s
	}
	panic("labeledStmt: a loop keyword did not give a loop")
}

func (p *Parser) stmt() Stmt {
	t := p.tok()
	if t.Kind == TokIdent && p.toks[p.i+1].Kind == TokPunct && p.toks[p.i+1].Text == ":" {
		return p.labeledStmt()
	}
	if t.Kind == TokIdent && t.Text == "async" && p.toks[p.i+1].Kind == TokKeyword && p.toks[p.i+1].Text == "function" {
		p.fail(p.toks[p.i+1].Pos, "a function inside a function is not supported; declare it at module level as a private function and pass the enclosing values as arguments")
	}
	switch {
	case p.is("function"):
		p.fail(t.Pos, "a function inside a function is not supported; declare it at module level as a private function and pass the enclosing values as arguments")
	case p.is("const") || p.is("let"):
		mutable := p.next().Text == "let"
		s := &VarStmt{Pos: t.Pos, NamePos: p.tok().Pos, Mutable: mutable}
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
		if p.is("yield") && s.Name != "" && s.Type == nil && !(p.toks[p.i+1].Kind == TokPunct && p.toks[p.i+1].Text == "*") {
			// `const input = yield value` in a coroutine.
			y := &YieldStmt{Pos: p.next().Pos, Bind: true, Mutable: s.Mutable, Name: s.Name, NamePos: s.Pos}
			if p.tok().Kind == TokNewline || p.tok().Kind == TokEOF || p.is("}") {
				p.fail(p.tok().Pos, "`yield` needs a value: write `yield value`, or `yield* generator`")
			}
			y.Value = p.expr()
			p.endStatement()
			return y
		}
		s.Value = p.expr()
		p.endStatement()
		return s
	case p.is("return"):
		p.next()
		s := &ReturnStmt{Pos: t.Pos}
		if p.tok().Kind != TokNewline && !p.is("}") {
			s.Value = p.expr()
			if p.isTernary(s.Value) {
				p.endStatement()
				return p.returnedTernary(s.Value)
			}
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
		if !p.is("const") && !p.is("let") {
			p.fail(p.tok().Pos, "write `for (const item of items)`")
		}
		mutable := p.next().Text == "let"
		headPos := p.tok().Pos
		next := p.toks[p.i+1] // the token stream always ends with EOF
		ctor := p.tok().Kind == TokIdent && next.Kind == TokPunct && (next.Text == "(" || next.Text == ".")
		var pattern *Pattern
		name := ""
		if p.is("{") || p.is("#[") || ctor {
			pattern = p.pattern()
		} else {
			name = p.ident().Text
		}
		if p.is("in") || (p.tok().Kind == TokIdent && p.tok().Text == "in") {
			p.fail(p.tok().Pos, "use `of`; `for ... in` is not part of the language")
		}
		p.expect("of")
		iter := p.expr()
		p.expect(")")
		body := p.block()
		if pattern != nil || mutable {
			// for (const <pattern> of xs) body becomes
			// for (const __for_L_C of xs) { const <pattern> = __for_L_C; body }
			// where L and C are the line and column of the pattern. A user
			// name cannot clash with it because each loop has its own position.
			hidden := fmt.Sprintf("__for_%d_%d", headPos.Line, headPos.Col)
			decl := &VarStmt{Pos: headPos, Mutable: mutable, Name: name, Pattern: pattern, NamePos: headPos,
				Value: &Ident{exprBase: exprBase{P: headPos}, Name: hidden}}
			body.Stmts = append([]Stmt{decl}, body.Stmts...)
			name = hidden
		}
		s := &ForStmt{Pos: t.Pos, NamePos: headPos, Name: name, Iter: iter, Body: body}
		p.skipNewlines()
		return s
	case p.is("yield"):
		p.next()
		s := &YieldStmt{Pos: t.Pos, Delegate: p.accept("*")}
		if p.tok().Kind == TokNewline || p.tok().Kind == TokEOF || p.is("}") {
			p.fail(p.tok().Pos, "`yield` needs a value: write `yield value`, or `yield* generator`")
		}
		s.Value = p.expr()
		p.endStatement()
		return s
	case p.is("break"):
		p.next()
		label := p.loopExitLabel()
		p.endStatement()
		return &BreakStmt{Pos: t.Pos, Label: label}
	case p.is("continue"):
		p.next()
		label := p.loopExitLabel()
		p.endStatement()
		return &ContinueStmt{Pos: t.Pos, Label: label}
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

// structLitAhead reports whether the `{` at the current token starts a struct
// literal whose type comes from context: `{}`, `{ ...spread }` or
// `{ name: value }`. Any other `{` in value position is a block expression.
// A label (`name: for`, `name: while`) starts a block.
func (p *Parser) structLitAhead() bool {
	j := p.i + 1
	for j < len(p.toks) && p.toks[j].Kind == TokNewline {
		j++
	}
	if j+1 >= len(p.toks) {
		return false
	}
	first := p.toks[j]
	if first.Kind == TokPunct && (first.Text == "}" || first.Text == "...") {
		return true
	}
	second := p.toks[j+1]
	if first.Kind == TokString && second.Kind == TokPunct && second.Text == ":" {
		return true // a TypeScript object literal: structLit reports it
	}
	if first.Kind != TokIdent || second.Kind != TokPunct || second.Text != ":" || j+2 >= len(p.toks) {
		return false
	}
	third := p.toks[j+2]
	return !(third.Kind == TokKeyword && (third.Text == "for" || third.Text == "while"))
}

// blockExpr parses a block in value position.
func (p *Parser) blockExpr() *BlockExpr {
	pos := p.tok().Pos
	return &BlockExpr{exprBase: exprBase{P: pos}, Block: p.block()}
}

// ifExpr parses `if (c) { ... } else { ... }` in value position. It becomes a
// block expression holding one if statement whose branches each return a
// block expression, so the checker and emitter need no node of their own.
// The else branch is required; `else if` chains are allowed.
func (p *Parser) ifExpr() Expr {
	pos := p.expect("if").Pos
	p.expect("(")
	cond := p.expr()
	p.expect(")")
	thenBranch := p.blockExpr()
	save := p.i
	p.skipNewlines()
	if !p.accept("else") {
		p.i = save
		p.expect("else")
	}
	var elseBranch Expr
	if p.is("if") {
		elseBranch = p.ifExpr()
	} else {
		elseBranch = p.blockExpr()
	}
	ifStmt := &IfStmt{Pos: pos, Cond: cond, Then: returnBlock(thenBranch), Else: returnBlock(elseBranch)}
	return &BlockExpr{exprBase: exprBase{P: pos}, Block: &Block{Pos: pos, Stmts: []Stmt{ifStmt}}}
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

// expr parses an expression: a binary expression, or `cond ? a : b`, which
// is looser than `||` and right-associative. It becomes a block expression
// holding one if statement whose branches each return one value, like an
// `if` in value position, so the checker and emitter need no node of their
// own. Only the chosen branch runs.
func (p *Parser) expr() Expr {
	start := p.tok().Pos
	cond := p.binary(0)
	if !p.is("?") {
		return cond
	}
	pos := p.next().Pos
	thenValue := p.expr()
	if !p.is(":") {
		// A `:` on the next line is a ternary that spans lines.
		saved := p.i
		p.skipNewlines()
		spans := p.is(":")
		p.i = saved
		if spans {
			p.fail(p.tok().Pos, "a conditional expression that spans lines must be wrapped in parentheses, as in `(cond ? a : b)`")
		}
		p.fail(p.tok().Pos, "a conditional expression needs `:` and a second value, as in `cond ? a : b`")
	}
	p.next()
	elseValue := p.expr()
	ifStmt := &IfStmt{Pos: pos, Cond: cond, Then: returnBlock(thenValue), Else: returnBlock(elseValue)}
	// The expression starts at the condition and its block at the `?`, which
	// is how isConditional tells it from a plain block and an `if` value.
	ternary := &BlockExpr{exprBase: exprBase{P: start}, Block: &Block{Pos: pos, Stmts: []Stmt{ifStmt}}}
	p.ternaries[ternary] = true
	return ternary
}

// returnBlock is a block holding `return value`, positioned at the value.
func returnBlock(value Expr) *Block {
	return &Block{Pos: value.position(), Stmts: []Stmt{&ReturnStmt{Pos: value.position(), Value: value}}}
}

func (p *Parser) isTernary(value Expr) bool {
	block, ok := value.(*BlockExpr)
	return ok && p.ternaries[block]
}

// returnedTernary turns `return c ? a : b` into `if (c) { return a } else {
// return b }`, so a self call in a branch is a tail call (tail-loops). A
// branch that is itself a ternary is turned the same way.
func (p *Parser) returnedTernary(value Expr) Stmt {
	if !p.isTernary(value) {
		return &ReturnStmt{Pos: value.position(), Value: value}
	}
	ifStmt := value.(*BlockExpr).Block.Stmts[0].(*IfStmt)
	branch := func(block *Block) *Block {
		branchValue := block.Stmts[0].(*ReturnStmt).Value
		return &Block{Pos: block.Pos, Stmts: []Stmt{p.returnedTernary(branchValue)}}
	}
	return &IfStmt{Pos: ifStmt.Pos, Cond: ifStmt.Cond, Then: branch(ifStmt.Then), Else: branch(ifStmt.Else.(*Block))}
}

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
			p.callArgs(call)
			e = call
		case p.is("["):
			p.next()
			idx := p.expr()
			p.expect("]")
			e = &IndexExpr{exprBase: exprBase{P: t.Pos}, Target: e, Index: idx}
		case p.is("<") && p.callTypeArgsAhead(e):
			// `f<int>(x)`: type arguments of a call.
			targs := p.litTypeArgs()
			open := p.expect("(")
			call := &CallExpr{exprBase: exprBase{P: open.Pos}, Callee: e, TypeArgs: targs}
			p.callArgs(call)
			e = call
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

// callArgs parses the arguments of a call after its `(`, and the closing `)`.
func (p *Parser) callArgs(call *CallExpr) {
	for !p.is(")") {
		arg := &Arg{Mutable: p.accept("mutable")}
		arg.Value = p.expr()
		call.Args = append(call.Args, arg)
		if !p.accept(",") {
			break
		}
	}
	closing := p.tok()
	p.expect(")")
	call.End = Pos{closing.Pos.Line, closing.Pos.Col + 1}
}

// callTypeArgsAhead reports whether the `<` at the current token starts the
// type arguments of a call on callee: a name or `Module.name`, then a
// balanced `<` ... `>` of type names followed directly by `(`. A name not
// followed by `.` must be a type: a type parameter in scope, a primitive
// type name, or a capitalized name that is not SCREAMING_CASE. So
// `a < b > (c)` and `x < MAX, MIN > (y)` stay comparisons.
func (p *Parser) callTypeArgsAhead(callee Expr) bool {
	switch callee := callee.(type) {
	case *Ident:
	case *FieldExpr:
		if _, ok := callee.Target.(*Ident); !ok {
			return false
		}
	default:
		return false
	}
	depth := 0
	for i := p.i; i < len(p.toks); i++ {
		t := p.toks[i]
		switch {
		case t.Kind == TokIdent:
			if depth == 0 {
				return false
			}
			next := p.toks[min(i+1, len(p.toks)-1)]
			if !p.isTypeNameAhead(t.Text) && !(next.Kind == TokPunct && next.Text == ".") {
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
				return n.Kind == TokPunct && n.Text == "("
			}
		case t.Text == "," || t.Text == "." || t.Text == "#[" || t.Text == "]":
		default:
			return false
		}
	}
	return false
}

// isTypeNameAhead reports whether a name in call type arguments can be a
// type: a type parameter in scope, a primitive type, or a capitalized name
// that is not a constant (two or more characters with no lowercase letter,
// such as MAX or LIMIT_2).
func (p *Parser) isTypeNameAhead(name string) bool {
	for _, n := range p.typeNames {
		if n == name {
			return true
		}
	}
	if isPrimitiveTypeName(name) {
		return true
	}
	return startsUpper(name) && !isConstantName(name)
}

// isConstantName reports whether name looks like a constant: it has two or
// more characters and no lowercase letter.
func isConstantName(name string) bool {
	if len(name) < 2 {
		return false
	}
	for i := 0; i < len(name); i++ {
		if name[i] >= 'a' && name[i] <= 'z' {
			return false
		}
	}
	return true
}

// isPrimitiveTypeName reports whether name is a type written in lowercase.
func isPrimitiveTypeName(name string) bool {
	switch name {
	case "int", "f64", "bool", "char", "string", "void":
		return true
	}
	return false
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
		p.habitExpr(t)
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
		case "if":
			return p.ifExpr()
		case "yield":
			p.fail(t.Pos, "`yield` is a statement, so it cannot be part of an expression")
		}
	case TokPunct:
		switch t.Text {
		case "{":
			if p.structLitAhead() {
				return p.structLit()
			}
			return p.blockExpr()
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
		}
	}
	if p.is("?") {
		p.fail(t.Pos, "a conditional expression that spans lines must be wrapped in parentheses, as in `(cond ? a : b)`")
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
			sub := &Parser{toks: inner, ternaries: map[*BlockExpr]bool{}}
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
				fp.Pattern = &Pattern{Pos: name.Pos, Kind: PatName, Name: name.Text, Shorthand: true}
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
	// Enum.Name, or Alias.Enum.Name through a namespace import.
	for p.accept(".") {
		pat.Kind = PatCtor
		if pat.Enum != "" {
			pat.Enum += "."
		}
		pat.Enum += pat.Name
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

// headerLine returns the line after the leading comment block that a blank
// line follows, or 1. A comment right above a declaration is its doc comment.
func headerLine(src string) int {
	header := 1
	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines) && strings.HasPrefix(lines[i], "//"); i++ {
		if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) == "" {
			header = i + 2
		}
	}
	return header
}
