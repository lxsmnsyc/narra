package narra

import "strings"

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
	savedMod := c.enter(g.mod)
	defer c.enter(savedMod)
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
	c.checkPolyRecursion(g)
}

// genericEdge is a reference from one generic declaration's template to
// another generic type (or to itself), with the arguments it gives.
type genericEdge struct {
	to   *genericType
	t    Type // the referenced instance, for the message
	grow bool // some argument contains a type parameter without being one
}

// templateEdges lists the generic types that t refers to, at any depth.
func (c *Checker) templateEdges(t Type, out []genericEdge) []genericEdge {
	switch t := t.(type) {
	case *ArrayType:
		return c.templateEdges(t.Elem, out)
	case *BoxType:
		return c.templateEdges(t.Elem, out)
	case *FuncType:
		for _, p := range t.Params {
			out = c.templateEdges(p, out)
		}
		return c.templateEdges(t.Result, out)
	case *StructType:
		if len(t.Args) == 0 {
			return out
		}
		if !t.Tuple {
			out = append(out, genericEdge{c.genericOf(t.Module, t.Name), t, growsArgs(t.Args)})
		}
		for _, a := range t.Args {
			out = c.templateEdges(a, out)
		}
	case *EnumType:
		if len(t.Args) == 0 {
			return out
		}
		out = append(out, genericEdge{c.genericOf(t.Module, t.Name), t, growsArgs(t.Args)})
		for _, a := range t.Args {
			out = c.templateEdges(a, out)
		}
	}
	return out
}

func growsArgs(args []Type) bool {
	for _, a := range args {
		if _, bare := a.(*TypeParamType); !bare && containsTypeParam(a) {
			return true
		}
	}
	return false
}

func (c *Checker) genericEdges(g *genericType) []genericEdge {
	var out []genericEdge
	for _, f := range g.fields {
		out = c.templateEdges(f.T, out)
	}
	for _, v := range g.variants {
		for _, f := range v.Fields {
			out = c.templateEdges(f.T, out)
		}
	}
	return out
}

// checkPolyRecursion rejects a generic type that reaches itself, directly or
// through other generic types, with arguments that strictly contain its
// parameters, such as struct G<T> { next: Array<G<Array<T>>> }. G<int> would
// need G<Array<int>>, then G<Array<Array<int>>>, without end, and every
// instance needs its own C type. Reaching itself with the same arguments
// (List<T> holding Array<List<T>>) is fine.
func (c *Checker) checkPolyRecursion(g *genericType) {
	type node struct {
		g    *genericType
		grew bool
	}
	seen := map[node]bool{}
	var culprit Type
	var walk func(h *genericType, grew bool, via Type) bool
	walk = func(h *genericType, grew bool, via Type) bool {
		for _, e := range c.genericEdges(h) {
			if e.to == nil {
				continue
			}
			ng := grew || e.grow
			v := via
			if e.grow && via == nil {
				v = e.t
			}
			if e.to == g && ng {
				culprit = v
				return true
			}
			n := node{e.to, ng}
			if seen[n] {
				continue
			}
			seen[n] = true
			if walk(e.to, ng, v) {
				return true
			}
		}
		return false
	}
	if !walk(g, false, nil) {
		return
	}
	var pos Pos
	if g.sdecl != nil {
		pos = g.sdecl.Pos
	} else {
		pos = g.edecl.Pos
	}
	c.errorf(pos, "generic type %s contains %s, so each instance needs a larger one without end; every instance needs its own C type, so polymorphic recursion is not supported (use the same type arguments, or an enum or Box)", g.name, culprit)
	// Drop the fields so a later instance cannot grow forever.
	g.fields = nil
	for _, v := range g.variants {
		v.Fields = nil
	}
}

