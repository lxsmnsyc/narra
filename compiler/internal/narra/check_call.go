package narra

import (
	"sort"
	"strconv"
	"strings"
)

func (c *Checker) isEnumName(name string) bool {
	if _, ok := c.types[name].(*EnumType); ok {
		return true
	}
	g, ok := c.generics[name]
	return ok && g.edecl != nil
}

// preludeCtors are the constructors that stay bare: prelude names of Option
// and Result (D71).
var preludeCtors = map[string]string{"Some": "Option", "None": "Option", "Ok": "Result", "Err": "Result"}

// ctor resolves an enum constructor and checks its arguments. enumName is the
// enum as written in the call, as in Shape.Circle or Geo.Shape.Circle. A bare
// constructor (enumName empty) is an error whose fix qualifies it, except for
// Some, None, Ok and Err. It returns nil when no enum has a constructor of
// that name.
func (c *Checker) ctor(pos Pos, enumName, name string, args []*Arg, want Type) (Type, *Variant) {
	if enumName == "" {
		if enum, ok := preludeCtors[name]; ok {
			return c.ctorOf(pos, c.enumByName(enum), name, args, want)
		}
		owner, owners := c.bareOwner(name, want)
		if owner == "" {
			return nil, nil
		}
		if len(owners) > 1 {
			c.errorf(pos, "constructor %s needs its enum; %s all have it, for example %s.%s", name, strings.Join(owners, ", "), owner, name)
		} else {
			c.errorfFix(pos, c.qualifyFix(pos, owner), "constructor %s needs its enum; write %s.%s", name, owner, name)
		}
		return c.ctorOf(pos, c.enumByName(owner), name, args, want)
	}
	return c.ctorOf(pos, c.enumByName(enumName), name, args, want)
}

// enumRef is an enum found by name: a plain enum or a generic template.
type enumRef struct {
	name string
	et   *EnumType
	g    *genericType
}

// enumByName finds the enum a name refers to in the current module. A dotted
// name Alias.Enum looks in the module the namespace import Alias names.
func (c *Checker) enumByName(name string) enumRef {
	types, generics := c.types, c.generics
	n := name
	if alias, rest, ok := strings.Cut(name, "."); ok {
		s := c.mods[c.modules[alias]]
		if s == nil || s.src == nil {
			return enumRef{name: name}
		}
		if _, ok := s.src.Exports[rest].(*EnumDecl); !ok {
			return enumRef{name: name}
		}
		types, generics, n = s.types, s.generics, rest
	}
	if et, ok := types[n].(*EnumType); ok {
		return enumRef{name: name, et: et}
	}
	if g := generics[n]; g != nil && g.edecl != nil {
		return enumRef{name: name, g: g}
	}
	return enumRef{name: name}
}

// bareOwner picks the enum a bare constructor name should be written with:
// the expected type's enum when it has the constructor, else the first enum
// that does (declared, imported, then prelude). It returns "" when none does.
// The second result lists the enums of that rank when more than one has it.
func (c *Checker) bareOwner(name string, want Type) (string, []string) {
	if et, ok := want.(*EnumType); ok && et.Variant(name) != nil {
		if n, ok := c.enumNameFor(et); ok {
			return n, nil
		}
	}
	owners := c.ctorOwners("", name)
	if len(owners) == 0 {
		return "", nil
	}
	return owners[0], owners
}

// qualifyFix writes the enum name before the bare constructor at pos.
func (c *Checker) qualifyFix(pos Pos, enum string) []Fix {
	return []Fix{{Title: "qualify the constructor with " + enum, Edits: []Edit{{Start: pos, End: pos, Text: enum + "."}}}}
}

