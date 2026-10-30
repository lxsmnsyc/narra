package narra

import "fmt"

// std/array: Array.length, Array.values, Array.push and Array.filled.

func init() {
	registerNamespaceModule("Array", "std/array", map[string]*Builtin{
		"length": {
			Check: checkArrayLength,
			Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
				return fmt.Sprintf("(%s)->len", e.argCode(x, 0)), false
			},
		},
		"values": {Check: checkArrayLength},
		"push": {
			Check: checkArrayPush, Mutable: []bool{true, false}, Effects: []string{"alloc"},
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
