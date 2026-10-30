package narra

import (
	"fmt"
	"strings"
)

// Traits. See roadmaps/selfhost/contracts/trait-model.md.
//
// Generic functions are compiled once per set of type arguments, so a trait
// call always resolves to a direct call: a user implementation (CallExpr.Func)
// or a function the compiler writes (CallExpr.Trait). Constraints are checked
// when the generic body is checked with its parameters open, and again at
// each call site against the type arguments.

// Trait is a declared trait. The standard traits are declared in the prelude.
type Trait struct {
	Name   string
	Funcs  []*TraitFunc
	Self   *SelfType
	Decl   *TraitDecl
	Std    bool   // Equatable, Hashable, Ordered or Show
	Module string // declaring module path; "" for the prelude
}

// TraitFunc is one function signature of a trait.
type TraitFunc struct {
	Name   string
	Params []Type // Self appears as *SelfType
	Result Type
	Decl   *FuncDecl // the signature, with Params[i].T and ResultT resolved
}

// SelfType stands for the implementing type inside a trait signature. It is
// a type parameter, so unify and substitute work on it unchanged.
type SelfType = TypeParamType

// TraitCall marks a call to a trait function the compiler implements for
// Self: a standard trait on a type without a user `implement` block.
type TraitCall struct {
	Trait *Trait
	Func  string
	Self  Type
}

// stdTraitFuncs are the functions of the standard traits, keyed
// "Trait.function", with the emitter operation that implements each.
var stdTraitFuncs = map[string]string{
	"Equatable.equals": "eq",
	"Hashable.hash":    "hash",
	"Ordered.compare":  "cmp",
	"Show.show":        "show",
}

// derivable are the traits `derives` accepts.
var derivable = map[string]bool{"Equatable": true, "Hashable": true, "Ordered": true, "Show": true}

// typeTraits is what a struct or enum declaration says about traits.
type typeTraits struct {
	name       string
	module     string // declaring module path; "" for the prelude
	pos        Pos
	generic    bool
	implExprs  []*TypeExpr
	derivExprs []*TypeExpr
	implements map[*Trait]bool
	derives    map[*Trait]bool
	impls      map[*Trait]*ImplDecl
}

type traitKey struct {
	tr *Trait
	t  Type
}

// typeKey names a struct or enum declaration: two modules may use one name.
type typeKey struct{ module, name string }

type traitState struct {
	// traits maps a module path to the traits visible in it by name: its own
	// and those it imports. "" holds the prelude's, visible everywhere.
	traits    map[string]map[string]*Trait
	types     map[typeKey]*typeTraits
	bounds    map[*TypeParamType][]*Trait
	busy      map[traitKey]bool
	implType  map[*ImplDecl]Type
	implFuncs map[Type]map[string]*FuncDecl // "Trait.function" -> implementation
}

func newTraitState() *traitState {
	return &traitState{
		traits:    map[string]map[string]*Trait{},
		types:     map[typeKey]*typeTraits{},
		bounds:    map[*TypeParamType][]*Trait{},
		busy:      map[traitKey]bool{},
		implType:  map[*ImplDecl]Type{},
		implFuncs: map[Type]map[string]*FuncDecl{},
	}
}

// ownsType reports whether the current module may write `implement` blocks
// for the type: only the module that declares it (SPEC section 4, D5).
func (c *Checker) ownsType(info *typeTraits) bool { return info.module == c.mod.path }

func (c *Checker) stdTrait(name string) *Trait {
	if tr := c.traits.traits[""][name]; tr != nil && tr.Std {
		return tr
	}
	return nil
}

// lookupTrait finds a trait by the name the current module uses for it.
func (c *Checker) lookupTrait(name string) *Trait {
	if tr := c.traits.traits[c.mod.path][name]; tr != nil {
		return tr
	}
	return c.traits.traits[""][name]
}

