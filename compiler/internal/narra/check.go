package narra

import (
	"fmt"
	"sort"
)

const (
	forRange forKind = iota
	forArray
	forMapKeys
	forMapEntries
)

type forKind int

// The prelude is written in Narra and checked once per program, as module "".
const preludeSrc = `
trait Equatable {
  function equals(a: Self, b: Self): bool
}

trait Hashable {
  function hash(value: Self): int
}

trait Ordered {
  function compare(a: Self, b: Self): Ordering
}

trait Show {
  function show(value: Self): string
}

enum Ordering derives Show, Ordered {
  Less
  Equal
  Greater
}

enum Option<T> derives Show, Ordered {
  Some(value: T)
  None
}

enum Result<T, E> derives Show, Ordered {
  Ok(value: T)
  Err(error: E)
}
`

type Program struct {
	Types  []Type // structs and enums in dependency order, so C defines inner types first
	Funcs  []*FuncDecl
	Tests  []*TestDecl
	Arrays []*ArrayType
	Boxes  []*BoxType
	Maps   []*MapType
	// Consts are the top-level constants, each after the constants it uses.
	Consts []*ConstDecl
	// Entries maps a binary name to its entry function. A binary whose entry
	// function does not exist has no key.
	Entries map[string]*FuncDecl
	// Files maps a module path to its file, for panic messages.
	Files map[string]string
	// Modules are the imported standard library modules, once each, in import order.
	// The emitter writes their Runtime C.
	Modules []*Module
	// Impls are the functions of `implement` blocks, by type and "Trait.function".
	Impls map[Type]map[string]*FuncDecl
	// FuncTypes are the function types the program uses, for C typedefs.
	FuncTypes []*FuncType
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
	building  bool       // fields are being resolved; see buildGeneric
	pending   []Type     // instances made while building, filled when it ends
	fields    []*Field   // struct template, with type parameters
	variants  []*Variant // enum template, with type parameters
	instances map[string]Type
	mod       *modScope // the declaring module
}

type pendingFunc struct {
	fd  *FuncDecl
	env map[string]Type
	mod *modScope
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
	boxes     map[string]*BoxType
	maps      map[string]*MapType
	mapKeys   []mapKeyCheck
	assigning Expr // the target of the assignment being checked (check_map.go)
	instTypes []Type
	localN    int
	isBinds   bool                   // an `is` pattern may bind names here
	tuples    map[string]*StructType // tuple types by element mangle (check_patterns.go)
	// The name tables above (types, generics, funcs, modules, imported)
	// belong to mod, the module being checked; enter swaps them (modules_check.go).
	mod     *modScope
	mods    map[string]*modScope // by module path; "" is the prelude
	funcMod map[*FuncDecl]*modScope
	pkg     *Package
	prelude *File // the parsed prelude, for DeclDump
	traits  *traitState
	closure *closureState // check_closure.go
	inTest  bool          // checking a test block, arrow functions inside it included (std_testing.go)
}

// Check checks one parsed file as the only module of package "main".
func Check(f *File) (*Program, error) {
	pkg := &Package{Name: "main", Binaries: []Binary{{Name: "main", Entry: "main.run"}}}
	pkg.Modules = []*SourceModule{newSourceModule("main", f)}
	return CheckPackage(pkg)
}

