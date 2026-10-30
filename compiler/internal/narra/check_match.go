package narra

import (
	"fmt"
	"strings"
)

// resolvePattern checks a pattern against the type of the value it matches
// and creates a Local for every name it binds, in p.Locals. The caller puts
// the locals in scope.
func (c *Checker) resolvePattern(p *Pattern, t Type) {
	p.Locals = nil
	c.resolvePat(p, t, p, map[string]bool{})
}

func (c *Checker) resolvePat(p *Pattern, t Type, root *Pattern, seen map[string]bool) {
	p.T = t
	switch p.Kind {
	case PatWildcard:
	case PatName:
		if et, ok := t.(*EnumType); ok {
			if v := et.Variant(p.Name); v != nil {
				p.Variant = v
				return
			}
			if startsUpper(p.Name) {
				c.errorf(p.Pos, "enum %s has no constructor %q", et, p.Name)
				return
			}
		} else if startsUpper(p.Name) {
			c.errorf(p.Pos, "%s is not a constructor of %s; a name that binds starts with a lowercase letter", p.Name, t)
			return
		}
		if seen[p.Name] {
			c.errorf(p.Pos, "%q is bound twice in this pattern", p.Name)
			return
		}
		seen[p.Name] = true
		if c.lookup(p.Name) != nil {
			c.errorf(p.Pos, "%q is already declared in this function; shadowing is not allowed", p.Name)
		}
		c.localN++
		p.Local = &Local{Name: p.Name, T: t, Pos: p.Pos, CName: fmt.Sprintf("v_%s_%d", p.Name, c.localN)}
		root.Locals = append(root.Locals, p.Local)
	case PatCtor:
		et, ok := t.(*EnumType)
		if !ok {
			c.errorf(p.Pos, "constructor pattern %s needs an enum value, found %s", p.Name, t)
			c.bindRest(p, root, seen)
			return
		}
		if p.Enum != "" && !c.namesEnum(p.Enum, et) {
			c.errorf(p.Pos, "%s.%s does not match a value of type %s", p.Enum, p.Name, et)
			c.bindRest(p, root, seen)
			return
		}
		v := et.Variant(p.Name)
		if v == nil {
			c.errorf(p.Pos, "enum %s has no constructor %q", et, p.Name)
			c.bindRest(p, root, seen)
			return
		}
		p.Variant = v
		if p.Args == nil {
			return
		}
		if len(p.Args) != len(v.Fields) {
			c.errorf(p.Pos, "%s has %d fields, the pattern lists %d", v.Name, len(v.Fields), len(p.Args))
			c.bindRest(p, root, seen)
			return
		}
		for i, a := range p.Args {
			c.resolvePat(a, v.Fields[i].T, root, seen)
		}
	case PatStruct:
		st, ok := t.(*StructType)
		if !ok || st.Tuple {
			c.errorf(p.Pos, "a struct pattern needs a struct value, found %s", t)
			c.bindRest(p, root, seen)
			return
		}
		if c.hiddenStruct(st) {
			c.errorf(p.Pos, "cannot match the fields of %s: it is an opaque struct of module %q", st, st.Module)
		}
		given := map[string]bool{}
		for _, fp := range p.Fields {
			f := st.Field(fp.Name)
			if f == nil {
				c.errorf(fp.Pos, "struct %s has no field %q", st, fp.Name)
				c.bindRest(fp.Pattern, root, seen)
				continue
			}
			if given[fp.Name] {
				c.errorf(fp.Pos, "field %q is listed twice", fp.Name)
				c.bindRest(fp.Pattern, root, seen)
				continue
			}
			given[fp.Name] = true
			c.resolvePat(fp.Pattern, f.T, root, seen)
		}
	case PatTuple:
		st, ok := t.(*StructType)
		if !ok || !st.Tuple {
			c.errorf(p.Pos, "a tuple pattern needs a tuple value, found %s", t)
			c.bindRest(p, root, seen)
			return
		}
		if len(p.Args) != len(st.Fields) {
			c.errorf(p.Pos, "%s has %d elements, the pattern lists %d", st, len(st.Fields), len(p.Args))
			c.bindRest(p, root, seen)
			return
		}
		for i, a := range p.Args {
			c.resolvePat(a, st.Fields[i].T, root, seen)
		}
	case PatLit:
		pt, ok := t.(*PrimType)
		if !ok || pt == TVoid {
			c.errorf(p.Pos, "a literal pattern needs an int, f64, bool, char or string value, found %s", t)
			return
		}
		c.expectType(p.Pos, t, c.expr(p.Lit, t))
	}
}

