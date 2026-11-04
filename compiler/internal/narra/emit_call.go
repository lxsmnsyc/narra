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

// selfTailCall returns the call of `return f(...)` when f is the function
// being emitted and the call can become a jump to the function start. A
// function-type parameter keeps the plain call, because a closure made for
// the call may live on the stack of the iteration that ends at the jump. A
// mutable parameter must be passed on as it is.
func selfTailCall(fn *FuncDecl, s *ReturnStmt) *CallExpr {
	if fn == nil || s.Value == nil {
		return nil
	}
	call, ok := s.Value.(*CallExpr)
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

// hasSelfTailCall reports whether the body of fn has a self tail call.
func hasSelfTailCall(fn *FuncDecl) bool {
	return stmtHasTailCall(fn, fn.Body)
}

func stmtHasTailCall(fn *FuncDecl, s Stmt) bool {
	switch s := s.(type) {
	case *Block:
		for _, inner := range s.Stmts {
			if stmtHasTailCall(fn, inner) {
				return true
			}
		}
	case *ReturnStmt:
		return selfTailCall(fn, s) != nil
	case *IfStmt:
		if stmtHasTailCall(fn, s.Then) {
			return true
		}
		if s.Else != nil {
			return stmtHasTailCall(fn, s.Else)
		}
	case *WhileStmt:
		return stmtHasTailCall(fn, s.Body)
	case *ForStmt:
		return stmtHasTailCall(fn, s.Body)
	case *ExprStmt:
		if m, ok := s.Expr.(*MatchExpr); ok {
			for _, arm := range m.Arms {
				if arm.Block != nil && stmtHasTailCall(fn, arm.Block) {
					return true
				}
			}
		}
	}
	return false
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
