package narra

import (
	"fmt"
)

func (e *Emitter) stmt(s Stmt) {
	switch s := s.(type) {
	case *Block:
		e.line("{")
		e.indent++
		e.block(s, false)
		e.indent--
		e.line("}")
	case *VarStmt:
		e.beginStmt()
		code, owned := e.expr(s.Value)
		v := e.consume(code, owned, s.Local.T)
		e.line("%s %s = %s;", s.Local.T.CName(), s.Local.CName, v)
		e.endStmt()
		e.addLocal(s.Local)
	case *AssignStmt:
		e.beginStmt()
		code, owned := e.expr(s.Value)
		v := e.consume(code, owned, s.Value.Type())
		loc := e.lvaluePtr(s.Target)
		t := s.Target.Type()
		if r := release(t, "(*"+loc+")"); r != "" {
			ptr := e.temp(&ptrType{t}, loc)
			e.line("%s", release(t, "(*"+ptr+")"))
			e.line("*%s = %s;", ptr, v)
		} else {
			e.line("*%s = %s;", loc, v)
		}
		e.endStmt()
	case *ExprStmt:
		e.beginStmt()
		if m, ok := s.Expr.(*MatchExpr); ok {
			e.matchStmt(m)
		} else {
			code, _ := e.expr(s.Expr)
			if code != "" && s.Expr.Type() == TVoid {
				e.line("%s;", code)
			}
		}
		e.endStmt()
	case *ReturnStmt:
		e.beginStmt()
		if s.Value != nil {
			code, owned := e.expr(s.Value)
			v := e.consume(code, owned, s.Value.Type())
			r := e.temp(s.Value.Type(), v)
			e.unwindAll()
			e.line("return %s;", r)
		} else {
			e.unwindAll()
			e.line("return;")
		}
		e.frames = e.frames[:len(e.frames)-1]
	case *IfStmt:
		if is, ok := s.Cond.(*IsExpr); ok {
			e.ifIs(s, is)
			return
		}
		e.beginStmt()
		cond, _ := e.expr(s.Cond)
		c := e.temp(TBool, cond)
		e.endStmt()
		e.line("if (%s) {", c)
		e.indent++
		e.block(s.Then, false)
		e.indent--
		if s.Else != nil {
			e.line("} else {")
			e.indent++
			if b, ok := s.Else.(*Block); ok {
				e.block(b, false)
			} else {
				e.stmt(s.Else)
			}
			e.indent--
		}
		e.line("}")
	case *WhileStmt:
		e.line("while (1) {")
		e.indent++
		e.beginStmt()
		cond, _ := e.expr(s.Cond)
		c := e.temp(TBool, cond)
		e.endStmt()
		e.line("if (!%s) break;", c)
		e.loopTemp = append(e.loopTemp, len(e.frames))
		e.block(s.Body, true)
		e.loopTemp = e.loopTemp[:len(e.loopTemp)-1]
		e.indent--
		e.line("}")
	case *ForStmt:
		e.forStmt(s)
	case *BreakStmt:
		e.unwindLoop()
		e.line("break;")
	case *ContinueStmt:
		e.unwindLoop()
		e.line("continue;")
	}
}

func (e *Emitter) forStmt(s *ForStmt) {
	call := s.Iter.(*CallExpr)
	e.beginStmt()
	switch s.Kind {
	case forRange:
		var lo, hi string
		if call.Builtin == "Range.upTo" {
			lo = "0"
			c, _ := e.expr(call.Args[0].Value)
			hi = e.temp(TInt, c)
		} else {
			a, _ := e.expr(call.Args[0].Value)
			lo = e.temp(TInt, a)
			b, _ := e.expr(call.Args[1].Value)
			hi = e.temp(TInt, b)
		}
		e.line("for (int64_t %s = %s; %s < %s; %s++) {", s.Local.CName, lo, s.Local.CName, hi, s.Local.CName)
		e.indent++
		e.loopTemp = append(e.loopTemp, len(e.frames))
		e.block(s.Body, true)
		e.loopTemp = e.loopTemp[:len(e.loopTemp)-1]
		e.indent--
		e.line("}")
	case forArray:
		at := call.Args[0].Value.Type().(*ArrayType)
		code, owned := e.expr(call.Args[0].Value)
		arr := e.consume(code, owned, at)
		e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{arr, at})
		e.tempN++
		idx := fmt.Sprintf("i%d", e.tempN)
		e.line("for (int64_t %s = 0; %s < %s->len; %s++) {", idx, idx, arr, idx)
		e.indent++
		e.loopTemp = append(e.loopTemp, len(e.frames))
		e.pushScope(true)
		e.line("%s %s = %s->items[%s];", at.Elem.CName(), s.Local.CName, arr, idx)
		if r := retain(at.Elem, s.Local.CName); r != "" {
			e.line("%s", r)
		}
		e.addLocal(s.Local)
		e.blockInScope(s.Body)
		e.popScope()
		e.loopTemp = e.loopTemp[:len(e.loopTemp)-1]
		e.indent--
		e.line("}")
	}
	e.endStmt()
}
