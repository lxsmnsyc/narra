package narra

// std/result: Result.map, mapErr, andThen, withDefault, ok and err. The
// callbacks follow std/option.

func init() {
	registerNamespaceModule("Result", "std/result", map[string]*Builtin{
		"map":         {Check: checkResultMap, Effects: []string{"alloc"}, Callbacks: []int{1}, Emit: emitResultMap},
		"mapErr":      {Check: checkResultMap, Effects: []string{"alloc"}, Callbacks: []int{1}, Emit: emitResultMap},
		"andThen":     {Check: checkResultAndThen, Effects: []string{"alloc"}, Callbacks: []int{1}, Emit: emitResultAndThen},
		"withDefault": {Check: checkResultWithDefault, Effects: []string{}, Emit: emitResultWithDefault},
		"ok":          {Check: checkResultToOption, Effects: []string{"alloc"}, Emit: emitResultToOption},
		"err":         {Check: checkResultToOption, Effects: []string{"alloc"}, Emit: emitResultToOption},
	})
}

// checkResultMap checks Result.map (the callback takes the Ok value) and
// Result.mapErr (it takes the Err value).
func checkResultMap(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	r := preludeArg(c, e, 0, "Result")
	if r == nil {
		c.expr(e.Args[1].Value, nil)
		return TVoid
	}
	if e.Builtin == "Result.mapErr" {
		f := checkCallback(c, e, 1, []Type{r.Args[1]}, nil)
		if f == nil {
			return TVoid
		}
		return preludeEnum(c, "Result", r.Args[0], f)
	}
	u := checkCallback(c, e, 1, []Type{r.Args[0]}, nil)
	if u == nil {
		return TVoid
	}
	return preludeEnum(c, "Result", u, r.Args[1])
}

func checkResultAndThen(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	r := preludeArg(c, e, 0, "Result")
	if r == nil {
		c.expr(e.Args[1].Value, nil)
		return TVoid
	}
	// The expected type of the call, when its error type is r's, gives the
	// callback its result, so `Err("e")` needs no annotation there.
	var wantRes Type
	if w := c.preludeEnumArg(want, "Result"); w != nil && sameType(w.Args[1], r.Args[1]) {
		wantRes = w
	}
	res := callbackReturns(c, e, checkCallback(c, e, 1, []Type{r.Args[0]}, wantRes), "Result")
	if res == nil {
		return TVoid
	}
	// The callback keeps the error type: an Err passes through unchanged.
	if !sameType(res.Args[1], r.Args[1]) {
		c.errorf(e.Args[1].Value.position(), "expected %s, found %s", r.Args[1], res.Args[1])
	}
	return res
}

func checkResultWithDefault(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 2) {
		return TVoid
	}
	r := preludeArg(c, e, 0, "Result")
	if r == nil {
		c.expr(e.Args[1].Value, nil)
		return TVoid
	}
	c.expectType(e.Args[1].Value.position(), r.Args[0], c.expr(e.Args[1].Value, r.Args[0]))
	return r.Args[0]
}

func checkResultToOption(c *Checker, e *CallExpr, want Type) Type {
	if !c.argCount(e, 1) {
		return TVoid
	}
	r := preludeArg(c, e, 0, "Result")
	if r == nil {
		return TVoid
	}
	if e.Builtin == "Result.err" {
		return preludeEnum(c, "Option", r.Args[1])
	}
	return preludeEnum(c, "Option", r.Args[0])
}

// passOn sets r to variant `to` of rt, holding a new reference to the
// payload of the borrowed variant `from` of o.
func passOn(e *Emitter, r, o string, from *Variant, rt *EnumType, to string) {
	v := retained(e, payloadOf(o, from), from.Fields[0].T)
	e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, to), []string{v}))
}

func emitResultMap(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	f, ft := funcArg(e, x, 1)
	rt := x.Type().(*EnumType)
	r := zeroResult(e, rt)
	hit, miss := "Ok", "Err"
	if x.Builtin == "Result.mapErr" {
		hit, miss = "Err", "Ok"
	}
	branch(e, o, variantOf(ot, hit), func() {
		v := callFn(e, f, ft, payloadOf(o, variantOf(ot, hit)))
		e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, hit), []string{v}))
	}, func() {
		passOn(e, r, o, variantOf(ot, miss), rt, miss)
	})
	return r, true
}

func emitResultAndThen(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	f, ft := funcArg(e, x, 1)
	rt := x.Type().(*EnumType)
	r := zeroResult(e, rt)
	branch(e, o, variantOf(ot, "Ok"), func() {
		e.line("%s = %s;", r, callFn(e, f, ft, payloadOf(o, variantOf(ot, "Ok"))))
	}, func() {
		passOn(e, r, o, variantOf(ot, "Err"), rt, "Err")
	})
	return r, true
}

func emitResultWithDefault(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	t := x.Type()
	code, _ := e.expr(x.Args[1].Value)
	d := e.temp(t, code)
	r := zeroResult(e, t)
	ok := variantOf(ot, "Ok")
	branch(e, o, ok, func() {
		assignCopy(e, r, payloadOf(o, ok), t)
	}, func() {
		assignCopy(e, r, d, t)
	})
	return r, true
}

// emitResultToOption emits Result.ok (the Ok value as Some) and Result.err
// (the Err value as Some).
func emitResultToOption(e *Emitter, x *CallExpr, w string) (string, bool) {
	o, ot := enumArg(e, x, 0)
	rt := x.Type().(*EnumType)
	r := zeroResult(e, rt)
	hit := "Ok"
	if x.Builtin == "Result.err" {
		hit = "Err"
	}
	branch(e, o, variantOf(ot, hit), func() {
		passOn(e, r, o, variantOf(ot, hit), rt, "Some")
	}, func() {
		e.line("%s = %s;", r, e.enumAlloc(variantOf(rt, "None"), nil))
	})
	return r, true
}