// ctorOf builds constructor name of an enum found by enumByName.
func (c *Checker) ctorOf(pos Pos, en enumRef, name string, args []*Arg, want Type) (Type, *Variant) {
	// A constructor stores its arguments.
	for _, a := range args {
		c.noMutableEscape(a.Value)
	}
	if en.et != nil {
		v := en.et.Variant(name)
		if v == nil {
			return c.noCtor(pos, en, name, args)
		}
		c.ctorArgs(pos, v, args)
		return en.et, v
	}
	g := en.g
	if g == nil {
		return c.noCtor(pos, en, name, args)
	}
	c.buildGeneric(g)
	for _, tv := range g.variants {
		if tv.Name != name {
			continue
		}
		m := map[*TypeParamType]Type{}
		if et, ok := want.(*EnumType); ok && et.Name == g.name && et.Module == g.mod.path {
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
	return c.noCtor(pos, en, name, args)
}

func (c *Checker) noCtor(pos Pos, en enumRef, name string, args []*Arg) (Type, *Variant) {
	c.errorf(pos, "enum %s has no constructor %q", en.name, name)
	for _, a := range args {
		c.expr(a.Value, nil)
	}
	return TVoid, nil
}

// ctorOwners lists the enums, by the name the current module uses for them,
// that have a constructor with this name. enumName, when set, restricts the
// search to one enum. Enums are ranked: those the module declares, then
// those it imports by name, then the prelude. Only the first rank with a
// match is returned, so a user `Less` hides `Ordering.Less`. Within a rank
// the order is fixed (declaration order, import order, then by name), so the
// result never depends on map order. More than one result is ambiguous.
func (c *Checker) ctorOwners(enumName, name string) []string {
	var ranks [3][]string
	for _, n := range c.visibleEnums(enumName) {
		module, has := "", false
		if et, ok := c.types[n].(*EnumType); ok {
			module, has = et.Module, et.Variant(name) != nil
		} else if g := c.generics[n]; g != nil && g.edecl != nil {
			c.buildGeneric(g)
			module = g.mod.path
			for _, v := range g.variants {
				if v.Name == name {
					has = true
				}
			}
		}
		if !has {
			continue
		}
		switch module {
		case c.mod.path:
			ranks[0] = append(ranks[0], n)
		case "":
			ranks[2] = append(ranks[2], n)
		default:
			ranks[1] = append(ranks[1], n)
		}
	}
	for _, r := range ranks {
		if len(r) > 0 {
			return r
		}
	}
	return nil
}

// visibleEnums lists the names of the enums visible in the current module
// without duplicates: its declarations in order, its imports in order, then
// the rest (the prelude) sorted by name.
func (c *Checker) visibleEnums(only string) []string {
	if only != "" {
		return []string{only}
	}
	var names []string
	seen := map[string]bool{}
	add := func(n string) {
		if !seen[n] && c.isEnumName(n) {
			seen[n] = true
			names = append(names, n)
		}
	}
	if c.mod.src != nil {
		for _, f := range c.mod.src.Files {
			for _, d := range f.Decls {
				if ed, ok := d.(*EnumDecl); ok {
					add(ed.Name)
				}
			}
		}
		for _, f := range c.mod.src.Files {
			for _, imp := range f.Imports {
				for _, n := range imp.Names {
					add(n.Local)
				}
			}
		}
	}
	var rest []string
	for n := range c.types {
		rest = append(rest, n)
	}
	for n := range c.generics {
		rest = append(rest, n)
	}
	sort.Strings(rest)
	for _, n := range rest {
		add(n)
	}
	return names
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
	case *GeneratorType:
		return containsOpen(t.Elem, params)
	case *BoxType:
		return containsOpen(t.Elem, params)
	case *MapType:
		return containsOpen(t.Key, params) || containsOpen(t.Val, params)
	case *FuncType:
		for _, p := range t.Params {
			if containsOpen(p, params) {
				return true
			}
		}
		return containsOpen(t.Result, params)
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

func (c *Checker) call(e *CallExpr, want Type) Type {
	switch callee := e.Callee.(type) {
	case *Ident:
		if c.lookup(callee.Name) != nil || c.mod.consts[callee.Name] != nil {
			return c.valueCall(e)
		}
		if b, ok := preludeFuncs[callee.Name]; ok {
			return c.builtinCall(e, b, want)
		}
		if b, ok := c.imported[callee.Name]; ok {
			return c.builtinCall(e, b, want)
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
			if tr := c.lookupTrait(id.Name); tr != nil {
				return c.traitCall(e, tr, callee.Name, want)
			}
			if path, ok := c.modules[id.Name]; ok {
				if s := c.mods[path]; s != nil && s.src != nil {
					return c.moduleCall(e, s, callee, want)
				}
				b := stdModules[path].Funcs[callee.Name]
				if b == nil {
					c.errorf(callee.P, "module %q has no function %q", path, callee.Name)
					c.argsLoose(e)
					return TVoid
				}
				return c.builtinCall(e, b, want)
			}
			// Array, Box, Map and Range are prelude types; their functions
			// need the import. A type the program declares named Box or Map
			// hides the builtin.
			if c.missingImport(id, callee.Name) {
				c.argsLoose(e)
				return TVoid
			}
			if d := c.distinctOf(callee.Target); d != nil {
				return c.distinctCall(e, d, callee)
			}
			if c.isEnumName(id.Name) {
				t, v := c.ctor(callee.P, id.Name, callee.Name, e.Args, want)
				e.Ctor = v
				return t
			}
			if _, ok := c.types[id.Name]; !ok && startsUpper(id.Name) {
				var fixes []Fix
				if path, ok := stdAliasPaths[id.Name]; ok {
					fixes = c.importFix(id.Name, path)
				}
				c.errorfFix(id.P, fixes, "unknown module %q; import it first, for example `import * as %s from \"std/...\"`", id.Name, id.Name)
				c.argsLoose(e)
				return TVoid
			}
		}
		if en := c.moduleEnumName(callee.Target); en != "" {
			t, v := c.ctor(callee.P, en, callee.Name, e.Args, want)
			e.Ctor = v
			return t
		}
		t := c.expr(callee.Target, nil)
		if ft := fieldFunc(callee, t); ft != nil {
			return c.valueCallOf(e, ft)
		}
		alias := c.aliasFor("std/array")
		fixes := pushFix(e, callee, t, alias)
		if alias == "" && len(fixes) > 0 {
			fixes[0].Edits = append(fixes[0].Edits, c.importFix("Arrays", "std/array")[0].Edits...)
		}
		c.errorfFix(callee.P, fixes, "there are no method calls; call a module function such as Module.%s(value, ...) (value has type %s)", callee.Name, t)
		c.argsLoose(e)
		return TVoid
	}
	return c.valueCall(e)
}

// shapeMismatch reports whether a concrete type cannot match a pattern that
// still holds open type parameters. Only the positions that hold an open
// parameter count; the rest is left to expectType.
func shapeMismatch(pat, act Type, tps []*TypeParamType) bool {
	if !containsOpen(pat, tps) {
		return false
	}
	switch p := pat.(type) {
	case *ArrayType:
		a, ok := act.(*ArrayType)
		return !ok || shapeMismatch(p.Elem, a.Elem, tps)
	case *BoxType:
		a, ok := act.(*BoxType)
		return !ok || shapeMismatch(p.Elem, a.Elem, tps)
	case *GeneratorType:
		a, ok := act.(*GeneratorType)
		return !ok || shapeMismatch(p.Elem, a.Elem, tps)
	case *MapType:
		a, ok := act.(*MapType)
		return !ok || shapeMismatch(p.Key, a.Key, tps) || shapeMismatch(p.Val, a.Val, tps)
	case *FuncType:
		a, ok := act.(*FuncType)
		if !ok {
			return true
		}
		if len(a.Params) != len(p.Params) {
			// An arrow function with the wrong arity already has its own error.
			return false
		}
		for i := range p.Params {
			if shapeMismatch(p.Params[i], a.Params[i], tps) {
				return true
			}
		}
		return shapeMismatch(p.Result, a.Result, tps)
	case *StructType:
		a, ok := act.(*StructType)
		if !ok || a.Name != p.Name || a.Module != p.Module || len(a.Args) != len(p.Args) {
			return true
		}
		for i := range p.Args {
			if shapeMismatch(p.Args[i], a.Args[i], tps) {
				return true
			}
		}
	case *EnumType:
		a, ok := act.(*EnumType)
		if !ok || a.Name != p.Name || a.Module != p.Module || len(a.Args) != len(p.Args) {
			return true
		}
		for i := range p.Args {
			if shapeMismatch(p.Args[i], a.Args[i], tps) {
				return true
			}
		}
	}
	return false
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
			c.hintArrow(e, i, fd, tps, m)
			return nil
		}
		return pw
	})
	for i, a := range e.Args {
		if i < len(fd.Params) {
			unify(fd.Params[i].T, a.Value.Type(), m)
		}
	}
	// An argument whose shape cannot fit its parameter is a mismatch, not a
	// missing type parameter.
	mismatch := false
	for i, a := range e.Args {
		if i >= len(fd.Params) {
			break
		}
		got := a.Value.Type()
		pw := c.substitute(fd.Params[i].T, m)
		if got != nil && got != TVoid && containsOpen(pw, tps) && shapeMismatch(pw, got, tps) {
			c.errorf(a.Value.position(), "expected %s, found %s; the argument does not fit the generic parameter", typeName(pw, got), typeName(got, pw))
			mismatch = true
		}
	}
	targs := make([]Type, len(tps))
	for i, p := range tps {
		t, ok := m[p]
		if mismatch && !ok {
			// Report the mismatch only; give back the result when it is known.
			result := c.substitute(fd.ResultT, m)
			if containsOpen(result, tps) {
				return TVoid
			}
			return result
		}
		if !ok {
			c.errorf(e.P, "cannot tell type parameter %s of %s here; add a type annotation", p.Name, fd.Name)
			return TVoid
		}
		targs[i] = t
	}
	c.checkBounds(e.P, fd, tps, targs)
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
	name := instanceName(fd.Module, fd.Name, targs)
	key := fd.Module + "." + name
	if inst, ok := c.funcInsts[key]; ok {
		return inst
	}
	inst := cloneFunc(fd)
	inst.Name = name
	inst.Module = fd.Module
	// Types in the signature are names in the declaring module.
	mod := c.funcMod[fd]
	c.funcMod[inst] = mod
	savedMod := c.enter(mod)
	defer c.enter(savedMod)
	env := envOf(tps, targs)
	saved := c.env
	c.env = env
	c.resolveSignature(inst, fd)
	c.env = saved
	c.funcInsts[key] = inst
	if len(c.funcInsts) > 10000 {
		c.errorf(fd.Pos, "too many instances of generic functions; %s may recurse with ever larger types", fd.Name)
		return inst
	}
	c.prog.Funcs = append(c.prog.Funcs, inst)
	c.pending = append(c.pending, pendingFunc{inst, env, mod})
	return inst
}

