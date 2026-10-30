package narra

import "fmt"

// Builtins that need no import: print, assert, panic, and the Array and Range namespaces.

func init() {
	registerPrelude("print", &Builtin{
		Params: []Type{TStr}, Result: TVoid, Effects: []string{"io"},
		Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
			return fmt.Sprintf("narra_print(%s)", e.argCode(x, 0)), false
		},
	})
	registerPrelude("assert", &Builtin{
		Params: []Type{TBool}, Result: TVoid,
		Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
			return fmt.Sprintf("narra_assert(%s, %s)", e.argCode(x, 0), w), false
		},
	})
	registerPrelude("panic", &Builtin{
		Params: []Type{TStr}, Result: TVoid,
		Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
			return fmt.Sprintf("narra_panic_str(%s, %s)", e.argCode(x, 0), w), false
		},
	})
	// Range.upTo, Range.between and Array.values are only valid as the source of
	// `for ... of`, which the statement checker and emitter handle directly.
	registerNamespace("Range", map[string]*Builtin{
		"upTo":    {Check: checkRangeUpTo},
		"between": {Check: checkRangeBetween},
	})
	registerNamespace("Box", map[string]*Builtin{
		"new": {
			Check: checkBoxNew, Effects: []string{"alloc"},
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
				bt := x.Type().(*BoxType)
				code, owned := e.expr(x.Args[0].Value)
				v := e.consume(code, owned, bt.Elem)
				return e.ownedTemp(bt, fmt.Sprintf("%s_new(%s)", boxCName(bt), v)), true
			},
		},
	})
	registerNamespace("Array", map[string]*Builtin{
		"length": {
			Check: checkArrayLength,
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
				return fmt.Sprintf("(%s)->len", e.argCode(x, 0)), false
			},
		},
		"values": {Check: checkArrayLength},
		"push": {
			Check: checkArrayPush, Mutable: []bool{true, false},
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
				at := x.Args[0].Value.Type().(*ArrayType)
				code, owned := e.expr(x.Args[1].Value)
				v := e.consume(code, owned, at.Elem)
				p := e.lvaluePtr(x.Args[0].Value)
				return fmt.Sprintf("%s_push(%s, %s)", arrName(at), p, v), false
			},
		},
		"filled": {
			Check: checkArrayFilled, Effects: []string{"alloc"},
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
				at := x.Type().(*ArrayType)
				v := e.argCode(x, 0)
				return e.ownedTemp(at, fmt.Sprintf("%s_filled(%s, %s, %s)", arrName(at), v, e.argCode(x, 1), w)), true
			},
		},
	})
}

// argCount reports a wrong argument count for a builtin that checks its own
// arguments, and checks the arguments loosely so their errors still show.
func (c *Checker) argCount(e *CallExpr, n int) bool {
	if len(e.Args) != n {
		c.errorf(e.P, "%s takes %d arguments, found %d", e.Builtin, n, len(e.Args))
		c.argsLoose(e)
		return false
	}
	return true
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

// checkBoxNew checks Box.new.
func checkBoxNew(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 1) {
		return TVoid
	}
	var elem Type
	if bt, ok := want.(*BoxType); ok {
		elem = bt.Elem
	}
	t := c.expr(e.Args[0].Value, elem)
	return c.boxOf(t)
}

// checkArrayLength checks Array.length and Array.values.
func checkArrayLength(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 1) {
		return TVoid
	}
	t := c.expr(e.Args[0].Value, nil)
	at, ok := t.(*ArrayType)
	if !ok {
		c.errorf(e.P, "%s needs an array, found %s", e.Builtin, t)
		return TVoid
	}
	if e.Builtin == "Array.length" {
		return TInt
	}
	return at
}

func checkArrayPush(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	if !e.Args[0].Mutable {
		c.errorf(e.P, "Array.push changes the array; write Array.push(mutable items, value)")
	}
	t := c.expr(e.Args[0].Value, nil)
	at, ok := t.(*ArrayType)
	if !ok {
		c.errorf(e.P, "Array.push needs an array, found %s", t)
		return TVoid
	}
	c.checkMutableArg(e.Args[0].Value)
	c.expectType(e.Args[1].Value.position(), at.Elem, c.expr(e.Args[1].Value, at.Elem))
	c.checkExclusive(e, []int{0})
	return TVoid
}

func checkArrayFilled(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	var elem Type
	if at, ok := want.(*ArrayType); ok {
		elem = at.Elem
	}
	t := c.expr(e.Args[0].Value, elem)
	c.expectType(e.Args[1].Value.position(), TInt, c.expr(e.Args[1].Value, TInt))
	return c.arrayOf(t)
}
