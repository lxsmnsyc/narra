package narra

import (
	_ "embed"
	"fmt"
)

// std/test: assertions for test blocks (SPEC.md section 13, D65).
//
// Every function may be called only inside a `test` block, arrow functions
// inside it included. Test blocks are not part of a non-test build, so no
// std/test code runs there. The file is not named std_test.go because Go
// would treat that name as a test file.

//go:embed runtime_test.h
var runtimeTestH string

func init() {
	registerModule(&Module{
		Path:    "std/test",
		Runtime: runtimeTestH,
		Funcs: map[string]*Builtin{
			// eq(actual, expected): panics with "expected <expected>, got <actual>".
			"eq": {Name: "Test.eq", Effects: []string{"alloc"}, Check: checkTestCompare, Emit: emitTestCompare},
			// ne(actual, other): panics with "expected a value other than <other>".
			"ne": {Name: "Test.ne", Effects: []string{"alloc"}, Check: checkTestCompare, Emit: emitTestCompare},
			// panics(f): panics unless calling f panics.
			"panics": {Name: "Test.panics", Effects: []string{"io"}, Check: checkTestPanics, Emit: emitTestPanics},
			// panicsWith(f, message): panics unless calling f panics with a
			// message that contains message.
			"panicsWith": {Name: "Test.panicsWith", Effects: []string{"io"}, Check: checkTestPanics, Emit: emitTestPanics},
			// golden(name, actual): compares actual with <testdata>/<name>.golden;
			// NARRA_UPDATE_GOLDEN=1 writes the file instead (D66).
			"golden": {Name: "Test.golden", Effects: []string{"io"}, Check: checkTestGolden, Emit: emitTestGolden},
		},
	})
}

func checkTestGolden(c *Checker, e *CallExpr, want Type) Type {
	testOnly(c, e)
	return checkFixed(c, e, []Type{TStr, TStr}, TVoid)
}

func emitTestGolden(e *Emitter, x *CallExpr, w string) (string, bool) {
	name := e.temp(TStr, e.argCode(x, 0))
	actual := e.temp(TStr, e.argCode(x, 1))
	return fmt.Sprintf("narra_test_golden(%s, %s, %s, %s)", name, actual, cString(e.prog.GoldenDir), w), false
}

// testOnly reports a std/test call outside a test block.
func testOnly(c *Checker, e *CallExpr) {
	if !c.inTest {
		c.errorf(e.P, "%s can only be called inside a `test` block", e.Builtin)
	}
}

// checkTestCompare checks Test.eq and Test.ne: two values of one type that
// implements Equatable and Show.
func checkTestCompare(c *Checker, e *CallExpr, want Type) Type {
	testOnly(c, e)
	if !c.argCount(e, 2) {
		return TVoid
	}
	for _, a := range e.Args {
		if a.Mutable {
			c.errorf(e.P, "%s does not change its arguments; remove `mutable`", e.Builtin)
		}
	}
	t := c.expr(e.Args[0].Value, nil)
	c.expectType(e.Args[1].Value.position(), t, c.expr(e.Args[1].Value, t))
	if t == TVoid {
		return TVoid
	}
	for _, name := range []string{"Equatable", "Show"} {
		if !c.implements(t, c.lookupTrait(name)) {
			c.errorf(e.P, "%s does not implement %s", t, name)
		}
	}
	return TVoid
}

func emitTestCompare(e *Emitter, x *CallExpr, w string) (string, bool) {
	t := x.Args[0].Value.Type()
	a := e.temp(t, e.argCode(x, 0))
	b := e.temp(t, e.argCode(x, 1))
	eq, show := e.useEq(t), e.useShow(t)
	if x.Builtin == "Test.eq" {
		e.line("if (!%s(%s, %s)) narra_test_fail_eq(%s(%s), %s(%s), %s);", eq, a, b, show, b, show, a, w)
	} else {
		e.line("if (%s(%s, %s)) narra_test_fail_ne(%s(%s), %s);", eq, a, b, show, b, w)
	}
	return "", false
}

// checkTestPanics checks Test.panics(f) and Test.panicsWith(f, message).
func checkTestPanics(c *Checker, e *CallExpr, want Type) Type {
	testOnly(c, e)
	params := []Type{c.funcOf(nil, TVoid)}
	if e.Builtin == "Test.panicsWith" {
		params = append(params, TStr)
	}
	return checkFixed(c, e, params, TVoid)
}

// emitTestPanics runs the closure in a child process (runtime_test.h).
func emitTestPanics(e *Emitter, x *CallExpr, w string) (string, bool) {
	ft := x.Args[0].Value.Type()
	f := e.temp(ft, e.argCode(x, 0))
	want := "NULL"
	if len(x.Args) == 2 {
		want = e.argCode(x, 1)
	}
	e.tempN++
	p := fmt.Sprintf("t%d", e.tempN)
	e.line("narra_test_proc %s;", p)
	e.line("if (narra_test_fork(&%s, %s)) {", p, w)
	e.line("  %s.fn(%s.env);", f, f)
	e.line("  narra_test_returned();")
	e.line("}")
	return fmt.Sprintf("narra_test_expect(&%s, %s, %s)", p, want, w), false
}
