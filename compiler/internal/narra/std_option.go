package narra

import (
	"fmt"
	"strings"
)

// std/option: Option.map, andThen, withDefault, orElse, isSome, isNone and
// okOr. The callbacks are non-escaping function parameters with an effect
// variable: a call has the effects of its callback (Builtin.Callbacks).
// None of these returns early from the caller; `is` and `match` stay the
// main style for error paths.

func init() {
	registerNamespaceModule("Option", "std/option", map[string]*Builtin{
		"map":         {Check: checkOptionMap, Effects: []string{"alloc"}, Callbacks: []int{1}, Emit: emitOptionMap},
		"andThen":     {Check: checkOptionAndThen, Effects: []string{}, Callbacks: []int{1}, Emit: emitOptionAndThen},
		"withDefault": {Check: checkOptionWithDefault, Effects: []string{}, Emit: emitOptionWithDefault},
		"orElse":      {Check: checkOptionOrElse, Effects: []string{}, Callbacks: []int{1}, Emit: emitOptionOrElse},
		"isSome":      {Check: checkOptionIs, Effects: []string{}, Emit: emitOptionIs},
		"isNone":      {Check: checkOptionIs, Effects: []string{}, Emit: emitOptionIs},
		"okOr":        {Check: checkOptionOkOr, Effects: []string{"alloc"}, Emit: emitOptionOkOr},
	})
}

// preludeEnumArg returns t as an instance of the prelude enum name (Option
// or Result), or nil.
func (c *Checker) preludeEnumArg(t Type, name string) *EnumType {
	et, ok := t.(*EnumType)
	if !ok || et.Name != name || len(et.Args) == 0 {
		return nil
	}
	if g := c.generics[name]; g == nil || g.mod == nil || g.mod.path != et.Module {
		return nil
	}
	return et
}

// preludeArg checks argument i and returns it as an instance of the prelude
// enum name, or nil after reporting the mismatch.
func preludeArg(c *Checker, e *CallExpr, i int, name string) *EnumType {
	t := c.expr(e.Args[i].Value, nil)
	if et := c.preludeEnumArg(t, name); et != nil {
		return et
	}
	if t == nil || t == TVoid {
		return nil
	}
	if name == "Option" {
		c.errorf(e.Args[i].Value.position(), "%s needs an Option<T>, found %s", e.Builtin, t)
	} else {
		c.errorf(e.Args[i].Value.position(), "%s needs a Result<T, E>, found %s", e.Builtin, t)
	}
	return nil
}

// checkCallback checks argument i, a function the builtin calls with params.
// An arrow function gets the parameter types as its hint and a stack
// environment, because the builtin never keeps the function. result is the
// required result type, or nil when the callback decides it. It returns the
// callback's result type, or nil after an error.
func checkCallback(c *Checker, e *CallExpr, i int, params []Type, result Type) Type {
	arg := e.Args[i].Value
	if e.Args[i].Mutable {
		c.errorf(e.P, "%s does not change its arguments; remove `mutable`", e.Builtin)
	}
	var want Type
	switch v := arg.(type) {
	case *ArrowExpr:
		clo := closureOf(v)
		clo.Stack = true
		clo.Hint = &FuncType{Params: params, Result: result}
	case *Ident:
		c.cl().allowed = v
	}
	if result != nil {
		want = &FuncType{Params: params, Result: result}
	}
	t := c.expr(arg, want)
	ft, ok := t.(*FuncType)
	if !ok {
		if t != nil && t != TVoid {
			c.errorf(arg.position(), "%s needs a function as argument %d, found %s", e.Builtin, i+1, t)
		}
		return nil
	}
	if ft.Result == TVoid {
		c.errorf(arg.position(), "%s needs a function that returns a value", e.Builtin)
		return nil
	}
	if result != nil {
		c.expectType(arg.position(), c.funcOf(params, result), ft)
		return result
	}
	c.expectType(arg.position(), c.funcOf(params, ft.Result), ft)
	return ft.Result
}

// callbackReturns checks that a callback result r is an instance of the
// prelude enum name. It returns the instance, or nil after an error.
func callbackReturns(c *Checker, e *CallExpr, r Type, name string) *EnumType {
	if r == nil {
		return nil
	}
	if et := c.preludeEnumArg(r, name); et != nil {
		return et
	}
	c.errorf(e.Args[1].Value.position(), "%s needs a function that returns %s, found one that returns %s", e.Builtin, name, r)
	return nil
}

func checkOptionMap(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	o := preludeArg(c, e, 0, "Option")
	if o == nil {
		c.expr(e.Args[1].Value, nil)
		return TVoid
	}
	u := checkCallback(c, e, 1, []Type{o.Args[0]}, nil)
	if u == nil {
		return TVoid
	}
	return preludeEnum(c, "Option", u)
}

func checkOptionAndThen(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	o := preludeArg(c, e, 0, "Option")
	if o == nil {
		c.expr(e.Args[1].Value, nil)
		return TVoid
	}
	if r := callbackReturns(c, e, checkCallback(c, e, 1, []Type{o.Args[0]}, nil), "Option"); r != nil {
		return r
	}
	return TVoid
}

func checkOptionWithDefault(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	o := preludeArg(c, e, 0, "Option")
	if o == nil {
		c.expr(e.Args[1].Value, nil)
		return TVoid
	}
	c.expectType(e.Args[1].Value.position(), o.Args[0], c.expr(e.Args[1].Value, o.Args[0]))
	return o.Args[0]
}