func (c *Checker) argsLoose(e *CallExpr) {
	for _, a := range e.Args {
		c.expr(a.Value, nil)
	}
}

func (c *Checker) builtinCall(e *CallExpr, b *Builtin, want Type) Type {
	e.Builtin = b.Name
	if b.Check != nil {
		return b.Check(c, e, want)
	}
	if len(e.Args) != len(b.Params) {
		c.errorf(e.P, "%s takes %d arguments, found %d", b.Name, len(b.Params), len(e.Args))
	}
	var mutables []int
	for i, a := range e.Args {
		mutable := i < len(b.Mutable) && b.Mutable[i]
		var want Type
		if i < len(b.Params) {
			want = b.Params[i]
		}
		// Check the argument first: checkMutableArg reads the resolved local.
		got := c.expr(a.Value, want)
		switch {
		case mutable && !a.Mutable:
			c.errorf(e.P, "%s changes its argument; write `mutable` before it", b.Name)
		case !mutable && a.Mutable:
			c.errorf(e.P, "%s does not change its arguments; remove `mutable`", b.Name)
		case mutable:
			c.checkMutableArg(a.Value)
			mutables = append(mutables, i)
		}
		c.expectType(a.Value.position(), want, got)
	}
	c.checkExclusive(e, mutables)
	return b.Result
}

