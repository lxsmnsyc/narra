package langc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type forKind int

const (
	forRange forKind = iota
	forArray
)

// Builtin describes a standard function the compiler implements directly.
type Builtin struct {
	Name    string // C-facing name, such as "Int.toString"
	Params  []Type // nil entries are checked by hand
	Result  Type
	Mutable []bool
}

var stdModules = map[string]map[string]*Builtin{
	"std/int": {
		"toString": {Name: "Int.toString", Params: []Type{TInt}, Result: TStr},
		"abs":      {Name: "Int.abs", Params: []Type{TInt}, Result: TInt},
	},
	"std/float": {
		"toString": {Name: "Float.toString", Params: []Type{TF64}, Result: TStr},
		"sqrt":     {Name: "Float.sqrt", Params: []Type{TF64}, Result: TF64},
		"fromInt":  {Name: "Float.fromInt", Params: []Type{TInt}, Result: TF64},
		"truncate": {Name: "Float.truncate", Params: []Type{TF64}, Result: TInt},
	},
	"std/string": {
		"length":   {Name: "Str.length", Params: []Type{TStr}, Result: TInt},
		"concat":   {Name: "Str.concat", Params: []Type{TStr, TStr}, Result: TStr},
		"fromChar": {Name: "Str.fromChar", Params: []Type{TChar}, Result: TStr},
		"repeat":   {Name: "Str.repeat", Params: []Type{TStr, TInt}, Result: TStr},
	},
}

type Program struct {
	File   *File
	Types  []Type // structs and enums in dependency order, so C defines inner types first
	Funcs  []*FuncDecl
	Tests  []*TestDecl
	Arrays []*ArrayType
	Entry  *FuncDecl
}

type scope struct {
	names  map[string]*Local
	parent *scope
}

type Checker struct {
	prog     *Program
	types    map[string]Type
	funcs    map[string]*FuncDecl
	modules  map[string]string // alias -> path
	imported map[string]*Builtin
	errs     []string
	scope    *scope
	fn       *FuncDecl
	loops    int
	arrays   map[string]*ArrayType
	localN   int
}

func Check(f *File) (*Program, error) {
	c := &Checker{
		prog:     &Program{File: f},
		types:    map[string]Type{"int": TInt, "i64": TInt, "f64": TF64, "bool": TBool, "char": TChar, "string": TStr},
		funcs:    map[string]*FuncDecl{},
		modules:  map[string]string{},
		imported: map[string]*Builtin{},
		arrays:   map[string]*ArrayType{},
	}
	c.run(f)
	if len(c.errs) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(c.errs, "\n"))
	}
	return c.prog, nil
}

func (c *Checker) errorf(pos Pos, format string, args ...any) {
	c.errs = append(c.errs, fmt.Sprintf("%s: %s", pos, fmt.Sprintf(format, args...)))
}

