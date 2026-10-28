package narra

import "sort"

// Arrow functions, function values and the escape rules (SPEC.md section 2,
// "Closures", and D46).
//
// - A function-type parameter without `escaping` (Local.NoEscape) can only
//   be called, or passed straight to another non-escaping function-type
//   parameter. Any other use is an error.
// - An arrow function passed straight to a non-escaping parameter gets its
//   environment on the caller's stack (Closure.Stack). Every other arrow
//   function with captures gets a reference-counted heap environment.
// - Closures capture only constants: `const` bindings, non-mutable
//   parameters, loop and pattern bindings. Capturing a `let` variable, a
//   mutable parameter or a non-escaping parameter is an error.
// - Parameters of a function type, and struct fields and array elements,
//   always hold escaping values, so a value passed through them must be
//   escaping too.

// Closure is what the checker records about an arrow function.
type Closure struct {
	Type     *FuncType
	Fn       *FuncDecl // Params with their Locals, ResultT, and the Body to emit
	Captures []*Local  // in order of first use
	Stack    bool      // passed straight to a non-escaping parameter
	// Hint is the expected type from a generic call whose type arguments
	// are not all known yet. A nil parameter or result is unknown.
	Hint *FuncType
}

type closureState struct {
	funcTypes map[string]*FuncType
	allowed   Expr // the one Ident that may name a non-escaping parameter
	arrows    []*arrowCtx
}

type arrowCtx struct {
	outer *scope // the scope the arrow function appears in
	clo   *Closure
}

func (c *Checker) cl() *closureState {
	if c.closure == nil {
		c.closure = &closureState{funcTypes: map[string]*FuncType{}}
	}
	return c.closure
}

func closureOf(a *ArrowExpr) *Closure {
	if a.Closure == nil {
		a.Closure = &Closure{}
	}
	return a.Closure
}

// funcOf returns the one FuncType for a signature, so the program's
// function types can be listed for the emitter.
func (c *Checker) funcOf(params []Type, result Type) *FuncType {
	ft := &FuncType{Params: params, Result: result}
	key := mangle(ft)
	if have, ok := c.cl().funcTypes[key]; ok {
		return have
	}
	c.cl().funcTypes[key] = ft
	return ft
}

// funcResult resolves the result of a function type, where `void` means
// the function returns nothing.
func (c *Checker) funcResult(te *TypeExpr) Type {
	if te != nil && te.Name == "void" && len(te.Args) == 0 && !te.Func {
		return TVoid
	}
	return c.resolveType(te)
}

func (c *Checker) finishFuncTypes() {
	keys := make([]string, 0, len(c.cl().funcTypes))
	for k, ft := range c.cl().funcTypes {
		if !containsTypeParam(ft) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		c.prog.FuncTypes = append(c.prog.FuncTypes, c.cl().funcTypes[k])
	}
}

// nonEscapingParam reports whether a parameter is written with a function
// type and without `escaping`. A parameter whose type is a type parameter is
// escaping even when the type argument is a function type.
func nonEscapingParam(p *Param) bool {
	return p.Type != nil && p.Type.Func && !p.Escaping && !p.Mutable
}

// funcParam sets up the local of a function parameter.
func (c *Checker) funcParam(p *Param, l *Local) {
	if p.Escaping && (p.Type == nil || !p.Type.Func) {
		c.errorf(p.Pos, "only a parameter with a function type can be escaping")
	}
	l.NoEscape = nonEscapingParam(p)
}

// funcArg prepares an argument for parameter p before it is checked.
func (c *Checker) funcArg(p *Param, v Expr) {
	if !nonEscapingParam(p) {
		return
	}
	switch v := v.(type) {
	case *Ident:
		c.cl().allowed = v
	case *ArrowExpr:
		closureOf(v).Stack = true
	}
}

// useLocal checks a use of a local: non-escaping parameters, and captures
// by the arrow functions being checked.
func (c *Checker) useLocal(e *Ident, l *Local) {
	st := c.cl()
	if l.NoEscape && st.allowed != e {
		c.errorf(e.P, "parameter %q cannot escape: it can only be called or passed to a function parameter that is not escaping; declare it as `escaping %s` to store, return or capture it", l.Name, l.Name)
		return
	}
	for i := len(st.arrows) - 1; i >= 0; i-- {
		ctx := st.arrows[i]
		if declaredBelow(c.scope, ctx.outer, l) {
			return
		}
		if i == len(st.arrows)-1 {
			switch {
			case l.NoEscape:
				c.errorf(e.P, "a closure cannot capture parameter %q because it is not escaping; declare it as `escaping %s`", l.Name, l.Name)
				return
			case l.Mutable && l.Param:
				c.errorf(e.P, "a closure cannot capture mutable parameter %q; copy it first with `const copy = %s`", l.Name, l.Name)
				return
			case l.Mutable:
				c.errorf(e.P, "a closure cannot capture %q because it is declared with let; closures capture only constants", l.Name)
				return
			}
		}
		have := false
		for _, x := range ctx.clo.Captures {
			if x == l {
				have = true
			}
		}
		if !have {
			ctx.clo.Captures = append(ctx.clo.Captures, l)
		}
	}
}

