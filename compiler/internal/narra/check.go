package narra

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
	Name   string // C-facing name, such as "Int.toString"
	Params []Type
	Result Type
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

// The prelude is written in Narra and checked with every program.
const preludeSrc = `
enum Option<T> {
  Some(value: T)
  None
}

enum Result<T, E> {
  Ok(value: T)
  Err(error: E)
}
`

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
	narrow map[*Local]*Variant
	parent *scope
}

// genericType is a generic struct or enum declaration.
type genericType struct {
	name      string
	params    []*TypeParamType
	sdecl     *StructDecl
	edecl     *EnumDecl
	built     bool
	fields    []*Field   // struct template, with type parameters
	variants  []*Variant // enum template, with type parameters
	instances map[string]Type
}

type pendingFunc struct {
	fd  *FuncDecl
	env map[string]Type
}

type Checker struct {
	prog      *Program
	types     map[string]Type // non-generic named types
	generics  map[string]*genericType
	funcs     map[string]*FuncDecl
	funcTP    map[*FuncDecl][]*TypeParamType
	funcInsts map[string]*FuncDecl
	pending   []pendingFunc
	env       map[string]Type // type parameter names in scope
	modules   map[string]string
	imported  map[string]*Builtin
	errs      []string
	scope     *scope
	fn        *FuncDecl
	loops     int
	arrays    map[string]*ArrayType
	instTypes []Type
	localN    int
	isBinds   bool // an `is` pattern may bind names here
}

func Check(f *File) (*Program, error) {
	prelude, err := Parse(preludeSrc)
	if err != nil {
		return nil, fmt.Errorf("prelude: %v", err)
	}
	f.Decls = append(prelude.Decls, f.Decls...)
	c := &Checker{
		prog:      &Program{File: f},
		types:     map[string]Type{"int": TInt, "i64": TInt, "f64": TF64, "bool": TBool, "char": TChar, "string": TStr},
		generics:  map[string]*genericType{},
		funcs:     map[string]*FuncDecl{},
		funcTP:    map[*FuncDecl][]*TypeParamType{},
		funcInsts: map[string]*FuncDecl{},
		modules:   map[string]string{},
		imported:  map[string]*Builtin{},
		arrays:    map[string]*ArrayType{},
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
	var concrete []Type
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *StructDecl:
			if len(d.TypeParams) > 0 {
				c.declareGeneric(d.Pos, &genericType{name: d.Name, params: makeParams(d.TypeParams), sdecl: d})
				continue
			}
			d.T = &StructType{Name: d.Name}
			c.declareType(d.Pos, d.Name, d.T)
			concrete = append(concrete, d.T)
		case *EnumDecl:
			if len(d.TypeParams) > 0 {
				c.declareGeneric(d.Pos, &genericType{name: d.Name, params: makeParams(d.TypeParams), edecl: d})
				continue
			}
			d.T = &EnumType{Name: d.Name}
			c.declareType(d.Pos, d.Name, d.T)
			concrete = append(concrete, d.T)
		}
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *StructDecl:
			if d.T != nil {
				d.T.Fields = c.resolveFields(d.Fields)
			}
		case *EnumDecl:
			if d.T != nil {
				d.T.Variants = c.resolveVariants(d.Variants, d.T)
			}
		case *FuncDecl:
			if _, dup := c.funcs[d.Name]; dup {
				c.errorf(d.Pos, "function %q is declared twice", d.Name)
			}
			c.funcs[d.Name] = d
			c.env = nil
			if len(d.TypeParams) > 0 {
				tps := makeParams(d.TypeParams)
				c.funcTP[d] = tps
				c.env = envOf(tps, nil)
			}
			for _, p := range d.Params {
				p.T = c.resolveType(p.Type)
			}
			d.ResultT = TVoid
			if d.Result != nil {
				d.ResultT = c.resolveType(d.Result)
			}
			c.env = nil
			if len(d.TypeParams) == 0 {
				c.prog.Funcs = append(c.prog.Funcs, d)
			}
			if d.Name == "run" && d.Exported {
				c.prog.Entry = d
			}
		case *TestDecl:
			c.prog.Tests = append(c.prog.Tests, d)
		}
	}
	for _, g := range c.generics {
		c.buildGeneric(g)
	}
	// Defaults are checked after every field type is known.
	for _, d := range f.Decls {
		if sd, ok := d.(*StructDecl); ok && sd.T != nil {
			c.checkDefaults(sd.T.Fields)
		}
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *FuncDecl:
			if tps, ok := c.funcTP[d]; ok {
				// Check the generic body once with the parameters left open,
				// so errors point at the declaration.
				c.funcBody(d, envOf(tps, nil))
			} else {
				c.funcBody(d, nil)
			}
		case *TestDecl:
			c.fn = &FuncDecl{Name: "test", ResultT: TVoid}
			c.env = nil
			c.scope = &scope{names: map[string]*Local{}}
			c.block(d.Body)
		}
	}
	// Instances of generic functions are checked as they are discovered.
	for len(c.pending) > 0 {
		p := c.pending[0]
		c.pending = c.pending[1:]
		c.funcBody(p.fd, p.env)
	}
	if c.prog.Entry != nil {
		e := c.prog.Entry
		ok := len(e.Params) == 1 && sameType(e.Params[0].T, &ArrayType{Elem: TStr}) && !e.Params[0].Mutable && e.ResultT == TInt
		if !ok {
			c.errorf(e.Pos, "the entry function must be `export function run(args: Array<string>): int`")
		}
	}
	for _, t := range c.instTypes {
		if !containsTypeParam(t) {
			concrete = append(concrete, t)
		}
	}
	c.orderTypes(concrete)
	keys := make([]string, 0, len(c.arrays))
	for k, a := range c.arrays {
		if !containsTypeParam(a) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		c.prog.Arrays = append(c.prog.Arrays, c.arrays[k])
	}
}

