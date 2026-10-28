package narra

import (
	"fmt"
	"strings"
)

func (e *Emitter) call(x *CallExpr) (string, bool) {
	w := e.where(x.P)
	switch {
	case x.Ctor != nil:
		v := x.Ctor
		var parts []string
		for i, a := range x.Args {
			code, owned := e.expr(a.Value)
			c := e.consume(code, owned, v.Fields[i].T)
			parts = append(parts, fmt.Sprintf(".%s = %s", fieldName(v.Fields[i].Name), c))
		}
		init := fmt.Sprintf("((%s){.tag = %d", v.Enum.CName(), v.Index)
		if len(parts) > 0 {
			init += fmt.Sprintf(", .u.%s = {%s}", variantName(v), strings.Join(parts, ", "))
		}
		init += "})"
		return e.ownedTemp(v.Enum, init), true
	case x.Builtin != "":
		return e.builtin(x, w)
	case x.Func != nil:
		var args []string
		for i, a := range x.Args {
			if x.Func.Params[i].Mutable {
				args = append(args, e.lvaluePtr(a.Value))
				continue
			}
			code, _ := e.expr(a.Value)
			args = append(args, code)
		}
		callCode := fmt.Sprintf("f_%s(%s)", x.Func.Name, strings.Join(args, ", "))
		if x.Func.ResultT == TVoid {
			return callCode, false
		}
		return e.ownedTemp(x.Func.ResultT, callCode), true
	}
	panic("emit: unresolved call")
}

// builtin emits a call to a registered builtin.
func (e *Emitter) builtin(x *CallExpr, w string) (string, bool) {
	b := builtinsByName[x.Builtin]
	if b == nil || b.Emit == nil {
		panic("emit: unknown builtin " + x.Builtin)
	}
	return b.Emit(e, x, w)
}

// argCode emits argument i of a builtin call and returns its C code.
func (e *Emitter) argCode(x *CallExpr, i int) string {
	code, _ := e.expr(x.Args[i].Value)
	return code
}