// declaredBelow reports whether l is declared in a scope from s up to, but
// not including, stop.
func declaredBelow(s, stop *scope, l *Local) bool {
	for ; s != nil && s != stop; s = s.parent {
		if s.names[l.Name] == l {
			return true
		}
	}
	return false
}

// funcValue checks a named function used as a value. dst receives the
// function, or its instance for a generic function.
func (c *Checker) funcValue(pos Pos, fd *FuncDecl, want Type, dst **FuncDecl) Type {
	for _, p := range fd.Params {
		if p.Mutable {
			c.errorf(pos, "function %s has a mutable parameter, so it cannot be used as a value", fd.Name)
			return TVoid
		}
	}
	tps, generic := c.funcTP[fd]
	if !generic {
		*dst = fd
		params := make([]Type, len(fd.Params))
		for i, p := range fd.Params {
			params[i] = p.T
		}
		return c.funcOf(params, fd.ResultT)
	}
	m := map[*TypeParamType]Type{}
	if wt, ok := want.(*FuncType); ok && len(wt.Params) == len(fd.Params) {
		for i, p := range fd.Params {
			unify(p.T, wt.Params[i], m)
		}
		unify(fd.ResultT, wt.Result, m)
	}
	targs := make([]Type, len(tps))
	open := false
	for i, p := range tps {
		t, ok := m[p]
		if !ok {
			c.errorf(pos, "cannot tell type parameter %s of %s here; add a type annotation", p.Name, fd.Name)
			return TVoid
		}
		targs[i] = t
		open = open || containsTypeParam(t)
	}
	c.checkBounds(pos, fd, tps, targs)
	*dst = fd
	if !open {
		*dst = c.funcInstance(fd, tps, targs)
	}
	params := make([]Type, len(fd.Params))
	for i, p := range fd.Params {
		params[i] = c.substitute(p.T, m)
	}
	return c.funcOf(params, c.substitute(fd.ResultT, m))
}

// moduleFuncValue checks `Alias.f` used as a value, where Alias is a
// package module. It returns nil when e is something else.
func (c *Checker) moduleFuncValue(e *FieldExpr, want Type) Type {
	id, ok := e.Target.(*Ident)
	if !ok || c.lookup(id.Name) != nil {
		return nil
	}
	path, ok := c.modules[id.Name]
	s := c.mods[path]
	if !ok || s == nil || s.src == nil {
		return nil
	}
	fd, ok := s.src.Exports[e.Name].(*FuncDecl)
	if !ok {
		return nil
	}
	id.setType(&ModuleType{Path: path})
	return c.funcValue(e.P, fd, want, &e.Func)
}

// fieldFunc returns the function type of `target.name` when the target is
// a struct whose field holds a function, so `value.field(args)` is a call.
func fieldFunc(callee *FieldExpr, t Type) *FuncType {
	st, ok := t.(*StructType)
	if !ok {
		return nil
	}
	f := st.Field(callee.Name)
	if f == nil {
		return nil
	}
	ft, ok := f.T.(*FuncType)
	if ok {
		callee.setType(ft)
	}
	return ft
}

// valueCall checks a call whose callee is an expression, such as a local,
// a parameter or a call that returns a function.
func (c *Checker) valueCall(e *CallExpr) Type {
	if id, ok := e.Callee.(*Ident); ok {
		c.cl().allowed = id
	}
	t := c.expr(e.Callee, nil)
	ft, ok := t.(*FuncType)
	if !ok {
		if id, isIdent := e.Callee.(*Ident); isIdent && id.Local != nil {
			c.errorf(e.P, "%q is a %s, not a function", id.Name, t)
		} else {
			c.errorf(e.P, "this expression cannot be called")
		}
		c.argsLoose(e)
		return TVoid
	}
	return c.valueCallOf(e, ft)
}