// bindRest binds the names inside a pattern that failed to check, with type
// void, so their uses do not also report "unknown name".
func (c *Checker) bindRest(p *Pattern, root *Pattern, seen map[string]bool) {
	if p.Kind == PatName && p.Variant == nil && p.Local == nil && !startsUpper(p.Name) && !seen[p.Name] {
		seen[p.Name] = true
		c.localN++
		p.Local = &Local{Name: p.Name, T: TVoid, Pos: p.Pos, CName: fmt.Sprintf("v_%s_%d", p.Name, c.localN)}
		root.Locals = append(root.Locals, p.Local)
	}
	for _, a := range p.Args {
		c.bindRest(a, root, seen)
	}
	for _, f := range p.Fields {
		c.bindRest(f.Pattern, root, seen)
	}
}

// irrefutable reports whether a pattern matches every value of its type.
func irrefutable(p *Pattern) bool {
	switch p.Kind {
	case PatWildcard:
		return true
	case PatName:
		return p.Variant == nil || len(p.Variant.Enum.Variants) == 1
	case PatCtor:
		return p.Variant != nil && len(p.Variant.Enum.Variants) == 1 && argsIrrefutable(p)
	case PatStruct:
		for _, f := range p.Fields {
			if !irrefutable(f.Pattern) {
				return false
			}
		}
		return true
	case PatTuple:
		return argsIrrefutable(p)
	}
	return false
}

// argsIrrefutable reports whether a constructor pattern matches every value
// of its variant, so `x is Some(_)` failing means x is None.
func argsIrrefutable(p *Pattern) bool {
	for _, a := range p.Args {
		if !irrefutable(a) {
			return false
		}
	}
	return true
}

func (c *Checker) isExpr(e *IsExpr) Type {
	binds := c.isBinds
	c.isBinds = false
	t := c.expr(e.Subject, nil)
	c.resolvePattern(e.Pattern, t)
	if len(e.Pattern.Locals) > 0 && !binds {
		c.errorf(e.P, "an `is` pattern can only bind names when it is the whole condition of an if")
	}
	return TBool
}

