package narra

func (c *Checker) funcBody(fd *FuncDecl, env map[string]Type) {
	c.fn = fd
	c.env = env
	c.loops = 0
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
	if !ok || id.Local == nil || id.Local.Mutable {
		return nil
	}
	return id.Local
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
	if isPattern && !is.Pattern.Wildcard {
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
	switch s := s.(type) {
	case *Block:
		c.block(s)
	case *VarStmt:
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
		t := c.expr(s.Target, nil)
		root := c.lvalueRoot(s.Target)
		if root == nil {
			c.errorf(s.Pos, "only variables, fields and array elements can be assigned")
		} else if !root.Mutable {
			c.errorf(s.Pos, "%q is constant; declare it with `let` to change it", root.Name)
		}
		v := c.expr(s.Value, t)
		c.expectType(s.Value.position(), t, v)
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
		t := c.expr(s.Value, want)
		if want == TVoid {
			c.errorf(s.Pos, "this function returns nothing")
			return
		}
		c.expectType(s.Value.position(), want, t)
	case *IfStmt:
		c.ifStmt(s)
	case *WhileStmt:
		c.expectType(s.Cond.position(), TBool, c.expr(s.Cond, TBool))
		c.loops++
		c.block(s.Body)
		c.loops--
	case *ForStmt:
		var elem Type
		call, isCall := s.Iter.(*CallExpr)
		t := c.expr(s.Iter, nil)
		switch {
		case isCall && (call.Builtin == "Range.upTo" || call.Builtin == "Range.between"):
			s.Kind = forRange
			elem = TInt
		case isCall && call.Builtin == "Array.values":
			s.Kind = forArray
			elem = TVoid
			if at, ok := call.Args[0].Value.Type().(*ArrayType); ok {
				elem = at.Elem
			}
		default:
			c.errorf(s.Iter.position(), "for ... of needs Range.upTo, Range.between or Array.values, found %s", t)
			elem = TInt
		}
		c.push()
		s.Local = c.declare(s.Pos, s.Name, elem, false)
		c.loops++
		c.block(s.Body)
		c.loops--
		c.pop()
	case *BreakStmt:
		if c.loops == 0 {
			c.errorf(s.Pos, "break must be inside a loop")
		}
	case *ContinueStmt:
		if c.loops == 0 {
			c.errorf(s.Pos, "continue must be inside a loop")
		}
	}
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
