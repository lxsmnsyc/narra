package narra

import (
	"fmt"
	"strings"
)

func (c *Checker) resolvePattern(p *Pattern, et *EnumType) {
	if p.Wildcard {
		return
	}
	v := et.Variant(p.Name)
	if v == nil {
		c.errorf(p.Pos, "enum %s has no constructor %q", et, p.Name)
		return
	}
	p.Variant = v
	if len(p.Binds) != len(v.Fields) && len(p.Binds) != 0 {
		c.errorf(p.Pos, "%s has %d fields, the pattern lists %d", v.Name, len(v.Fields), len(p.Binds))
		return
	}
	p.Locals = nil
	for i, b := range p.Binds {
		if b == "_" {
			p.Locals = append(p.Locals, nil)
			continue
		}
		if c.lookup(b) != nil {
			c.errorf(p.Pos, "%q is already declared in this function; shadowing is not allowed", b)
		}
		c.localN++
		p.Locals = append(p.Locals, &Local{Name: b, T: v.Fields[i].T, Pos: p.Pos, CName: fmt.Sprintf("v_%s_%d", b, c.localN)})
	}
}

func (c *Checker) match(m *MatchExpr, want Type, isValue bool) Type {
	m.IsValue = isValue
	st := c.expr(m.Subject, nil)
	et, ok := st.(*EnumType)
	if !ok {
		c.errorf(m.P, "match needs an enum value, found %s", st)
		return TVoid
	}
	covered := map[*Variant]bool{}
	hasWildcard := false
	result := want
	for _, arm := range m.Arms {
		c.push()
		p := arm.Pattern
		if p.Wildcard {
			hasWildcard = true
		} else {
			c.resolvePattern(p, et)
			for _, l := range p.Locals {
				if l != nil {
					c.scope.names[l.Name] = l
				}
			}
			if p.Variant != nil && arm.Guard == nil {
				covered[p.Variant] = true
			}
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
	if !hasWildcard {
		var missing []string
		for _, v := range et.Variants {
			if !covered[v] {
				missing = append(missing, v.Name)
			}
		}
		if len(missing) > 0 {
			c.errorf(m.P, "match does not cover %s; add the cases or a `_` arm", strings.Join(missing, ", "))
		}
	}
	if !isValue || result == nil {
		return TVoid
	}
	return result
}
