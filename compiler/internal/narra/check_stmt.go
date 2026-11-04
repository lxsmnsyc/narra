package narra

func (c *Checker) funcBody(fd *FuncDecl, env map[string]Type) {
	c.fn = fd
	c.env = env
	c.loops = 0
	c.labels = nil
	c.scope = &scope{names: map[string]*Local{}}
	for _, p := range fd.Params {
		l := c.declare(p.Pos, p.Name, p.T, p.Mutable)
		l.Param = true
		l.ByRef = p.Mutable
		c.funcParam(p, l)
		p.Local = l
	}
	c.block(fd.Body)
	c.env = nil
}

// inGenericTemplate reports whether the current body is a generic function
// checked with its type parameters left open. Such a body is never emitted.
func (c *Checker) inGenericTemplate() bool {
	for _, t := range c.env {
		if _, ok := t.(*TypeParamType); ok {
			return true
		}
	}
	return false
}

func (c *Checker) block(b *Block) {
	c.push()
	for _, s := range b.Stmts {
		c.stmt(s)
	}
	c.pop()
}

// diverges reports whether a block always leaves by return, break, continue
// or panic.
func diverges(b *Block) bool {
	if b == nil || len(b.Stmts) == 0 {
		return false
	}
	switch s := b.Stmts[len(b.Stmts)-1].(type) {
	case *ReturnStmt, *BreakStmt, *ContinueStmt:
		return true
	case *ExprStmt:
		if call, ok := s.Expr.(*CallExpr); ok {
			if id, ok := call.Callee.(*Ident); ok && id.Name == "panic" {
				return true
			}
		}
	case *IfStmt:
		if s.Else == nil || !diverges(s.Then) {
			return false
		}
		if eb, ok := s.Else.(*Block); ok {
			return diverges(eb)
		}
		return diverges(&Block{Stmts: []Stmt{s.Else}})
	case *Block:
		return diverges(s)
	}
	return false
}

// otherVariant returns the variant an enum value must be when it is not v,
// for enums with exactly two variants.
func otherVariant(v *Variant) *Variant {
	if v == nil || len(v.Enum.Variants) != 2 {
		return nil
	}
	return v.Enum.Variants[1-v.Index]
}

func (c *Checker) narrowable(e Expr) *Local {
	id, ok := e.(*Ident)
	if !ok || id.Local == nil {
		return nil
	}
	return id.Local
}

// endNarrowing forgets what an `is` test told about l, in every scope. A
// `let` or `mutable` local is narrowed only until it is assigned, passed as
// `mutable` to a call, or reached again by a loop that assigns it.
func (c *Checker) endNarrowing(l *Local) {
	if l == nil || !l.Mutable {
		return
	}
	for s := c.scope; s != nil; s = s.parent {
		delete(s.narrow, l)
	}
}

// endLoopNarrowing ends the narrowing of every local that the condition or
// the body of a loop assigns or passes as `mutable`: the next iteration sees
// the changed value. Names are matched without scopes, which can only end
// more narrowing than needed.
func (c *Checker) endLoopNarrowing(cond Expr, body *Block) {
	names := map[string]bool{}
	assignedInExpr(cond, names)
	assignedInBlock(body, names)
	for name := range names {
		c.endNarrowing(c.lookup(name))
	}
}

// endExprNarrowing ends the narrowing of every local that an expression
// passes as `mutable` (or assigns, inside a match arm block).
func (c *Checker) endExprNarrowing(x Expr) {
	names := map[string]bool{}
	assignedInExpr(x, names)
	for name := range names {
		c.endNarrowing(c.lookup(name))
	}
}

// rootName is the variable name at the root of an assignable expression.
func rootName(e Expr) string {
	switch e := e.(type) {
	case *Ident:
		return e.Name
	case *FieldExpr:
		return rootName(e.Target)
	case *IndexExpr:
		return rootName(e.Target)
	}
	return ""
}

func assignedInBlock(b *Block, names map[string]bool) {
	if b == nil {
		return
	}
	for _, st := range b.Stmts {
		assignedInStmt(st, names)
	}
}

func assignedInStmt(st Stmt, names map[string]bool) {
	switch st := st.(type) {
	case *Block:
		assignedInBlock(st, names)
	case *VarStmt:
		assignedInExpr(st.Value, names)
	case *AssignStmt:
		names[rootName(st.Target)] = true
		assignedInExpr(st.Target, names)
		assignedInExpr(st.Value, names)
	case *ExprStmt:
		assignedInExpr(st.Expr, names)
	case *ReturnStmt:
		assignedInExpr(st.Value, names)
	case *IfStmt:
		assignedInExpr(st.Cond, names)
		assignedInBlock(st.Then, names)
		if st.Else != nil {
			assignedInStmt(st.Else, names)
		}
	case *WhileStmt:
		assignedInExpr(st.Cond, names)
		assignedInBlock(st.Body, names)
	case *ForStmt:
		assignedInExpr(st.Iter, names)
		assignedInBlock(st.Body, names)
	}
}