func (c *Checker) errorf(pos Pos, format string, args ...any) {
	file := ""
	if c.mod != nil {
		file = c.mod.file
	}
	c.report(Diagnostic{Code: codeFor(format), Pos: pos, File: file, Message: fmt.Sprintf(format, args...)})
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

// declareModule runs the declaration passes of the current module: imports,
// type names, fields, function signatures and generic templates. It returns
// the module's non-generic types.
func (c *Checker) declareModule(f *File) []Type {
	if len(f.Imports) > 0 && c.mod != nil && c.mod.importAt.Line == 0 {
		c.mod.importAt = Pos{f.Imports[0].Pos.Line, 1}
	}
	for _, imp := range f.Imports {
		if c.importPackageModule(imp) {
			continue
		}
		mod, ok := stdModules[imp.Path]
		if !ok {
			c.errorf(imp.Pos, "unknown module %q (this compiler supports %s)", imp.Path, moduleList())
			continue
		}
		c.useModule(mod)
		if imp.Path == "std/string" {
			// The type is visible only where its runtime C is emitted.
			c.types["StringBuilder"] = TStringBuilder
		}
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
	var own []*genericType
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *StructDecl:
			if len(d.TypeParams) > 0 {
				g := &genericType{name: d.Name, params: makeParams(d.TypeParams), sdecl: d, mod: c.mod}
				c.declareGeneric(d.Pos, g)
				own = append(own, g)
				continue
			}
			d.T = &StructType{Name: d.Name, Module: c.mod.path, Opaque: d.Opaque}
			c.declareType(d.Pos, d.Name, d.T)
			concrete = append(concrete, d.T)
		case *DistinctDecl:
			c.declareDistinct(d)
		case *EnumDecl:
			if len(d.TypeParams) > 0 {
				g := &genericType{name: d.Name, params: makeParams(d.TypeParams), edecl: d, mod: c.mod}
				c.declareGeneric(d.Pos, g)
				own = append(own, g)
				continue
			}
			d.T = &EnumType{Name: d.Name, Module: c.mod.path}
			c.declareType(d.Pos, d.Name, d.T)
			concrete = append(concrete, d.T)
		}
	}
	c.declareTraits(f)
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
		case *DistinctDecl:
			c.resolveDistinct(d)
		case *FuncDecl:
			if _, dup := c.funcs[d.Name]; dup {
				c.errorf(d.Pos, "function %q is declared twice", d.Name)
			}
			c.funcs[d.Name] = d
			d.Module = c.mod.path
			c.funcMod[d] = c.mod
			c.env = nil
			if len(d.TypeParams) > 0 {
				tps := makeParams(d.TypeParams)
				c.funcTP[d] = tps
				c.funcBounds(d, tps)
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
		case *TestDecl:
			d.Module = c.mod.path
			c.prog.Tests = append(c.prog.Tests, d)
		}
	}
	c.declareConsts(f)
	for _, g := range own {
		c.buildGeneric(g)
	}
	c.resolveTraits(f)
	// Defaults are checked after every field type is known.
	for _, d := range f.Decls {
		if sd, ok := d.(*StructDecl); ok && sd.T != nil {
			c.checkDefaults(sd.T.Fields)
		}
	}
	return concrete
}

// checkBodies checks the function and test bodies of the current module.
func (c *Checker) checkBodies(f *File) {
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
		case *ConstDecl:
			c.constType(d, d.Pos)
		case *ImplDecl:
			c.implBodies(d)
		case *TestDecl:
			c.fn = &FuncDecl{Name: "test", ResultT: TVoid}
			c.env = nil
			c.scope = &scope{names: map[string]*Local{}}
			c.inTest = true
			c.block(d.Body)
			c.inTest = false
		}
	}
}

// checkPending checks instances of generic functions as they are
// discovered, each in the module that declares the generic function.
func (c *Checker) checkPending() {
	for len(c.pending) > 0 {
		p := c.pending[0]
		c.pending = c.pending[1:]
		saved := c.enter(p.mod)
		c.funcBody(p.fd, p.env)
		c.enter(saved)
	}
}

// finish orders the program's types and arrays for the emitter.
func (c *Checker) finish(concrete []Type) {
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
	var bkeys []string
	for k, b := range c.boxes {
		if !containsTypeParam(b) {
			bkeys = append(bkeys, k)
		}
	}
	sort.Strings(bkeys)
	for _, k := range bkeys {
		c.prog.Boxes = append(c.prog.Boxes, c.boxes[k])
	}
	var mkeys []string
	for k, m := range c.maps {
		if !containsTypeParam(m) {
			mkeys = append(mkeys, k)
		}
	}
	sort.Strings(mkeys)
	for _, k := range mkeys {
		c.prog.Maps = append(c.prog.Maps, c.maps[k])
	}
	c.finishFuncTypes()
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