func (c *Checker) checkMutableArg(v Expr) {
	if c.constWrite(v.position(), v) {
		return
	}
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
		c.funcArg(p, a.Value)
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
	c.checkBorrowed(e)
}

// checkBorrowed rejects a call with a stack closure that assigns a captured
// variable when another argument uses that variable, or another closure
// argument captures it. The closure borrows the variable for the call, and
// constant parameters are passed by pointer (D50), so the callee could read
// what the closure frees (D77).
func (c *Checker) checkBorrowed(e *CallExpr) {
	for i, a := range e.Args {
		arrow, ok := a.Value.(*ArrowExpr)
		if !ok || arrow.Closure == nil || !arrow.Closure.Stack {
			continue
		}
		names := map[string]bool{}
		assignedInExpr(arrow, names)
		for _, l := range arrow.Closure.Captures {
			if !l.Mutable || !names[l.Name] {
				continue
			}
			for j, other := range e.Args {
				if j != i && usesLocal(other.Value, l) {
					c.errorf(other.Value.position(), "%q is assigned by a closure passed to this call, so no other argument can use it", l.Name)
				}
			}
		}
	}
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
		case *ArrowExpr:
			// A closure uses what it captures.
			if e.Closure != nil {
				for _, x := range e.Closure.Captures {
					if x == l {
						found = true
					}
				}
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
		case *MapLit:
			for i := range e.Keys {
				walk(e.Keys[i])
				walk(e.Vals[i])
			}
		case *MatchExpr:
			walk(e.Subject)
			for _, arm := range e.Arms {
				walk(arm.Guard)
				walk(arm.Body)
				eachBlockExpr(arm.Block, walk)
			}
		case *BlockExpr:
			eachBlockExpr(e.Block, walk)
		}
	}
	walk(e)
	return found
}

