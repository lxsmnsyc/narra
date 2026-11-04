package narra

import "fmt"

// std/range: Range.upTo and Range.between build a Range value, which is two
// ints. As the source of `for ... of` the statement emitter reads the bounds
// from the call directly; any other Range value is read from the struct.

func init() {
	registerNamespaceModule("Range", "std/range", map[string]*Builtin{
		"upTo": {Check: checkRangeUpTo,
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
				code, _ := e.expr(x.Args[0].Value)
				high := e.temp(TInt, code)
				return fmt.Sprintf("((narra_range){0, %s})", high), false
			}},
		"between": {Check: checkRangeBetween,
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
				first, _ := e.expr(x.Args[0].Value)
				low := e.temp(TInt, first)
				second, _ := e.expr(x.Args[1].Value)
				high := e.temp(TInt, second)
				return fmt.Sprintf("((narra_range){%s, %s})", low, high), false
			}},
	})
}

func checkRangeUpTo(c *Checker, e *CallExpr, want Type) Type {
	if c.argCount(e, 1) {
		c.expectType(e.Args[0].Value.position(), TInt, c.expr(e.Args[0].Value, TInt))
	}
	return TRange
}

func checkRangeBetween(c *Checker, e *CallExpr, want Type) Type {
	if c.argCount(e, 2) {
		c.expectType(e.Args[0].Value.position(), TInt, c.expr(e.Args[0].Value, TInt))
		c.expectType(e.Args[1].Value.position(), TInt, c.expr(e.Args[1].Value, TInt))
	}
	return TRange
}