func (c *Checker) run(f *File) {
	for _, imp := range f.Imports {
		mod, ok := stdModules[imp.Path]
		if !ok {
			c.errorf(imp.Pos, "unknown module %q (this compiler supports std/int, std/float and std/string)", imp.Path)
			continue
		}
		if imp.Alias != "" {
			c.modules[imp.Alias] = imp.Path
		}
		for _, n := range imp.Names {
			b, ok := mod[n]
			if !ok {
				c.errorf(imp.Pos, "module %q has no export %q", imp.Path, n)
				continue
			}
			c.imported[n] = b
		}
	}
	// Declare type names first, so declarations can refer to each other.
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *StructDecl:
			d.T = &StructType{Name: d.Name}
			c.declareType(d.Pos, d.Name, d.T)
		case *EnumDecl:
			d.T = &EnumType{Name: d.Name}
			c.declareType(d.Pos, d.Name, d.T)
		}
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *StructDecl:
			for _, fd := range d.Fields {
				if d.T.Field(fd.Name) != nil {
					c.errorf(fd.Pos, "field %q is declared twice", fd.Name)
				}
				d.T.Fields = append(d.T.Fields, &Field{Name: fd.Name, T: c.resolveType(fd.Type)})
			}
		case *EnumDecl:
			for i, vd := range d.Variants {
				v := &Variant{Name: vd.Name, Index: i, Enum: d.T}
				for _, fd := range vd.Fields {
					v.Fields = append(v.Fields, &Field{Name: fd.Name, T: c.resolveType(fd.Type)})
				}
				d.T.Variants = append(d.T.Variants, v)
			}
		case *FuncDecl:
			if _, dup := c.funcs[d.Name]; dup {
				c.errorf(d.Pos, "function %q is declared twice", d.Name)
			}
			c.funcs[d.Name] = d
			for _, p := range d.Params {
				p.T = c.resolveType(p.Type)
			}
			d.ResultT = TVoid
			if d.Result != nil {
				d.ResultT = c.resolveType(d.Result)
			}
			c.prog.Funcs = append(c.prog.Funcs, d)
			if d.Name == "run" && d.Exported {
				c.prog.Entry = d
			}
		case *TestDecl:
			c.prog.Tests = append(c.prog.Tests, d)
		}
	}
	// Defaults are checked after every field type is known.
	for _, d := range f.Decls {
		if sd, ok := d.(*StructDecl); ok {
			for i, fd := range sd.Fields {
				if fd.Default != nil {
					c.scope = &scope{names: map[string]*Local{}}
					t := c.expr(fd.Default, sd.T.Fields[i].T)
					c.expectType(fd.Default.position(), sd.T.Fields[i].T, t)
					if !isConstant(fd.Default) {
						c.errorf(fd.Default.position(), "a field default must be a constant")
					}
					sd.T.Fields[i].Default = fd.Default
				}
			}
		}
	}
	c.orderStructs(f)
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *FuncDecl:
			c.funcBody(d)
		case *TestDecl:
			c.fn = &FuncDecl{Name: "test", ResultT: TVoid}
			c.scope = &scope{names: map[string]*Local{}}
			c.block(d.Body)
		}
	}
	if c.prog.Entry != nil {
		e := c.prog.Entry
		ok := len(e.Params) == 1 && sameType(e.Params[0].T, &ArrayType{Elem: TStr}) && !e.Params[0].Mutable && e.ResultT == TInt
		if !ok {
			c.errorf(e.Pos, "the entry function must be `export function run(args: Array<string>): int`")
		}
	}
	keys := make([]string, 0, len(c.arrays))
	for k := range c.arrays {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		c.prog.Arrays = append(c.prog.Arrays, c.arrays[k])
	}
}

func isConstant(e Expr) bool {
	switch e := e.(type) {
	case *IntLit, *FloatLit, *StringLit, *CharLit, *BoolLit:
		return true
	case *UnaryExpr:
		return isConstant(e.Operand)
	case *ArrayLit:
		for _, el := range e.Elems {
			if !isConstant(el) {
				return false
			}
		}
		return true
	}
	return false
}

func (c *Checker) declareType(pos Pos, name string, t Type) {
	if _, dup := c.types[name]; dup {
		c.errorf(pos, "type %q is declared twice", name)
	}
	c.types[name] = t
}

// orderStructs sorts structs and enums so that each one is defined after the
// types it holds inline, and rejects types that contain themselves.
func (c *Checker) orderStructs(f *File) {
	state := map[Type]int{}
	var visit func(t Type, pos Pos)
	visit = func(t Type, pos Pos) {
		switch t := t.(type) {
		case *StructType, *EnumType:
			if state[t] == 2 {
				return
			}
			if state[t] == 1 {
				c.errorf(pos, "type %s contains itself; use an array or Box for recursive data", t)
				return
			}
			state[t] = 1
			var fields []*Field
			if st, ok := t.(*StructType); ok {
				fields = st.Fields
			} else {
				for _, v := range t.(*EnumType).Variants {
					fields = append(fields, v.Fields...)
				}
			}
			for _, fl := range fields {
				visit(fl.T, pos)
			}
			state[t] = 2
			c.prog.Types = append(c.prog.Types, t)
		}
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *StructDecl:
			visit(d.T, d.Pos)
		case *EnumDecl:
			visit(d.T, d.Pos)
		}
	}
}