func (c *Checker) match(m *MatchExpr, want Type, isValue bool) Type {
	m.IsValue = isValue
	st := c.expr(m.Subject, nil)
	if st == TVoid {
		c.errorf(m.P, "match needs a value, found void")
		return TVoid
	}
	var rows [][]*Pattern
	patternErr := false
	result := want
	for _, arm := range m.Arms {
		c.push()
		nerrs := len(c.errs)
		c.resolvePattern(arm.Pattern, st)
		if len(c.errs) > nerrs {
			// Checking the guard and body would only report errors that
			// follow from this one.
			patternErr = true
			c.pop()
			continue
		}
		for _, l := range arm.Pattern.Locals {
			c.scope.names[l.Name] = l
		}
		if arm.Guard == nil {
			rows = append(rows, []*Pattern{arm.Pattern})
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
	limit := 10
	var missing [][]string
	if !patternErr {
		missing = c.missingCases(rows, []Type{st}, &limit)
	}
	if len(missing) > 0 {
		var names []string
		other := false
		for _, w := range missing {
			if w[0] == "_" {
				other = true
			}
			names = append(names, w[0])
		}
		if other {
			c.errorf(m.P, "match does not cover every %s value; add a `_` arm", st)
		} else {
			c.errorf(m.P, "match does not cover %s; add the cases or a `_` arm", strings.Join(names, ", "))
		}
	}
	if !isValue || result == nil {
		return TVoid
	}
	return result
}

// Exhaustiveness follows the usefulness algorithm: a match is exhaustive
// when no value vector escapes the rows. missingCases returns example values
// that no row matches, one string per column, at most *limit of them.
//
// Enums, bools, structs and tuples have a finite set of constructors. Other
// types (int, string, arrays, type parameters) have infinitely many values,
// so only wildcards and bindings cover them.

var wildPattern = &Pattern{Kind: PatWildcard}

func isWild(p *Pattern) bool {
	switch p.Kind {
	case PatWildcard:
		return true
	case PatName:
		return p.Variant == nil
	}
	return false
}

// ctorCount returns the number of constructors of t, or -1 when t has
// infinitely many values.
func ctorCount(t Type) int {
	switch t := t.(type) {
	case *EnumType:
		return len(t.Variants)
	case *StructType:
		return 1
	case *PrimType:
		if t == TBool {
			return 2
		}
	}
	return -1
}

func ctorArgTypes(t Type, k int) []Type {
	var out []Type
	switch t := t.(type) {
	case *EnumType:
		for _, f := range t.Variants[k].Fields {
			out = append(out, f.T)
		}
	case *StructType:
		for _, f := range t.Fields {
			out = append(out, f.T)
		}
	}
	return out
}

// specialize returns the argument patterns of p when p matches constructor
// k of t, with n arguments, and false when it matches another constructor.
func specialize(p *Pattern, t Type, k, n int) ([]*Pattern, bool) {
	wilds := func() []*Pattern {
		out := make([]*Pattern, n)
		for i := range out {
			out[i] = wildPattern
		}
		return out
	}
	if isWild(p) {
		return wilds(), true
	}
	switch p.Kind {
	case PatName:
		if p.Variant.Index != k {
			return nil, false
		}
		return wilds(), true
	case PatCtor:
		if p.Variant == nil {
			return wilds(), true
		}
		if p.Variant.Index != k {
			return nil, false
		}
		if p.Args == nil || len(p.Args) != n {
			return wilds(), true
		}
		return p.Args, true
	case PatTuple:
		if len(p.Args) != n {
			return wilds(), true
		}
		return p.Args, true
	case PatStruct:
		out := wilds()
		st := t.(*StructType)
		for _, fp := range p.Fields {
			for i, f := range st.Fields {
				if f.Name == fp.Name {
					out[i] = fp.Pattern
				}
			}
		}
		return out, true
	case PatLit:
		if b, ok := p.Lit.(*BoolLit); ok {
			if (k == 1) != b.Value {
				return nil, false
			}
			return nil, true
		}
	}
	return wilds(), true
}

func (c *Checker) missingCases(rows [][]*Pattern, tys []Type, limit *int) [][]string {
	if *limit <= 0 {
		return nil
	}
	if len(rows) == 0 {
		*limit--
		w := make([]string, len(tys))
		for i := range w {
			w[i] = "_"
		}
		return [][]string{w}
	}
	if len(tys) == 0 {
		return nil
	}
	t := tys[0]
	n := ctorCount(t)
	allWild := true
	for _, r := range rows {
		if !isWild(r[0]) {
			allWild = false
		}
	}
	if n < 0 || allWild {
		var def [][]*Pattern
		for _, r := range rows {
			if isWild(r[0]) {
				def = append(def, r[1:])
			}
		}
		var out [][]string
		for _, w := range c.missingCases(def, tys[1:], limit) {
			out = append(out, append([]string{"_"}, w...))
		}
		return out
	}
	var out [][]string
	for k := 0; k < n; k++ {
		args := ctorArgTypes(t, k)
		var spec [][]*Pattern
		for _, r := range rows {
			if ps, ok := specialize(r[0], t, k, len(args)); ok {
				row := append(append([]*Pattern{}, ps...), r[1:]...)
				spec = append(spec, row)
			}
		}
		sub := append(append([]Type{}, args...), tys[1:]...)
		for _, w := range c.missingCases(spec, sub, limit) {
			head := ctorText(t, k, w[:len(args)])
			out = append(out, append([]string{head}, w[len(args):]...))
		}
	}
	return out
}

// ctorText prints constructor k of t applied to the example arguments.
func ctorText(t Type, k int, args []string) string {
	allWild := true
	for _, a := range args {
		if a != "_" {
			allWild = false
		}
	}
	switch t := t.(type) {
	case *EnumType:
		v := t.Variants[k]
		if len(v.Fields) == 0 {
			return v.Name
		}
		return v.Name + "(" + strings.Join(args, ", ") + ")"
	case *StructType:
		if allWild {
			return "_"
		}
		if t.Tuple {
			return "#[" + strings.Join(args, ", ") + "]"
		}
		var parts []string
		for i, a := range args {
			if a != "_" {
				parts = append(parts, t.Fields[i].Name+": "+a)
			}
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	case *PrimType:
		if k == 1 {
			return "true"
		}
		return "false"
	}
	return "_"
}
