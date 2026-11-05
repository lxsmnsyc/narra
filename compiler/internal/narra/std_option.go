package narra

import (
	"fmt"
	"strings"
)

// std/option is Narra source (std/src/option.nra): map, andThen,
// withDefault, orElse, isSome, isNone and okOr are generic functions with
// effect variables. The helpers below serve std/result, which is still a
// builtin module.

func init() {
	registerSourceModule("Option", "std/option")
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
	before := len(c.errs)
	t := c.expr(arg, want)
	ft, ok := t.(*FuncType)
	if !ok {
		if t != nil && t != TVoid {
			c.errorf(arg.position(), "%s needs a function as argument %d, found %s", e.Builtin, i+1, t)
		}
		return nil
	}
	if ft.Result == TVoid {
		// An arrow whose body failed to check also ends up void; its error is
		// already reported.
		if len(c.errs) == before {
			c.errorf(arg.position(), "%s needs a function that returns a value", e.Builtin)
		}
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
