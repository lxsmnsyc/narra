package narra

import "fmt"

func init() {
	registerModule(&Module{
		Path: "std/char",
		Funcs: map[string]*Builtin{
			// code(c): the Unicode scalar value.
			"code": {
				Name: "Char.code", Params: []Type{TChar}, Result: TInt,
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return fmt.Sprintf("((int64_t)%s)", e.argCode(x, 0)), false
				},
			},
			// fromCode(n): None for surrogates, negatives and values above 0x10FFFF.
			"fromCode": {
				Name: "Char.fromCode",
				Check: func(c *Checker, x *CallExpr, want Type) Type {
					return checkFixed(c, x, []Type{TInt}, preludeEnum(c, "Option", TChar))
				},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					ot := x.Type().(*EnumType)
					some, none := variantOf(ot, "Some"), variantOf(ot, "None")
					n := e.temp(TInt, e.argCode(x, 0))
					c := e.temp(TChar, fmt.Sprintf("(uint32_t)%s", n))
					return ioEither(e, ot, fmt.Sprintf("%s >= 0 && %s <= 0x10FFFF && !(%s >= 0xD800 && %s <= 0xDFFF)", n, n, n, n), some, []string{c}, none, nil), true
				},
			},
		},
	})
}