func assignedInExpr(x Expr, names map[string]bool) {
	switch x := x.(type) {
	case *TemplateLit:
		for _, p := range x.Parts {
			assignedInExpr(p, names)
		}
	case *FieldExpr:
		assignedInExpr(x.Target, names)
	case *IsExpr:
		assignedInExpr(x.Subject, names)
	case *IndexExpr:
		assignedInExpr(x.Target, names)
		assignedInExpr(x.Index, names)
	case *MapLit:
		for i := range x.Keys {
			assignedInExpr(x.Keys[i], names)
			assignedInExpr(x.Vals[i], names)
		}
	case *UnaryExpr:
		assignedInExpr(x.Operand, names)
	case *BinaryExpr:
		assignedInExpr(x.Left, names)
		assignedInExpr(x.Right, names)
	case *StructLit:
		for _, f := range x.Fields {
			assignedInExpr(f.Value, names)
		}
		assignedInExpr(x.Spread, names)
	case *ArrayLit:
		for _, e := range x.Elems {
			assignedInExpr(e, names)
		}
	case *MatchExpr:
		assignedInExpr(x.Subject, names)
		for _, arm := range x.Arms {
			assignedInExpr(arm.Guard, names)
			assignedInExpr(arm.Body, names)
			assignedInBlock(arm.Block, names)
		}
	case *CallExpr:
		assignedInExpr(x.Callee, names)
		for _, a := range x.Args {
			if a.Mutable {
				names[rootName(a.Value)] = true
			}
			assignedInExpr(a.Value, names)
		}
	}
}

func (c *Checker) ifStmt(s *IfStmt) {
	is, isPattern := s.Cond.(*IsExpr)
	if isPattern {
		c.isBinds = true
	}
	c.expectType(s.Cond.position(), TBool, c.expr(s.Cond, TBool))
	c.isBinds = false
	var subject *Local
	var v *Variant
	if isPattern && is.Pattern.Variant != nil {
		subject = c.narrowable(is.Subject)
		v = is.Pattern.Variant
	}
	c.push()
	if isPattern {
		for _, l := range is.Pattern.Locals {
			if l != nil {
				c.scope.names[l.Name] = l
			}
		}
		if subject != nil && v != nil {
			c.setNarrow(subject, v)
		}
	}
	for _, st := range s.Then.Stmts {
		c.stmt(st)
	}
	c.pop()
	other := otherVariant(v)
	if isPattern && !argsIrrefutable(is.Pattern) {
		other = nil // `x is Some(1)` failing does not make x None
	}
	if s.Else != nil {
		c.push()
		if subject != nil && other != nil {
			c.setNarrow(subject, other)
		}
		c.stmt(s.Else)
		c.pop()
	} else if subject != nil && other != nil && diverges(s.Then) {
		c.setNarrow(subject, other)
	}
}

