package narra

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed runtime_string.h
var runtimeString string

// strFn builds a builtin with fixed parameter types whose Emit is one runtime
// call. When withPos, the panic position is the last argument.
func strFn(name string, params []Type, result Type, effects []string, call string, withPos bool, owned bool) *Builtin {
	return &Builtin{
		Name: name, Params: params, Result: result, Effects: effects,
		Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
			var args []string
			for i := range params {
				args = append(args, e.argCode(x, i))
			}
			if withPos {
				args = append(args, w)
			}
			code := fmt.Sprintf("%s(%s)", call, strings.Join(args, ", "))
			if owned {
				return e.ownedTemp(result, code), true
			}
			return code, false
		},
	}
}

func init() {
	registerModule(&Module{
		Path:    "std/string",
		Runtime: runtimeString,
		Funcs: map[string]*Builtin{
			// Positions are byte offsets (type int for now; see README).
			"slice":      strFn("Str.slice", []Type{TStr, TInt, TInt}, TStr, []string{"alloc"}, "narra_str_slice", true, true),
			"sliceFrom":  strFn("Str.sliceFrom", []Type{TStr, TInt}, TStr, []string{"alloc"}, "narra_str_slice_from", true, true),
			"byteAt":     strFn("Str.byteAt", []Type{TStr, TInt}, TInt, nil, "narra_str_byte_at", true, false),
			"charAt":     strFn("Str.charAt", []Type{TStr, TInt}, TChar, nil, "narra_str_char_at", true, false),
			"nextChar":   strFn("Str.nextChar", []Type{TStr, TInt}, TInt, nil, "narra_str_next_char", true, false),
			"charCount":  strFn("Str.charCount", []Type{TStr}, TInt, nil, "narra_str_char_count", false, false),
			"trim":       strFn("Str.trim", []Type{TStr}, TStr, []string{"alloc"}, "narra_str_trim", false, true),
			"startsWith": strFn("Str.startsWith", []Type{TStr, TStr}, TBool, nil, "narra_str_starts_with", false, false),
			"endsWith": strFn("Str.endsWith", []Type{TStr, TStr}, TBool, nil, "narra_str_ends_with", false, false),
			"contains": strFn("Str.contains", []Type{TStr, TStr}, TBool, nil, "narra_str_contains", false, false),
			// replace(s, from, to): every match of from, left to right and
			// non-overlapping. An empty from returns s unchanged.
			"replace": strFn("Str.replace", []Type{TStr, TStr, TStr}, TStr, []string{"alloc"}, "narra_str_replace", false, true),
			// join(parts, sep): the parts with sep between them.
			"join": {
				Name: "Str.join", Effects: []string{"alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{c.arrayOf(TStr), TStr}, TStr)
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					a := e.argCode(x, 0)
					return e.ownedTemp(TStr, fmt.Sprintf("narra_str_join(%s->items, %s->len, %s)", a, a, e.argCode(x, 1))), true
				},
			},
			// indexOf(s, needle): the first position of needle, or None.
			"indexOf": {
				Name: "Str.indexOf", Effects: []string{"alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TStr, TStr}, preludeEnum(c, "Option", TInt))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					ot := x.Type().(*EnumType)
					some, none := variantOf(ot, "Some"), variantOf(ot, "None")
					a := e.argCode(x, 0)
					p := e.temp(TInt, fmt.Sprintf("narra_str_index_of(%s, %s)", a, e.argCode(x, 1)))
					return ioEither(e, ot, p+" >= 0", some, []string{p}, none, nil), true
				},
			},
			// split(s, sep): the pieces between occurrences of sep; panics
			// when sep is empty.
			"split": {
				Name: "Str.split", Effects: []string{"alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TStr, TStr}, c.arrayOf(TStr))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					at := x.Type().(*ArrayType)
					a := e.argCode(x, 0)
					n := e.temp(TInt, "0")
					parts := e.temp(&ptrType{&ptrType{TStr}}, fmt.Sprintf("narra_str_split(%s, %s, &%s, %s)", a, e.argCode(x, 1), n, w))
					arr := e.ownedTemp(at, fmt.Sprintf("%s_new(%s)", arrName(at), n))
					e.line("for (int64_t i = 0; i < %s; i++) %s->items[%s->len++] = %s[i];", n, arr, arr, parts)
					e.line("narra_free(%s);", parts)
					return arr, true
				},
			},
			// chars(s): the characters of s.
			"chars": {
				Name: "Str.chars", Effects: []string{"alloc"},
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TStr}, c.arrayOf(TChar))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					at := x.Type().(*ArrayType)
					s := e.argCode(x, 0)
					arr := e.ownedTemp(at, fmt.Sprintf("%s_new(narra_str_char_count(%s))", arrName(at), s))
					e.line("for (int64_t i = 0; i < %s->len; i += narra_str_char_len(%s, i)) %s->items[%s->len++] = narra_str_decode(%s, i);", s, s, arr, arr, s)
					return arr, true
				},
			},
			// StringBuilder: builder() makes an empty one, append and
			// appendChar add to it, build copies out the string.
			"builder": strFn("Str.builder", []Type{}, TStringBuilder, []string{"alloc"}, "narra_sb_new", false, true),
			"build":   strFn("Str.build", []Type{TStringBuilder}, TStr, []string{"alloc"}, "narra_sb_build", false, true),
			"append": {
				Name: "Str.append", Params: []Type{TStringBuilder, TStr}, Mutable: []bool{true, false}, Result: TVoid, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					s := e.argCode(x, 1)
					return fmt.Sprintf("narra_sb_append(%s, %s)", e.lvaluePtr(x.Args[0].Value), s), false
				},
			},
			"appendChar": {
				Name: "Str.appendChar", Params: []Type{TStringBuilder, TChar}, Mutable: []bool{true, false}, Result: TVoid, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					c := e.argCode(x, 1)
					return fmt.Sprintf("narra_sb_append_char(%s, %s)", e.lvaluePtr(x.Args[0].Value), c), false
				},
			},
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
