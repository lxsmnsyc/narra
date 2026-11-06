package narra

import "fmt"

// ioVoid builds a builtin that is one runtime call with no result.
func ioVoid(name string, params []Type, call string) *Builtin {
	return &Builtin{
		Name: name, Params: params, Result: TVoid, Effects: []string{"io"},
		Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
			args := ""
			if len(params) > 0 {
				args = e.argCode(x, 0)
			}
			return fmt.Sprintf("%s(%s)", call, args), false
		},
	}
}

func init() {
	registerModule(&Module{
		Path:    "std/io",
		Runtime: ioRuntime("io"),
		Funcs: map[string]*Builtin{
			// readLine(): the next line of stdin without its "\n" (a "\r"
			// stays), or None at the end of the input.
			"readLine": {
				Name: "Io.readLine", Effects: []string{"io", "alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, nil, resultOf(c, preludeEnum(c, "Option", TStr)))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					rt := x.Type().(*EnumType)
					ot := variantOf(rt, "Ok").Fields[0].T.(*EnumType)
					line, bad := stringOut(e), stringOut(e)
					ok := e.temp(TBool, fmt.Sprintf("narra_io_read_line(&%s, &%s)", line, bad))
					r := e.ownedTemp(rt, fmt.Sprintf("((%s){0})", rt.CName()))
					e.line("if (%s) {", ok)
					e.indent++
					// The Option moves into Ok, so it is a plain temporary.
					opt := e.temp(ot, fmt.Sprintf("((%s){0})", ot.CName()))
					e.line("if (%s) {", line)
					e.indent++
					e.line("%s = %s;", opt, e.enumAlloc(variantOf(ot, "Some"), []string{line}))
					e.indent--
					e.line("} else {")
					e.indent++
					e.line("%s = %s;", opt, e.enumAlloc(variantOf(ot, "None"), nil))
					e.indent--
					e.line("}")
					e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, "Ok"), []string{opt}))
					e.indent--
					e.line("} else {")
					e.indent++
					e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, "Err"), []string{bad}))
					e.indent--
					e.line("}")
					return r, true
				},
			},
			// readBytes(count): exactly count bytes of stdin.
			"readBytes": {
				Name: "Io.readBytes", Effects: []string{"io", "alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TInt}, resultOf(c, TStr))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					count := e.argCode(x, 0)
					out, bad := stringOut(e), stringOut(e)
					ok := e.temp(TBool, fmt.Sprintf("narra_io_read_exact(%s, &%s, &%s)", count, out, bad))
					return ioResult(e, x.Type().(*EnumType), ok, out, bad), true
				},
			},
			// readAll(): the rest of stdin.
			"readAll": {
				Name: "Io.readAll", Effects: []string{"io", "alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, nil, resultOf(c, TStr))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					out, bad := stringOut(e), stringOut(e)
					ok := e.temp(TBool, fmt.Sprintf("narra_io_read_all(&%s, &%s)", out, bad))
					return ioResult(e, x.Type().(*EnumType), ok, out, bad), true
				},
			},
			// write(text): stdout, no newline.
			"write": ioVoid("Io.write", []Type{TStr}, "narra_io_write"),
			// writeError(text): stderr, no newline.
			"writeError": ioVoid("Io.writeError", []Type{TStr}, "narra_io_write_error"),
			// flush(): flushes stdout.
			"flush": ioVoid("Io.flush", nil, "narra_io_flush"),
		},
	})
}
