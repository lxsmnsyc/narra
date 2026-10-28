package narra

import (
	"fmt"
	"sort"
)

const (
	forRange forKind = iota
	forArray
)

type forKind int

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
	// Modules are the imported standard library modules, once each, in import order.
	// The emitter writes their Runtime C.
	Modules []*Module
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
	errs      []Diagnostic
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
		return nil, Diagnostics(c.errs)
	}
	return c.prog, nil
}

func (c *Checker) errorf(pos Pos, format string, args ...any) {
	c.report(Diagnostic{Code: "E0000", Pos: pos, Message: fmt.Sprintf(format, args...)})
}

// useModule records an imported standard library module.
func (c *Checker) useModule(m *Module) {
	for _, have := range c.prog.Modules {
		if have == m {
			return
		}
	}
	c.prog.Modules = append(c.prog.Modules, m)
}

func (c *Checker) run(f *File) {
	for _, imp := range f.Imports {
		mod, ok := stdModules[imp.Path]
		if !ok {
			c.errorf(imp.Pos, "unknown module %q (this compiler supports %s)", imp.Path, moduleList())
			continue
		}
		c.useModule(mod)
		if imp.Alias != "" {
			c.modules[imp.Alias] = imp.Path
		}
		for _, n := range imp.Names {
			b, ok := mod.Funcs[n]
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

func (c *Checker) pop() { c.scope = c.scope.parent }
