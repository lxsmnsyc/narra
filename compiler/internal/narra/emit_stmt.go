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
		if s.Pattern != nil {
			e.destructure(s)
			return
		}
		e.beginStmt()
		code, owned := e.expr(s.Value)
		v := e.consume(code, owned, s.Local.T)
		e.declareLocal(s.Local, v)
		e.endStmt()
		e.addLocal(s.Local)
	case *AssignStmt:
		e.beginStmt()
		if ix, ok := s.Target.(*IndexExpr); ok && ix.MapSet {
			e.mapSet(ix, s.Value)
			e.endStmt()
			return
		}
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
		if call := e.tailCallOf(s.Expr); call != nil {
			e.tailJump(call)
			return
		}
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
		if len(e.blocks) > 0 {
			e.blockReturn(s)
			return
		}
		if e.gen != nil {
			e.generatorReturn()
			return
		}
		if call := e.tailCallOf(s.Value); call != nil {
			e.tailJump(call)
			return
		}
		if m, ok := s.Value.(*MatchExpr); ok && e.matchHasTailCall(m) {
			e.returnMatch(m)
			return
		}
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
		e.enterLoop(s.Label, s.Pos)
		e.block(s.Body, true)
		e.continueLabel()
		info := e.leaveLoop()
		e.indent--
		e.line("}")
		e.breakLabel(info)
	case *ForStmt:
		e.forStmt(s)
	case *YieldStmt:
		e.yieldStmt(s)
	case *BreakStmt:
		if s.Label != "" {
			e.labeledExit(s.Label, true)
			return
		}
		e.unwindLoop()
		e.line("break;")
	case *ContinueStmt:
		if s.Label != "" {
			e.labeledExit(s.Label, false)
			return
		}
		e.unwindLoop()
		e.line("continue;")
	}
}

func (e *Emitter) forStmt(s *ForStmt) {
	call, _ := s.Iter.(*CallExpr)
	e.beginStmt()
	switch s.Kind {
	case forRange:
		var lo, hi string
		if call == nil || (call.Builtin != "Range.upTo" && call.Builtin != "Range.between") {
			// A Range value: read both bounds once.
			code, _ := e.expr(s.Iter)
			r := e.temp(TRange, code)
			lo, hi = r+".low", r+".high"
		} else if call.Builtin == "Range.upTo" {
			lo = "0"
			c, _ := e.expr(call.Args[0].Value)
			hi = e.temp(TInt, c)
		} else {
			a, _ := e.expr(call.Args[0].Value)
			lo = e.temp(TInt, a)
			b, _ := e.expr(call.Args[1].Value)
			hi = e.temp(TInt, b)
		}
		decl := e.declareCounter(s.Local)
		v := e.localRef(s.Local)
		e.line("for (%s = %s; %s < %s; %s++) {", decl, lo, v, hi, v)
		e.indent++
		e.enterLoop(s.Label, s.Pos)
		e.block(s.Body, true)
		e.continueLabel()
		info := e.leaveLoop()
		e.indent--
		e.line("}")
		e.breakLabel(info)
	case forMapKeys:
		e.mapFor(s, call, false)
	case forMapEntries:
		e.mapFor(s, call, true)
	case forGenerator:
		e.generatorFor(s)
	case forArray, forArrayEntries:
		source := s.Iter
		if call != nil && (call.Builtin == "Array.values" || call.Builtin == "Array.entries") {
			source = call.Args[0].Value
		}
		at := source.Type().(*ArrayType)
		code, owned := e.expr(source)
		arr := e.consume(code, owned, at)
		e.frames[len(e.frames)-1] = append(e.frames[len(e.frames)-1], tempVar{arr, at})
		e.tempN++
		decl, idx := e.cVar("int64_t", fmt.Sprintf("i%d", e.tempN))
		e.line("for (%s = 0; %s < %s->len; %s++) {", decl, idx, arr, idx)
		e.indent++
		e.enterLoop(s.Label, s.Pos)
		e.pushScope(true)
		if s.Kind == forArrayEntries {
			// Array.entries gives #[index, item].
			st := s.Local.T.(*StructType)
			init := fmt.Sprintf("{.%s = %s, .%s = %s->items[%s]}", fieldName("0"), idx, fieldName("1"), arr, idx)
			if e.gen != nil {
				init = "(" + st.CName() + ")" + init // an assignment needs a compound literal
			}
			e.declareLocal(s.Local, init)
		} else {
			e.declareLocal(s.Local, fmt.Sprintf("%s->items[%s]", arr, idx))
		}
		if r := retain(s.Local.T, e.localRef(s.Local)); r != "" {
			e.line("%s", r)
		}
		e.addLocal(s.Local)
		e.blockInScope(s.Body)
		e.popScope()
		e.continueLabel()
		info := e.leaveLoop()
		e.indent--
		e.line("}")
		e.breakLabel(info)
	}
	e.endStmt()
}
