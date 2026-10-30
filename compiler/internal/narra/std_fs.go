package narra

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed runtime_io.h
var runtimeIO string

// ioRuntime returns the C for one module of runtime_io.h: the common helpers
// (which guard themselves against a second copy) plus the module's section.
func ioRuntime(section string) string {
	var common, own string
	for _, part := range strings.Split(runtimeIO, "// @section ")[1:] {
		name, body, _ := strings.Cut(part, "\n")
		switch strings.TrimSpace(name) {
		case "common":
			common = body
		case section:
			own = body
		}
	}
	return common + "\n" + own
}

// checkFixed checks the arguments of a builtin against params and returns
// result. It is the Check of builtins whose result type needs the checker.
func checkFixed(c *Checker, e *CallExpr, params []Type, result Type) Type {
	if !c.argCount(e, len(params)) {
		return result
	}
	for i, a := range e.Args {
		if a.Mutable {
			c.errorf(e.P, "%s does not change its arguments; remove `mutable`", e.Builtin)
		}
		c.expectType(a.Value.position(), params[i], c.expr(a.Value, params[i]))
	}
	return result
}

// preludeEnum instantiates a generic enum of the prelude, such as Result.
func preludeEnum(c *Checker, name string, args ...Type) Type {
	g := c.generics[name]
	if g == nil {
		return TVoid
	}
	return c.instantiate(g, args)
}

// resultOf is Result<ok, string>, the type of every fallible std/fs and
// std/process function.
func resultOf(c *Checker, ok Type) Type { return preludeEnum(c, "Result", ok, TStr) }

// variantOf finds a variant of an enum by name.
func variantOf(et *EnumType, name string) *Variant {
	for _, v := range et.Variants {
		if v.Name == name {
			return v
		}
	}
	panic("no variant " + name + " in " + et.String())
}

// ioResult builds a Result value from the pieces a runtime function filled in.
// okVal is moved into Ok when success is true, errVal into Err otherwise; the
// result is an owned temporary that releases whichever one it holds.
func ioResult(e *Emitter, rt *EnumType, success, okVal, errVal string) string {
	ok, bad := variantOf(rt, "Ok"), variantOf(rt, "Err")
	code := fmt.Sprintf("(%[1]s ? ((%[2]s){.tag = %[3]d, .u.%[4]s.%[5]s = %[6]s}) : ((%[2]s){.tag = %[7]d, .u.%[8]s.%[9]s = %[10]s}))",
		success, rt.CName(),
		ok.Index, variantName(ok), fieldName("value"), okVal,
		bad.Index, variantName(bad), fieldName("error"), errVal)
	return e.ownedTemp(rt, code)
}

// stringOut declares a string out-parameter for a runtime call.
func stringOut(e *Emitter) string { return e.temp(TStr, "NULL") }

func init() {
	registerModule(&Module{
		Path:    "std/fs",
		Runtime: ioRuntime("fs"),
		Funcs: map[string]*Builtin{
			// readText(path): the file's text, or Err when it cannot be
			// read or is not valid UTF-8.
			"readText": {
				Name: "Fs.readText", Effects: []string{"io", "alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TStr}, resultOf(c, TStr))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					path := e.argCode(x, 0)
					out, bad := stringOut(e), stringOut(e)
					ok := e.temp(TBool, fmt.Sprintf("narra_fs_read(%s, &%s, &%s)", path, out, bad))
					return ioResult(e, x.Type().(*EnumType), ok, out, bad), true
				},
			},
			// writeText(path, text): creates or replaces the file; Ok holds
			// the number of bytes written.
			"writeText": {
				Name: "Fs.writeText", Effects: []string{"io"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TStr, TStr}, resultOf(c, TInt))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					path := e.argCode(x, 0)
					text := e.argCode(x, 1)
					n := e.temp(TInt, "0")
					bad := stringOut(e)
					ok := e.temp(TBool, fmt.Sprintf("narra_fs_write(%s, %s, &%s, &%s)", path, text, n, bad))
					return ioResult(e, x.Type().(*EnumType), ok, n, bad), true
				},
			},
			"exists": {
				Name: "Fs.exists", Params: []Type{TStr}, Result: TBool, Effects: []string{"io"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return fmt.Sprintf("narra_fs_exists(%s)", e.argCode(x, 0)), false
				},
			},
			// listDir(path): the entry names, sorted, without "." and "..".
			"listDir": {
				Name: "Fs.listDir", Effects: []string{"io", "alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TStr}, resultOf(c, c.arrayOf(TStr)))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					rt := x.Type().(*EnumType)
					at := variantOf(rt, "Ok").Fields[0].T.(*ArrayType)
					path := e.argCode(x, 0)
					e.tempN++
					names := fmt.Sprintf("t%d", e.tempN)
					e.line("narra_fs_names %s = {0};", names)
					bad := stringOut(e)
					ok := e.temp(TBool, fmt.Sprintf("narra_fs_list(%s, &%s, &%s)", path, names, bad))
					// Move the names into an array; the Result owns it.
					arr := e.temp(at, "NULL")
					e.line("if (%s) {", ok)
					e.line("  %s = %s_new(%s.len);", arr, arrName(at), names)
					e.line("  for (int64_t i = 0; i < %s.len; i++) %s->items[%s->len++] = %s.items[i];", names, arr, arr, names)
					e.line("  free(%s.items);", names)
					e.line("}")
					return ioResult(e, rt, ok, arr, bad), true
				},
			},
		},
	})
}
