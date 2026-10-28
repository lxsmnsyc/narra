package narra

import "strconv"

// destructure checks `const { a, b: c } = s` and `const [a, b] = t`.
func (c *Checker) destructure(s *VarStmt) {
	var want Type
	if s.Type != nil {
		want = c.resolveType(s.Type)
	}
	t := c.expr(s.Value, want)
	if want != nil {
		c.expectType(s.Value.position(), want, t)
		t = want
	}
	c.resolvePattern(s.Pattern, t)
	if !irrefutable(s.Pattern) {
		c.errorf(s.Pattern.Pos, "this pattern does not match every %s; use match or `is` to test a value", t)
	}
	for _, l := range s.Pattern.Locals {
		l.Mutable = s.Mutable
		c.scope.names[l.Name] = l
	}
}

// spreadLit checks `{ ...s, field: v }`. The result has the type of s.
func (c *Checker) spreadLit(e *StructLit, want Type) Type {
	var named Type
	if e.TypeName != "" {
		if g := c.generics[e.TypeName]; g != nil && g.sdecl != nil {
			named = nil // compared by name below
		} else if t, ok := c.types[e.TypeName].(*StructType); ok {
			named = t
		} else {
			c.errorf(e.P, "%q is not a struct", e.TypeName)
			return TVoid
		}
	}
	hint := want
	if named != nil {
		hint = named
	}
	t := c.expr(e.Spread, hint)
	st, ok := t.(*StructType)
	if !ok || st.Tuple {
		c.errorf(e.Spread.position(), "a spread needs a struct value, found %s", t)
		return TVoid
	}
	target := hint
	if e.TypeName != "" && named == nil {
		g := c.generics[e.TypeName]
		if st.Name != g.name || st.Module != g.mod.path {
			c.errorf(e.Spread.position(), "cannot spread %s into a %s literal; the spread must have the literal's type", st, e.TypeName)
			return TVoid
		}
		target = nil
	}
	if ws, ok := target.(*StructType); ok && ws != st {
		c.errorf(e.Spread.position(), "cannot spread %s into a %s literal; the spread must have the literal's type", typeName(st, ws), typeName(ws, st))
		return ws
	}
	e.Struct = st
	seen := map[string]bool{}
	for _, fi := range e.Fields {
		f := st.Field(fi.Name)
		if f == nil {
			c.errorf(fi.Pos, "struct %s has no field %q", st, fi.Name)
			c.expr(fi.Value, nil)
			continue
		}
		if seen[fi.Name] {
			c.errorf(fi.Pos, "field %q is given twice", fi.Name)
		}
		seen[fi.Name] = true
		c.expectType(fi.Value.position(), f.T, c.expr(fi.Value, f.T))
	}
	return st
}

// tupleType resolves Tuple<A, B, ...>.
func (c *Checker) tupleType(te *TypeExpr) Type {
	if len(te.Args) < 2 {
		c.errorf(te.Pos, "Tuple takes at least two type arguments")
		return TVoid
	}
	elems := make([]Type, len(te.Args))
	for i, a := range te.Args {
		elems[i] = c.resolveType(a)
	}
	return c.tupleOf(elems)
}

// tupleOf returns the one tuple type with these element types. A tuple is a
// prelude struct named Tuple whose fields are "0", "1", ..., so the struct
// code paths (layout, reference counts, traits) handle it.
func (c *Checker) tupleOf(elems []Type) *StructType {
	key := typeArgsMangle(elems)
	if c.tuples == nil {
		c.tuples = map[string]*StructType{}
	}
	if t, ok := c.tuples[key]; ok {
		return t
	}
	t := &StructType{Name: "Tuple", Args: elems, Tuple: true}
	for i, e := range elems {
		t.Fields = append(t.Fields, &Field{Name: strconv.Itoa(i), T: e})
	}
	c.tuples[key] = t
	c.instTypes = append(c.instTypes, t)
	// Tuples derive Show and Ordered, and get Equatable and Hashable like
	// any struct, when their elements implement them.
	k := typeKey{"", "Tuple"}
	if c.traits != nil && c.traits.types[k] == nil {
		info := &typeTraits{name: "Tuple", generic: true, derives: map[*Trait]bool{}, implements: map[*Trait]bool{}, impls: map[*Trait]*ImplDecl{}}
		for _, n := range []string{"Ordered", "Show"} {
			if tr := c.stdTrait(n); tr != nil {
				info.derives[tr] = true
			}
		}
		c.traits.types[k] = info
	}
	return t
}

// tupleLit checks `[a, b]` with a tuple as the expected type, and Tuple(a, b).
func (c *Checker) tupleLit(e *ArrayLit, want Type) Type {
	wt, _ := want.(*StructType)
	if wt != nil && (!wt.Tuple || len(wt.Fields) != len(e.Elems)) {
		wt = nil
	}
	if wt == nil && !e.TupleCall {
		c.errorf(e.P, "expected %s, found a literal with %d elements", want, len(e.Elems))
		return TVoid
	}
	if len(e.Elems) < 2 {
		c.errorf(e.P, "a tuple needs at least two elements")
		return TVoid
	}
	elems := make([]Type, len(e.Elems))
	for i, el := range e.Elems {
		var ew Type
		if wt != nil {
			ew = wt.Fields[i].T
		}
		elems[i] = c.expr(el, ew)
		if ew != nil {
			c.expectType(el.position(), ew, elems[i])
			elems[i] = ew
		}
	}
	e.Tuple = c.tupleOf(elems)
	return e.Tuple
}

func isTupleType(t Type) bool {
	st, ok := t.(*StructType)
	return ok && st.Tuple
}
