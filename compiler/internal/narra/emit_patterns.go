package narra

import (
	"fmt"
	"strings"
)

// destructure emits `const { a, b: c } = s` and `const [a, b] = t`. The
// value is held in a temporary for the statement; each bound name retains
// its part, and an owned value is released at the end of the statement.
func (e *Emitter) destructure(s *VarStmt) {
	e.beginStmt()
	code, _ := e.expr(s.Value)
	subj := e.temp(s.Pattern.T, code)
	e.bindPattern(s.Pattern, subj)
	e.endStmt()
}

// spreadLit emits `{ ...s, field: v }`. Given fields are consumed; the
// others are copied from s and retained, so the result owns every field.
func (e *Emitter) spreadLit(x *StructLit) (string, bool) {
	st := x.Struct
	code, _ := e.expr(x.Spread)
	src := e.temp(st, code)
	given := map[string]string{}
	for _, fi := range x.Fields {
		c, owned := e.expr(fi.Value)
		given[fi.Name] = e.consume(c, owned, fi.Value.Type())
	}
	var parts []string
	for _, f := range st.Fields {
		v, ok := given[f.Name]
		if !ok {
			v = fmt.Sprintf("%s.%s", src, fieldName(f.Name))
			if r := retain(f.T, v); r != "" {
				e.line("%s", r)
			}
		}
		parts = append(parts, fmt.Sprintf(".%s = %s", fieldName(f.Name), v))
	}
	return e.ownedTemp(st, fmt.Sprintf("((%s){%s})", st.CName(), strings.Join(parts, ", "))), true
}

// tupleLit emits `[a, b]` or `Tuple(a, b)` building a tuple.
func (e *Emitter) tupleLit(x *ArrayLit) (string, bool) {
	st := x.Tuple
	var parts []string
	for i, el := range x.Elems {
		code, owned := e.expr(el)
		v := e.consume(code, owned, st.Fields[i].T)
		parts = append(parts, fmt.Sprintf(".%s = %s", fieldName(st.Fields[i].Name), v))
	}
	return e.ownedTemp(st, fmt.Sprintf("((%s){%s})", st.CName(), strings.Join(parts, ", "))), true
}