var typeParamID int

func makeParams(names []string) []*TypeParamType {
	tps := make([]*TypeParamType, len(names))
	for i, n := range names {
		typeParamID++
		tps[i] = &TypeParamType{Name: n, ID: typeParamID}
	}
	return tps
}

func envOf(tps []*TypeParamType, args []Type) map[string]Type {
	env := map[string]Type{}
	for i, tp := range tps {
		if args != nil {
			env[tp.Name] = args[i]
		} else {
			env[tp.Name] = tp
		}
	}
	return env
}

func (c *Checker) declareType(pos Pos, name string, t Type) {
	if _, dup := c.types[name]; dup {
		c.errorf(pos, "type %q is declared twice", name)
	}
	if _, dup := c.generics[name]; dup {
		c.errorf(pos, "type %q is declared twice", name)
	}
	c.types[name] = t
}

func (c *Checker) declareGeneric(pos Pos, g *genericType) {
	if _, dup := c.types[g.name]; dup {
		c.errorf(pos, "type %q is declared twice", g.name)
	}
	if _, dup := c.generics[g.name]; dup {
		c.errorf(pos, "type %q is declared twice", g.name)
	}
	g.instances = map[string]Type{}
	c.generics[g.name] = g
}

func (c *Checker) resolveFields(decls []*FieldDecl) []*Field {
	var fields []*Field
	seen := map[string]bool{}
	for _, fd := range decls {
		if seen[fd.Name] {
			c.errorf(fd.Pos, "field %q is declared twice", fd.Name)
		}
		seen[fd.Name] = true
		fields = append(fields, &Field{Name: fd.Name, T: c.resolveType(fd.Type), Default: fd.Default})
	}
	return fields
}

func (c *Checker) resolveVariants(decls []*VariantDecl, et *EnumType) []*Variant {
	var vs []*Variant
	for i, vd := range decls {
		vs = append(vs, &Variant{Name: vd.Name, Index: i, Enum: et, Fields: c.resolveFields(vd.Fields)})
	}
	return vs
}