func (c *Checker) addTrait(name string, tr *Trait) {
	m := c.traits.traits[c.mod.path]
	if m == nil {
		m = map[string]*Trait{}
		c.traits.traits[c.mod.path] = m
	}
	m[name] = tr
}

// importTrait makes an exported trait of another module visible by name.
func (c *Checker) importTrait(pos Pos, name string, d *TraitDecl) {
	if d.T == nil {
		return
	}
	if c.lookupTrait(name) != nil || c.types[name] != nil || c.generics[name] != nil {
		c.errorf(pos, "%q is declared twice", name)
		return
	}
	c.addTrait(name, d.T)
}

func typeKeyOf(t Type) typeKey {
	switch t := t.(type) {
	case *StructType:
		return typeKey{t.Module, t.Name}
	case *EnumType:
		return typeKey{t.Module, t.Name}
	}
	return typeKey{}
}

// declareTraits records trait names and what each struct and enum lists in
// `implements` and `derives`. It runs after type names are declared and
// before function signatures, so constraints can name traits.
func (c *Checker) declareTraits(f *File) {
	ts := c.traits
	mod := c.mod.path
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *TraitDecl:
			_, isType := c.types[d.Name]
			_, isGeneric := c.generics[d.Name]
			if c.lookupTrait(d.Name) != nil || isType || isGeneric {
				c.errorf(d.Pos, "%q is declared twice", d.Name)
				continue
			}
			typeParamID++
			tr := &Trait{Name: d.Name, Decl: d, Module: mod, Self: &SelfType{Name: "Self", ID: typeParamID}}
			tr.Std = mod == "" && derivable[d.Name]
			d.T = tr
			c.addTrait(d.Name, tr)
		case *StructDecl:
			ts.types[typeKey{mod, d.Name}] = &typeTraits{name: d.Name, module: mod, pos: d.Pos, generic: len(d.TypeParams) > 0, implExprs: d.Implements, derivExprs: d.Derives}
		case *EnumDecl:
			ts.types[typeKey{mod, d.Name}] = &typeTraits{name: d.Name, module: mod, pos: d.Pos, generic: len(d.TypeParams) > 0, implExprs: d.Implements, derivExprs: d.Derives}
		}
	}
	for _, d := range f.Decls {
		var info *typeTraits
		switch d := d.(type) {
		case *StructDecl:
			info = ts.types[typeKey{mod, d.Name}]
		case *EnumDecl:
			info = ts.types[typeKey{mod, d.Name}]
		default:
			continue
		}
		info.implements = map[*Trait]bool{}
		info.derives = map[*Trait]bool{}
		info.impls = map[*Trait]*ImplDecl{}
		for _, te := range info.implExprs {
			tr := c.traitNamed(te)
			if tr == nil {
				continue
			}
			if info.implements[tr] {
				c.errorf(te.Pos, "%s is listed twice in `implements`", tr.Name)
			}
			info.implements[tr] = true
		}
		for _, te := range info.derivExprs {
			if te.Name == "Shape" {
				c.errorf(te.Pos, "derives Shape is not supported by this compiler yet")
				continue
			}
			tr := c.traitNamed(te)
			if tr == nil {
				continue
			}
			if !tr.Std {
				c.errorf(te.Pos, "%s cannot be derived; only Equatable, Hashable, Ordered and Show can. List it in `implements` and write `implement %s for %s`", tr.Name, tr.Name, info.name)
				continue
			}
			if info.derives[tr] {
				c.errorf(te.Pos, "%s is listed twice in `derives`", tr.Name)
			}
			info.derives[tr] = true
		}
	}
}

func (c *Checker) traitNamed(te *TypeExpr) *Trait {
	tr := c.lookupTrait(te.Name)
	if tr == nil {
		c.errorf(te.Pos, "unknown trait %q", te.Name)
	}
	return tr
}