// valueCallOf checks the arguments of a call through a value of type ft.
// Parameters of a function type are escaping.
func (c *Checker) valueCallOf(e *CallExpr, ft *FuncType) Type {
	e.FuncValue = ft
	if len(e.Args) != len(ft.Params) {
		c.errorf(e.P, "this call passes %d arguments, the function takes %d", len(e.Args), len(ft.Params))
	}
	for i, a := range e.Args {
		if a.Mutable {
			c.errorf(a.Value.position(), "a function value takes no mutable arguments; remove `mutable`")
		}
		var want Type
		if i < len(ft.Params) {
			want = ft.Params[i]
		}
		c.expectType(a.Value.position(), want, c.expr(a.Value, want))
	}
	c.checkCalleeExclusive(e)
	return ft.Result
}

// checkCalleeExclusive rejects a call through a value whose arguments pass,
// at any depth, a variable as mutable that the callee expression reads, as
// in `f(swap(mutable f))`. The call could otherwise change or free the
// function value it is calling (SPEC section 6, exclusive access).
func (c *Checker) checkCalleeExclusive(e *CallExpr) {
	for _, a := range e.Args {
		for _, root := range c.mutableRoots(a.Value) {
			if usesLocal(e.Callee, root) {
				c.errorf(a.Value.position(), "%q is passed as mutable inside this call, so it cannot also hold the function being called", root.Name)
			}
		}
	}
}

// mutableRoots lists the variables passed as mutable arguments by the calls
// inside x, outside arrow functions.
func (c *Checker) mutableRoots(x Expr) []*Local {
	var out []*Local
	var walk func(x Expr)
	walk = func(x Expr) {
		switch x := x.(type) {
		case *CallExpr:
			walk(x.Callee)
			for _, a := range x.Args {
				if a.Mutable {
					if root := c.lvalueRoot(a.Value); root != nil {
						out = append(out, root)
					}
				}
				walk(a.Value)
			}
		case *FieldExpr:
			walk(x.Target)
		case *IndexExpr:
			walk(x.Target)
			walk(x.Index)
		case *UnaryExpr:
			walk(x.Operand)
		case *BinaryExpr:
			walk(x.Left)
			walk(x.Right)
		case *IsExpr:
			walk(x.Subject)
		case *TemplateLit:
			for _, p := range x.Parts {
				walk(p)
			}
		case *StructLit:
			for _, f := range x.Fields {
				walk(f.Value)
			}
		case *ArrayLit:
			for _, el := range x.Elems {
				walk(el)
			}
		case *MatchExpr:
			walk(x.Subject)
			for _, arm := range x.Arms {
				walk(arm.Guard)
				walk(arm.Body)
			}
		}
	}
	walk(x)
	return out
}

// hintArrow gives an arrow function argument of a generic call the parts
// of its parameter type that the earlier arguments already decide, so
// `map(items, (x) => x * 2)` knows the type of x.
func (c *Checker) hintArrow(e *CallExpr, i int, fd *FuncDecl, tps []*TypeParamType, m map[*TypeParamType]Type) {
	if i >= len(e.Args) || i >= len(fd.Params) {
		return
	}
	a, ok := e.Args[i].Value.(*ArrowExpr)
	if !ok {
		return
	}
	m2 := map[*TypeParamType]Type{}
	for k, v := range m {
		m2[k] = v
	}
	for j := 0; j < i && j < len(fd.Params); j++ {
		if t := e.Args[j].Value.Type(); t != nil {
			unify(fd.Params[j].T, t, m2)
		}
	}
	ft, ok := c.substitute(fd.Params[i].T, m2).(*FuncType)
	if !ok {
		return
	}
	h := &FuncType{Params: make([]Type, len(ft.Params))}
	for k, p := range ft.Params {
		if !containsOpen(p, tps) {
			h.Params[k] = p
		}
	}
	if !containsOpen(ft.Result, tps) {
		h.Result = ft.Result
	}
	closureOf(a).Hint = h
}

