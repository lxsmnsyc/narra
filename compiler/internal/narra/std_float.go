package narra

import "fmt"

func init() {
	registerModule(&Module{
		Path: "std/float",
		Funcs: map[string]*Builtin{
			"toString": {
				Name: "Float.toString", Params: []Type{TF64}, Result: TStr, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return e.ownedTemp(TStr, fmt.Sprintf("lstr_from_f64(%s)", e.argCode(x, 0))), true
				},
			},
			"sqrt": {
				Name: "Float.sqrt", Params: []Type{TF64}, Result: TF64,
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return fmt.Sprintf("sqrt(%s)", e.argCode(x, 0)), false
				},
			},
			"fromInt": {
				Name: "Float.fromInt", Params: []Type{TInt}, Result: TF64,
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return fmt.Sprintf("((double)%s)", e.argCode(x, 0)), false
				},
			},
			"truncate": {
				Name: "Float.truncate", Params: []Type{TF64}, Result: TInt,
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return e.temp(TInt, fmt.Sprintf("narra_truncate(%s, %s)", e.argCode(x, 0), w)), false
				},
			},
		},
	})
}