func (c *Checker) resolveType(te *TypeExpr) Type {
	if te == nil {
		return TVoid
	}
	if te.Name == "Array" {
		if len(te.Args) != 1 {
			c.errorf(te.Pos, "Array takes one type argument")
			return TVoid
		}
		return c.arrayOf(c.resolveType(te.Args[0]))
	}
	switch te.Name {
	case "number":
		c.errorf(te.Pos, "there is no `number`; use `int` or `f64`")
		return TInt
	case "boolean":
		c.errorf(te.Pos, "use `bool`")
		return TBool
	}
	if len(te.Args) > 0 {
		c.errorf(te.Pos, "generic type %s is not supported by this compiler yet", te.Name)
		return TVoid
	}
	t, ok := c.types[te.Name]
	if !ok {
		c.errorf(te.Pos, "unknown type %q", te.Name)
		return TVoid
	}
	return t
}

func (c *Checker) arrayOf(elem Type) *ArrayType {
	key := mangle(elem)
	if a, ok := c.arrays[key]; ok {
		return a
	}
	a := &ArrayType{Elem: elem}
	c.arrays[key] = a
	return a
}

func (c *Checker) expectType(pos Pos, want, got Type) {
	if want == nil || got == nil {
		return
	}
	if !sameType(want, got) {
		c.errorf(pos, "expected %s, found %s", want, got)
	}
}

func (c *Checker) lookup(name string) *Local {
	for s := c.scope; s != nil; s = s.parent {
		if l, ok := s.names[name]; ok {
			return l
		}
	}
	return nil
}

func (c *Checker) declare(pos Pos, name string, t Type, mutable bool) *Local {
	if name != "_" && c.lookup(name) != nil {
		c.errorf(pos, "%q is already declared in this function; shadowing is not allowed", name)
	}
	c.localN++
	l := &Local{Name: name, T: t, Mutable: mutable, Pos: pos, CName: fmt.Sprintf("v_%s_%d", name, c.localN)}
	c.scope.names[name] = l
	return l
}

func (c *Checker) push() { c.scope = &scope{names: map[string]*Local{}, parent: c.scope} }
func (c *Checker) pop()  { c.scope = c.scope.parent }

func (c *Checker) funcBody(fd *FuncDecl) {
	c.fn = fd
	c.scope = &scope{names: map[string]*Local{}}
	for _, p := range fd.Params {
		l := c.declare(p.Pos, p.Name, p.T, p.Mutable)
		l.Param = true
		l.ByRef = p.Mutable
		p.Local = l
	}
	c.block(fd.Body)
}

func (c *Checker) block(b *Block) {
	c.push()
	for _, s := range b.Stmts {
		c.stmt(s)
	}
	c.pop()
}

func (c *Checker) stmt(s Stmt) {
	switch s := s.(type) {
	case *Block:
		c.block(s)
	case *VarStmt:
		var want Type
		if s.Type != nil {
			want = c.resolveType(s.Type)
		}
		t := c.expr(s.Value, want)
		if want != nil {
			c.expectType(s.Value.position(), want, t)
			t = want
		}
		if t == TVoid {
			c.errorf(s.Pos, "this expression has no value")
		}
		s.Local = c.declare(s.Pos, s.Name, t, s.Mutable)
	case *AssignStmt:
		t := c.expr(s.Target, nil)
		root := c.lvalueRoot(s.Target)
		if root == nil {
			c.errorf(s.Pos, "only variables, fields and array elements can be assigned")
		} else if !root.Mutable {
			c.errorf(s.Pos, "%q is constant; declare it with `let` to change it", root.Name)
		}
		v := c.expr(s.Value, t)
		c.expectType(s.Value.position(), t, v)
	case *ExprStmt:
		if m, ok := s.Expr.(*MatchExpr); ok {
			c.match(m, nil, false)
			return
		}
		t := c.expr(s.Expr, nil)
		if _, ok := s.Expr.(*CallExpr); !ok {
			c.errorf(s.Pos, "this expression does nothing")
		}
		_ = t
	case *ReturnStmt:
		want := c.fn.ResultT
		if s.Value == nil {
			if want != TVoid {
				c.errorf(s.Pos, "this function must return a %s", want)
			}
			return
		}
		t := c.expr(s.Value, want)
		if want == TVoid {
			c.errorf(s.Pos, "this function returns nothing")
			return
		}
		c.expectType(s.Value.position(), want, t)
	case *IfStmt:
		c.expectType(s.Cond.position(), TBool, c.expr(s.Cond, TBool))
		c.block(s.Then)
		if s.Else != nil {
			c.stmt(s.Else)
		}
	case *WhileStmt:
		c.expectType(s.Cond.position(), TBool, c.expr(s.Cond, TBool))
		c.loops++
		c.block(s.Body)
		c.loops--
	case *ForStmt:
		var elem Type
		call, isCall := s.Iter.(*CallExpr)
		t := c.expr(s.Iter, nil)
		switch {
		case isCall && (call.Builtin == "Range.upTo" || call.Builtin == "Range.between"):
			s.Kind = forRange
			elem = TInt
		case isCall && call.Builtin == "Array.values":
			s.Kind = forArray
			elem = call.Args[0].Value.Type().(*ArrayType).Elem
		default:
			c.errorf(s.Iter.position(), "for ... of needs Range.upTo, Range.between or Array.values, found %s", t)
			elem = TInt
		}
		c.push()
		s.Local = c.declare(s.Pos, s.Name, elem, false)
		c.loops++
		c.block(s.Body)
		c.loops--
		c.pop()
	case *BreakStmt, *ContinueStmt:
		if c.loops == 0 {
			c.errorf(stmtPos(s), "break and continue must be inside a loop")
		}
	}
}

