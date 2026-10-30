package narra

import "strings"

// Top-level constants (SPEC section 10). A constant is checked the first
// time something refers to it, or when its module's bodies are checked, so
// the order of declarations does not matter and a cycle shows up as a
// reference to a constant that is still being checked.

const (
	constNew = iota
	constBusy
	constDone
)

// declareConsts registers the constants of f in the current module.
func (c *Checker) declareConsts(f *File) {
	// An imported constant was added before this module's types existed.
	for _, imp := range f.Imports {
		for _, n := range imp.Names {
			if d := c.mod.consts[n]; d != nil && d.Module != c.mod.path && c.typeNamed(n) {
				c.errorf(imp.Pos, "constant %q is declared twice", n)
			}
		}
	}
	for _, d := range f.Decls {
		cd, ok := d.(*ConstDecl)
		if !ok {
			continue
		}
		cd.Module = c.mod.path
		_, dupFunc := c.funcs[cd.Name]
		if _, dup := c.mod.consts[cd.Name]; dup || dupFunc || c.typeNamed(cd.Name) {
			c.errorf(cd.Pos, "constant %q is declared twice", cd.Name)
			continue
		}
		c.mod.consts[cd.Name] = cd
	}
}

// importConst makes an exported constant of another module visible.
func (c *Checker) importConst(pos Pos, d *ConstDecl) {
	if _, dup := c.mod.consts[d.Name]; dup || c.typeNamed(d.Name) {
		c.errorf(pos, "constant %q is declared twice", d.Name)
		return
	}
	c.mod.consts[d.Name] = d
}

// constType returns the type of d, checking its value first when needed.
// pos is where the constant is referred to.
func (c *Checker) constType(d *ConstDecl, pos Pos) Type {
	switch d.state {
	case constDone:
		return d.T
	case constBusy:
		c.errorf(pos, "constant %q is defined in terms of itself", d.Name)
		return TVoid
	}
	d.state = constBusy
	// The value is checked as a body of its own module, outside any function.
	saved := c.enter(c.mods[d.Module])
	fn, sc, env, loops := c.fn, c.scope, c.env, c.loops
	c.fn = &FuncDecl{Name: "const " + d.Name, ResultT: TVoid, Module: d.Module}
	c.scope = &scope{names: map[string]*Local{}}
	c.env, c.loops = nil, 0

	var want Type
	if d.Type != nil {
		want = c.resolveType(d.Type)
	}
	nerrs := len(c.errs)
	t := c.expr(d.Value, want)
	if want != nil {
		c.expectType(d.Value.position(), want, t)
		t = want
	}
	if len(c.errs) == nerrs {
		if t == TVoid {
			c.errorf(d.Pos, "this expression has no value")
		} else {
			c.checkConstValue(d, d.Value)
		}
	}
	d.T = t
	d.state = constDone
	c.prog.Consts = append(c.prog.Consts, d)

	c.fn, c.scope, c.env, c.loops = fn, sc, env, loops
	c.enter(saved)
	return t
}

// moduleConstValue checks `Alias.NAME` where Alias names a package module
// and NAME is one of its constants. It returns nil for anything else.
func (c *Checker) moduleConstValue(e *FieldExpr) Type {
	id, ok := e.Target.(*Ident)
	if !ok || c.lookup(id.Name) != nil {
		return nil
	}
	path, ok := c.modules[id.Name]
	s := c.mods[path]
	if !ok || s == nil || s.src == nil {
		return nil
	}
	var d *ConstDecl
	for _, f := range s.src.Files {
		for _, x := range f.Decls {
			if cd, ok := x.(*ConstDecl); ok && cd.Name == e.Name {
				d = cd
			}
		}
	}
	if d == nil {
		return nil
	}
	if c.exported(e.P, s, e.Name) == nil {
		return TVoid
	}
	id.setType(&ModuleType{Path: path})
	e.Const = d
	return c.constType(d, e.P)
}

