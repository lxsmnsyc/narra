package narra

import "fmt"

// Builtins that need no import: print, assert, panic.

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