// missingImport reports Array.f, Box.f, Map.f or Range.f used without the
// import of the module that holds f, and returns true when it did.
func (c *Checker) missingImport(id *Ident, name string) bool {
	path, ok := namespaceModules[id.Name]
	if _, imported := c.modules[id.Name]; imported || !ok || (id.Name == "Box" && c.userBox()) || (id.Name == "Map" && c.userMap()) {
		return false
	}
	rename := Edit{Start: id.P, End: Pos{id.P.Line, id.P.Col + len(id.Name)}, Text: pluralAlias[id.Name]}
	if have := c.aliasFor(path); have != "" {
		rename.Text = have
		c.errorfFix(id.P, []Fix{{Title: "call " + have + "." + name, Edits: []Edit{rename}}}, "%s.%s names a type; this module imports %q as %s", id.Name, name, path, have)
		return true
	}
	fix := c.importFix(pluralAlias[id.Name], path)
	fix[0].Edits = append(fix[0].Edits, rename)
	c.errorfFix(id.P, fix, "%s.%s needs an import; add `import * as %s from %q`", id.Name, name, pluralAlias[id.Name], path)
	return true
}

// moduleEnumName returns "Alias.Enum" when x is that expression, Alias being a
// namespace import and Enum an enum the module exports. It returns "" for
// anything else, so `Geo.Shape.Circle` can name a constructor.
func (c *Checker) moduleEnumName(x Expr) string {
	fe, ok := x.(*FieldExpr)
	if !ok {
		return ""
	}
	id, ok := fe.Target.(*Ident)
	if !ok || c.lookup(id.Name) != nil || c.mod.consts[id.Name] != nil {
		return ""
	}
	name := id.Name + "." + fe.Name
	if en := c.enumByName(name); en.et != nil || en.g != nil {
		id.setType(&ModuleType{Path: c.modules[id.Name]})
		return name
	}
	return ""
}

// localUse is one use of a local in a statement's expression: a read, an
// assignment target, or a `mutable` argument.
type localUse struct {
	local *Local
	pos   Pos
	kind  int       // useRead, useAssign or useMutable
	call  *CallExpr // for useMutable: the call it is an argument of
	chain []*CallExpr
	hoist bool // the call can move to its own statement
	arms  []armStep
	order int
	// calleeOf is the call whose callee holds the use, when it does.
	calleeOf *CallExpr
}

