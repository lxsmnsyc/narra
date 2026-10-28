package narra

import "fmt"

func init() {
	registerModule(&Module{
		Path: "std/int",
		Funcs: map[string]*Builtin{
			"toString": {
				Name: "Int.toString", Params: []Type{TInt}, Result: TStr, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return e.ownedTemp(TStr, fmt.Sprintf("lstr_from_int(%s)", e.argCode(x, 0))), true
				},
			},
			"abs": {
				Name: "Int.abs", Params: []Type{TInt}, Result: TInt,
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return e.temp(TInt, fmt.Sprintf("narra_abs(%s, %s)", e.argCode(x, 0), w)), false
				},
			},
		},
	})
}