func stmtPos(s Stmt) Pos {
	switch s := s.(type) {
	case *BreakStmt:
		return s.Pos
	case *ContinueStmt:
		return s.Pos
	}
	return Pos{}
}

// lvalueRoot returns the variable at the root of an assignable expression.
func (c *Checker) lvalueRoot(e Expr) *Local {
	switch e := e.(type) {
	case *Ident:
		return e.Local
	case *FieldExpr:
		if e.Builtin != "" || e.Ctor != nil {
			return nil
		}
		return c.lvalueRoot(e.Target)
	case *IndexExpr:
		return c.lvalueRoot(e.Target)
	}
	return nil
}

func (c *Checker) expr(e Expr, want Type) Type {
	t := c.exprInner(e, want)
	if t == nil {
		t = TVoid
	}
	e.setType(t)
	return t
}

func (c *Checker) exprInner(e Expr, want Type) Type {
	switch e := e.(type) {
	case *IntLit:
		if _, err := strconv.ParseInt(e.Text, 0, 64); err != nil {
			c.errorf(e.P, "integer literal %s does not fit in int", e.Text)
		}
		return TInt
	case *FloatLit:
		return TF64
	case *StringLit:
		return TStr
	case *CharLit:
		return TChar
	case *BoolLit:
		return TBool
	case *TemplateLit:
		for _, part := range e.Parts {
			t := c.expr(part, TStr)
			if t != TStr {
				c.errorf(part.position(), "a template string part must be a string, found %s; convert it, for example with Int.toString", t)
			}
		}
		return TStr
	case *Ident:
		if l := c.lookup(e.Name); l != nil {
			e.Local = l
			l.Used = true
			return l.T
		}
		if fd, ok := c.funcs[e.Name]; ok {
			e.Func = fd
			c.errorf(e.P, "functions as values are not supported by this compiler yet")
			return TVoid
		}
		if v := c.findCtor(e.Name, want); v != nil {
			if len(v.Fields) > 0 {
				c.errorf(e.P, "%s takes %d fields", v.Name, len(v.Fields))
			}
			e.Ctor = v
			return v.Enum
		}
		if _, ok := c.modules[e.Name]; ok {
			c.errorf(e.P, "a module is not a value")
			return TVoid
		}
		c.errorf(e.P, "unknown name %q", e.Name)
		return TVoid
	case *FieldExpr:
		if id, ok := e.Target.(*Ident); ok && c.lookup(id.Name) == nil {
			if et, ok := c.types[id.Name].(*EnumType); ok {
				v := et.Variant(e.Name)
				if v == nil {
					c.errorf(e.P, "enum %s has no constructor %q", et, e.Name)
					return et
				}
				if len(v.Fields) > 0 {
					c.errorf(e.P, "%s takes %d fields", v.Name, len(v.Fields))
				}
				e.Ctor = v
				return et
			}
		}
		t := c.expr(e.Target, nil)
		switch tt := t.(type) {
		case *StructType:
			f := tt.Field(e.Name)
			if f == nil {
				c.errorf(e.P, "struct %s has no field %q", tt, e.Name)
				return TVoid
			}
			return f.T
		case *ArrayType, *PrimType:
			if e.Name == "length" {
				c.errorf(e.P, "there is no .length; use Array.length(x) or Str.length(x)")
			} else {
				c.errorf(e.P, "%s has no fields", t)
			}
			return TVoid
		}
		c.errorf(e.P, "%s has no fields", t)
		return TVoid
	case *IndexExpr:
		t := c.expr(e.Target, nil)
		c.expectType(e.Index.position(), TInt, c.expr(e.Index, TInt))
		at, ok := t.(*ArrayType)
		if !ok {
			c.errorf(e.P, "only arrays can be indexed, found %s", t)
			return TVoid
		}
		return at.Elem
	case *UnaryExpr:
		t := c.expr(e.Operand, want)
		switch e.Op {
		case "-":
			if !isNumeric(t) {
				c.errorf(e.P, "- needs a number, found %s", t)
			}
		case "!":
			c.expectType(e.P, TBool, t)
		case "~":
			c.expectType(e.P, TInt, t)
		}
		return t
	case *BinaryExpr:
		return c.binary(e)
	case *CallExpr:
		return c.call(e, want)
	case *StructLit:
		return c.structLit(e, want)
	case *ArrayLit:
		var elem Type
		if at, ok := want.(*ArrayType); ok {
			elem = at.Elem
		}
		for _, el := range e.Elems {
			t := c.expr(el, elem)
			if elem == nil {
				elem = t
			}
			c.expectType(el.position(), elem, t)
		}
		if elem == nil {
			c.errorf(e.P, "the type of [] is unknown; add a type annotation")
			return TVoid
		}
		return c.arrayOf(elem)
	case *MatchExpr:
		return c.match(e, want, true)
	}
	c.errorf(e.position(), "unsupported expression")
	return TVoid
}