// funcBounds resolves the constraints of a generic function's type parameters.
func (c *Checker) funcBounds(fd *FuncDecl, tps []*TypeParamType) {
	for i, bs := range fd.Bounds {
		seen := map[*Trait]bool{}
		for _, te := range bs {
			if tr := c.traitNamed(te); tr != nil {
				if seen[tr] {
					c.errorf(te.Pos, "%s is listed twice in the constraint of %s", tr.Name, fd.TypeParams[i])
					continue
				}
				seen[tr] = true
				c.traits.bounds[tps[i]] = append(c.traits.bounds[tps[i]], tr)
			}
		}
	}
}

// resolveTraits resolves trait signatures and `implement` blocks, then checks
// that every listed trait is implemented and every derived trait can be.
// It runs once every type's fields are known.
func (c *Checker) resolveTraits(f *File) {
	ts := c.traits
	c.prog.Impls = ts.implFuncs
	for _, d := range f.Decls {
		if td, ok := d.(*TraitDecl); ok && td.T != nil {
			c.resolveTraitDecl(td)
		}
	}
	for _, d := range f.Decls {
		if id, ok := d.(*ImplDecl); ok {
			c.resolveImpl(id)
		}
	}
	if c.mod.path == "" {
		return
	}
	for _, d := range f.Decls {
		var info *typeTraits
		var t Type
		switch d := d.(type) {
		case *StructDecl:
			info = ts.types[typeKey{c.mod.path, d.Name}]
			if d.T != nil {
				t = d.T
			}
		case *EnumDecl:
			info = ts.types[typeKey{c.mod.path, d.Name}]
			if d.T != nil {
				t = d.T
			}
		default:
			continue
		}
		for _, te := range info.implExprs {
			tr := c.lookupTrait(te.Name)
			if tr != nil && info.impls[tr] == nil && !info.derives[tr] {
				c.errorf(te.Pos, "%s implements %s but has no `implement %s for %s` block or `derives %s`", info.name, tr.Name, tr.Name, info.name, tr.Name)
			}
		}
		for _, te := range info.derivExprs {
			tr := c.lookupTrait(te.Name)
			if tr == nil || !info.derives[tr] {
				continue
			}
			if info.impls[tr] != nil {
				c.errorf(te.Pos, "%s both derives and implements %s; keep one", info.name, tr.Name)
				continue
			}
			// A generic type derives a trait for the instances whose
			// contents implement it, so only concrete types are checked here.
			if t == nil {
				continue
			}
			for _, fl := range fieldsOf(t) {
				if !c.implements(fl.T, tr) {
					c.errorf(te.Pos, "cannot derive %s for %s: field %q has type %s, which does not implement %s", tr.Name, info.name, fl.Name, fl.T, tr.Name)
				}
			}
		}
	}
}

func (c *Checker) resolveTraitDecl(td *TraitDecl) {
	tr := td.T
	c.env = map[string]Type{"Self": tr.Self}
	seen := map[string]bool{}
	for _, fd := range td.Funcs {
		if len(fd.TypeParams) > 0 {
			c.errorf(fd.Pos, "trait functions cannot have type parameters in this compiler yet")
		}
		if seen[fd.Name] {
			c.errorf(fd.Pos, "%s declares function %q twice", tr.Name, fd.Name)
		}
		seen[fd.Name] = true
		tf := &TraitFunc{Name: fd.Name, Decl: fd}
		mentionsSelf := false
		for _, p := range fd.Params {
			p.T = c.resolveType(p.Type)
			tf.Params = append(tf.Params, p.T)
			if containsOpen(p.T, []*TypeParamType{tr.Self}) {
				mentionsSelf = true
			}
		}
		fd.ResultT = TVoid
		if fd.Result != nil {
			fd.ResultT = c.resolveType(fd.Result)
		}
		tf.Result = fd.ResultT
		if !mentionsSelf {
			c.errorf(fd.Pos, "a parameter of %s.%s must use Self, so a call can tell which implementation to run", tr.Name, fd.Name)
		}
		tr.Funcs = append(tr.Funcs, tf)
	}
	c.env = nil
}

