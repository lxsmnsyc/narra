package narra

import "fmt"

// std/box: Box.new.

func init() {
	registerNamespaceModule("Box", "std/box", map[string]*Builtin{
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