// armStep says a use sits in one arm of a match, or on one side of && or ||.
// Uses in different arms of the same match never run together, and the left
// side of && and || is finished before the right side starts.
type armStep struct {
	key     any // the match, the && or || expression, or the block
	arm     int
	isMatch bool // an arm of a match or a block expression
	calls   int  // how many calls enclosed the step when it began
}

const (
	useRead = iota
	useAssign
	useMutable
)

// useCollector walks the expressions of one statement in evaluation order.
type useCollector struct {
	c    *Checker
	uses []*localUse
	arms []armStep
}

func (u *useCollector) add(l *Local, pos Pos, kind int, call *CallExpr, chain []*CallExpr, hoist bool) {
	if l != nil {
		u.uses = append(u.uses, &localUse{local: l, pos: pos, kind: kind, call: call, chain: append([]*CallExpr(nil), chain...), hoist: hoist, arms: append([]armStep(nil), u.arms...), order: len(u.uses)})
	}
}

// root records the local at the root of an assignable expression and walks
// the index expressions on the way.
func (u *useCollector) root(x Expr, kind int, call *CallExpr, chain []*CallExpr, hoist bool) {
	switch x := x.(type) {
	case *Ident:
		u.add(x.Local, x.P, kind, call, chain, hoist)
	case *FieldExpr:
		if u.c.lvalueRoot(x) != nil {
			u.root(x.Target, kind, call, chain, hoist)
		} else {
			u.walk(x.Target, chain, hoist)
		}
	case *IndexExpr:
		u.root(x.Target, kind, call, chain, hoist)
		u.walk(x.Index, chain, hoist)
	default:
		u.walk(x, chain, hoist)
	}
}

func (u *useCollector) walk(x Expr, chain []*CallExpr, hoist bool) {
	switch x := x.(type) {
	case *Ident:
		u.add(x.Local, x.P, useRead, nil, chain, hoist)
	case *FieldExpr:
		u.walk(x.Target, chain, hoist)
	case *IndexExpr:
		u.walk(x.Target, chain, hoist)
		u.walk(x.Index, chain, hoist)
	case *IsExpr:
		u.walk(x.Subject, chain, hoist)
	case *UnaryExpr:
		u.walk(x.Operand, chain, hoist)
	case *BinaryExpr:
		short := x.Op == "&&" || x.Op == "||"
		if short {
			u.arms = append(u.arms, armStep{key: x, arm: 0, calls: len(chain)})
		}
		u.walk(x.Left, chain, hoist)
		if short {
			u.arms[len(u.arms)-1].arm = 1
		}
		// The right side of && and || may not run, so a call there stays put.
		u.walk(x.Right, chain, hoist && !short)
		if short {
			u.arms = u.arms[:len(u.arms)-1]
		}
	case *CallExpr:
		inner := append(append([]*CallExpr(nil), chain...), x)
		// A mutable argument that is also the function called is reported
		// by the closure check, so the callee's uses are marked.
		first := len(u.uses)
		u.walk(x.Callee, chain, hoist)
		for _, use := range u.uses[first:] {
			if use.calleeOf == nil {
				use.calleeOf = x
			}
		}
		for _, a := range x.Args {
			if a.Mutable {
				u.root(a.Value, useMutable, x, inner, hoist)
			} else {
				u.walk(a.Value, inner, hoist)
			}
		}
	case *TemplateLit:
		for _, part := range x.Parts {
			u.walk(part, chain, hoist)
		}
	case *StructLit:
		for _, f := range x.Fields {
			u.walk(f.Value, chain, hoist)
		}
		if x.Spread != nil {
			u.walk(x.Spread, chain, hoist)
		}
	case *ArrayLit:
		for _, el := range x.Elems {
			u.walk(el, chain, hoist)
		}
	case *MapLit:
		for i := range x.Keys {
			u.walk(x.Keys[i], chain, hoist)
			u.walk(x.Vals[i], chain, hoist)
		}
	case *BlockExpr:
		// A block expression is one value. checkExclusive looks inside it
		// too, but its statements run one after the other.
		u.arms = append(u.arms, armStep{key: x, arm: 0, isMatch: true, calls: len(chain)})
		u.block(x.Block, chain)
		u.arms = u.arms[:len(u.arms)-1]
	case *MatchExpr:
		// The subject is finished before the first arm starts.
		u.arms = append(u.arms, armStep{key: x, arm: -1, calls: len(chain)})
		u.walk(x.Subject, chain, hoist)
		u.arms = u.arms[:len(u.arms)-1]
		for i, arm := range x.Arms {
			u.arms = append(u.arms, armStep{key: x, arm: i, isMatch: true, calls: len(chain)})
			if arm.Guard != nil {
				u.walk(arm.Guard, chain, false)
			}
			if arm.Body != nil {
				u.walk(arm.Body, chain, false)
			}
			if arm.Block != nil {
				u.block(arm.Block, chain)
			}
			u.arms = u.arms[:len(u.arms)-1]
		}
	}
}

