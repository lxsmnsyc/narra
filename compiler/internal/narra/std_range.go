package narra

// std/range: Range.upTo and Range.between. Both are only valid as the source of
// `for ... of`, which the statement checker and emitter handle directly.

func init() {
	registerNamespaceModule("Range", "std/range", map[string]*Builtin{
		"upTo":    {Check: checkRangeUpTo},
		"between": {Check: checkRangeBetween},
	})
}

func checkRangeUpTo(c *Checker, e *CallExpr, want Type) Type {
	if c.argCount(e, 1) {
		c.expectType(e.Args[0].Value.position(), TInt, c.expr(e.Args[0].Value, TInt))
	}
	return &RangeType{}
}

func checkRangeBetween(c *Checker, e *CallExpr, want Type) Type {
	if c.argCount(e, 2) {
		c.expectType(e.Args[0].Value.position(), TInt, c.expr(e.Args[0].Value, TInt))
		c.expectType(e.Args[1].Value.position(), TInt, c.expr(e.Args[1].Value, TInt))
	}
	return &RangeType{}
}
