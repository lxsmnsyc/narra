package narra

import (
	"fmt"
	"strings"
)

func (e *Emitter) call(x *CallExpr) (string, bool) {
	w := e.where(x.P)
	switch {
	case x.Distinct != nil:
		// from and to keep the representation: the argument is the result.
		return e.expr(x.Args[0].Value)
	case x.Ctor != nil:
		v := x.Ctor
		var parts []string
		for i, a := range x.Args {
			code, owned := e.expr(a.Value)
			parts = append(parts, e.consume(code, owned, v.Fields[i].T))
		}
		return e.ownedTemp(v.Enum, e.enumAlloc(v, parts)), true
	case x.Trait != nil:
		return e.traitCall(x)
	case x.FuncValue != nil:
		return e.valueCall(x)
	case x.Builtin != "":
		return e.builtin(x, w)
	case x.Func != nil:
		var args []string
		for i, a := range x.Args {
			if x.Func.Params[i].Mutable {
				args = append(args, e.lvaluePtr(a.Value))
				continue
			}
			code, _ := e.expr(a.Value)
			args = append(args, code)
		}
		callCode := fmt.Sprintf("%s(%s)", funcCName(x.Func), strings.Join(args, ", "))
		if x.Func.ResultT == TVoid {
			return callCode, false
		}
		return e.ownedTemp(x.Func.ResultT, callCode), true
	}
	panic("emit: unresolved call")
}

// builtin emits a call to a registered builtin.
func (e *Emitter) builtin(x *CallExpr, w string) (string, bool) {
	b := builtinsByName[x.Builtin]
	if b == nil || b.Emit == nil {
		panic("emit: unknown builtin " + x.Builtin)
	}
	return b.Emit(e, x, w)
}

// argCode emits argument i of a builtin call and returns its C code.
func (e *Emitter) argCode(x *CallExpr, i int) string {
	code, _ := e.expr(x.Args[i].Value)
	return code
}

// selfCall returns expr when it is a call of fn that can become a jump to
// the function start. A function-type parameter keeps the plain call,
// because a closure made for the call may live on the stack of the iteration
// that ends at the jump. A mutable parameter must be passed on as it is.
func selfCall(fn *FuncDecl, expr Expr) *CallExpr {
	call, ok := expr.(*CallExpr)
	if !ok || call.Func != fn || call.Builtin != "" || call.Ctor != nil || call.Trait != nil || call.FuncValue != nil || call.Distinct != nil {
		return nil
	}
	if len(call.Args) != len(fn.Params) {
		return nil
	}
	for i, p := range fn.Params {
		if _, isFunc := p.T.(*FuncType); isFunc {
			return nil
		}
		if p.Mutable {
			id, isIdent := call.Args[i].Value.(*Ident)
			if !isIdent || id.Local != p.Local {
				return nil
			}
		}
	}
	return call
}

// findTailCalls lists the self calls in tail position: the value of a
// `return`, an arm of a returned `match`, and the last statement of a void
// function, also at the end of an `if` branch or a `match` arm that is in
// tail position. Calls inside closures are never listed.
func findTailCalls(fn *FuncDecl) map[*CallExpr]bool {
	found := map[*CallExpr]bool{}
	scanTailBlock(fn, fn.Body, fn.ResultT == TVoid, found)
	return found
}

func scanTailBlock(fn *FuncDecl, b *Block, tail bool, found map[*CallExpr]bool) {
	for i, s := range b.Stmts {
		scanTailStmt(fn, s, tail && i == len(b.Stmts)-1, found)
	}
}

func scanTailStmt(fn *FuncDecl, s Stmt, tail bool, found map[*CallExpr]bool) {
	switch s := s.(type) {
	case *Block:
		scanTailBlock(fn, s, tail, found)
	case *ReturnStmt:
		if s.Value == nil {
			return
		}
		if call := selfCall(fn, s.Value); call != nil {
			found[call] = true
		}
		if m, ok := s.Value.(*MatchExpr); ok {
			for _, arm := range m.Arms {
				if call := selfCall(fn, arm.Body); call != nil {
					found[call] = true
				}
			}
		}
	case *IfStmt:
		scanTailBlock(fn, s.Then, tail, found)
		if s.Else != nil {
			scanTailStmt(fn, s.Else, tail, found)
		}
	case *WhileStmt:
		scanTailBlock(fn, s.Body, false, found)
	case *ForStmt:
		scanTailBlock(fn, s.Body, false, found)
	case *ExprStmt:
		if !tail {
			if m, ok := s.Expr.(*MatchExpr); ok {
				for _, arm := range m.Arms {
					if arm.Block != nil {
						scanTailBlock(fn, arm.Block, false, found)
					}
				}
			}
			return
		}
		if call := selfCall(fn, s.Expr); call != nil {
			found[call] = true
		}
		if m, ok := s.Expr.(*MatchExpr); ok {
			for _, arm := range m.Arms {
				if arm.Block != nil {
					scanTailBlock(fn, arm.Block, true, found)
				} else if call := selfCall(fn, arm.Body); call != nil {
					found[call] = true
				}
			}
		}
	}
}

// tailCallOf returns expr when it is a listed tail call.
func (e *Emitter) tailCallOf(expr Expr) *CallExpr {
	if call, ok := expr.(*CallExpr); ok && e.tailCalls[call] {
		return call
	}
	return nil
}

// matchHasTailCall reports whether an arm of the returned match is a tail call.
func (e *Emitter) matchHasTailCall(m *MatchExpr) bool {
	for _, arm := range m.Arms {
		if e.tailCallOf(arm.Body) != nil {
			return true
		}
	}
	return false
}

// returnMatch emits `return match ...` with each arm returning its own
// value, so an arm that is a tail call becomes a jump.
func (e *Emitter) returnMatch(m *MatchExpr) {
	t := m.Type()
	e.beginStmt()
	e.matchArms(m, func(arm *MatchArm) {
		if call := e.tailCallOf(arm.Body); call != nil {
			e.tailJump(call)
			return
		}
		e.beginStmt()
		var code string
		var owned bool
		if arm.Block != nil {
			code, owned = e.blockBody(arm.Block, t)
		} else {
			code, owned = e.expr(arm.Body)
		}
		v := e.consume(code, owned, t)
		r := e.temp(t, v)
		e.unwindAll()
		e.line("return %s;", r)
		e.frames = e.frames[:len(e.frames)-1]
	})
	e.frames = e.frames[:len(e.frames)-1]
}

// tailJump emits a self tail call: the new arguments are evaluated and
// owned, everything the function holds is released, then the parameters
// take the new values and control jumps to the function start.
func (e *Emitter) tailJump(call *CallExpr) {
	e.beginStmt()
	values := make([]string, len(call.Args))
	for i, a := range call.Args {
		p := e.fn.Params[i]
		if p.Mutable {
			continue
		}
		code, owned := e.expr(a.Value)
		v := e.consume(code, owned, p.T)
		if !needsRC(p.T) {
			v = e.temp(p.T, v)
		}
		values[i] = v
	}
	e.unwindAll()
	for i, p := range e.fn.Params {
		if !p.Mutable {
			e.line("%s = %s;", p.Local.CName, values[i])
		}
	}
	e.line("goto narra_tail;")
	e.frames = e.frames[:len(e.frames)-1]
}