func (c *Checker) resolveImpl(d *ImplDecl) {
	ts := c.traits
	tr := c.traitNamed(d.Trait)
	if tr == nil {
		return
	}
	te := d.Type
	var info *typeTraits
	if !te.Func {
		switch t := c.types[te.Name].(type) {
		case *StructType, *EnumType:
			info = ts.types[typeKeyOf(t)]
		case nil:
			if g := c.generics[te.Name]; g != nil {
				info = ts.types[typeKey{g.mod.path, g.name}]
			} else if te.Name != "Array" {
				c.errorf(te.Pos, "unknown type %q", te.Name)
				return
			}
		}
	}
	if info == nil || !c.ownsType(info) {
		c.errorf(d.Pos, "cannot implement %s for %s: the type is declared in another module; wrap it in a struct", tr.Name, typeExprString(te))
		return
	}
	if info.generic {
		c.errorf(d.Pos, "implement blocks for generic types are not supported by this compiler yet; use derives")
		return
	}
	if len(te.Args) > 0 {
		c.errorf(te.Pos, "type %s is not generic", te.Name)
		return
	}
	t := c.types[te.Name]
	if !info.implements[tr] {
		c.errorf(d.Pos, "%s does not list %s in `implements`; add `implements %s` to its declaration", info.name, tr.Name, tr.Name)
		return
	}
	if info.impls[tr] != nil {
		c.errorf(d.Pos, "%s is implemented twice for %s", tr.Name, info.name)
		return
	}
	info.impls[tr] = d
	ts.implType[d] = t
	if ts.implFuncs[t] == nil {
		ts.implFuncs[t] = map[string]*FuncDecl{}
	}
	selfMap := map[*TypeParamType]Type{tr.Self: t}
	c.env = map[string]Type{"Self": t}
	got := map[string]bool{}
	for _, fd := range d.Funcs {
		if len(fd.TypeParams) > 0 {
			c.errorf(fd.Pos, "functions in an implement block cannot have type parameters in this compiler yet")
		}
		for _, p := range fd.Params {
			p.T = c.resolveType(p.Type)
		}
		fd.ResultT = TVoid
		if fd.Result != nil {
			fd.ResultT = c.resolveType(fd.Result)
		}
		var tf *TraitFunc
		for _, f := range tr.Funcs {
			if f.Name == fd.Name {
				tf = f
			}
		}
		if tf == nil {
			c.errorf(fd.Pos, "trait %s has no function %q", tr.Name, fd.Name)
			continue
		}
		if got[fd.Name] {
			c.errorf(fd.Pos, "function %q is implemented twice", fd.Name)
			continue
		}
		got[fd.Name] = true
		if !c.sameSignature(fd, tf, selfMap) {
			c.errorf(fd.Pos, "%s.%s for %s must be `%s`", tr.Name, fd.Name, info.name, c.signatureText(tf, selfMap))
		}
		key := tr.Name + "." + tf.Name
		fd.Name = implName(tr, tf.Name, t)
		fd.Module = c.mod.path
		c.funcMod[fd] = c.mod
		ts.implFuncs[t][key] = fd
		c.prog.Funcs = append(c.prog.Funcs, fd)
	}
	c.env = nil
	for _, tf := range tr.Funcs {
		if !got[tf.Name] {
			c.errorf(d.Pos, "implement %s for %s is missing function %s", tr.Name, info.name, tf.Name)
		}
	}
}

func (c *Checker) sameSignature(fd *FuncDecl, tf *TraitFunc, self map[*TypeParamType]Type) bool {
	want := tf.Decl
	if len(fd.Params) != len(want.Params) || !sameType(fd.ResultT, c.substitute(want.ResultT, self)) {
		return false
	}
	for i, p := range fd.Params {
		if p.Mutable != want.Params[i].Mutable || p.Escaping != want.Params[i].Escaping || !sameType(p.T, c.substitute(want.Params[i].T, self)) {
			return false
		}
	}
	return true
}

