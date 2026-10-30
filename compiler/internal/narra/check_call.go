package narra

import (
	"sort"
	"strings"
)

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
	if et, ok := want.(*EnumType); ok && (enumName == "" || c.namesEnum(enumName, et)) {
		if v := et.Variant(name); v != nil {
			c.ctorArgs(pos, v, args)
			return et, v
		}
	}
	owners := c.ctorOwners(enumName, name)
	if len(owners) > 1 {
		c.errorf(pos, "constructor %s is ambiguous: %s all have it; write %s.%s", name, strings.Join(owners, ", "), owners[0], name)
		for _, a := range args {
			c.expr(a.Value, nil)
		}
		return TVoid, nil
	}
	for _, gn := range owners {
		if et, ok := c.types[gn].(*EnumType); ok {
			v := et.Variant(name)
			c.ctorArgs(pos, v, args)
			return et, v
		}
		g := c.generics[gn]
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
	}
	if enumName != "" {
		c.errorf(pos, "enum %s has no constructor %q", enumName, name)
		return TVoid, nil
	}
	return nil, nil
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
					add(n)
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
	case *BoxType:
		return containsOpen(t.Elem, params)
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
		if c.lookup(callee.Name) != nil {
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
			if namespaces[id.Name] {
				name := id.Name + "." + callee.Name
				b := namespaceFuncs[name]
				if b == nil {
					c.errorf(e.P, "unknown function %s", name)
					c.argsLoose(e)
					return TVoid
				}
				return c.builtinCall(e, b, want)
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
		if ft := fieldFunc(callee, t); ft != nil {
			return c.valueCallOf(e, ft)
		}
		c.errorf(callee.P, "there are no method calls; call a module function such as Module.%s(value, ...) (value has type %s)", callee.Name, t)
		c.argsLoose(e)
		return TVoid
	}
	return c.valueCall(e)
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
	targs := make([]Type, len(tps))
	for i, p := range tps {
		t, ok := m[p]
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
		switch {
		case mutable && !a.Mutable:
			c.errorf(e.P, "%s changes its argument; write `mutable` before it", b.Name)
		case !mutable && a.Mutable:
			c.errorf(e.P, "%s does not change its arguments; remove `mutable`", b.Name)
		case mutable:
			c.checkMutableArg(a.Value)
			mutables = append(mutables, i)
		}
		var want Type
		if i < len(b.Params) {
			want = b.Params[i]
		}
		c.expectType(a.Value.position(), want, c.expr(a.Value, want))
	}
	c.checkExclusive(e, mutables)
	return b.Result
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
