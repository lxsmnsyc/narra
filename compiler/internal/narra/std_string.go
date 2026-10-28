package narra

import "fmt"

func init() {
	registerModule(&Module{
		Path: "std/string",
		Funcs: map[string]*Builtin{
			"length": {
				Name: "Str.length", Params: []Type{TStr}, Result: TInt,
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return fmt.Sprintf("(%s)->len", e.argCode(x, 0)), false
				},
			},
			"concat": {
				Name: "Str.concat", Params: []Type{TStr, TStr}, Result: TStr, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					a := e.argCode(x, 0)
					return e.ownedTemp(TStr, fmt.Sprintf("lstr_concat(%s, %s)", a, e.argCode(x, 1))), true
				},
			},
			"fromChar": {
				Name: "Str.fromChar", Params: []Type{TChar}, Result: TStr, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return e.ownedTemp(TStr, fmt.Sprintf("lstr_from_char(%s)", e.argCode(x, 0))), true
				},
			},
			"repeat": {
				Name: "Str.repeat", Params: []Type{TStr, TInt}, Result: TStr, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					a := e.argCode(x, 0)
					return e.ownedTemp(TStr, fmt.Sprintf("lstr_repeat(%s, %s, %s)", a, e.argCode(x, 1), w)), true
				},
			},
		},
	})
}