func (c *Checker) signatureText(tf *TraitFunc, self map[*TypeParamType]Type) string {
	var parts []string
	for _, p := range tf.Decl.Params {
		m := ""
		if p.Mutable {
			m = "mutable "
		}
		if p.Escaping {
			m = "escaping "
		}
		parts = append(parts, fmt.Sprintf("%s%s: %s", m, p.Name, c.substitute(p.T, self)))
	}
	s := fmt.Sprintf("function %s(%s)", tf.Name, strings.Join(parts, ", "))
	if r := c.substitute(tf.Result, self); r != TVoid {
		s += ": " + r.String()
	}
	return s
}

func typeExprString(te *TypeExpr) string {
	if te.Func {
		return "a function type"
	}
	if len(te.Args) == 0 && !te.Tuple {
		return te.Name
	}
	parts := make([]string, len(te.Args))
	for i, a := range te.Args {
		parts[i] = typeExprString(a)
	}
	if te.Tuple {
		return "#[" + strings.Join(parts, ", ") + "]"
	}
	return te.Name + "<" + strings.Join(parts, ", ") + ">"
}

// implBodies checks the functions of an `implement` block.
func (c *Checker) implBodies(d *ImplDecl) {
	t, ok := c.traits.implType[d]
	if !ok {
		return
	}
	for _, fd := range d.Funcs {
		c.funcBody(fd, map[string]Type{"Self": t})
	}
}

// declaredName is the declared name of a struct or enum, without type arguments.
func declaredName(t Type) string {
	switch t := t.(type) {
	case *StructType:
		return t.Name
	case *EnumType:
		return t.Name
	}
	return ""
}

// fieldsOf lists the fields of a struct, or of every variant of an enum.
func fieldsOf(t Type) []*Field {
	switch t := t.(type) {
	case *StructType:
		return t.Fields
	case *EnumType:
		var fs []*Field
		for _, v := range t.Variants {
			fs = append(fs, v.Fields...)
		}
		return fs
	}
	return nil
}

// implements reports whether t implements the trait.
//
// Equatable and Hashable are automatic for types whose contents implement
// them (Hashable excludes f64). Arrays and the prelude enums implement a
// standard trait when their contents do. A struct or enum implements a trait
// through an `implement` block or `derives`. A type parameter implements the
// traits in its constraint.
func (c *Checker) implements(t Type, tr *Trait) bool {
	if tr == nil || t == nil {
		return false
	}
	key := traitKey{tr, t}
	if c.traits.busy[key] {
		// Reached again through an array, as in `struct Tree { kids: Array<Tree> }`.
		return true
	}
	c.traits.busy[key] = true
	defer delete(c.traits.busy, key)
	switch t := t.(type) {
	case *TypeParamType:
		for _, b := range c.traits.bounds[t] {
			if b == tr {
				return true
			}
		}
		return false
	case *PrimType:
		if !tr.Std || t == TVoid {
			return false
		}
		return tr.Name != "Hashable" || t != TF64
	case *ArrayType:
		return tr.Std && c.implements(t.Elem, tr)
	case *StructType, *EnumType:
		info := c.traits.types[typeKeyOf(t)]
		if info == nil {
			return false
		}
		if info.impls[tr] != nil {
			return true
		}
		auto := tr.Std && (tr.Name == "Equatable" || tr.Name == "Hashable")
		if !auto && !info.derives[tr] {
			return false
		}
		for _, f := range fieldsOf(t) {
			if !c.implements(f.T, tr) {
				return false
			}
		}
		return true
	}
	return false
}

