package narra

import "fmt"

func init() {
	registerModule(&Module{
		Path:    "std/env",
		Runtime: ioRuntime("env"),
		Funcs: map[string]*Builtin{
			// get(name): the value of an environment variable, or None.
			"get": {
				Name: "Env.get", Effects: []string{"io", "alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TStr}, preludeEnum(c, "Option", TStr))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					ot := x.Type().(*EnumType)
					some, none := variantOf(ot, "Some"), variantOf(ot, "None")
					name := e.argCode(x, 0)
					v := e.temp(TStr, fmt.Sprintf("narra_env_get(%s)", name))
					code := fmt.Sprintf("(%[1]s ? ((%[2]s){.tag = %[3]d, .u.%[4]s.%[5]s = %[1]s}) : ((%[2]s){.tag = %[6]d}))",
						v, ot.CName(), some.Index, variantName(some), fieldName("value"), none.Index)
					return e.ownedTemp(ot, code), true
				},
			},
			// args(): the program's arguments, as the entry function's
			// `args` parameter holds them (the program name comes first).
			"args": {
				Name: "Env.args", Effects: []string{"io", "alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, nil, c.arrayOf(TStr))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					at := x.Type().(*ArrayType)
					arr := e.ownedTemp(at, fmt.Sprintf("%s_new(narra_env_argc())", arrName(at)))
					e.line("for (int64_t i = 0, n = narra_env_argc(); i < n; i++) %s->items[%s->len++] = narra_env_arg(i);", arr, arr)
					return arr, true
				},
			},
		},
	})
}