func checkOptionOrElse(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	o := preludeArg(c, e, 0, "Option")
	if o == nil {
		c.expr(e.Args[1].Value, nil)
		return TVoid
	}
	checkCallback(c, e, 1, nil, o)
	return o
}

func checkOptionIs(c *Checker, e *CallExpr, want Type) Type {
	if c.argCount(e, 1) {
		preludeArg(c, e, 0, "Option")
	}
	return TBool
}

func checkOptionOkOr(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	o := preludeArg(c, e, 0, "Option")
	var hint Type
	if w, ok := want.(*EnumType); ok && w.Name == "Result" && len(w.Args) == 2 {
		hint = w.Args[1]
	}
	et := c.expr(e.Args[1].Value, hint)
	if o == nil {
		return TVoid
	}
	return preludeEnum(c, "Result", o.Args[0], et)
}

// enumArg emits argument i into a temporary and returns its name. The value
// is borrowed: the argument's own owner keeps it alive for the statement.
func enumArg(e *Emitter, x *CallExpr, i int) (string, *EnumType) {
	t := x.Args[i].Value.Type().(*EnumType)
	return e.temp(t, e.argCode(x, i)), t
}

// funcArg is enumArg for a callback.
func funcArg(e *Emitter, x *CallExpr, i int) (string, *FuncType) {
	ft := x.Args[i].Value.Type().(*FuncType)
	return e.temp(ft, e.argCode(x, i)), ft
}

// payloadOf reads the first field of variant v from the borrowed value o.
func payloadOf(o string, v *Variant) string {
	return fmt.Sprintf("%s.u->%s.%s", o, variantName(v), fieldName(v.Fields[0].Name))
}

// retained copies a borrowed value of type t into a new reference.
func retained(e *Emitter, code string, t Type) string {
	return e.consume(code, false, t)
}

// zeroResult declares the owned temporary a branching builtin fills.
func zeroResult(e *Emitter, t Type) string {
	return e.ownedTemp(t, fmt.Sprintf("((%s){0})", t.CName()))
}

// branch emits `if (o is variant v) { body } else { other }`.
func branch(e *Emitter, o string, v *Variant, body, other func()) {
	e.line("if (%s.tag == %d) {", o, v.Index)
	e.indent++
	body()
	e.indent--
	e.line("} else {")
	e.indent++
	other()
	e.indent--
	e.line("}")
}

// callFn calls function value f with args and returns the owned result in a
// block-local temporary, to be moved into a payload or the result.
func callFn(e *Emitter, f string, ft *FuncType, args ...string) string {
	all := append([]string{f + ".env"}, args...)
	return e.temp(ft.Result, fmt.Sprintf("%s.fn(%s)", f, strings.Join(all, ", ")))
}

// assignCopy sets r to a new reference to the borrowed value v.
func assignCopy(e *Emitter, r, v string, t Type) {
	e.line("%s = %s;", r, v)
	if s := retain(t, r); s != "" {
		e.line("%s", s)
	}
}

func emitOptionMap(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	f, ft := funcArg(e, x, 1)
	rt := x.Type().(*EnumType)
	r := zeroResult(e, rt)
	some := variantOf(ot, "Some")
	branch(e, o, some, func() {
		v := callFn(e, f, ft, payloadOf(o, some))
		e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, "Some"), []string{v}))
	}, func() {
		e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, "None"), nil))
	})
	return r, true
}

func emitOptionAndThen(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	f, ft := funcArg(e, x, 1)
	rt := x.Type().(*EnumType)
	r := zeroResult(e, rt)
	some := variantOf(ot, "Some")
	branch(e, o, some, func() {
		e.line("%s = %s;", r, callFn(e, f, ft, payloadOf(o, some)))
	}, func() {
		e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, "None"), nil))
	})
	return r, true
}

func emitOptionWithDefault(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	t := x.Type()
	code, _ := e.expr(x.Args[1].Value)
	d := e.temp(t, code)
	r := zeroResult(e, t)
	some := variantOf(ot, "Some")
	branch(e, o, some, func() {
		assignCopy(e, r, payloadOf(o, some), t)
	}, func() {
		assignCopy(e, r, d, t)
	})
	return r, true
}

func emitOptionOrElse(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	f, ft := funcArg(e, x, 1)
	r := zeroResult(e, ot)
	branch(e, o, variantOf(ot, "Some"), func() {
		assignCopy(e, r, o, ot)
	}, func() {
		e.line("%s = %s;", r, callFn(e, f, ft))
	})
	return r, true
}

func emitOptionIs(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	name := "Some"
	if x.Builtin == "Option.isNone" {
		name = "None"
	}
	return fmt.Sprintf("(%s.tag == %d)", o, variantOf(ot, name).Index), false
}

func emitOptionOkOr(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	et := x.Args[1].Value.Type()
	code, _ := e.expr(x.Args[1].Value)
	errv := e.temp(et, code)
	rt := x.Type().(*EnumType)
	r := zeroResult(e, rt)
	some := variantOf(ot, "Some")
	branch(e, o, some, func() {
		v := retained(e, payloadOf(o, some), some.Fields[0].T)
		e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, "Ok"), []string{v}))
	}, func() {
		v := retained(e, errv, et)
		e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, "Err"), []string{v}))
	})
	return r, true
}