func (c *Checker) findCtor(name string, want Type) *Variant {
	if et, ok := want.(*EnumType); ok {
		if v := et.Variant(name); v != nil {
			return v
		}
	}
	var found *Variant
	for _, t := range c.types {
		if et, ok := t.(*EnumType); ok {
			if v := et.Variant(name); v != nil {
				if found != nil {
					return found // ambiguous names resolve to the first; the type check reports mismatches
				}
				found = v
			}
		}
	}
	return found
}

func (c *Checker) binary(e *BinaryExpr) Type {
	l := c.expr(e.Left, nil)
	r := c.expr(e.Right, l)
	switch e.Op {
	case "&&", "||":
		c.expectType(e.Left.position(), TBool, l)
		c.expectType(e.Right.position(), TBool, r)
		return TBool
	case "==", "!=":
		if !sameType(l, r) {
			c.errorf(e.P, "cannot compare %s with %s", l, r)
		} else if _, ok := l.(*PrimType); !ok {
			c.errorf(e.P, "== on %s is not supported by this compiler yet", l)
		}
		return TBool
	case "<", "<=", ">", ">=":
		if !sameType(l, r) || !(isNumeric(l) || l == TChar || l == TStr) {
			c.errorf(e.P, "cannot order %s and %s", l, r)
		}
		return TBool
	case "+", "-", "*", "/", "%":
		if !sameType(l, r) || !isNumeric(l) {
			if l == TStr || r == TStr {
				c.errorf(e.P, "strings are joined with a template string, not +")
			} else {
				c.errorf(e.P, "%s needs two numbers of the same type, found %s and %s", e.Op, l, r)
			}
		}
		return l
	case "+%", "-%", "*%", "&", "|", "^", "<<", ">>":
		if l != TInt || r != TInt {
			c.errorf(e.P, "%s needs two integers, found %s and %s", e.Op, l, r)
		}
		return TInt
	}
	c.errorf(e.P, "unknown operator %s", e.Op)
	return TVoid
}

