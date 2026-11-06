package narra

import "fmt"

// std/coroutine is Narra source (std/src/coroutine.nra): the Step enum and
// two resume functions that Co.resume calls by argument count (moduleCall).
// The primitives are in std/internal/coroutine: next resumes a generator,
// resume a coroutine; both return Some(value), or None when the generator is
// done or panicked (test builds only), told apart by panicked and
// panicMessage. See SPEC.md section 8 and D79.

func init() {
	registerSourceModule("", "std/coroutine")
	registerModule(&Module{
		Path: "std/internal/coroutine",
		Funcs: map[string]*Builtin{
			"next": {
				Name: "Co.next", Check: checkCoResume, Mutable: []bool{true}, Effects: []string{"alloc"},
				Emit: emitCoResume,
			},
			"resume": {
				Name: "Co.resume", Check: checkCoResume, Mutable: []bool{true, false}, Effects: []string{"alloc"},
				Emit: emitCoResume,
			},
			"panicked": {
				Name: "Co.panicked", Params: []Type{}, Result: TBool,
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return "narra_co_panicked", false
				},
			},
			"panicMessage": {
				Name: "Co.panicMessage", Params: []Type{}, Result: TStr, Effects: []string{"alloc"},
				Emit: func(e *Emitter, x *CallExpr, w string) (string, bool) {
					return e.ownedTemp(TStr, "lstr_from_cstr(narra_panic_text)"), true
				},
			},
		},
	})
}

// checkCoResume checks Co.next(mutable g) and Co.resume(mutable g, input):
// the result is Option<Out>.
func checkCoResume(c *Checker, e *CallExpr, want Type) Type {
	n := 1
	if e.Builtin == "Co.resume" {
		n = 2
	}
	if !c.argCount(e, n) {
		return TVoid
	}
	t := c.expr(e.Args[0].Value, nil)
	gt, ok := t.(*GeneratorType)
	if !ok || (gt.In != nil) != (n == 2) {
		// Only std/coroutine calls these, with the right arguments.
		panic("internal error: " + e.Builtin + " needs a " + map[int]string{1: "Generator", 2: "Coroutine"}[n])
	}
	c.checkMutableArg(e.Args[0].Value)
	if n == 2 {
		c.expectType(e.Args[1].Value.position(), gt.In, c.expr(e.Args[1].Value, gt.In))
	}
	c.checkExclusive(e, []int{0})
	return preludeEnum(c, "Option", gt.Elem)
}

// emitCoResume emits a resume through narra_co_next. Only a test build
// catches a panic of the coroutine.
func emitCoResume(e *Emitter, x *CallExpr, w string) (string, bool) {
	opt := x.Type().(*EnumType)
	gt := x.Args[0].Value.Type().(*GeneratorType)
	in := "NULL"
	if gt.In != nil {
		code, _ := e.expr(x.Args[1].Value)
		in = "&" + e.temp(gt.In, code)
	}
	g := e.lvaluePtr(x.Args[0].Value)
	e.tempN++
	out := e.declareUnset(gt.Elem.CName(), fmt.Sprintf("t%d", e.tempN))
	e.tempN++
	status := e.declareUnset("int", fmt.Sprintf("t%d", e.tempN))
	guard := "false"
	if e.tests {
		guard = "true"
	}
	e.line("%s = narra_co_next(%s, &%s, %s, %s);", status, g, out, in, guard)
	return ioEither(e, opt, status+" > 0", variantOf(opt, "Some"), []string{out}, variantOf(opt, "None"), nil), true
}