func (c *Checker) checkDefaults(fields []*Field) {
	for _, f := range fields {
		if f.Default == nil {
			continue
		}
		c.scope = &scope{names: map[string]*Local{}}
		t := c.expr(f.Default, f.T)
		c.expectType(f.Default.position(), f.T, t)
		if !isConstant(f.Default) {
			c.errorf(f.Default.position(), "a field default must be a constant")
		}
	}
}

// buildGeneric resolves a generic declaration's fields with its type
// parameters left open. Instances substitute into this template.
func (c *Checker) buildGeneric(g *genericType) {
	if g.built {
		return
	}
	g.built = true
	saved := c.env
	c.env = envOf(g.params, nil)
	if g.sdecl != nil {
		g.fields = c.resolveFields(g.sdecl.Fields)
		for i, f := range g.fields {
			if f.Default != nil && containsTypeParam(f.T) {
				c.errorf(g.sdecl.Fields[i].Pos, "a default for a field whose type uses a type parameter is not supported by this compiler yet")
				f.Default = nil
			}
		}
		c.checkDefaults(g.fields)
	} else {
		g.variants = c.resolveVariants(g.edecl.Variants, nil)
	}
	c.env = saved
}

func (c *Checker) instantiate(g *genericType, args []Type) Type {
	c.buildGeneric(g)
	key := typeArgsMangle(args)
	if t, ok := g.instances[key]; ok {
		return t
	}
	m := map[*TypeParamType]Type{}
	for i, p := range g.params {
		m[p] = args[i]
	}
	if g.sdecl != nil {
		st := &StructType{Name: g.name, Args: args}
		g.instances[key] = st
		for _, f := range g.fields {
			st.Fields = append(st.Fields, &Field{Name: f.Name, T: c.substitute(f.T, m), Default: f.Default})
		}
		c.instTypes = append(c.instTypes, st)
		return st
	}
	et := &EnumType{Name: g.name, Args: args}
	g.instances[key] = et
	for _, v := range g.variants {
		nv := &Variant{Name: v.Name, Index: v.Index, Enum: et}
		for _, f := range v.Fields {
			nv.Fields = append(nv.Fields, &Field{Name: f.Name, T: c.substitute(f.T, m)})
		}
		et.Variants = append(et.Variants, nv)
	}
	c.instTypes = append(c.instTypes, et)
	return et
}

func (c *Checker) substitute(t Type, m map[*TypeParamType]Type) Type {
	switch t := t.(type) {
	case *TypeParamType:
		if r, ok := m[t]; ok {
			return r
		}
		return t
	case *ArrayType:
		return c.arrayOf(c.substitute(t.Elem, m))
	case *StructType:
		if len(t.Args) == 0 {
			return t
		}
		return c.instantiate(c.generics[t.Name], c.substituteAll(t.Args, m))
	case *EnumType:
		if len(t.Args) == 0 {
			return t
		}
		return c.instantiate(c.generics[t.Name], c.substituteAll(t.Args, m))
	}
	return t
}

func (c *Checker) substituteAll(ts []Type, m map[*TypeParamType]Type) []Type {
	out := make([]Type, len(ts))
	for i, t := range ts {
		out[i] = c.substitute(t, m)
	}
	return out
}

// unify matches a type that contains parameters against a concrete type and
// records what each parameter must be.
func unify(pat, act Type, m map[*TypeParamType]Type) {
	if pat == nil || act == nil {
		return
	}
	switch p := pat.(type) {
	case *TypeParamType:
		if _, known := m[p]; !known {
			if _, isVoid := act.(*PrimType); !isVoid || act != TVoid {
				m[p] = act
			}
		}
	case *ArrayType:
		if a, ok := act.(*ArrayType); ok {
			unify(p.Elem, a.Elem, m)
		}
	case *StructType:
		if a, ok := act.(*StructType); ok && a.Name == p.Name && len(a.Args) == len(p.Args) {
			for i := range p.Args {
				unify(p.Args[i], a.Args[i], m)
			}
		}
	case *EnumType:
		if a, ok := act.(*EnumType); ok && a.Name == p.Name && len(a.Args) == len(p.Args) {
			for i := range p.Args {
				unify(p.Args[i], a.Args[i], m)
			}
		}
	}
}