// arrow checks an arrow function. Its body is checked as its own function
// whose scope sees the enclosing locals; the ones it uses become captures.
func (c *Checker) arrow(e *ArrowExpr, want Type) Type {
	clo := closureOf(e)
	hint := clo.Hint
	if wt, ok := want.(*FuncType); ok {
		hint = wt
	}
	if hint != nil && len(hint.Params) != len(e.Params) {
		c.errorf(e.P, "this arrow function takes %d parameters, but the expected function takes %d", len(e.Params), len(hint.Params))
		hint = nil
	}
	params := make([]Type, len(e.Params))
	for i, p := range e.Params {
		switch {
		case p.Type != nil:
			params[i] = c.resolveType(p.Type)
		case hint != nil && hint.Params[i] != nil:
			params[i] = hint.Params[i]
		default:
			c.errorf(p.Pos, "cannot tell the type of parameter %q; write it as `%s: Type`", p.Name, p.Name)
			params[i] = TVoid
		}
		p.T = params[i]
	}
	var result Type // nil until known
	if e.Result != nil {
		result = c.funcResult(e.Result)
	} else if hint != nil && hint.Result != nil {
		result = hint.Result
	}

	st := c.cl()
	savedFn, savedLoops, savedScope, savedBinds := c.fn, c.loops, c.scope, c.isBinds
	st.arrows = append(st.arrows, &arrowCtx{outer: c.scope, clo: clo})
	c.scope = &scope{names: map[string]*Local{}, parent: savedScope}
	c.loops, c.isBinds = 0, false
	fn := &FuncDecl{Pos: e.P, Name: "arrow", Params: e.Params, Module: c.mod.path}
	for _, p := range e.Params {
		l := c.declare(p.Pos, p.Name, p.T, false)
		l.Param = true
		p.Local = l
	}
	c.fn = fn
	if e.Block != nil {
		if result == nil {
			result = blockResult(e.Block)
		}
		fn.ResultT = result
		c.block(e.Block)
		if fn.ResultT == nil {
			fn.ResultT = c.inferReturns(e.Block)
		}
		fn.Body = e.Block
	} else {
		switch result {
		case nil:
			result = c.expr(e.Body, nil)
		case TVoid:
			c.expr(e.Body, nil)
		default:
			c.expectType(e.Body.position(), result, c.expr(e.Body, result))
		}
		fn.ResultT = result
		var s Stmt = &ReturnStmt{Pos: e.Body.position(), Value: e.Body}
		if result == TVoid {
			s = &ExprStmt{Pos: e.Body.position(), Expr: e.Body}
		}
		fn.Body = &Block{Pos: e.P, Stmts: []Stmt{s}}
	}
	c.fn, c.loops, c.scope, c.isBinds = savedFn, savedLoops, savedScope, savedBinds
	st.arrows = st.arrows[:len(st.arrows)-1]
	clo.Fn = fn
	clo.Type = c.funcOf(params, fn.ResultT)
	return clo.Type
}

// returnsOf lists the return statements of a block, outside nested arrow
// functions.
func returnsOf(b *Block) []*ReturnStmt {
	var out []*ReturnStmt
	var walk func(s Stmt)
	walkMatch := func(e Expr) {
		if m, ok := e.(*MatchExpr); ok {
			for _, a := range m.Arms {
				if a.Block != nil {
					walk(a.Block)
				}
			}
		}
	}
	walk = func(s Stmt) {
		switch s := s.(type) {
		case *Block:
			for _, x := range s.Stmts {
				walk(x)
			}
		case *ReturnStmt:
			out = append(out, s)
		case *IfStmt:
			walk(s.Then)
			if s.Else != nil {
				walk(s.Else)
			}
		case *WhileStmt:
			walk(s.Body)
		case *ForStmt:
			walk(s.Body)
		case *ExprStmt:
			walkMatch(s.Expr)
		case *VarStmt:
			walkMatch(s.Value)
		}
	}
	walk(b)
	return out
}

// blockResult gives the result type of a block-bodied arrow function with
// no written or expected result: void when no return has a value, or nil
// to infer it from the returns.
func blockResult(b *Block) Type {
	for _, r := range returnsOf(b) {
		if r.Value == nil {
			return TVoid
		}
	}
	if len(returnsOf(b)) == 0 {
		return TVoid
	}
	return nil
}

// inferReturns gives the result type from the returns of a checked block.
func (c *Checker) inferReturns(b *Block) Type {
	var result Type
	for _, r := range returnsOf(b) {
		t := r.Value.Type()
		if result == nil {
			result = t
			continue
		}
		c.expectType(r.Value.position(), result, t)
	}
	if result == nil {
		return TVoid
	}
	return result
}

// cloneArrow copies the syntax of an arrow function for a generic instance.
func cloneArrow(e *ArrowExpr) *ArrowExpr {
	a := &ArrowExpr{exprBase: exprBase{P: e.P}, Result: e.Result, Body: cloneExpr(e.Body), Block: cloneBlock(e.Block)}
	for _, p := range e.Params {
		a.Params = append(a.Params, &Param{Pos: p.Pos, Name: p.Name, Type: p.Type})
	}
	return a
}