func (c *Checker) instantiate(g *genericType, args []Type) Type {
	c.buildGeneric(g)
	key := typeArgsKey(args)
	if t, ok := g.instances[key]; ok {
		return t
	}
	m := map[*TypeParamType]Type{}
	for i, p := range g.params {
		m[p] = args[i]
	}
	if g.sdecl != nil {
		st := &StructType{Name: g.name, Module: g.mod.path, Args: args}
		g.instances[key] = st
		for _, f := range g.fields {
			st.Fields = append(st.Fields, &Field{Name: f.Name, T: c.substitute(f.T, m), Default: f.Default})
		}
		c.instTypes = append(c.instTypes, st)
		return st
	}
	et := &EnumType{Name: g.name, Module: g.mod.path, Args: args}
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
	case *BoxType:
		return c.boxOf(c.substitute(t.Elem, m))
	case *FuncType:
		if !containsTypeParam(t) {
			return t
		}
		return c.funcOf(c.substituteAll(t.Params, m), c.substitute(t.Result, m))
	case *StructType:
		if len(t.Args) == 0 {
			return t
		}
		if t.Tuple {
			return c.tupleOf(c.substituteAll(t.Args, m))
		}
		return c.instantiate(c.genericOf(t.Module, t.Name), c.substituteAll(t.Args, m))
	case *EnumType:
		if len(t.Args) == 0 {
			return t
		}
		return c.instantiate(c.genericOf(t.Module, t.Name), c.substituteAll(t.Args, m))
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
	case *BoxType:
		if a, ok := act.(*BoxType); ok {
			unify(p.Elem, a.Elem, m)
		}
	case *FuncType:
		if a, ok := act.(*FuncType); ok && len(a.Params) == len(p.Params) {
			for i := range p.Params {
				unify(p.Params[i], a.Params[i], m)
			}
			unify(p.Result, a.Result, m)
		}
	case *StructType:
		if a, ok := act.(*StructType); ok && a.Name == p.Name && a.Module == p.Module && len(a.Args) == len(p.Args) {
			for i := range p.Args {
				unify(p.Args[i], a.Args[i], m)
			}
		}
	case *EnumType:
		if a, ok := act.(*EnumType); ok && a.Name == p.Name && a.Module == p.Module && len(a.Args) == len(p.Args) {
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
			c.typeErrorf(t, "type %s contains itself; put the recursive field in an enum, Box or array", t)
			return
		}
		if et, ok := t.(*EnumType); ok && isBoxedEnum(et) {
			// The value is a tag and a pointer, so the enum is complete
			// before its fields are; this is what lets an enum hold itself.
			state[t] = 2
			c.prog.Types = append(c.prog.Types, t)
			for _, v := range et.Variants {
				for _, f := range v.Fields {
					visit(f.T)
				}
			}
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
	if te.Func {
		var params []Type
		for _, a := range te.Args {
			params = append(params, c.resolveType(a))
		}
		return c.funcOf(params, c.funcResult(te.Result))
	}
	if te.Tuple {
		return c.tupleType(te)
	}
	if t, ok := c.env[te.Name]; ok && len(te.Args) == 0 {
		return t
	}
	if strings.Contains(te.Name, ".") {
		return c.qualifiedType(te)
	}
	if te.Name == "Box" && c.generics["Box"] == nil && c.types["Box"] == nil {
		if len(te.Args) != 1 {
			c.errorf(te.Pos, "Box takes one type argument")
			return TVoid
		}
		return c.boxOf(c.resolveType(te.Args[0]))
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
	if _, ok := c.types[te.Name]; !ok && te.Name == "Tuple" {
		c.errorf(te.Pos, "there is no Tuple type; write a tuple type as #[A, B]")
		return TVoid
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
	key := typeID(elem)
	if a, ok := c.arrays[key]; ok {
		return a
	}
	a := &ArrayType{Elem: elem}
	c.arrays[key] = a
	return a
}

func (c *Checker) boxOf(elem Type) *BoxType {
	key := typeID(elem)
	if b, ok := c.boxes[key]; ok {
		return b
	}
	b := &BoxType{Elem: elem}
	c.boxes[key] = b
	return b
}

func (c *Checker) expectType(pos Pos, want, got Type) {
	if want == nil || got == nil {
		return
	}
	if !sameType(want, got) {
		c.errorf(pos, "expected %s, found %s", typeName(want, got), typeName(got, want))
	}
}
