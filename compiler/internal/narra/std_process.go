package narra

import "fmt"

func init() {
	registerModule(&Module{
		Path:    "std/process",
		Runtime: ioRuntime("process"),
		Funcs: map[string]*Builtin{
			// run(command, args): starts the command (found through PATH)
			// with the arguments, waits, and returns #[exit code, stdout,
			// stderr]. Err means the command could not be started. The
			// command's standard input is closed.
			"run": {
				Name: "Process.run", Effects: []string{"io", "alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					out := c.tupleOf([]Type{TInt, TStr, TStr})
					return checkFixed(c, x, []Type{TStr, c.arrayOf(TStr)}, resultOf(c, out))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					rt := x.Type().(*EnumType)
					tuple := variantOf(rt, "Ok").Fields[0].T.(*StructType)
					cmd := e.argCode(x, 0)
					at := x.Args[1].Value.Type().(*ArrayType)
					args := e.temp(at, e.argCode(x, 1))
					code := e.temp(TInt, "0")
					out, errout, bad := stringOut(e), stringOut(e), stringOut(e)
					ok := e.temp(TBool, fmt.Sprintf("narra_proc_run(%s, %s->items, %s->len, &%s, &%s, &%s, &%s)",
						cmd, args, args, code, out, errout, bad))
					value := fmt.Sprintf("((%s){.%s = %s, .%s = %s, .%s = %s})", tuple.CName(),
						fieldName("0"), code, fieldName("1"), out, fieldName("2"), errout)
					return ioResult(e, rt, ok, value, bad), true
				},
			},
		},
	})
}
