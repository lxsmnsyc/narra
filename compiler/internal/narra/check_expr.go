package narra

import (
	"strconv"
)

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
		} else {
			c.checkEquatable(e.P, l)
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
		if wt, ok := want.(*StructType); ok && wt.Name == g.name && wt.Module == g.mod.path {
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
