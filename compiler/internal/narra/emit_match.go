package narra

import (
	"fmt"
	"strings"
)

// Patterns read from a subject held in a C variable for the whole match or
// if statement. A test is a chain of conditions on C paths into the subject,
// joined with &&, so a field of a union member is read only after its tag
// is checked. A binding copies the value at its path and retains it, so the
// local owns one reference and is released with its scope.

// patternCond returns a C condition that is true when the value at path
// matches p. It may emit lines for literal values.
func (e *Emitter) patternCond(p *Pattern, path string) string {
	var conds []string
	e.patternConds(p, path, &conds)
	if len(conds) == 0 {
		return "true"
	}
	return strings.Join(conds, " && ")
}

func (e *Emitter) patternConds(p *Pattern, path string, conds *[]string) {
	switch p.Kind {
	case PatName:
		if p.Variant != nil {
			*conds = append(*conds, fmt.Sprintf("%s.tag == %d", path, p.Variant.Index))
		}
	case PatCtor:
		*conds = append(*conds, fmt.Sprintf("%s.tag == %d", path, p.Variant.Index))
		for i, a := range p.Args {
			e.patternConds(a, variantPath(path, p.Variant, i), conds)
		}
	case PatStruct:
		for _, fp := range p.Fields {
			e.patternConds(fp.Pattern, fmt.Sprintf("%s.%s", path, fieldName(fp.Name)), conds)
		}
	case PatTuple:
		st := p.T.(*StructType)
		for i, a := range p.Args {
			e.patternConds(a, fmt.Sprintf("%s.%s", path, fieldName(st.Fields[i].Name)), conds)
		}
	case PatLit:
		lit, _ := e.expr(p.Lit)
		if p.T == TStr {
			*conds = append(*conds, fmt.Sprintf("lstr_eq(%s, %s)", path, lit))
		} else {
			*conds = append(*conds, fmt.Sprintf("%s == %s", path, lit))
		}
	}
}

func variantPath(path string, v *Variant, i int) string {
	return fmt.Sprintf("%s.u->%s.%s", path, variantName(v), fieldName(v.Fields[i].Name))
}

// bindPattern declares the names a pattern binds, reading from path. Each
// local takes its own reference.
func (e *Emitter) bindPattern(p *Pattern, path string) {
	switch p.Kind {
	case PatName:
		if l := p.Local; l != nil {
			e.line("%s %s = %s;", l.T.CName(), l.CName, path)
			if r := retain(l.T, l.CName); r != "" {
				e.line("%s", r)
			}
			e.addLocal(l)
		}
	case PatCtor:
		for i, a := range p.Args {
			e.bindPattern(a, variantPath(path, p.Variant, i))
		}
	case PatStruct:
		for _, fp := range p.Fields {
			e.bindPattern(fp.Pattern, fmt.Sprintf("%s.%s", path, fieldName(fp.Name)))
		}
	case PatTuple:
		st := p.T.(*StructType)
		for i, a := range p.Args {
			e.bindPattern(a, fmt.Sprintf("%s.%s", path, fieldName(st.Fields[i].Name)))
		}
	}
}

// isCond emits `x is Pattern` outside the condition of an if.
func (e *Emitter) isCond(x *IsExpr) string {
	code, _ := e.expr(x.Subject)
	subj := e.temp(x.Subject.Type(), code)
	return "(" + e.patternCond(x.Pattern, subj) + ")"
}

// ifIs emits `if (x is Pattern) { ... }`. The subject is held for the whole
// statement, because the pattern's names read from it.
func (e *Emitter) ifIs(s *IfStmt, is *IsExpr) {
	t := is.Subject.Type()
	e.line("{")
	e.indent++
	e.pushScope(false)
	e.beginStmt()
	code, owned := e.expr(is.Subject)
	held := e.consume(code, owned, t)
	subj := e.temp(t, held)
	e.endStmt()
	e.addLocal(&Local{Name: "subject", T: t, CName: subj})
	e.beginStmt()
	cond := e.temp(TBool, e.patternCond(is.Pattern, subj))
	e.endStmt()
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

// matchArms emits an if-chain over the arms. body emits each arm's action.
func (e *Emitter) matchArms(m *MatchExpr, body func(arm *MatchArm)) {
	code, _ := e.expr(m.Subject)
	subj := e.temp(m.Subject.Type(), code)
	e.tempN++
	done := fmt.Sprintf("m%d", e.tempN)
	e.line("bool %s = false;", done)
	for _, arm := range m.Arms {
		p := arm.Pattern
		e.line("if (!%s) {", done)
		e.indent++
		e.beginStmt()
		cond := e.temp(TBool, e.patternCond(p, subj))
		e.endStmt()
		e.line("if (%s) {", cond)
		e.indent++
		e.pushScope(false)
		e.bindPattern(p, subj)
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