func (c *Checker) stmt(s Stmt) {
	// The emitter runs `mutable` calls before the rest of the statement, so
	// a read of the local anywhere in the statement may see the new value.
	switch s := s.(type) {
	case *VarStmt:
		c.endExprNarrowing(s.Value)
	case *AssignStmt:
		c.endExprNarrowing(s.Target)
		c.endExprNarrowing(s.Value)
	case *ExprStmt:
		c.endExprNarrowing(s.Expr)
	case *ReturnStmt:
		c.endExprNarrowing(s.Value)
	case *IfStmt:
		c.endExprNarrowing(s.Cond)
	case *ForStmt:
		c.endExprNarrowing(s.Iter)
	}
	switch s := s.(type) {
	case *Block:
		c.block(s)
	case *VarStmt:
		if s.Pattern != nil {
			c.destructure(s)
			return
		}
		var want Type
		if s.Type != nil {
			want = c.resolveType(s.Type)
		}
		nerrs := len(c.errs)
		t := c.expr(s.Value, want)
		if want != nil {
			c.expectType(s.Value.position(), want, t)
			t = want
		}
		// A value whose check already failed is void; do not report it twice.
		if t == TVoid && len(c.errs) == nerrs {
			c.errorf(s.Pos, "this expression has no value")
		}
		s.Local = c.declare(s.Pos, s.Name, t, s.Mutable)
	case *AssignStmt:
		if c.constWrite(s.Pos, s.Target) {
			c.expr(s.Value, nil)
			return
		}
		c.assigning = s.Target
		t := c.expr(s.Target, nil)
		c.assigning = nil
		root := c.lvalueRoot(s.Target)
		if root == nil {
			c.errorf(s.Pos, "only variables, fields and array elements can be assigned")
		} else if !root.Mutable {
			c.errorf(s.Pos, "%q is constant; declare it with `let` to change it", root.Name)
		}
		if _, plain := s.Target.(*Ident); !plain {
			c.noMutableEscape(s.Value)
		}
		v := c.expr(s.Value, t)
		c.expectType(s.Value.position(), t, v)
		c.endNarrowing(root)
	case *ExprStmt:
		if m, ok := s.Expr.(*MatchExpr); ok {
			c.match(m, nil, false)
			return
		}
		c.expr(s.Expr, nil)
		if call, ok := s.Expr.(*CallExpr); !ok {
			c.errorf(s.Pos, "this expression does nothing")
		} else if isResult(call.Type()) {
			c.errorf(s.Pos, "this call returns a Result that is not checked; handle it, or discard it with `const _ = ...`")
		}
	case *ReturnStmt:
		want := c.fn.ResultT
		if s.Value == nil {
			if want != TVoid {
				c.errorf(s.Pos, "this function must return a %s", want)
			}
			return
		}
		c.noMutableEscape(s.Value)
		t := c.expr(s.Value, want)
		if want == TVoid {
			c.errorf(s.Pos, "this function returns nothing")
			return
		}
		c.expectType(s.Value.position(), want, t)
	case *IfStmt:
		c.ifStmt(s)
	case *WhileStmt:
		c.endLoopNarrowing(s.Cond, s.Body)
		c.expectType(s.Cond.position(), TBool, c.expr(s.Cond, TBool))
		c.enterLoop(s.Pos, s.Label)
		c.block(s.Body)
		c.leaveLoop(s.Label)
	case *ForStmt:
		var elem Type
		call, isCall := s.Iter.(*CallExpr)
		outerSource := c.forSource
		c.forSource = s.Iter
		t := c.expr(s.Iter, nil)
		c.forSource = outerSource
		switch {
		case isCall && (call.Builtin == "Range.upTo" || call.Builtin == "Range.between"):
			s.Kind = forRange
			elem = TInt
		case isCall && (call.Builtin == "Array.values" || call.Builtin == "Array.entries"):
			s.Kind = forArray
			elem = TVoid
			if at, ok := call.Args[0].Value.Type().(*ArrayType); ok {
				elem = at.Elem
				if call.Builtin == "Array.entries" {
					s.Kind = forArrayEntries
					elem = c.tupleOf([]Type{TInt, at.Elem})
				}
			}
		case isCall && (call.Builtin == "Map.keys" || call.Builtin == "Map.entries"):
			s.Kind = forMapKeys
			if call.Builtin == "Map.entries" {
				s.Kind = forMapEntries
			}
			elem = TVoid
			if at, ok := t.(*ArrayType); ok {
				elem = at.Elem
			}
		case t == Type(TRange):
			// A Range value: its bounds are read once, before the loop.
			s.Kind = forRange
			elem = TInt
		default:
			if at, ok := t.(*ArrayType); ok {
				// An array value, as with Array.values.
				s.Kind = forArray
				elem = at.Elem
				break
			}
			c.errorf(s.Iter.position(), "for ... of needs a Range, an Array, Str.chars, Array.entries, Map.keys or Map.entries, found %s", t)
			elem = TVoid
		}
		c.endLoopNarrowing(nil, s.Body)
		c.push()
		s.Local = c.declare(s.Pos, s.Name, elem, false)
		c.enterLoop(s.Pos, s.Label)
		c.block(s.Body)
		c.leaveLoop(s.Label)
		c.pop()
	case *BreakStmt:
		c.loopExit(s.Pos, s.Label, "break")
	case *ContinueStmt:
		c.loopExit(s.Pos, s.Label, "continue")
	}
}

// enterLoop opens a loop body; its label is in scope inside the body only.
func (c *Checker) enterLoop(pos Pos, label string) {
	c.loops++
	if label == "" {
		return
	}
	for _, outer := range c.labels {
		if outer == label {
			c.errorf(pos, "label %q is already used by an enclosing loop", label)
			break
		}
	}
	c.labels = append(c.labels, label)
}

func (c *Checker) leaveLoop(label string) {
	c.loops--
	if label != "" {
		c.labels = c.labels[:len(c.labels)-1]
	}
}

// loopExit checks a break or continue, with or without a label.
func (c *Checker) loopExit(pos Pos, label, word string) {
	if label == "" {
		if c.loops == 0 && word == "break" {
			c.errorf(pos, "break must be inside a loop")
		} else if c.loops == 0 {
			c.errorf(pos, "continue must be inside a loop")
		}
		return
	}
	for _, name := range c.labels {
		if name == label {
			return
		}
	}
	c.errorf(pos, "no enclosing loop is labeled %q", label)
}

func isResult(t Type) bool {
	et, ok := t.(*EnumType)
	return ok && et.Name == "Result" && len(et.Args) == 2
}

// lvalueRoot returns the variable at the root of an assignable expression.
func (c *Checker) lvalueRoot(e Expr) *Local {
	switch e := e.(type) {
	case *Ident:
		return e.Local
	case *FieldExpr:
		if e.Builtin != "" || e.Ctor != nil || e.Narrow != nil {
			return nil
		}
		return c.lvalueRoot(e.Target)
	case *IndexExpr:
		return c.lvalueRoot(e.Target)
	}
	return nil
}