// implFunc returns the user implementation of a trait function for t, or nil
// when the compiler provides it (automatic Equatable and Hashable, derives).
func (c *Checker) implFunc(t Type, tr *Trait, name string) *FuncDecl {
	return c.traits.implFuncs[t][tr.Name+"."+name]
}

// traitCall checks `Trait.function(args)`.
func (c *Checker) traitCall(e *CallExpr, tr *Trait, name string, want Type) Type {
	var tf *TraitFunc
	for _, f := range tr.Funcs {
		if f.Name == name {
			tf = f
		}
	}
	if tf == nil {
		c.errorf(e.P, "trait %s has no function %q", tr.Name, name)
		c.argsLoose(e)
		return TVoid
	}
	fd := tf.Decl
	self := []*TypeParamType{tr.Self}
	m := map[*TypeParamType]Type{}
	if want != nil {
		unify(fd.ResultT, want, m)
	}
	c.args(e, fd.Params, func(i int) Type {
		pw := c.substitute(fd.Params[i].T, m)
		if containsOpen(pw, self) {
			return nil
		}
		return pw
	})
	for i, a := range e.Args {
		if i < len(fd.Params) {
			unify(fd.Params[i].T, a.Value.Type(), m)
		}
	}
	st, ok := m[tr.Self]
	if !ok {
		c.errorf(e.P, "cannot tell which type's %s.%s to call here; add a type annotation", tr.Name, name)
		return TVoid
	}
	for i, a := range e.Args {
		if i < len(fd.Params) {
			c.expectType(a.Value.position(), c.substitute(fd.Params[i].T, m), a.Value.Type())
		}
	}
	result := c.substitute(fd.ResultT, m)
	if !c.implements(st, tr) {
		if tp, ok := st.(*TypeParamType); ok {
			c.errorf(e.P, "%s.%s needs %s to implement %s; declare it as <%s: %s>", tr.Name, name, tp.Name, tr.Name, tp.Name, tr.Name)
		} else {
			c.errorf(e.P, "%s does not implement %s", st, tr.Name)
		}
		return result
	}
	if containsTypeParam(st) {
		// A generic body checked with its parameters open; each instance
		// resolves the call again with concrete types.
		return result
	}
	if impl := c.implFunc(st, tr, name); impl != nil {
		e.Func = impl
	} else {
		e.Trait = &TraitCall{Trait: tr, Func: name, Self: st}
	}
	return result
}

// checkBounds checks a generic call's type arguments against the
// constraints of the function's type parameters.
func (c *Checker) checkBounds(pos Pos, fd *FuncDecl, tps []*TypeParamType, targs []Type) {
	for i, tp := range tps {
		for _, tr := range c.traits.bounds[tp] {
			if !c.implements(targs[i], tr) {
				c.errorf(pos, "%s does not implement %s, which %s needs for type parameter %s", targs[i], tr.Name, fd.Name, tp.Name)
			}
		}
	}
}

// checkEquatable checks that `==` and `!=` work on t.
func (c *Checker) checkEquatable(pos Pos, t Type) {
	if t == TVoid || c.implements(t, c.stdTrait("Equatable")) {
		return
	}
	switch {
	case containsFunc(t, map[Type]bool{}):
		c.errorf(pos, "cannot use == on %s: it contains a function", t)
	case containsTypeParam(t):
		c.errorf(pos, "== on %s needs Equatable; declare the type parameter as <T: Equatable>", t)
	default:
		c.errorf(pos, "%s does not implement Equatable", t)
	}
}

func containsFunc(t Type, seen map[Type]bool) bool {
	if seen[t] {
		return false
	}
	seen[t] = true
	switch t := t.(type) {
	case *FuncType:
		return true
	case *ArrayType:
		return containsFunc(t.Elem, seen)
	case *StructType, *EnumType:
		for _, f := range fieldsOf(t) {
			if containsFunc(f.T, seen) {
				return true
			}
		}
	}
	return false
}