// orderTypes sorts structs and enums so that each one is defined after the
// types it holds inline, and rejects types that contain themselves.
func (c *Checker) orderTypes(types []Type) {
	state := map[Type]int{}
	var visit func(t Type)
	visit = func(t Type) {
		switch t.(type) {
		case *StructType, *EnumType:
		default:
			return
		}
		if state[t] == 2 {
			return
		}
		if state[t] == 1 {
			c.errorf(Pos{}, "type %s contains itself; use an array for recursive data", t)
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
		for _, f := range fields {
			visit(f.T)
		}
		state[t] = 2
		c.prog.Types = append(c.prog.Types, t)
	}
	for _, t := range types {
		visit(t)
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
	case *Ident:
		return e.Ctor != nil
	case *CallExpr:
		if e.Ctor == nil {
			return false
		}
		for _, a := range e.Args {
			if !isConstant(a.Value) {
				return false
			}
		}
		return true
	}
	return false
}

func (c *Checker) resolveType(te *TypeExpr) Type {
	if te == nil {
		return TVoid
	}
	if t, ok := c.env[te.Name]; ok && len(te.Args) == 0 {
		return t
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
	if g, ok := c.generics[te.Name]; ok {
		if len(te.Args) != len(g.params) {
			c.errorf(te.Pos, "%s takes %d type arguments", te.Name, len(g.params))
			return TVoid
		}
		args := make([]Type, len(te.Args))
		for i, a := range te.Args {
			args[i] = c.resolveType(a)
		}
		return c.instantiate(g, args)
	}
	if len(te.Args) > 0 {
		c.errorf(te.Pos, "type %s is not generic", te.Name)
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

func (c *Checker) narrowed(l *Local) *Variant {
	for s := c.scope; s != nil; s = s.parent {
		if v, ok := s.narrow[l]; ok {
			return v
		}
	}
	return nil
}

func (c *Checker) setNarrow(l *Local, v *Variant) {
	if c.scope.narrow == nil {
		c.scope.narrow = map[*Local]*Variant{}
	}
	c.scope.narrow[l] = v
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

func (c *Checker) funcBody(fd *FuncDecl, env map[string]Type) {
	c.fn = fd
	c.env = env
	c.loops = 0
	c.scope = &scope{names: map[string]*Local{}}
	for _, p := range fd.Params {
		l := c.declare(p.Pos, p.Name, p.T, p.Mutable)
		l.Param = true
		l.ByRef = p.Mutable
		p.Local = l
	}
	c.block(fd.Body)
	c.env = nil
}

// inGenericTemplate reports whether the current body is a generic function
// checked with its type parameters left open. Such a body is never emitted.
func (c *Checker) inGenericTemplate() bool {
	for _, t := range c.env {
		if _, ok := t.(*TypeParamType); ok {
			return true
		}
	}
	return false
}

func (c *Checker) block(b *Block) {
	c.push()
	for _, s := range b.Stmts {
		c.stmt(s)
	}
	c.pop()
}

// diverges reports whether a block always leaves by return, break, continue
// or panic.
func diverges(b *Block) bool {
	if b == nil || len(b.Stmts) == 0 {
		return false
	}
	switch s := b.Stmts[len(b.Stmts)-1].(type) {
	case *ReturnStmt, *BreakStmt, *ContinueStmt:
		return true
	case *ExprStmt:
		if call, ok := s.Expr.(*CallExpr); ok {
			if id, ok := call.Callee.(*Ident); ok && id.Name == "panic" {
				return true
			}
		}
	case *IfStmt:
		if s.Else == nil || !diverges(s.Then) {
			return false
		}
		if eb, ok := s.Else.(*Block); ok {
			return diverges(eb)
		}
		return diverges(&Block{Stmts: []Stmt{s.Else}})
	case *Block:
		return diverges(s)
	}
	return false
}

// otherVariant returns the variant an enum value must be when it is not v,
// for enums with exactly two variants.
func otherVariant(v *Variant) *Variant {
	if v == nil || len(v.Enum.Variants) != 2 {
		return nil
	}
	return v.Enum.Variants[1-v.Index]
}

func (c *Checker) narrowable(e Expr) *Local {
	id, ok := e.(*Ident)
	if !ok || id.Local == nil || id.Local.Mutable {
		return nil
	}
	return id.Local
}

func (c *Checker) ifStmt(s *IfStmt) {
	is, isPattern := s.Cond.(*IsExpr)
	if isPattern {
		c.isBinds = true
	}
	c.expectType(s.Cond.position(), TBool, c.expr(s.Cond, TBool))
	c.isBinds = false
	var subject *Local
	var v *Variant
	if isPattern && !is.Pattern.Wildcard {
		subject = c.narrowable(is.Subject)
		v = is.Pattern.Variant
	}
	c.push()
	if isPattern {
		for _, l := range is.Pattern.Locals {
			if l != nil {
				c.scope.names[l.Name] = l
			}
		}
		if subject != nil && v != nil {
			c.setNarrow(subject, v)
		}
	}
	for _, st := range s.Then.Stmts {
		c.stmt(st)
	}
	c.pop()
	other := otherVariant(v)
	if s.Else != nil {
		c.push()
		if subject != nil && other != nil {
			c.setNarrow(subject, other)
		}
		c.stmt(s.Else)
		c.pop()
	} else if subject != nil && other != nil && diverges(s.Then) {
		c.setNarrow(subject, other)
	}
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
		c.expr(s.Expr, nil)
		if call, ok := s.Expr.(*CallExpr); !ok {
			c.errorf(s.Pos, "this expression does nothing")
		} else if isResult(call.Type()) {
			c.errorf(s.Pos, "this call returns a Result that is not checked; handle it, or discard it with `const _ = ...`")
		}
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
		c.ifStmt(s)
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
			elem = TVoid
			if at, ok := call.Args[0].Value.Type().(*ArrayType); ok {
				elem = at.Elem
			}
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
	case *BreakStmt:
		if c.loops == 0 {
			c.errorf(s.Pos, "break must be inside a loop")
		}
	case *ContinueStmt:
		if c.loops == 0 {
			c.errorf(s.Pos, "continue must be inside a loop")
		}
	}
}

func isResult(t Type) bool {
	et, ok := t.(*EnumType)
	return ok && et.Name == "Result" && len(et.Args) == 2
}

// lvalueRoot returns the variable at the root of an assignable expression.
func (c *Checker) lvalueRoot(e Expr) *Local {
	switch e := e.(type) {
	case *Ident:
		return e.Local
	case *FieldExpr:
		if e.Builtin != "" || e.Ctor != nil || e.Narrow != nil {
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
		if _, ok := c.funcs[e.Name]; ok {
			c.errorf(e.P, "functions as values are not supported by this compiler yet")
			return TVoid
		}
		if t, v := c.ctor(e.P, "", e.Name, nil, want); t != nil {
			e.Ctor = v
			return t
		}
		if _, ok := c.modules[e.Name]; ok {
			c.errorf(e.P, "a module is not a value")
			return TVoid
		}
		c.errorf(e.P, "unknown name %q", e.Name)
		return TVoid
	case *FieldExpr:
		if id, ok := e.Target.(*Ident); ok && c.lookup(id.Name) == nil && c.isEnumName(id.Name) {
			t, v := c.ctor(e.P, id.Name, e.Name, nil, want)
			e.Ctor = v
			return t
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
		case *EnumType:
			var v *Variant
			if l := c.narrowable(e.Target); l != nil {
				v = c.narrowed(l)
			}
			if v == nil {
				c.errorf(e.P, "this %s may hold any of its variants; check it with `is` or `match` before reading %q", tt, e.Name)
				return TVoid
			}
			for _, f := range v.Fields {
				if f.Name == e.Name {
					e.Narrow = v
					return f.T
				}
			}
			c.errorf(e.P, "%s has no field %q", v.Name, e.Name)
			return TVoid
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
	case *IsExpr:
		binds := c.isBinds
		c.isBinds = false
		t := c.expr(e.Subject, nil)
		et, ok := t.(*EnumType)
		if !ok {
			c.errorf(e.P, "`is` needs an enum value, found %s", t)
			return TBool
		}
		c.resolvePattern(e.Pattern, et)
		if len(e.Pattern.Binds) > 0 && !binds {
			c.errorf(e.P, "an `is` pattern can only bind names when it is the whole condition of an if")
		}
		return TBool
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

func (c *Checker) isEnumName(name string) bool {
	if _, ok := c.types[name].(*EnumType); ok {
		return true
	}
	g, ok := c.generics[name]
	return ok && g.edecl != nil
}

// ctor resolves an enum constructor by name and checks its arguments.
// enumName restricts the search to one enum when the call is qualified, as
// in Shape.Circle. It returns nil when no enum has a constructor of that name.
func (c *Checker) ctor(pos Pos, enumName, name string, args []*Arg, want Type) (Type, *Variant) {
	// The expected type decides first, so `None` takes its type from context.
	if et, ok := want.(*EnumType); ok && (enumName == "" || enumName == et.Name) {
		if v := et.Variant(name); v != nil {
			c.ctorArgs(pos, v, args)
			return et, v
		}
	}
	for tn, t := range c.types {
		et, ok := t.(*EnumType)
		if !ok || (enumName != "" && tn != enumName) {
			continue
		}
		if v := et.Variant(name); v != nil {
			c.ctorArgs(pos, v, args)
			return et, v
		}
	}
	names := make([]string, 0, len(c.generics))
	for n := range c.generics {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, gn := range names {
		g := c.generics[gn]
		if g.edecl == nil || (enumName != "" && gn != enumName) {
			continue
		}
		c.buildGeneric(g)
		for _, tv := range g.variants {
			if tv.Name != name {
				continue
			}
			m := map[*TypeParamType]Type{}
			if et, ok := want.(*EnumType); ok && et.Name == g.name {
				for i, p := range g.params {
					m[p] = et.Args[i]
				}
			}
			if len(args) != len(tv.Fields) {
				c.errorf(pos, "%s takes %d fields, found %d", name, len(tv.Fields), len(args))
				for _, a := range args {
					c.expr(a.Value, nil)
				}
				return TVoid, nil
			}
			for i, a := range args {
				fw := c.substitute(tv.Fields[i].T, m)
				if containsOpen(fw, g.params) {
					fw = nil
				}
				at := c.expr(a.Value, fw)
				unify(tv.Fields[i].T, at, m)
			}
			targs := make([]Type, len(g.params))
			for i, p := range g.params {
				t, ok := m[p]
				if !ok {
					c.errorf(pos, "cannot tell the type of %s here; add a type annotation", name)
					return TVoid, nil
				}
				targs[i] = t
			}
			et := c.instantiate(g, targs).(*EnumType)
			v := et.Variants[tv.Index]
			for i, a := range args {
				c.expectType(a.Value.position(), v.Fields[i].T, a.Value.Type())
				if a.Mutable {
					c.errorf(pos, "constructor fields are not mutable")
				}
			}
			return et, v
		}
	}
	if enumName != "" {
		c.errorf(pos, "enum %s has no constructor %q", enumName, name)
		return TVoid, nil
	}
	return nil, nil
}

// containsOpen reports whether t still mentions one of the given parameters.
func containsOpen(t Type, params []*TypeParamType) bool {
	switch t := t.(type) {
	case *TypeParamType:
		for _, p := range params {
			if p == t {
				return true
			}
		}
	case *ArrayType:
		return containsOpen(t.Elem, params)
	case *StructType:
		for _, a := range t.Args {
			if containsOpen(a, params) {
				return true
			}
		}
	case *EnumType:
		for _, a := range t.Args {
			if containsOpen(a, params) {
				return true
			}
		}
	}
	return false
}

func (c *Checker) ctorArgs(pos Pos, v *Variant, args []*Arg) {
	if len(args) != len(v.Fields) {
		c.errorf(pos, "%s takes %d fields, found %d", v.Name, len(v.Fields), len(args))
	}
	for i, a := range args {
		if a.Mutable {
			c.errorf(pos, "constructor fields are not mutable")
		}
		if i < len(v.Fields) {
			c.expectType(a.Value.position(), v.Fields[i].T, c.expr(a.Value, v.Fields[i].T))
		} else {
			c.expr(a.Value, nil)
		}
	}
}

func (c *Checker) resolvePattern(p *Pattern, et *EnumType) {
	if p.Wildcard {
		return
	}
	v := et.Variant(p.Name)
	if v == nil {
		c.errorf(p.Pos, "enum %s has no constructor %q", et, p.Name)
		return
	}
	p.Variant = v
	if len(p.Binds) != len(v.Fields) && len(p.Binds) != 0 {
		c.errorf(p.Pos, "%s has %d fields, the pattern lists %d", v.Name, len(v.Fields), len(p.Binds))
		return
	}
	p.Locals = nil
	for i, b := range p.Binds {
		if b == "_" {
			p.Locals = append(p.Locals, nil)
			continue
		}
		if c.lookup(b) != nil {
			c.errorf(p.Pos, "%q is already declared in this function; shadowing is not allowed", b)
		}
		c.localN++
		p.Locals = append(p.Locals, &Local{Name: b, T: v.Fields[i].T, Pos: p.Pos, CName: fmt.Sprintf("v_%s_%d", b, c.localN)})
	}
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
			c.errorf(e.P, "== on %s is not supported by this compiler yet; use `is` or `match` for enums", l)
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
	switch {
	case e.TypeName != "" && c.generics[e.TypeName] != nil && c.generics[e.TypeName].sdecl != nil:
		g := c.generics[e.TypeName]
		c.buildGeneric(g)
		m := map[*TypeParamType]Type{}
		if wt, ok := want.(*StructType); ok && wt.Name == g.name {
			for i, p := range g.params {
				m[p] = wt.Args[i]
			}
		}
		for _, fi := range e.Fields {
			for _, f := range g.fields {
				if f.Name == fi.Name {
					fw := c.substitute(f.T, m)
					if containsOpen(fw, g.params) {
						fw = nil
					}
					unify(f.T, c.expr(fi.Value, fw), m)
				}
			}
		}
		targs := make([]Type, len(g.params))
		for i, p := range g.params {
			t, ok := m[p]
			if !ok {
				c.errorf(e.P, "cannot tell type parameter %s of %s here; add a type annotation", p.Name, g.name)
				return TVoid
			}
			targs[i] = t
		}
		st = c.instantiate(g, targs).(*StructType)
	case e.TypeName != "":
		t, ok := c.types[e.TypeName]
		st, _ = t.(*StructType)
		if !ok || st == nil {
			c.errorf(e.P, "%q is not a struct", e.TypeName)
			return TVoid
		}
	default:
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
		if fi.Value.Type() == nil {
			c.expr(fi.Value, f.T)
		}
		c.expectType(fi.Value.position(), f.T, fi.Value.Type())
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
			c.resolvePattern(p, et)
			for _, l := range p.Locals {
				if l != nil {
					c.scope.names[l.Name] = l
				}
			}
			if p.Variant != nil && arm.Guard == nil {
				covered[p.Variant] = true
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
	if !isValue || result == nil {
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
		if c.lookup(callee.Name) != nil {
			c.errorf(callee.P, "calling a function value is not supported by this compiler yet")
			c.argsLoose(e)
			return TVoid
		}
		if b, ok := preludeFuncs[callee.Name]; ok {
			return c.builtinCall(e, b)
		}
		if b, ok := c.imported[callee.Name]; ok {
			return c.builtinCall(e, b)
		}
		if fd, ok := c.funcs[callee.Name]; ok {
			if _, generic := c.funcTP[fd]; generic {
				return c.genericCall(e, fd, want)
			}
			e.Func = fd
			c.args(e, fd.Params, func(i int) Type { return fd.Params[i].T })
			return fd.ResultT
		}
		if t, v := c.ctor(callee.P, "", callee.Name, e.Args, want); t != nil {
			e.Ctor = v
			return t
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
			if c.isEnumName(id.Name) {
				t, v := c.ctor(callee.P, id.Name, callee.Name, e.Args, want)
				e.Ctor = v
				return t
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

func (c *Checker) genericCall(e *CallExpr, fd *FuncDecl, want Type) Type {
	tps := c.funcTP[fd]
	m := map[*TypeParamType]Type{}
	if want != nil {
		unify(fd.ResultT, want, m)
	}
	c.args(e, fd.Params, func(i int) Type {
		pw := c.substitute(fd.Params[i].T, m)
		if containsOpen(pw, tps) {
			return nil
		}
		return pw
	})
	for i, a := range e.Args {
		if i < len(fd.Params) {
			unify(fd.Params[i].T, a.Value.Type(), m)
		}
	}
	targs := make([]Type, len(tps))
	for i, p := range tps {
		t, ok := m[p]
		if !ok {
			c.errorf(e.P, "cannot tell type parameter %s of %s here; add a type annotation", p.Name, fd.Name)
			return TVoid
		}
		targs[i] = t
	}
	for i, a := range e.Args {
		if i < len(fd.Params) {
			c.expectType(a.Value.position(), c.substitute(fd.Params[i].T, m), a.Value.Type())
		}
	}
	result := c.substitute(fd.ResultT, m)
	open := false
	for _, t := range targs {
		if containsTypeParam(t) {
			open = true
		}
	}
	if open {
		// Inside a generic template the call stays generic; it is resolved
		// again when the template is instantiated.
		e.Func = fd
		return result
	}
	e.Func = c.funcInstance(fd, tps, targs)
	return result
}

func (c *Checker) funcInstance(fd *FuncDecl, tps []*TypeParamType, targs []Type) *FuncDecl {
	key := fd.Name + typeArgsMangle(targs)
	if inst, ok := c.funcInsts[key]; ok {
		return inst
	}
	inst := cloneFunc(fd)
	inst.Name = key
	env := envOf(tps, targs)
	saved := c.env
	c.env = env
	for _, p := range inst.Params {
		p.T = c.resolveType(p.Type)
	}
	inst.ResultT = TVoid
	if inst.Result != nil {
		inst.ResultT = c.resolveType(inst.Result)
	}
	c.env = saved
	c.funcInsts[key] = inst
	if len(c.funcInsts) > 10000 {
		c.errorf(fd.Pos, "too many instances of generic functions; %s may recurse with ever larger types", fd.Name)
		return inst
	}
	c.prog.Funcs = append(c.prog.Funcs, inst)
	c.pending = append(c.pending, pendingFunc{inst, env})
	return inst
}

func (c *Checker) argsLoose(e *CallExpr) {
	for _, a := range e.Args {
		c.expr(a.Value, nil)
	}
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

// args checks the arguments of a call. paramType gives the expected type of
// each argument, or nil when it is not known yet.
func (c *Checker) args(e *CallExpr, params []*Param, paramType func(int) Type) {
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
		want := paramType(i)
		t := c.expr(a.Value, want)
		if want != nil {
			c.expectType(a.Value.position(), want, t)
		}
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
		case *IsExpr:
			walk(e.Subject)
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