// block walks the statements of a block inside an expression, such as a match
// arm. The statements run one after the other, so each has a step of its own.
func (u *useCollector) block(b *Block, chain []*CallExpr) {
	for i, st := range b.Stmts {
		u.arms = append(u.arms, armStep{key: b, arm: i, calls: len(chain)})
		u.stmt(st, chain)
		u.arms = u.arms[:len(u.arms)-1]
	}
}

func (u *useCollector) stmt(st Stmt, chain []*CallExpr) {
	switch st := st.(type) {
	case *Block:
		u.block(st, chain)
	case *VarStmt:
		if st.Value != nil {
			u.walk(st.Value, chain, false)
		}
	case *AssignStmt:
		u.root(st.Target, useAssign, nil, chain, false)
		u.walk(st.Value, chain, false)
	case *ExprStmt:
		u.walk(st.Expr, chain, false)
	case *ReturnStmt:
		if st.Value != nil {
			u.walk(st.Value, chain, false)
		}
	case *IfStmt:
		// The condition is finished before the branches start.
		u.arms = append(u.arms, armStep{key: st, arm: 0, calls: len(chain)})
		u.walk(st.Cond, chain, false)
		u.arms[len(u.arms)-1].arm = 1
		u.block(st.Then, chain)
		// The two branches never run together.
		u.arms[len(u.arms)-1].arm = 2
		if st.Else != nil {
			u.stmt(st.Else, chain)
		}
		u.arms = u.arms[:len(u.arms)-1]
	case *WhileStmt:
		u.arms = append(u.arms, armStep{key: st, arm: 0, calls: len(chain)})
		u.walk(st.Cond, chain, false)
		u.arms[len(u.arms)-1].arm = 1
		u.block(st.Body, chain)
		u.arms = u.arms[:len(u.arms)-1]
	case *ForStmt:
		u.arms = append(u.arms, armStep{key: st, arm: 0, calls: len(chain)})
		u.walk(st.Iter, chain, false)
		u.arms[len(u.arms)-1].arm = 1
		u.block(st.Body, chain)
		u.arms = u.arms[:len(u.arms)-1]
	case *YieldStmt:
		u.walk(st.Value, chain, false)
	}
}

// checkMutableConflicts rejects a statement whose expressions pass a local as
// `mutable` and also use it anywhere else, at any depth. The emitter runs a
// `mutable` call before the rest of its statement, so the other use would
// see the new value. Uses inside the arguments of the call itself belong to
// checkExclusive.
func (c *Checker) checkMutableConflicts(s Stmt) {
	u := &useCollector{c: c}
	fixable := true
	switch s := s.(type) {
	case *VarStmt:
		if s.Value != nil {
			u.walk(s.Value, nil, true)
		}
	case *AssignStmt:
		u.root(s.Target, useAssign, nil, nil, true)
		u.walk(s.Value, nil, true)
	case *ExprStmt:
		u.walk(s.Expr, nil, true)
	case *ReturnStmt:
		if s.Value != nil {
			u.walk(s.Value, nil, true)
		}
	case *IfStmt:
		u.walk(s.Cond, nil, true)
		fixable = false
	case *WhileStmt:
		u.walk(s.Cond, nil, false)
		fixable = false
	case *ForStmt:
		u.walk(s.Iter, nil, true)
		fixable = false
	case *YieldStmt:
		u.walk(s.Value, nil, true)
		fixable = false
	default:
		return
	}
	reported := map[*Local]bool{}
	for _, v := range u.uses {
		if reported[v.local] {
			continue
		}
		for _, m := range u.uses {
			if m == v || m.local != v.local || m.kind != useMutable || (v.kind == useMutable && m.order > v.order) {
				continue
			}
			if (v.calleeOf != nil && inChain(m.chain, v.calleeOf)) || differentArms(v.arms, m.arms) || insideCall(v, m.call) || (v.kind == useMutable && insideCall(m, v.call)) {
				continue
			}
			reported[v.local] = true
			var fixes []Fix
			if fixable && m.hoist {
				fixes = c.hoistFix(s, m.call)
			}
			c.errorfFix(v.pos, fixes, "%q is passed as mutable in one call, so the rest of the expression cannot use it", v.local.Name)
			break
		}
	}
}