func (c *Checker) structLit(e *StructLit, want Type) Type {
	var st *StructType
	if e.TypeName != "" {
		t, ok := c.types[e.TypeName]
		st, _ = t.(*StructType)
		if !ok || st == nil {
			c.errorf(e.P, "%q is not a struct", e.TypeName)
			return TVoid
		}
	} else {
		st, _ = want.(*StructType)
		if st == nil {
			c.errorf(e.P, "the struct type is unknown here; write `TypeName { ... }`")
			return TVoid
		}
	}
	e.Struct = st
	seen := map[string]bool{}
	for _, fi := range e.Fields {
		f := st.Field(fi.Name)
		if f == nil {
			c.errorf(fi.Pos, "struct %s has no field %q", st, fi.Name)
			continue
		}
		if seen[fi.Name] {
			c.errorf(fi.Pos, "field %q is given twice", fi.Name)
		}
		seen[fi.Name] = true
		c.expectType(fi.Value.position(), f.T, c.expr(fi.Value, f.T))
	}
	for _, f := range st.Fields {
		if !seen[f.Name] && f.Default == nil {
			c.errorf(e.P, "missing field %q of %s", f.Name, st)
		}
	}
	return st
}

func (c *Checker) match(m *MatchExpr, want Type, isValue bool) Type {
	m.IsValue = isValue
	st := c.expr(m.Subject, nil)
	et, ok := st.(*EnumType)
	if !ok {
		c.errorf(m.P, "match needs an enum value, found %s", st)
		return TVoid
	}
	covered := map[*Variant]bool{}
	hasWildcard := false
	result := want
	for _, arm := range m.Arms {
		c.push()
		p := arm.Pattern
		if p.Wildcard {
			hasWildcard = true
		} else {
			v := et.Variant(p.Name)
			if v == nil {
				c.errorf(p.Pos, "enum %s has no constructor %q", et, p.Name)
			} else {
				p.Variant = v
				if len(p.Binds) != len(v.Fields) && !(len(p.Binds) == 0 && len(v.Fields) == 0) {
					c.errorf(p.Pos, "%s has %d fields, the pattern lists %d", v.Name, len(v.Fields), len(p.Binds))
				} else {
					for i, b := range p.Binds {
						if b == "_" {
							p.Locals = append(p.Locals, nil)
							continue
						}
						p.Locals = append(p.Locals, c.declare(p.Pos, b, v.Fields[i].T, false))
					}
				}
				if arm.Guard == nil {
					covered[v] = true
				}
			}
		}
		if arm.Guard != nil {
			c.expectType(arm.Guard.position(), TBool, c.expr(arm.Guard, TBool))
		}
		if arm.Block != nil {
			if isValue {
				c.errorf(arm.Pos, "a match used as a value needs a single expression in each arm")
			}
			c.block(arm.Block)
		} else {
			t := c.expr(arm.Body, result)
			if isValue {
				if result == nil {
					result = t
				}
				c.expectType(arm.Body.position(), result, t)
			}
		}
		c.pop()
	}
	if !hasWildcard {
		var missing []string
		for _, v := range et.Variants {
			if !covered[v] {
				missing = append(missing, v.Name)
			}
		}
		if len(missing) > 0 {
			c.errorf(m.P, "match does not cover %s; add the cases or a `_` arm", strings.Join(missing, ", "))
		}
	}
	if !isValue {
		return TVoid
	}
	if result == nil {
		return TVoid
	}
	return result
}

// Builtins that live in the prelude, with no import.
var preludeFuncs = map[string]*Builtin{
	"print":  {Name: "print", Params: []Type{TStr}, Result: TVoid},
	"assert": {Name: "assert", Params: []Type{TBool}, Result: TVoid},
	"panic":  {Name: "panic", Params: []Type{TStr}, Result: TVoid},
}

