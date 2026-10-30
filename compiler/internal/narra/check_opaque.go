package narra

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Opaque structs and distinct types (SPEC.md section 4, D39, D40).
//
// A distinct type is a DistinctType with its own identity. Its C type is the
// base type's, so it costs nothing, but composite names (arrays, maps) are
// made from its own typeID. `T.from(x)` and `T.toInt(x)` are checked in
// distinctCall and emitted as the argument itself.

// DistinctType is `distinct type Name = Base`.
type DistinctType struct {
	Name   string
	Module string // declaring module path
	Base   Type
	Opaque bool
}

func (t *DistinctType) String() string { return t.Name }
func (t *DistinctType) CName() string  { return t.Base.CName() }

// underlying strips distinct wrappers.
func underlying(t Type) Type {
	for {
		d, ok := t.(*DistinctType)
		if !ok {
			return t
		}
		t = d.Base
	}
}

func (p *Parser) isWord(w string) bool {
	t := p.tok()
	return t.Kind == TokIdent && t.Text == w
}

// opaqueOrDistinct parses `opaque struct`, `distinct type` and
// `opaque distinct type` after an optional `export`. It returns nil when the
// declaration is none of them.
func (p *Parser) opaqueOrDistinct(exported bool) Decl {
	opaque := false
	if p.isWord("opaque") {
		if !exported {
			p.fail(p.tok().Pos, "only an exported type can be opaque; write `export opaque`")
		}
		p.next()
		opaque = true
	}
	switch {
	case p.is("struct") && opaque:
		sd := p.structDecl(exported)
		sd.Opaque = true
		return sd
	case p.isWord("distinct"):
		pos := p.next().Pos
		if !p.isWord("type") {
			p.fail(p.tok().Pos, "expected `type` after `distinct`, found %s", describe(p.tok()))
		}
		p.next()
		d := &DistinctDecl{Pos: pos, Exported: exported, Opaque: opaque, Name: p.ident().Text}
		p.expect("=")
		d.Base = p.typeExpr()
		p.endTypeDecl()
		return d
	case opaque:
		p.fail(p.tok().Pos, "expected `struct` or `distinct type` after `opaque`, found %s", describe(p.tok()))
	}
	return nil
}

// declareDistinct creates the type of a distinct declaration.
func (c *Checker) declareDistinct(d *DistinctDecl) {
	d.T = &DistinctType{Name: d.Name, Module: c.mod.path, Opaque: d.Opaque}
	c.declareType(d.Pos, d.Name, d.T)
}

// resolveDistinct resolves the base type of a distinct declaration.
func (c *Checker) resolveDistinct(d *DistinctDecl) {
	base := c.resolveType(d.Base)
	if b, ok := base.(*DistinctType); ok && b.Base == nil {
		c.errorf(d.Pos, "distinct type %s cannot be defined in terms of itself", d.Name)
		base = TInt
	}
	d.T.Base = base
}

// hiddenStruct reports whether st is an opaque struct of another module.
func (c *Checker) hiddenStruct(st *StructType) bool {
	return st.Opaque && st.Module != c.mod.path
}

// distinctOf finds the distinct type named by the callee target of
// `T.f(...)` or `Alias.T.f(...)`.
func (c *Checker) distinctOf(target Expr) *DistinctType {
	switch t := target.(type) {
	case *Ident:
		if c.lookup(t.Name) != nil {
			return nil
		}
		d, _ := c.types[t.Name].(*DistinctType)
		return d
	case *FieldExpr:
		id, ok := t.Target.(*Ident)
		if !ok || c.lookup(id.Name) != nil {
			return nil
		}
		path, ok := c.modules[id.Name]
		if !ok {
			return nil
		}
		s := c.mods[path]
		if s == nil || s.src == nil {
			return nil
		}
		if _, ok := s.src.Exports[t.Name].(*DistinctDecl); !ok {
			return nil
		}
		d, _ := s.types[t.Name].(*DistinctType)
		return d
	}
	return nil
}

// unwrapName is the name of the generated unwrap function: toInt for int,
// toString for string.
func unwrapName(d *DistinctType) string {
	s := underlying(d).String()
	if i := strings.IndexByte(s, '<'); i >= 0 {
		s = s[:i]
	}
	r, n := utf8.DecodeRuneInString(s)
	return "to" + string(unicode.ToUpper(r)) + s[n:]
}

// distinctCall checks `T.from(x)` and `T.toInt(x)`.
func (c *Checker) distinctCall(e *CallExpr, d *DistinctType, callee *FieldExpr) Type {
	e.Distinct = d
	from := false
	switch callee.Name {
	case "from":
		from = true
		if d.Opaque && d.Module != c.mod.path {
			c.errorf(callee.P, "%s.from is not available outside module %q: %s is an opaque distinct type", d.Name, d.Module, d.Name)
		}
	case unwrapName(d):
	default:
		c.errorf(callee.P, "%s has no function %q; a distinct type has from and %s", d.Name, callee.Name, unwrapName(d))
		c.argsLoose(e)
		return TVoid
	}
	if len(e.Args) != 1 {
		c.errorf(e.P, "%s takes %d arguments, found %d", d.Name+"."+callee.Name, 1, len(e.Args))
		c.argsLoose(e)
		return TVoid
	}
	arg := e.Args[0].Value
	if from {
		c.expectType(arg.position(), d.Base, c.expr(arg, d.Base))
		return d
	}
	c.expectType(arg.position(), d, c.expr(arg, d))
	return d.Base
}

// memberName reads the name after `.`. `from` is a keyword, but
// `UserId.from(x)` needs it as a member name.
func (p *Parser) memberName() Token {
	if p.is("from") {
		return p.next()
	}
	return p.ident()
}