// constWrite reports a write through a constant and returns true, or
// returns false when the root of e is not a constant. It reads the syntax
// only, so it can run before e is checked.
func (c *Checker) constWrite(pos Pos, e Expr) bool {
	for {
		switch x := e.(type) {
		case *Ident:
			if c.lookup(x.Name) != nil || c.mod.consts[x.Name] == nil {
				return false
			}
			c.errorf(pos, "constant %q cannot be changed; copy it into a `let` variable first", x.Name)
			return true
		case *FieldExpr:
			if id, ok := x.Target.(*Ident); ok && c.lookup(id.Name) == nil && c.moduleHasConst(id.Name, x.Name) {
				c.errorf(pos, "constant %q cannot be changed; copy it into a `let` variable first", x.Name)
				return true
			}
			e = x.Target
		case *IndexExpr:
			e = x.Target
		default:
			return false
		}
	}
}

// moduleHasConst reports whether alias names a package module that declares
// the constant name.
func (c *Checker) moduleHasConst(alias, name string) bool {
	s := c.mods[c.modules[alias]]
	if s == nil || s.src == nil {
		return false
	}
	for _, f := range s.src.Files {
		for _, d := range f.Decls {
			if cd, ok := d.(*ConstDecl); ok && cd.Name == name {
				return true
			}
		}
	}
	return false
}

// typeNamed reports whether name is a type or trait of the current module.
func (c *Checker) typeNamed(name string) bool {
	_, isType := c.types[name]
	return isType || c.generics[name] != nil || c.lookupTrait(name) != nil
}

// checkConstValue reports the first part of x that is not allowed in the
// value of a top-level constant.
func (c *Checker) checkConstValue(d *ConstDecl, x Expr) bool {
	bad := func(what string) bool {
		c.errorf(x.position(), "constant %q cannot use %s; a top-level constant holds literals, constructors, structs, arrays, function names and other constants", d.Name, what)
		return false
	}
	switch x := x.(type) {
	case *IntLit, *FloatLit, *StringLit, *CharLit, *BoolLit:
		return true
	case *TemplateLit:
		for _, p := range x.Parts {
			if !c.checkConstValue(d, p) {
				return false
			}
		}
	case *Ident:
		return true
	case *FieldExpr:
		if x.Ctor != nil || x.Func != nil || x.Const != nil {
			return true
		}
		if x.Narrow != nil {
			return bad("a variant field")
		}
		return c.checkConstValue(d, x.Target)
	case *UnaryExpr:
		return c.checkConstValue(d, x.Operand)
	case *BinaryExpr:
		return c.checkConstValue(d, x.Left) && c.checkConstValue(d, x.Right)
	case *CallExpr:
		if x.Ctor == nil {
			c.errorf(x.P, "constant %q cannot call function %q; a top-level constant holds literals, constructors, structs, arrays, function names and other constants", d.Name, calleeName(x.Callee))
			return false
		}
		for _, a := range x.Args {
			if !c.checkConstValue(d, a.Value) {
				return false
			}
		}
	case *StructLit:
		if x.Spread != nil && !c.checkConstValue(d, x.Spread) {
			return false
		}
		for _, f := range x.Fields {
			if !c.checkConstValue(d, f.Value) {
				return false
			}
		}
	case *ArrayLit:
		for _, el := range x.Elems {
			if !c.checkConstValue(d, el) {
				return false
			}
		}
	case *MapLit:
		for i := range x.Keys {
			if !c.checkConstValue(d, x.Keys[i]) || !c.checkConstValue(d, x.Vals[i]) {
				return false
			}
		}
	case *ArrowExpr:
		return bad("a closure")
	case *MatchExpr:
		return bad("a match")
	case *IsExpr:
		return bad("an `is` test")
	case *IndexExpr:
		return bad("an element read")
	default:
		return bad("this expression")
	}
	return true
}

// calleeName prints the callee of a call for a message.
func calleeName(e Expr) string {
	switch e := e.(type) {
	case *Ident:
		return e.Name
	case *FieldExpr:
		return strings.TrimPrefix(calleeName(e.Target)+"."+e.Name, ".")
	}
	return "a function value"
}