func (c *Checker) call(e *CallExpr, want Type) Type {
	switch callee := e.Callee.(type) {
	case *Ident:
		if b, ok := preludeFuncs[callee.Name]; ok && c.lookup(callee.Name) == nil {
			return c.builtinCall(e, b)
		}
		if b, ok := c.imported[callee.Name]; ok && c.lookup(callee.Name) == nil {
			return c.builtinCall(e, b)
		}
		if fd, ok := c.funcs[callee.Name]; ok && c.lookup(callee.Name) == nil {
			e.Func = fd
			c.args(e, fd.Params)
			return fd.ResultT
		}
		if v := c.findCtor(callee.Name, want); v != nil && c.lookup(callee.Name) == nil {
			return c.ctorCall(e, v)
		}
		c.errorf(callee.P, "unknown function %q", callee.Name)
		c.argsLoose(e)
		return TVoid
	case *FieldExpr:
		if id, ok := callee.Target.(*Ident); ok && c.lookup(id.Name) == nil {
			if path, ok := c.modules[id.Name]; ok {
				b := stdModules[path][callee.Name]
				if b == nil {
					c.errorf(callee.P, "module %q has no function %q", path, callee.Name)
					c.argsLoose(e)
					return TVoid
				}
				return c.builtinCall(e, b)
			}
			if id.Name == "Array" || id.Name == "Range" {
				return c.preludeNamespaceCall(e, id.Name+"."+callee.Name, want)
			}
			if et, ok := c.types[id.Name].(*EnumType); ok {
				v := et.Variant(callee.Name)
				if v == nil {
					c.errorf(callee.P, "enum %s has no constructor %q", et, callee.Name)
					c.argsLoose(e)
					return et
				}
				return c.ctorCall(e, v)
			}
			if _, ok := c.types[id.Name]; !ok && startsUpper(id.Name) {
				c.errorf(id.P, "unknown module %q; import it first, for example `import * as %s from \"std/...\"`", id.Name, id.Name)
				c.argsLoose(e)
				return TVoid
			}
		}
		t := c.expr(callee.Target, nil)
		c.errorf(callee.P, "there are no method calls; call a module function such as Module.%s(value, ...) (value has type %s)", callee.Name, t)
		c.argsLoose(e)
		return TVoid
	}
	c.errorf(e.P, "this expression cannot be called")
	c.argsLoose(e)
	return TVoid
}

func (c *Checker) argsLoose(e *CallExpr) {
	for _, a := range e.Args {
		c.expr(a.Value, nil)
	}
}

func (c *Checker) ctorCall(e *CallExpr, v *Variant) Type {
	e.Ctor = v
	if len(e.Args) != len(v.Fields) {
		c.errorf(e.P, "%s takes %d fields, found %d", v.Name, len(v.Fields), len(e.Args))
	}
	for i, a := range e.Args {
		if a.Mutable {
			c.errorf(e.P, "constructor fields are not mutable")
		}
		if i < len(v.Fields) {
			c.expectType(a.Value.position(), v.Fields[i].T, c.expr(a.Value, v.Fields[i].T))
		} else {
			c.expr(a.Value, nil)
		}
	}
	return v.Enum
}

func (c *Checker) builtinCall(e *CallExpr, b *Builtin) Type {
	e.Builtin = b.Name
	if len(e.Args) != len(b.Params) {
		c.errorf(e.P, "%s takes %d arguments, found %d", b.Name, len(b.Params), len(e.Args))
	}
	for i, a := range e.Args {
		if a.Mutable {
			c.errorf(e.P, "%s does not change its arguments; remove `mutable`", b.Name)
		}
		var want Type
		if i < len(b.Params) {
			want = b.Params[i]
		}
		c.expectType(a.Value.position(), want, c.expr(a.Value, want))
	}
	return b.Result
}