func inChain(chain []*CallExpr, call *CallExpr) bool {
	for _, x := range chain {
		if x == call {
			return true
		}
	}
	return false
}

// callStart is where a call's text begins: the start of its callee.
func callStart(x Expr) Pos {
	switch x := x.(type) {
	case *FieldExpr:
		return callStart(x.Target)
	case *IndexExpr:
		return callStart(x.Target)
	case *CallExpr:
		return callStart(x.Callee)
	}
	return x.position()
}

// hoistFix binds a `mutable` call to a fresh const on the line before its
// statement and puts the name in its place. It needs the source text, a
// statement that starts its line and a call that returns a value.
func (c *Checker) hoistFix(s Stmt, call *CallExpr) []Fix {
	if c.mod == nil || c.mod.src == nil || len(c.mod.src.Files) != 1 || call.End.Line == 0 || call.Type() == TVoid {
		return nil
	}
	start, end := callStart(call), call.End
	if start.Line != end.Line || start.Col == 1 {
		return nil
	}
	var stmtPos Pos
	switch s := s.(type) {
	case *VarStmt:
		stmtPos = s.Pos
	case *AssignStmt:
		stmtPos = s.Pos
	case *ExprStmt:
		stmtPos = s.Pos
	case *ReturnStmt:
		stmtPos = s.Pos
	}
	lines := strings.Split(c.mod.src.Files[0].Src, "\n")
	if stmtPos.Line < 1 || stmtPos.Line > len(lines) || start.Line > len(lines) {
		return nil
	}
	stmtLine := []rune(lines[stmtPos.Line-1])
	if stmtPos.Col-1 > len(stmtLine) || strings.TrimSpace(string(stmtLine[:stmtPos.Col-1])) != "" {
		return nil
	}
	indent := string(stmtLine[:stmtPos.Col-1])
	callLine := []rune(lines[start.Line-1])
	if end.Col-1 > len(callLine) {
		return nil
	}
	text := string(callLine[start.Col-1 : end.Col-1])
	base := "result"
	switch callee := call.Callee.(type) {
	case *Ident:
		base = callee.Name + "Result"
	case *FieldExpr:
		base = callee.Name + "Result"
	}
	name := base
	for n := 2; c.lookup(name) != nil; n++ {
		name = base + strconv.Itoa(n)
	}
	return []Fix{{Title: "bind the call to a const before the statement", Edits: []Edit{
		{Start: Pos{stmtPos.Line, 1}, End: Pos{stmtPos.Line, 1}, Text: indent + "const " + name + " = " + text + "\n"},
		{Start: start, End: end, Text: name},
	}}}
}

// insideCall tells whether a use is in the arguments of a call and not in a
// match arm there: checkExclusive reports those. A match arm in the arguments
// is not covered by it.
func insideCall(use *localUse, call *CallExpr) bool {
	for i, x := range use.chain {
		if x != call {
			continue
		}
		for _, step := range use.arms {
			if step.isMatch && step.calls > i {
				return false
			}
		}
		return true
	}
	return false
}

func differentArms(a, b []armStep) bool {
	for _, x := range a {
		for _, y := range b {
			if x.key == y.key && x.arm != y.arm {
				return true
			}
		}
	}
	return false
}
