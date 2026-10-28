package narra

import (
	"fmt"
)

// ifIs emits `if (x is Pattern) { ... }`. The subject is held for the whole
// statement, because the pattern's names read from it.
func (e *Emitter) ifIs(s *IfStmt, is *IsExpr) {
	et := is.Subject.Type().(*EnumType)
	e.line("{")
	e.indent++
	e.pushScope(false)
	e.beginStmt()
	code, owned := e.expr(is.Subject)
	held := e.consume(code, owned, et)
	subj := e.temp(et, held)
	e.endStmt()
	e.addLocal(&Local{Name: "subject", T: et, CName: subj})
	cond := "true"
	if !is.Pattern.Wildcard {
		cond = fmt.Sprintf("%s.tag == %d", subj, is.Pattern.Variant.Index)
	}
	e.line("if (%s) {", cond)
	e.indent++
	e.pushScope(false)
	e.bindPattern(is.Pattern, subj)
	e.blockInScope(s.Then)
	e.popScope()
	e.indent--
	if s.Else != nil {
		e.line("} else {")
		e.indent++
		if b, ok := s.Else.(*Block); ok {
			e.block(b, false)
		} else {
			e.stmt(s.Else)
		}
		e.indent--
	}
	e.line("}")
	e.popScope()
	e.indent--
	e.line("}")
}

// bindPattern declares the names a pattern binds, reading from subj.
func (e *Emitter) bindPattern(p *Pattern, subj string) {
	for i, l := range p.Locals {
		if l == nil {
			continue
		}
		f := p.Variant.Fields[i]
		e.line("%s %s = %s.u.%s.%s;", l.T.CName(), l.CName, subj, variantName(p.Variant), fieldName(f.Name))
		if r := retain(l.T, l.CName); r != "" {
			e.line("%s", r)
		}
		e.addLocal(l)
	}
}

// matchArms emits an if-chain over the arms. body emits each arm's action.
func (e *Emitter) matchArms(m *MatchExpr, body func(arm *MatchArm)) {
	code, _ := e.expr(m.Subject)
	et := m.Subject.Type().(*EnumType)
	subj := e.temp(et, code)
	e.tempN++
	done := fmt.Sprintf("m%d", e.tempN)
	e.line("bool %s = false;", done)
	for _, arm := range m.Arms {
		p := arm.Pattern
		if p.Wildcard {
			e.line("if (!%s) {", done)
		} else {
			e.line("if (!%s && %s.tag == %d) {", done, subj, p.Variant.Index)
		}
		e.indent++
		e.pushScope(false)
		if !p.Wildcard {
			e.bindPattern(p, subj)
		}
		if arm.Guard != nil {
			e.beginStmt()
			g, _ := e.expr(arm.Guard)
			gv := e.temp(TBool, g)
			e.endStmt()
			e.line("if (%s) {", gv)
			e.indent++
		}
		e.line("%s = true;", done)
		body(arm)
		if arm.Guard != nil {
			e.indent--
			e.line("}")
		}
		e.popScope()
		e.indent--
		e.line("}")
	}
}

func (e *Emitter) matchStmt(m *MatchExpr) {
	e.matchArms(m, func(arm *MatchArm) {
		if arm.Block != nil {
			e.block(arm.Block, false)
			return
		}
		e.beginStmt()
		code, _ := e.expr(arm.Body)
		if code != "" && arm.Body.Type() == TVoid {
			e.line("%s;", code)
		}
		e.endStmt()
	})
}

func (e *Emitter) matchValue(m *MatchExpr) (string, bool) {
	t := m.Type()
	e.tempN++
	result := fmt.Sprintf("r%d", e.tempN)
	e.line("%s %s;", t.CName(), result)
	e.matchArms(m, func(arm *MatchArm) {
		e.beginStmt()
		code, owned := e.expr(arm.Body)
		v := e.consume(code, owned, t)
		e.line("%s = %s;", result, v)
		e.endStmt()
	})
	if needsRC(t) {
		e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{result, t})
		return result, true
	}
	return result, false
}