func (c *Checker) preludeNamespaceCall(e *CallExpr, name string, want Type) Type {
	e.Builtin = name
	argc := func(n int) bool {
		if len(e.Args) != n {
			c.errorf(e.P, "%s takes %d arguments, found %d", name, n, len(e.Args))
			c.argsLoose(e)
			return false
		}
		return true
	}
	switch name {
	case "Range.upTo":
		if argc(1) {
			c.expectType(e.Args[0].Value.position(), TInt, c.expr(e.Args[0].Value, TInt))
		}
		return &RangeType{}
	case "Range.between":
		if argc(2) {
			c.expectType(e.Args[0].Value.position(), TInt, c.expr(e.Args[0].Value, TInt))
			c.expectType(e.Args[1].Value.position(), TInt, c.expr(e.Args[1].Value, TInt))
		}
		return &RangeType{}
	case "Array.length", "Array.values":
		if !argc(1) {
			return TVoid
		}
		t := c.expr(e.Args[0].Value, nil)
		at, ok := t.(*ArrayType)
		if !ok {
			c.errorf(e.P, "%s needs an array, found %s", name, t)
			return TVoid
		}
		if name == "Array.length" {
			return TInt
		}
		return at
	case "Array.push":
		if !argc(2) {
			return TVoid
		}
		if !e.Args[0].Mutable {
			c.errorf(e.P, "Array.push changes the array; write Array.push(mutable items, value)")
		}
		t := c.expr(e.Args[0].Value, nil)
		at, ok := t.(*ArrayType)
		if !ok {
			c.errorf(e.P, "Array.push needs an array, found %s", t)
			return TVoid
		}
		c.checkMutableArg(e.Args[0].Value)
		c.expectType(e.Args[1].Value.position(), at.Elem, c.expr(e.Args[1].Value, at.Elem))
		c.checkExclusive(e, []int{0})
		return TVoid
	case "Array.filled":
		if !argc(2) {
			return TVoid
		}
		var elem Type
		if at, ok := want.(*ArrayType); ok {
			elem = at.Elem
		}
		t := c.expr(e.Args[0].Value, elem)
		c.expectType(e.Args[1].Value.position(), TInt, c.expr(e.Args[1].Value, TInt))
		return c.arrayOf(t)
	}
	c.errorf(e.P, "unknown function %s", name)
	c.argsLoose(e)
	return TVoid
}

func (c *Checker) checkMutableArg(v Expr) {
	root := c.lvalueRoot(v)
	if root == nil {
		c.errorf(v.position(), "a mutable argument must be a variable, a field or an element")
		return
	}
	if !root.Mutable {
		c.errorf(v.position(), "%q is constant; declare it with `let` to pass it as mutable", root.Name)
	}
}

func (c *Checker) args(e *CallExpr, params []*Param) {
	if len(e.Args) != len(params) {
		c.errorf(e.P, "this call passes %d arguments, the function takes %d", len(e.Args), len(params))
	}
	var mutables []int
	for i, a := range e.Args {
		if i >= len(params) {
			c.expr(a.Value, nil)
			continue
		}
		p := params[i]
		c.expectType(a.Value.position(), p.T, c.expr(a.Value, p.T))
		switch {
		case p.Mutable && !a.Mutable:
			c.errorf(a.Value.position(), "parameter %q is mutable; write `mutable` before the argument", p.Name)
		case !p.Mutable && a.Mutable:
			c.errorf(a.Value.position(), "parameter %q is not mutable; remove `mutable`", p.Name)
		case p.Mutable:
			c.checkMutableArg(a.Value)
			mutables = append(mutables, i)
		}
	}
	c.checkExclusive(e, mutables)
}

// checkExclusive rejects a call that passes a variable as mutable and also
// passes it, or reads it, in another argument.
func (c *Checker) checkExclusive(e *CallExpr, mutables []int) {
	for _, mi := range mutables {
		root := c.lvalueRoot(e.Args[mi].Value)
		if root == nil {
			continue
		}
		for j, a := range e.Args {
			if j == mi {
				continue
			}
			if usesLocal(a.Value, root) {
				c.errorf(a.Value.position(), "%q is passed as mutable in this call, so no other argument can use it", root.Name)
			}
		}
	}
}

func usesLocal(e Expr, l *Local) bool {
	found := false
	var walk func(e Expr)
	walk = func(e Expr) {
		if e == nil || found {
			return
		}
		switch e := e.(type) {
		case *Ident:
			if e.Local == l {
				found = true
			}
		case *FieldExpr:
			walk(e.Target)
		case *IndexExpr:
			walk(e.Target)
			walk(e.Index)
		case *UnaryExpr:
			walk(e.Operand)
		case *BinaryExpr:
			walk(e.Left)
			walk(e.Right)
		case *CallExpr:
			for _, a := range e.Args {
				walk(a.Value)
			}
		case *TemplateLit:
			for _, p := range e.Parts {
				walk(p)
			}
		case *StructLit:
			for _, f := range e.Fields {
				walk(f.Value)
			}
		case *ArrayLit:
			for _, el := range e.Elems {
				walk(el)
			}
		case *MatchExpr:
			walk(e.Subject)
		}
	}
	walk(e)
	return found
}
